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
	"fmt"
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

// Down stops and removes a running local ECS task.
// If the user stops the last running task in the local network then also remove the network.
func Down(c *cli.Context) error {
	defer func() {
		client := docker.NewClient()
		network.Teardown(client)
	}()

	taskPath := c.String(flags.TaskDefinitionFileFlag)
	taskARN := c.String(flags.TaskDefinitionArnFlag)

	// TaskDefinitionFileFlag flag has priority over TaskDefinitionArnFlag if both are present
	if taskPath != "" {
		return handleDownWithFile(taskPath)
	}
	if taskARN != "" {
		return handleDownWithARN(taskARN)
	}
	return handleDownWithCompose()
}

func handleDownWithFile(path string) error {
	return handleDownWithFilters(filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("%s=%s", taskFilePathLabelKey, path)),
	))
}

func handleDownWithARN(value string) error {
	return handleDownWithFilters(filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("%s=%s", taskDefinitionARNLabelKey, value)),
	))
}

func handleDownWithCompose() error {
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

func handleDownWithFilters(args filters.Args) error {
	client := docker.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	containers, err := client.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
		All:     true,
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers with args %v due to %v", args, err)
	}
	if len(containers) == 0 {
		logrus.Warnf("No containers found with label %v", args.Get("label"))
		return nil
	}

	for _, container := range containers {
		if err = client.ContainerStop(ctx, container.ID, nil); err != nil {
			logrus.Fatalf("Failed to stop container %s due to %v", container.ID, err)
		}
		logrus.Infof("Stopped container with id %s", container.ID)

		if err = client.ContainerRemove(ctx, container.ID, types.ContainerRemoveOptions{}); err != nil {
			logrus.Fatalf("Failed to remove container %s due to %v", container.ID, err)
		}
		logrus.Infof("Removed container with id %s", container.ID)
	}
	return nil
}
