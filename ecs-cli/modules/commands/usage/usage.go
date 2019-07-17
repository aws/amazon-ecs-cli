// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Package usage aggregates the usage documentation for all ECS-CLI commands and subcommands
package usage

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
)

// Strings displayed as usage for command. Constant should match the command
// prefix for corresponding command or combined command + subcommand, e.g.
// const Local contains the docstring for the `local` command; `LocalUp`
// contains the docstring for `local up`.

// AttributeChecker
const (
	Attributechecker = "Checks if a given list of container instances can run a given task definition by checking their attributes. Outputs attributes that are required by the task definition but not present on the container instances."
)

// Cluster
const (
	ClusterUp    = "Creates the ECS cluster (if it does not already exist) and the AWS resources required to set up the cluster."
	ClusterDown  = "Deletes the CloudFormation stack that was created by ecs-cli up and the associated resources."
	ClusterScale = "Modifies the number of container instances in your cluster. This command changes the desired and maximum instance count in the Auto Scaling group created by the ecs-cli up command. You can use this command to scale up (increase the number of instances) or scale down (decrease the number of instances) your cluster."
	ClusterPs    = "Lists all of the running containers in your ECS cluster"
)

// Compose
const (
	Compose       = "Executes docker-compose-style commands on an ECS cluster."
	ComposeCreate = "Creates an ECS task definition from your compose file. Note that we do not recommend using plain text environment variables for sensitive information, such as credential data."
	ComposePs     = "Lists all the containers in your cluster that were started by the compose project."
	ComposeUp     = "Creates an ECS task definition from your compose file (if it does not already exist) and runs one instance of that task on your cluster (a combination of create and start)."
	ComposeStart  = "Starts a single task from the task definition created from your compose file."
	ComposeRun    = "Starts all containers overriding commands with the supplied one-off commands for the containers."
	ComposeStop   = "Stops all the running tasks created by the compose project."
	ComposeScale  = "Scales the number of running tasks to the specified count."
)

// Compose Service
const (
	Service       = "Manage Amazon ECS services with docker-compose-style commands on an ECS cluster."
	ServiceCreate = "Creates an ECS service from your compose file. The service is created with a desired count of 0, so no containers are started by this command. Note that we do not recommend using plain text environment variables for sensitive information, such as credential data."
	ServiceStart  = "Starts one copy of each of the containers on an existing ECS service by setting the desired count to 1 (only if the current desired count is 0)."
	ServiceUp     = "Creates a new ECS service or updates an existing one according to your compose file. For new services or existing services with a current desired count of 0, the desired count for the service is set to 1. For existing services with non-zero desired counts, a new task definition is created to reflect any changes to the compose file and the service is updated to use that task definition. In this case, the desired count does not change."
	ServicePs     = "Lists all the containers in your cluster that belong to the service created with the compose project."
	ServiceScale  = "Scales the desired count of the service to the specified count."
	ServiceStop   = "Stops the running tasks that belong to the service created with the compose project. This command updates the desired count of the service to 0."
	ServiceRm     = "Updates the desired count of the service to 0 and then deletes the service."
)

// Configure
const (
	Configure               = "Stores a single cluster configuration."
	ConfigureDefault        = "Sets the default cluster config."
	ConfigureMigrate        = "Migrates a legacy ECS CLI configuration file to the current YAML format."
	ConfigureProfile        = "Stores a single profile."
	ConfigureProfileDefault = "Sets the default profile."
)

// Image
const (
	Push   = "Pushes an image to an Amazon ECR repository."
	Pull   = "Pulls an image from an Amazon ECR repository."
	Images = "Lists images an Amazon ECR repository."
)

// License
const (
	License = "Prints the LICENSE files for the ECS CLI and its dependencies."
)

// Local
var (
	Local       = "Runs your ECS tasks locally."
	LocalCreate = "Creates a Compose file from an ECS task definition."
	LocalUp     = fmt.Sprintf("Runs containers locally from an ECS Task Definition. NOTE: Creates a docker-compose file in current directory and a %s if one doesn't exist. ", network.EcsLocalNetworkName)
	LocalPs     = "Lists locally running ECS task containers."
	LocalDown   = fmt.Sprintf("Stops and removes a running ECS task. NOTE: Removes the %s if it has no more running tasks. ", network.EcsLocalNetworkName)
)
