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
	"reflect"
	"testing"

	"github.com/kylelemons/go-gypsy/yaml"
)

type testYml struct {
	// scalar
	IntKey int64  `yaml:"int_key"`
	StrKey string `yaml:"str_key"`

	// scalar: untagged
	BoolKey  bool
	FloatKey float64

	// scalar: unsettable field
	privateIntKey int64

	// slice
	StrSliceKey  []string `yaml:"strSlice_key"`
	BoolSliceKey []bool   `yaml:"boolSlice_key"`

	// struct
	InnerStructKey innerTestYml `yaml:"innerStruct_key"`

	// map
	MapKey map[string]innerTestYml

	// ptr
	StrKeyPtr         *string
	StrSliceKeyPtr    []*string
	InnerStructKeyPtr *innerTestYml

	// unsupported
	UnsupportedOption int64 `yaml:"unsupported"`
}

type innerTestYml struct {
	IntKey int64  `yaml:"innerStruct_int"`
	StrKey string `yaml:"innerStruct_str"`

	// unsupported
	UnsupportedOption string `yaml:"also_unsupported"`
}

func getSupportedYamlTagsForTest() map[string]bool {
	return map[string]bool{
		"int_key":           true,
		"str_key":           true,
		"BoolKey":           true,
		"FloatKey":          true,
		"privateIntKey":     true,
		"strSlice_key":      true,
		"boolSlice_key":     true,
		"innerStruct_key":   true,
		"MapKey":            true,
		"StrKeyPtr":         true,
		"StrSliceKeyPtr":    true,
		"InnerStructKeyPtr": true,
		"innerStruct_int":   true,
		"innerStruct_str":   true,
	}
}

func TestUnmarshalEmptyString(t *testing.T) {
	testUnmarshal(t, "")
}

func TestUnmarshalNilObject(t *testing.T) {
	err := unmarshal(nil, getNilTestYml(), getSupportedYamlTagsForTest())
	if err == nil {
		t.Error("Expected error while unmarshaling nil objects")
	}
}

func getNilTestYml() *testYml {
	return nil
}

func TestUnmarshalScalarValues(t *testing.T) {
	// yaml tagged fields
	intValue := int64(10)
	strValue := "str"

	// untagged fields
	boolValue := true
	floatValue := float64(3.14)

	// test string
	testYmlString := `root:
  int_key: 10
  str_key: str
  BoolKey: true
  FloatKey: 3.14`

	out := testUnmarshal(t, testYmlString)

	// verify yaml tagged fields
	if intValue != out.IntKey {
		t.Errorf("Expected intValue [%s] But was [%s]", intValue, out.IntKey)
	}
	if strValue != out.StrKey {
		t.Errorf("Expected strValue [%s] But was [%s]", strValue, out.StrKey)
	}

	// veify untagged fields
	if boolValue != out.BoolKey {
		t.Errorf("Expected boolValue [%s] But was [%s]", boolValue, out.BoolKey)
	}
	if floatValue != out.FloatKey {
		t.Errorf("Expected floatValue [%s] But was [%s]", floatValue, out.FloatKey)
	}
}

func TestUnmarshalScalarValuesUnexportedField(t *testing.T) {
	// test string with unexported field
	testYmlString := `root:
  privateIntKey: 10`
	rootNode := getRootNode(t, testYmlString)
	err := unmarshal(rootNode, &testYml{}, getSupportedYamlTagsForTest())
	if err == nil {
		t.Errorf("Expected error while setting unexported field from yaml string [%s]", testYmlString)
	}
}

func TestUnmarshalSliceValues(t *testing.T) {
	strSlice := []string{"part1", "part2"}
	boolSlice := []bool{true}

	testYmlString := `root:
  strSlice_key:
   - part1
   - part2
  boolSlice_key:
   - true`
	out := testUnmarshal(t, testYmlString)
	if !reflect.DeepEqual(strSlice, out.StrSliceKey) {
		t.Errorf("Expected strSlice [%v] But was [%v]", strSlice, out.StrSliceKey)
	}
	if !reflect.DeepEqual(boolSlice, out.BoolSliceKey) {
		t.Errorf("Expected boolSlice [%v] But was [%v]", boolSlice, out.BoolSliceKey)
	}
}

func TestUnmarshalStructValues(t *testing.T) {
	intValue := int64(10)
	strValue := "str"

	// test string
	testYmlString := `root:
  innerStruct_key:
    innerStruct_int: 10
    innerStruct_str: str`

	out := testUnmarshal(t, testYmlString)
	if intValue != out.InnerStructKey.IntKey {
		t.Errorf("Expected intValue [%s] But was [%s]", intValue, out.InnerStructKey.IntKey)
	}
	if strValue != out.InnerStructKey.StrKey {
		t.Errorf("Expected strValue [%s] But was [%s]", strValue, out.InnerStructKey.StrKey)
	}
}

func TestUnmarshalMapValues(t *testing.T) {
	intValues := []int64{10, 20}
	strValues := []string{"", "str"}
	keys := []string{"struct1", "struct2"}

	// test string
	testYmlString := `root:
  MapKey:
  	struct1:
      innerStruct_int: 10
  	struct2:
      innerStruct_int: 20
      innerStruct_str: str`

	out := testUnmarshal(t, testYmlString)
	i := 0
	for key, val := range out.MapKey {
		if keys[i] != key {
			t.Errorf("Expected key [%s] But was [%s]", keys[i], key)
		}
		if intValues[i] != val.IntKey {
			t.Errorf("Expected intValue [%s] But was [%s]", intValues[i], val.IntKey)
		}
		if strValues[i] != val.StrKey {
			t.Errorf("Expected strValue [%s] But was [%s]", strValues[i], val.StrKey)
		}
		i++
	}
}

func TestUnmarshalPointerValues(t *testing.T) {
	intValue := int64(10)
	strValue := "str"
	strSlice := []string{"part1", "part2"}

	// test string
	testYmlString := `root:
  StrKeyPtr: str
  StrSliceKeyPtr:
    - part1
    - part2
  InnerStructKeyPtr:
    innerStruct_int: 10
    innerStruct_str: str`

	out := testUnmarshal(t, testYmlString)

	if strValue != *out.StrKeyPtr {
		t.Errorf("Expected *out.StrKeyPtr [%s] But was [%s]", strValue, *out.StrKeyPtr)
	}
	for index, val := range out.StrSliceKeyPtr {
		if strSlice[index] != *val {
			t.Errorf("Expected strSlice [%v] But was [%v]", strSlice, *val)
		}
	}
	structPtr := *out.InnerStructKeyPtr
	if intValue != structPtr.IntKey {
		t.Errorf("Expected structPtr.intValue [%s] But was [%s]", intValue, structPtr.IntKey)
	}
	if strValue != structPtr.StrKey {
		t.Errorf("Expected structPtr.strValue [%s] But was [%s]", strValue, structPtr.StrKey)
	}
}

func TestUnmarshalUnsupportedValues(t *testing.T) {
	intValue := int64(10)
	strValue := "str"

	// test string
	testYmlString := `root:
  unsupported: 10
  InnerStructKeyPtr:
    innerStruct_int: 10
    innerStruct_str: str
    also_unsupported: str`

	out := testUnmarshal(t, testYmlString)

	// verify UnsupportedOption to not have been set
	if intValue == out.UnsupportedOption {
		t.Errorf("Expected *out.UnsupportedOption [%s] But was [%s]", 0, out.UnsupportedOption)
	}

	// verify struct to not have also_unsupported
	structPtr := *out.InnerStructKeyPtr
	if intValue != structPtr.IntKey {
		t.Errorf("Expected structPtr.intValue [%s] But was [%s]", intValue, structPtr.IntKey)
	}
	if strValue != structPtr.StrKey {
		t.Errorf("Expected structPtr.strValue [%s] But was [%s]", strValue, structPtr.StrKey)
	}
	if strValue == structPtr.UnsupportedOption {
		t.Errorf("Expected structPtr.UnsupportedOption to be empty But was [%s]", structPtr.UnsupportedOption)
	}
}

func testUnmarshal(t *testing.T, ymlString string) *testYml {
	rootNode := getRootNode(t, ymlString)
	out := &testYml{}
	err := unmarshal(rootNode, out, getSupportedYamlTagsForTest())
	if err != nil {
		t.Errorf("Expected to unmarshal the yaml string [%s]. But got error [%v]", ymlString, err)
	}
	return out
}

func getRootNode(t *testing.T, ymlString string) yaml.Node {
	config := yaml.Config(ymlString)
	if config == nil {
		t.Fatal("Got errors trying to parse test input yaml string [%s]", ymlString)
		return nil
	}
	configMap, _ := nodeToMap(config.Root)
	return configMap["root"]
}

func TestNodeToMap(t *testing.T) {
	// test if we can transform the parsed config to a map
	testStr := `key1:
  subKey1: value	
key2:
  subKey1: value`

	config := parseYamlString(testStr)
	if config == nil {
		t.Fatalf("Unable to parse the file string [%s]", testStr)
	}

	transformedMap, err := nodeToMap(config)
	if err != nil {
		t.Errorf("Unexpected error [%s] while transforming the config head node to a map", err)
	}

	expectedKeys := []string{"key1", "key2"}
	for _, key := range expectedKeys {
		if _, ok := transformedMap[key]; !ok {
			t.Errorf("Expected key [%s] but its missing", key)
		}
	}
}

func TestNodeToMapForIncorrectNode(t *testing.T) {
	// test when the node is not of type map
	testStr := "test"
	config := parseYamlString(testStr)
	if config == nil {
		t.Fatalf("Unable to parse the yaml string [%s]", testStr)
	}

	_, err := nodeToMap(config)
	if err == nil {
		t.Error("Expected nodeToMap to fail since node is not of type map")
	}
}

func TestNodeToMapForNilNode(t *testing.T) {
	// test when the node is nil
	config, err := nodeToMap(nil)
	if err == nil {
		t.Error("Expected nodeToMap to fail since node is nil")
	}
	if config != nil {
		t.Errorf("Expected config to be nil but was [%v]", config)
	}
}
