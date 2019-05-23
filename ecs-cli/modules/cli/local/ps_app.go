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
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

const (
	// ECSLocalLabelKey is the Docker object label associated with containers created with "ecs-cli local"
	// TODO This label should belong to local.Create
	ECSLocalLabelKey = "ECSLocalTask"
)

// Table formatting settings used by the Docker CLI.
// See https://github.com/docker/cli/blob/0904fbfc77dbd4b6296c56e68be573b889d049e3/cli/command/formatter/formatter.go#L74
const (
	cellWidthInSpaces         = 20
	widthBetweenCellsInSpaces = 1
	cellPaddingInSpaces       = 3
	paddingCharacter          = ' '
	noFormatting              = 0

	maxContainerIDLength = 12
)

// Ps lists the status of the ECS task containers running locally.
//
// Defaults to listing the container metadata in a table format to stdout. If the --json flag is provided,
// then output the content as JSON instead.
func Ps(c *cli.Context) {
	docker := newDockerClient()

	containers := listECSLocalContainers(docker)

	if c.Bool(flags.JsonFlag) {
		displayAsJSON(containers)
	} else {
		displayAsTable(containers)
	}
}

func listECSLocalContainers(docker *client.Client) []types.Container {
	// ECS Task containers running locally all have an ECS local label
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", ECSLocalLabelKey),
		),
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers with label=%s due to %v", ECSLocalLabelKey, err)
	}
	return containers
}

func displayAsJSON(containers []types.Container) {
	data, err := json.MarshalIndent(containers, "", "  ")
	if err != nil {
		logrus.Fatalf("Failed to marshal containers to JSON due to %v", err)
	}
	fmt.Fprintln(os.Stdout, string(data))
}

func displayAsTable(containers []types.Container) {
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, cellWidthInSpaces, widthBetweenCellsInSpaces, cellPaddingInSpaces, paddingCharacter, noFormatting)
	fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tSTATUS\tPORTS\tNAMES\tTASKDEFINITIONARN\tTASKFILEPATH")
	for _, container := range containers {
		row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s",
			container.ID[:maxContainerIDLength],
			container.Image,
			container.Status,
			prettifyPorts(container.Ports),
			prettifyNames(container.Names),
			container.Labels["taskDefinitionARN"],
			container.Labels["taskFilePath"])
		fmt.Fprintln(w, row)
	}
	w.Flush()
}

func prettifyPorts(ports []types.Port) string {
	var prettyPorts []string
	for _, port := range ports {
		prettyPorts = append(prettyPorts, fmt.Sprintf("%s:%d->%d/%s", port.IP, port.PublicPort, port.PrivatePort, port.Type))
	}
	return strings.Join(prettyPorts, ", ")
}

func prettifyNames(names []string) string {
	return strings.Join(names, ", ")
}
