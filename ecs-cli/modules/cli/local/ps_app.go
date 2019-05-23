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
	"strings"
	"text/tabwriter"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

// Ps lists the status of the ECS tasks running locally.
func Ps(c *cli.Context) {
	docker := newDockerClient()

	containers := listECSLocalContainers(docker)

	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 20, 1, 3, ' ', 0) // FIXME
	fmt.Fprintln(w, "ID\tNAMES\tIMAGE\tTASKDEFINITIONARN\tTASKFILEPATH")
	for _, container := range containers {
		fmt.Fprintln(w, "%s\t%s\t%s\t%s\t%s",
			container.ID,
			strings.Join(container.Names, ", "),
			container.Image,
			container.Labels["taskDefinitionARN"],
			container.Labels["taskFilePath"])
	}
	w.Flush()
}

func listECSLocalContainers(docker *client.Client) []types.Container {
	// ECS Tasks running locally all have the label "ECSLocalTask=true"
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "ECSLocalTask"),
		),
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers due to %v", err)
	}
	return containers
}
