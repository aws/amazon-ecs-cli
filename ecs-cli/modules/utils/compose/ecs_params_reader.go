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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

type ECSParams struct {
	Version        string
	TaskDefinition EcsTaskDef `yaml:"task_definition"`
	RunParams      RunParams  `yaml:"run_params"`
}

// EcsTaskDef corresponds to fields in an ECS TaskDefinition
type EcsTaskDef struct {
	NetworkMode          string        `yaml:"ecs_network_mode"`
	TaskRoleArn          string        `yaml:"task_role_arn"`
	ContainerDefinitions ContainerDefs `yaml:"services"`
	ExecutionRole        string        `yaml:"task_execution_role"`
	TaskSize             TaskSize      `yaml:"task_size"`           // Needed to run FARGATE tasks
}

type ContainerDefs map[string]ContainerDef

type ContainerDef struct {
	Essential bool `yaml:"essential"`
}

type TaskSize struct {
	Cpu    string `yaml:"cpu_limit"`
	Memory string `yaml:"mem_limit"`
}

// RunParams specifies non-TaskDefinition specific parameters
type RunParams struct {
	NetworkConfiguration NetworkConfiguration `yaml:"network_configuration"`
}

type NetworkConfiguration struct {
	AwsVpcConfiguration AwsVpcConfiguration `yaml:"awsvpc_configuration"`
}

type AwsVpcConfiguration struct {
	Subnets        []string       `yaml:"subnets"`
	SecurityGroups []string       `yaml:"security_groups"`
	AssignPublicIp AssignPublicIp `yaml:"assign_public_ip"` // Needed to run FARGATE tasks
}

type AssignPublicIp string

const (
	Enabled  AssignPublicIp = "ENABLED"
	Disabled AssignPublicIp = "DISABLED"
)

// ReadECSParams parses the ecs-params.yml file and puts it into an ECSParams struct.
func ReadECSParams(filename string) (*ECSParams, error) {
	if filename == "" {
		defaultFilename := "ecs-params.yml"
		if _, err := os.Stat(defaultFilename); err == nil {
			filename = defaultFilename
		} else {
			return nil, nil
		}
	}

	// NOTE: Readfile reads all data into memory and closes file. Could
	// eventually refactor this to read different sections separately.
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

// ConvertToECSNetworkConfiguration extracts out the NetworkConfiguration from
// the ECSParams into a format that is compatible with ECSClient calls.
func ConvertToECSNetworkConfiguration(ecsParams *ECSParams) (*ecs.NetworkConfiguration, error) {
	if ecsParams == nil {
		return nil, nil
	}

	networkMode := ecsParams.TaskDefinition.NetworkMode

	if networkMode != "awsvpc" {
		return nil, nil
	}

	awsvpcConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration

	subnets := awsvpcConfig.Subnets

	if len(subnets) < 1 {
		return nil, errors.New("at least one subnet is required in the network configuration")
	}

	securityGroups := awsvpcConfig.SecurityGroups
	assignPublicIp := string(awsvpcConfig.AssignPublicIp)

	ecsSubnets := make([]*string, len(subnets))
	for i, subnet := range subnets {
		ecsSubnets[i] = aws.String(subnet)
	}

	ecsSecurityGroups := make([]*string, len(securityGroups))
	for i, sg := range securityGroups {
		ecsSecurityGroups[i] = aws.String(sg)
	}

	ecsAwsVpcConfig := &ecs.AwsVpcConfiguration{
		Subnets:        ecsSubnets,
		SecurityGroups: ecsSecurityGroups,
	}

	// For tasks launched with network config in EC2 mode, assign_pubic_ip field is not accepted
	if assignPublicIp != "" {
		ecsAwsVpcConfig.AssignPublicIp = aws.String(assignPublicIp)
	}

	ecsNetworkConfig := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: ecsAwsVpcConfig,
	}

	return ecsNetworkConfig, nil
}
