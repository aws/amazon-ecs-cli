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
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/converter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/docker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/localproject"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/options"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
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
	if err := options.ValidateFlagPairs(c); err != nil {
		logrus.Fatal(err.Error())
	}
	containers, pathname, err := listContainers(c)
	if err != nil {
		logrus.Fatalf("Failed to list containers for %s due to:\n%v", pathname, err)
	}
	if err = displayContainers(c, pathname, containers); err != nil {
		logrus.Fatalf("Failed to display containers for %s due to:\n%v", pathname, err)
	}
}

func listContainers(c *cli.Context) ([]types.Container, string, error) {
	if c.IsSet(flags.TaskDefinitionFile) {
		file, err := filepath.Abs(c.String(flags.TaskDefinitionFile))
		if err != nil {
			return nil, "", err
		}
		containers, err := listContainersWithFilters(filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", converter.TaskDefinitionLabelValue, file)),
			filters.Arg("label", fmt.Sprintf("%s=%s", converter.TaskDefinitionLabelType, localproject.LocalTaskDefType)),
		))
		return containers, file, err
	}
	if c.IsSet(flags.TaskDefinitionRemote) {
		file := c.String(flags.TaskDefinitionRemote)
		containers, err := listContainersWithFilters(filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("%s=%s", converter.TaskDefinitionLabelValue,
				c.String(flags.TaskDefinitionRemote))),
			filters.Arg("label", fmt.Sprintf("%s=%s", converter.TaskDefinitionLabelType, localproject.RemoteTaskDefType)),
		))
		return containers, file, err
	}
	if c.Bool(flags.All) {
		file := "ALL containers"
		containers, err := listContainersWithFilters(filters.NewArgs(
			filters.Arg("label", converter.TaskDefinitionLabelValue),
		))
		return containers, file, err
	}

	defaultFile, err := filepath.Abs(localproject.LocalInFileName)
	if err != nil {
		return nil, "", err
	}

	containers, err := listContainersWithFilters(filters.NewArgs(
		filters.Arg("label", fmt.Sprintf("%s=%s", converter.TaskDefinitionLabelValue, defaultFile)),
		filters.Arg("label", fmt.Sprintf("%s=%s", converter.TaskDefinitionLabelType, localproject.LocalTaskDefType)),
	))

	return containers, defaultFile, err
}

func listContainersWithFilters(args filters.Args) ([]types.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	cl := docker.NewClient()
	containers, err := cl.ContainerList(ctx, types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list containers with args=%v", args)
	}
	return containers, nil
}

func displayContainers(c *cli.Context, pathname string, containers []types.Container) error {
	logrus.Infof("Displaying containers for %s", pathname)
	if c.Bool(flags.JSON) {
		return displayAsJSON(containers)
	} else {
		return displayAsTable(containers)
	}
}

func displayAsJSON(containers []types.Container) error {
	data, err := json.MarshalIndent(containers, jsonPrefix, jsonIndent)
	if err != nil {
		return errors.Wrap(err, "failed to marshal containers to JSON")
	}
	fmt.Fprintln(os.Stdout, string(data))
	return nil
}

func displayAsTable(containers []types.Container) error {
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
			container.Labels[converter.TaskDefinitionLabelValue])
		fmt.Fprintln(w, row)
	}
	return w.Flush()
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
