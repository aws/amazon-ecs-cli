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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/options"
	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

// Down stops and removes running local ECS tasks.
// If the user stops the last running task in the local network then also remove the network.
func Down(c *cli.Context) {
	defer func() {
		client := docker.NewClient()
		network.Teardown(client)
	}()

	if err := options.ValidateCombinations(c); err != nil {
		logrus.Fatal(err.Error())
	}
	containers, err := listContainers(c)
	if err != nil {
		logrus.Fatalf("Failed to list containers due to:\n%v", err)
	}
	downContainers(containers)
}

func downContainers(containers []types.Container) {
	if len(containers) == 0 {
		logrus.Warn("No running ECS local tasks found")
		return
	}
	client := docker.NewClient()
	logrus.Infof("Stop and remove %d container(s)", len(containers))
	for _, container := range containers {
		ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
		if err := client.ContainerStop(ctx, container.ID, nil); err != nil {
			logrus.Fatalf("Failed to stop container %s due to:\n%v", container.ID[:maxContainerIDLength], err)
		}
		logrus.Infof("Stopped container with id %s", container.ID[:maxContainerIDLength])

		if err := client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
			logrus.Fatalf("Failed to remove container %s due to:\n%v", container.ID[:maxContainerIDLength], err)
		}
		logrus.Infof("Removed container with id %s", container.ID[:maxContainerIDLength])
		cancel()
	}
}
