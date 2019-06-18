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

package local

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/docker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/urfave/cli"
)

// Up creates a Compose file from an ECS task definition and runs it locally.
//
// The Amazon ECS Local Endpoints container needs to be running already for any local ECS task to work
// (see https://github.com/awslabs/amazon-ecs-local-container-endpoints).
// If the container is not running, this command creates a new network for all local ECS tasks to join
// and communicate with the Amazon ECS Local Endpoints container.
func Up(c *cli.Context) {
	client := docker.NewClient()
	network.Setup(client)
}
