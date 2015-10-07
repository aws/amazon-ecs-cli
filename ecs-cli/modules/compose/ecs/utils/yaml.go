// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/kylelemons/go-gypsy/yaml"
)

const (
	yamlTag = "yaml"
)

// Unmarshaler provides a fuction to customize the yaml transformation
type Unmarshaler interface {
	UnmarshalYAML(unmarshal func(interface{}) error) error
}

// unmarshal parses the yaml data in node and stores the result in the value pointed to by out,
// filtering out the fields not listed in the supportedYamlTags map.
// For each field in the interface out, this function finds the yaml data from node (based on the field name or yaml tag),
// checks if the field type implements a custom unmarshaler and invokes that
// else it parses the node data and sets the value to field appropriately.
// Supports fields of type: Int, String, Float, Bool, Slice, Struct, Ptr, Map
func unmarshal(node yaml.Node, out interface{}, supportedYamlTags map[string]bool) error {
	field := reflect.ValueOf(out)
	if field.IsNil() {
		return fmt.Errorf("Error unmarshaling: nil object")
	}

	// dereferences pointers
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	return setValue(node, field.Type(), field, supportedYamlTags)
}

// setValue decodes a YAML value from node into the value field,
// filtering out the fields not listed in the supportedYamlTags map.
func setValue(node yaml.Node, fieldType reflect.Type, field reflect.Value, supportedYamlTags map[string]bool) error {
	if !field.CanSet() {
		return fmt.Errorf("Unable to set the value [%v] for the field [%v] of kind [%v]", node, field, field.Kind())
	}

	// invoke custom unmarshaler if implemented, else proceed with best-effort parsing using reflections
	if field.CanAddr() {
		if u, ok := field.Addr().Interface().(Unmarshaler); ok {
			return u.UnmarshalYAML(func(val interface{}) error {
				return unmarshal(node, val, supportedYamlTags)
			})
		}
	}

	switch field.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		return unmarshalScalar(node, field, supportedYamlTags)
	case reflect.Slice:
		return unmarshalSlice(node, field, supportedYamlTags)
	case reflect.Struct:
		return unmarshalStruct(node, fieldType, field, supportedYamlTags)
	case reflect.Map:
		return unmarshalMap(node, field, supportedYamlTags)
	case reflect.Ptr:
		fieldValue := reflect.New(fieldType.Elem())
		if err := setValue(node, fieldType.Elem(), fieldValue.Elem(), supportedYamlTags); err != nil {
			return err
		}
		field.Set(fieldValue)
		return nil
	// TODO, case reflect.Interface:
	default:
		return fmt.Errorf("Cannot setValue for the field [%v] of kind [%v]. Undefined operation.", field, field.Kind())
	}
}

// unmarshalStruct iterates through all the fields in the field struct interface
// and for each key=(yamlTagName or fieldName), finds the value in the node,
// filters out the fields not listed in the supportedYamlTags map
// and sets the value in the out struct using reflections
func unmarshalStruct(node yaml.Node, fieldType reflect.Type, field reflect.Value, supportedYamlTags map[string]bool) error {
	for i := 0; i < field.NumField(); i++ {
		f := field.Field(i)
		ft := fieldType.Field(i)

		// get the tag name (if any), defaults to fieldName
		tagName := ft.Name
		yamlTag := ft.Tag.Get(yamlTag) // Expected format `yaml:"tagName,omitempty"` // TODO, handle omitempty
		if yamlTag != "" {
			tags := strings.Split(yamlTag, ",")
			if len(tags) > 0 {
				tagName = tags[0]
			}
		}

		// getNode with key=tagName
		childNode, err := yaml.Child(node, tagName)
		if err != nil || childNode == nil {
			// not found
			continue
		}

		// filter out the node with key=tagName if not in supportedYamlTags map
		if !supportedYamlTags[tagName] {
			log.WithFields(log.Fields{"option name": tagName}).Warn("Skipping unsupported YAML option...")
			continue
		}

		// set value
		if err = setValue(childNode, ft.Type, f, supportedYamlTags); err != nil {
			return err
		}
	}
	return nil
}

// unmarshalScalar sets the scalar value of the node to field based on its type.
// This returns an error if the node cannot be converted to a scalar type
func unmarshalScalar(node yaml.Node, field reflect.Value, supportedYamlTags map[string]bool) error {
	value, ok := node.(yaml.Scalar)
	if !ok {
		return fmt.Errorf("Error parsing: node is not of type string %v", node)
	}
	strValue := value.String()

	// unmarshal double-quoted string (best-effort)
	// TODO, single-quoted string http://yaml.org/spec/1.2/spec.html#id2786942
	unquotedStrValue, err := strconv.Unquote(strValue)
	if err == nil { // strconv.Unquote returns no error only if the string has valid quotes
		strValue = unquotedStrValue
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(strValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := strconv.Atoi(strValue)
		if err != nil {
			return err
		}
		if !field.OverflowInt(int64(intValue)) {
			field.SetInt(int64(intValue))
		}
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(strValue)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(strValue, 64)
		if err != nil {
			return err
		}
		field.SetFloat(float64(floatValue))
	default:
		return fmt.Errorf("Unknown type kind [%v] to setValue from scalar", field.Kind())
	}
	return nil
}

// unmarshalSlice transforms the list node to a slice represented by field.
// This returns an error if the node cannot be converted to a list type
func unmarshalSlice(node yaml.Node, field reflect.Value, supportedYamlTags map[string]bool) error {
	list, ok := node.(yaml.List)
	if !ok {
		return fmt.Errorf("Error parsing: node is not of type list %v", node)
	}
	listLen := list.Len()
	field.Set(reflect.MakeSlice(field.Type(), listLen, listLen))

	elemType := field.Type().Elem()
	for i := 0; i < listLen; i++ {
		elem := reflect.New(elemType).Elem()
		if err := setValue(list.Item(i), elemType, elem, supportedYamlTags); err != nil {
			return err
		}
		field.Index(i).Set(elem)
	}
	return nil
}

// unmarshalMap transforms the map node to a map represented by field.
// This returns an error if the node cannot be converted to a map type
func unmarshalMap(node yaml.Node, field reflect.Value, supportedYamlTags map[string]bool) error {
	mapNode, err := nodeToMap(node)
	if err != nil {
		return err
	}

	fieldType := field.Type()
	keyType := fieldType.Key()
	elemType := fieldType.Elem()

	// go-gypsy yaml has Map structure as map[string]Node, cannot process any other key type
	if keyType.Kind() != reflect.String {
		return fmt.Errorf("Unable to unmarshal map [%v]. Only string key type is supported.", field)
	}
	if field.IsNil() {
		field.Set(reflect.MakeMap(fieldType))
	}
	for keyNode, valueNode := range mapNode {
		k := reflect.New(keyType).Elem()
		k.SetString(keyNode)

		e := reflect.New(elemType).Elem()
		if err := setValue(valueNode, elemType, e, supportedYamlTags); err != nil {
			return err
		}

		field.SetMapIndex(k, e)
	}
	return nil
}

// parseYamlString transforms a yaml string into readable yaml structure
func parseYamlString(yamlString string) yaml.Node {
	// Note : yaml.Config doesn't return an error message to bubble up, instead it panics
	config := yaml.Config(yamlString)

	if config == nil {
		return nil
	}
	return config.Root
}

// nodeToMap converts the yaml node to yaml map
func nodeToMap(node yaml.Node) (yaml.Map, error) {
	m, ok := node.(yaml.Map)
	if !ok {
		msg := "Error parsing: node is not of type map"
		logErrorWithFields(nil, msg, log.Fields{
			"node": node,
		})
		return nil, fmt.Errorf("%s %v", msg, node)
	}
	return m, nil
}
