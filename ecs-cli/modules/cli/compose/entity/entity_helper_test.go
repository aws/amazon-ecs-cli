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

package entity

import (
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/stretchr/testify/assert"
)

func TestValidateFargateParams_HappyFargatePath(t *testing.T) {
	taskDef := utils.EcsTaskDef{NetworkMode: "awsvpc"}
	subnets := []string{"subnet-feedface"}
	awsVpconfig := utils.AwsVpcConfiguration{
		Subnets: subnets,
	}

	networkConfig := utils.NetworkConfiguration{
		AwsVpcConfiguration: awsVpconfig,
	}

	ecsParams := &utils.ECSParams{
		TaskDefinition: taskDef,
		RunParams: utils.RunParams{
			NetworkConfiguration: networkConfig,
		},
	}

	launchType := config.LaunchTypeFargate

	err := ValidateFargateParams(ecsParams, launchType)
	assert.NoError(t, err)
}

func TestValidateFargateParams_NotFargateLaunchType(t *testing.T) {
	taskDef := utils.EcsTaskDef{NetworkMode: "awsvpc"}
	subnets := []string{"subnet-feedface"}
	awsVpconfig := utils.AwsVpcConfiguration{
		Subnets: subnets,
	}

	networkConfig := utils.NetworkConfiguration{
		AwsVpcConfiguration: awsVpconfig,
	}

	ecsParams := &utils.ECSParams{
		TaskDefinition: taskDef,
		RunParams: utils.RunParams{
			NetworkConfiguration: networkConfig,
		},
	}

	launchType := config.LaunchTypeEC2

	err := ValidateFargateParams(ecsParams, launchType)

	assert.NoError(t, err)
}

func TestValidateFargateParams_NoEcsParams_EC2mode(t *testing.T) {
	launchType := config.LaunchTypeEC2
	err := ValidateFargateParams(nil, launchType)

	assert.NoError(t, err)
}

func TestValidateFargateParams_NoEcsParams_FargateMode(t *testing.T) {
	launchType := config.LaunchTypeFargate
	err := ValidateFargateParams(nil, launchType)

	assert.Error(t, err, "Launch Type FARGATE requires network configuration to be set. Set network configuration using an ECS Params file.")
}

func TestValidateFargateParams_WrongNetworkMode(t *testing.T) {
	taskDef := utils.EcsTaskDef{NetworkMode: "host"}
	subnets := []string{"subnet-feedface"}
	awsVpconfig := utils.AwsVpcConfiguration{
		Subnets: subnets,
	}

	networkConfig := utils.NetworkConfiguration{
		AwsVpcConfiguration: awsVpconfig,
	}

	ecsParams := &utils.ECSParams{
		TaskDefinition: taskDef,
		RunParams: utils.RunParams{
			NetworkConfiguration: networkConfig,
		},
	}

	launchType := config.LaunchTypeFargate

	err := ValidateFargateParams(ecsParams, launchType)

	assert.Error(t, err, "Launch Type FARGATE requires network mode to be 'awsvpc'. Set network mode using an ECS Params file.")
}

// NOTE: ValidateFargateParams should technically also check for the presence
// of subnets, but this check already exists in
// utils#ConvertToECSNetworkConfiguration, since it also applies to non-Fargate
// tasks with Task Networking. TODO: refactor

// TODO: Backfill other tests
