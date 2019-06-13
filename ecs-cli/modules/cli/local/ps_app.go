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
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/docker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

// TODO These labels should be defined part of the local.Create workflow.
// Refactor to import these constants instead of re-defining them here.
// Docker object labels associated with containers created with "ecs-cli local".
const (
	// taskDefinitionLabelType represents the type of option used to
	// transform a task definition to a compose file e.g. remoteFile, localFile.
	// taskDefinitionLabelValue represents the value of the option
	// e.g. file path, arn, family.
	taskDefinitionLabelKey   = "ecsLocalTaskDefinition"
	taskDefinitionLabelType  = "ecsLocalTaskDefType"
	taskDefinitionLabelValue = "ecsLocalTaskDefVal"
)

// Table formatting settings used by the Docker CLI.
// See https://github.com/docker/cli/blob/0904fbfc77dbd4b6296c56e68be573b889d049e3/cli/command/formatter/formatter.go#L74
const (
	cellWidthInSpaces         = 20
	widthBetweenCellsInSpaces = 1
	cellPaddingInSpaces       = 3
	paddingCharacter          = ' '
	noFormatting              = 0
	maxContainerIDLength      = 12
)

// JSON formatting settings.
const (
	jsonPrefix = ""
	jsonIndent = "  "
)

// Ps lists the status of the ECS task containers running locally as a table.
//
// Defaults to listing containers from the local Compose file.
// If the --all flag is provided, then list all local ECS task containers.
// If the --json flag is provided, then output the format as JSON instead.
func Ps(c *cli.Context) {
	if err := psOptionsPreCheck(c); err != nil {
		logrus.Fatalf("Tasks can be either created by local files or remote files")
	}
	containers := listContainers(c)
	displayContainers(c, containers)
}

func psOptionsPreCheck(c *cli.Context) error {
	if (c.String(flags.TaskDefinitionFileFlag) != "") && (c.String(flags.TaskDefinitionTaskFlag) != "") {
		return errors.New("Tasks can be either created by local files or remote files")
	}
	return nil
}

func listContainers(c *cli.Context) []types.Container {
	if c.String(flags.TaskDefinitionFileFlag) != "" {
		return listContainersWithFilters(filters.NewArgs(
			filters.Arg("label", taskDefinitionLabelValue+"="+c.String(flags.TaskDefinitionFileFlag)),
			filters.Arg("label", taskDefinitionLabelType+"="+"localFile"),
		))
	}
	if c.String(flags.TaskDefinitionTaskFlag) != "" {
		return listContainersWithFilters(filters.NewArgs(
			filters.Arg("label", taskDefinitionLabelValue+"="+c.String(flags.TaskDefinitionTaskFlag)),
			filters.Arg("label", taskDefinitionLabelType+"="+"remoteFile"),
		))
	}
	if c.Bool(flags.AllFlag) {
		return listContainersWithFilters(filters.NewArgs(
			filters.Arg("label", taskDefinitionLabelValue),
		))
	}
	return listLocalComposeContainers()
}

func listLocalComposeContainers() []types.Container {
	wd, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(wd, ecsLocalDockerComposeFileName)); os.IsNotExist(err) {
		logrus.Fatalf("Compose file %s does not exist in current directory", ecsLocalDockerComposeFileName)
	}

	// The -q flag displays the ID of the containers instead of the default "Name, Command, State, Ports" metadata.
	cmd := exec.Command("docker-compose", "-f", ecsLocalDockerComposeFileName, "ps", "-q")
	composeOut, err := cmd.Output()
	if err != nil {
		logrus.Fatalf("Failed to run docker-compose ps due to %v", err)
	}

	containerIDs := strings.Split(string(composeOut), "\n")
	if len(containerIDs) == 0 {
		return []types.Container{}
	}

	var args []filters.KeyValuePair
	for _, containerID := range containerIDs {
		args = append(args, filters.Arg("id", containerID))
	}
	return listContainersWithFilters(filters.NewArgs(args...))
}

func listContainersWithFilters(args filters.Args) []types.Container {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	cl := docker.NewClient()
	containers, err := cl.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers with args=%v due to %v", args, err)
	}
	return containers
}

func displayContainers(c *cli.Context, containers []types.Container) {
	if c.Bool(flags.JsonFlag) {
		displayAsJSON(containers)
	} else {
		displayAsTable(containers)
	}
}

func displayAsJSON(containers []types.Container) {
	data, err := json.MarshalIndent(containers, jsonPrefix, jsonIndent)
	if err != nil {
		logrus.Fatalf("Failed to marshal containers to JSON due to %v", err)
	}
	fmt.Fprintln(os.Stdout, string(data))
}

func displayAsTable(containers []types.Container) {
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, cellWidthInSpaces, widthBetweenCellsInSpaces, cellPaddingInSpaces, paddingCharacter, noFormatting)
	fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tSTATUS\tPORTS\tNAMES\tTASKDEFINITION")
	for _, container := range containers {
		row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
			container.ID[:maxContainerIDLength],
			container.Image,
			container.Status,
			prettifyPorts(container.Ports),
			prettifyNames(container.Names),
			container.Labels[taskDefinitionLabelValue])
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

func prettifyPorts(containerPorts []types.Port) string {
	var prettyPorts []string
	for _, port := range containerPorts {
		// See https://github.com/docker/cli/blob/0904fbfc77dbd4b6296c56e68be573b889d049e3/cli/command/formatter/container.go#L268
		prettyPorts = append(prettyPorts, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
	}
	return strings.Join(prettyPorts, ", ")
}

func prettifyNames(containerNames []string) string {
	return strings.Join(containerNames, ", ")
}
