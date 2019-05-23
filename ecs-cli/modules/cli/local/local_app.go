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

// Package local implements the subcommands to run ECS task definitions locally
// (See: https://github.com/aws/containers-roadmap/issues/180).
package local

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func Create(c *cli.Context) {
	// 1. read in task def (from file or arn)
	// 2. parse task def into go object
	// 3. write to docker-compose.local.yml file
	fmt.Println("foo") // placeholder
}

// Up creates a Compose file from an ECS task definition and runs it locally.
//
// The Amazon ECS Local Endpoints container needs to be running already for any local ECS task to work
// (see https://github.com/awslabs/amazon-ecs-local-container-endpoints).
// If the container is not running, this command creates a new network for all local ECS tasks to join
// and communicate with the Amazon ECS Local Endpoints container.
func Up(c *cli.Context) {
	docker := newDockerClient()
	network.Setup(docker)
}

// Stop stops a running local ECS task.
//
// If the user stops the last running task in the local network then also remove the network.
func Stop(c *cli.Context) {
	// TODO move these clients to a separate file leveraging the DOCKER_API_VERSION,
	// these clients are created to test the local network for now.
	// See https://github.com/awslabs/amazon-ecs-local-container-endpoints/blob/master/local-container-endpoints/clients/docker/client.go#L49
	docker, err := client.NewEnvClient()
	if err != nil {
		logrus.Fatal("Could not connect to docker", err)
	}
	defer network.Teardown(docker)
}
