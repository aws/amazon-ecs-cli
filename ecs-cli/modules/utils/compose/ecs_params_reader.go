// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// ECS Params Reader is used to parse the ecs-params.yml file and marshal the data into the ECSParams struct

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ECSParams struct {
	Version        string
	TaskDefinition EcsTaskDef `yaml:"task_definition"`
}

type EcsTaskDef struct {
	NetworkMode          string        `yaml:"ecs_network_mode"`
	TaskRoleArn          string        `yaml:"task_role_arn"`
	ContainerDefinitions ContainerDefs `yaml:"services"`
}

type ContainerDefs map[string]ContainerDef

type ContainerDef struct {
	Essential bool `yaml:"essential"`
}

func readECSParams(filename string) (*ECSParams, error) {
	if filename == "" {
		defaultFilename := "ecs-params.yml"
		if _, err := os.Stat(defaultFilename); err == nil {
			filename = defaultFilename
		} else {
			return nil, nil
		}
	}
	ecsParamsData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading file '%v'", filename)
	}

	ecsParams := &ECSParams{}

	if err = yaml.Unmarshal([]byte(ecsParamsData), &ecsParams); err != nil {
		return nil, errors.Wrapf(err, "Error unmarshalling yaml data from ECS params file: %v", filename)
	}

	return ecsParams, nil
}
