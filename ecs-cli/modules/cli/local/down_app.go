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
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/docker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

// TODO These labels should be defined part of the local.Create workflow.
// Refactor to import these constants instead of re-defining them here.
const (
	ecsLocalDockerComposeFileName = "docker-compose.local.yml"
)

// Down stops and removes running local ECS tasks.
// If the user stops the last running task in the local network then also remove the network.
func Down(c *cli.Context) error {
	defer func() {
		client := docker.NewClient()
		network.Teardown(client)
	}()

	if c.Bool(flags.AllFlag) {
		return downAllLocalContainers()
	}
	return downComposeLocalContainers()
}

func downComposeLocalContainers() error {
	wd, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(wd, ecsLocalDockerComposeFileName)); os.IsNotExist(err) {
		logrus.Warnf("Compose file %s does not exist in current directory", ecsLocalDockerComposeFileName)
		return nil
	}

	logrus.Infof("Running Compose down on %s", ecsLocalDockerComposeFileName)
	cmd := exec.Command("docker-compose", "-f", ecsLocalDockerComposeFileName, "down")
	_, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Fatalf("Failed to run docker-compose down due to %v", err)
	}

	logrus.Info("Stopped and removed containers successfully")
	return nil
}

func downAllLocalContainers() error {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	client := docker.NewClient()
	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", ecsLocalLabelKey),
		),
		All: true,
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers with label=%s due to %v", ecsLocalLabelKey, err)
	}
	if len(containers) == 0 {
		logrus.Warn("No running ECS local tasks found")
		return nil
	}

	logrus.Infof("Stop and remove %d container(s)", len(containers))
	for _, container := range containers {
		if err = client.ContainerStop(ctx, container.ID, nil); err != nil {
			logrus.Fatalf("Failed to stop container %s due to %v", container.ID[:maxContainerIDLength], err)
		}
		logrus.Infof("Stopped container with id %s", container.ID[:maxContainerIDLength])

		if err = client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
			logrus.Fatalf("Failed to remove container %s due to %v", container.ID[:maxContainerIDLength], err)
		}
		logrus.Infof("Removed container with id %s", container.ID[:maxContainerIDLength])
	}
	return nil
}
