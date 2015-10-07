// This file is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found here :
// https://github.com/docker/libcompose/blob/master/project/types_yaml.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Modifications are Copyright 2015 Amazon.com, Inc. or its affiliates. Licensed under the Apache License 2.0
package libcompose

import (
	"strings"
)

type Stringorslice struct {
	parts []string
}

func (s Stringorslice) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

func (s *Stringorslice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var sliceType []string
	err := unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	var stringType string
	err = unmarshal(&stringType)
	if err == nil {
		sliceType = make([]string, 0, 1)
		s.parts = append(sliceType, string(stringType))
		return nil
	}
	return err
}

func (s *Stringorslice) Len() int {
	if s == nil {
		return 0
	}
	return len(s.parts)
}

func (s *Stringorslice) Slice() []string {
	if s == nil {
		return nil
	}
	return s.parts
}

func NewStringorslice(parts ...string) Stringorslice {
	return Stringorslice{parts}
}

type Command struct {
	parts []string
}

func (s Command) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

func (s *Command) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var stringType string
	err := unmarshal(&stringType)
	if err == nil {
		sliceType := make([]string, 0, 1)
		s.parts = append(sliceType, string(stringType))
		return nil
	}

	var sliceType []string
	err = unmarshal(&sliceType)
	if err == nil {
		s.parts = sliceType
		return nil
	}

	return err
}

func (s *Command) ToString() string {
	return strings.Join(s.parts, " ")
}

func (s *Command) Slice() []string {
	return s.parts
}

func NewCommand(parts ...string) Command {
	return Command{parts}
}

type SliceorMap struct {
	parts map[string]string
}

func (s SliceorMap) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

func (s *SliceorMap) UnmarshalYAML(unmarshal func(interface{}) error) error {
	mapType := make(map[string]string)
	err := unmarshal(&mapType)
	if err == nil {
		s.parts = mapType
		return nil
	}

	var sliceType []string
	var key string
	var value string

	err = unmarshal(&sliceType)
	if err != nil {
		return err
	}

	mapType = make(map[string]string)
	for _, slice := range sliceType {
		slice = strings.TrimSpace(slice)
		keyValueSlice := strings.SplitN(slice, "=", 2)

		key = keyValueSlice[0]
		value = ""
		if len(keyValueSlice) == 2 {
			value = keyValueSlice[1]
		}

		mapType[key] = value
	}
	s.parts = mapType
	return nil
}

func (s *SliceorMap) MapParts() map[string]string {
	if s == nil {
		return nil
	}
	return s.parts
}

func NewSliceorMap(parts map[string]string) SliceorMap {
	return SliceorMap{parts}
}

type MaporEqualSlice struct {
	parts []string
}

func (s MaporEqualSlice) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

func (s *MaporEqualSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&s.parts)
	if err == nil {
		return nil
	}

	var mapType map[string]string

	err = unmarshal(&mapType)
	if err != nil {
		return err
	}

	for k, v := range mapType {
		s.parts = append(s.parts, strings.Join([]string{k, v}, "="))
	}

	return nil
}

func (s *MaporEqualSlice) Slice() []string {
	return s.parts
}

func NewMaporEqualSlice(parts []string) MaporEqualSlice {
	return MaporEqualSlice{parts}
}

type MaporColonSlice struct {
	parts []string
}

func (s MaporColonSlice) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

func (s *MaporColonSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&s.parts)
	if err == nil {
		return nil
	}

	var mapType map[string]string

	err = unmarshal(&mapType)
	if err != nil {
		return err
	}

	for k, v := range mapType {
		s.parts = append(s.parts, strings.Join([]string{k, v}, ":"))
	}

	return nil
}

func (s *MaporColonSlice) Slice() []string {
	return s.parts
}

func NewMaporColonSlice(parts []string) MaporColonSlice {
	return MaporColonSlice{parts}
}

type MaporSpaceSlice struct {
	parts []string
}

func (s MaporSpaceSlice) MarshalYAML() (interface{}, error) {
	return s.parts, nil
}

func (s *MaporSpaceSlice) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&s.parts)
	if err == nil {
		return nil
	}

	var mapType map[string]string

	err = unmarshal(&mapType)
	if err != nil {
		return err
	}

	for k, v := range mapType {
		s.parts = append(s.parts, strings.Join([]string{k, v}, " "))
	}

	return nil
}

func (s *MaporSpaceSlice) Slice() []string {
	return s.parts
}

func NewMaporSpaceSlice(parts []string) MaporSpaceSlice {
	return MaporSpaceSlice{parts}
}
