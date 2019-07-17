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
// "fmt"
)

// String displayed as usage for command. Constant should match the command
// prefix for corresponding command or combined command + subcommand, e.g.
// const Local contains the docstring for the `local` command; `LocalUp`
// contains the docstring for `local up`.
const (
	// AttributeChecker
	Attributechecker = "Checks if a given list of container instances can run a given task definition by checking their attributes. Outputs attributes that are required by the task definition but not present on the container instances."

	// Cluster
	ClusterUp    = "Creates the ECS cluster (if it does not already exist) and the AWS resources required to set up the cluster."
	ClusterDown  = "Deletes the CloudFormation stack that was created by ecs-cli up and the associated resources."
	ClusterScale = "Modifies the number of container instances in your cluster. This command changes the desired and maximum instance count in the Auto Scaling group created by the ecs-cli up command. You can use this command to scale up (increase the number of instances) or scale down (decrease the number of instances) your cluster."
	ClusterPs    = "Lists all of the running containers in your ECS cluster"

	// Compose
	Compose       = "Executes docker-compose-style commands on an ECS cluster."
	ComposeCreate = "Creates an ECS task definition from your compose file. Note that we do not recommend using plain text environment variables for sensitive information, such as credential data."
	ComposePs     = "Lists all the containers in your cluster that were started by the compose project."
	ComposeUp     = "Creates an ECS task definition from your compose file (if it does not already exist) and runs one instance of that task on your cluster (a combination of create and start)."
	ComposeStart  = "Starts a single task from the task definition created from your compose file."
	ComposeRun    = "Starts all containers overriding commands with the supplied one-off commands for the containers."
	ComposeStop   = "Stops all the running tasks created by the compose project."
	ComposeScale  = "Scales the number of running tasks to the specified count."
)
