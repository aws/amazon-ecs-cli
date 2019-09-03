// +build integ

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package e2e

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

func TestECSLocal(t *testing.T) {

	type commandTest struct {
		args    []string
		execute func(t *testing.T, args []string)
	}

	tests := map[string]struct {
		sequence []commandTest
	}{
		"clean state": {
			sequence: []commandTest{
				{
					args: []string{"local", "ps"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"CONTAINER ID",
							"IMAGE",
							"STATUS",
							"PORTS",
							"NAMES",
							"TASKDEFINITION",
						})
					},
				},
				{
					args: []string{"local", "ps", "--all"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"CONTAINER ID",
							"IMAGE",
							"STATUS",
							"PORTS",
							"NAMES",
							"TASKDEFINITION",
						})
					},
				},
				{
					args: []string{"local", "ps", "--all", "--json"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"[]",
						})
					},
				},
				{
					args: []string{"local", "down"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"No running ECS local tasks found",
						})
					},
				},
				{
					args: []string{"local", "down", "--all"},
					execute: func(t *testing.T, args []string) {

						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"No running ECS local tasks found",
						})
					},
				},
				{
					args: []string{"local", "down", "-f", "task-definition.json"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"Searching for containers from local file task-definition.json",
							"No running ECS local tasks found",
						})
					},
				},
			},
		},
		"from task def": {
			sequence: []commandTest{
				{
					args: []string{"local", "create", "-f", ""},
					execute: func(t *testing.T, args []string) {
						tempFileName := createLocalTaskDefFile(t)
						defer cleanUp(tempFileName)
						args[3] = tempFileName
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"Successfully wrote docker-compose",
						})
						checkDockerComposeContents(
							t,
							"docker-compose.ecs-local.yml",
							"docker-compose.ecs-local.override.yml",
							tempFileName,
						)

					},
				},
				{
					args: []string{"local", "up", "-f", ""},
					execute: func(t *testing.T, args []string) {
						tempFileName := createLocalTaskDefFile(t)
						defer cleanUp(tempFileName)
						args[3] = tempFileName
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"Started container with ID",
						})

						checkDockerComposeContents(t,
							"docker-compose.ecs-local.yml",
							"docker-compose.ecs-local.override.yml",
							tempFileName,
						)

						dockerCli := getDockerClient(t)
						containers := checkNumberRunningContainers(t, dockerCli, 2)
						containerNames := getContainerNames(t, containers)
						require.Contains(t, containerNames, "httpd:2.4")

						downArgs := []string{"local", "down", "-f", tempFileName}
						stdout, err = integ.RunCmd(t, downArgs)

						checkNumberRunningContainers(t, dockerCli, 0)
					},
				},
				{
					args: []string{"local", "up", "-c", "docker-compose.ecs-local.yml"},
					execute: func(t *testing.T, args []string) {
						tempFileName := createLocalTaskDefFile(t)
						defer cleanUp(tempFileName)
						createArgs := []string{"local", "create", "-f", tempFileName}
						stdout, err := integ.RunCmd(t, createArgs)
						stdout, err = integ.RunCmd(t, args)
						require.NoError(t, err)
						checkDockerComposeContents(t,
							"docker-compose.ecs-local.yml",
							"docker-compose.ecs-local.override.yml",
							tempFileName,
						)
						stdout.TestHasAllSubstrings(t, []string{
							"Created the amazon-ecs-local-container-endpoints container",
						})
						dockerCli := getDockerClient(t)
						containers := checkNumberRunningContainers(t, dockerCli, 2)

						containerNames := getContainerNames(t, containers)
						require.Contains(t, containerNames, "httpd:2.4")

						downArgs := []string{"local", "down", "-f", tempFileName}
						stdout, err = integ.RunCmd(t, downArgs)
					},
				},
				{
					args: []string{"local", "up"},
					execute: func(t *testing.T, args []string) {
						tempFileName := createLocalTaskDefFile(t)
						os.Rename(tempFileName, "task-definition.json")
						defer cleanUp("task-definition.json")
						stdout, err := integ.RunCmd(t, args)
						stdout.TestHasAllSubstrings(t, []string{
							"Successfully wrote docker-compose.ecs-local.yml",
							"Started container with ID",
						})
						checkDockerComposeContents(
							t,
							"docker-compose.ecs-local.yml",
							"docker-compose.ecs-local.override.yml",
							"task-definition.json",
						)
						dockerCli := getDockerClient(t)
						containers := checkNumberRunningContainers(t, dockerCli, 2)
						containerNames := getContainerNames(t, containers)
						require.Contains(t, containerNames, "httpd:2.4")

						downArgs := []string{"local", "down", "--all"}
						stdout, err = integ.RunCmd(t, downArgs)
						stdout.TestHasAllSubstrings(t, []string{
							"Stopped container with name",
							"Removed container with name",
							"Removed network with name",
						})
						require.NoError(t, err)
						checkNumberRunningContainers(t, dockerCli, 0)
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, cmd := range tc.sequence {
				cmd.execute(t, cmd.args)
			}
		})
	}
}

func cleanUp(taskDefFile string) {
	err := os.Remove(taskDefFile)
	err = os.Remove("docker-compose.ecs-local.yml")
	err = os.Remove("docker-compose.ecs-local.override.yml")
	if err != nil {
		panic(fmt.Sprintf("error %v removing files!", err))
	}
}

func createLocalTaskDefFile(t *testing.T) string {
	content := `
{
	"containerDefinitions": [
		{
			"entryPoint": [
				"sh",
				"-c"
			],
			"essential": true,
			"image": "httpd:2.4",
			"name": "simple-test-app",
			"portMappings": [
				{
				"containerPort": 80,
				"hostPort": 80,
				"protocol": "tcp"
				}
			]
		}
	],
	"cpu": "256",
	"memory": "512",
	"networkMode": "bridge"
	}`

	tmpfile, err := ioutil.TempFile(".", "task-definition-*.json")
	require.NoError(t, err, "Failed to create task-definition-*.json")

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err, "Failed to write to %s", tmpfile.Name())

	err = tmpfile.Close()
	require.NoError(t, err, "Failed to close %s", tmpfile.Name())

	return tmpfile.Name()
}

func getFullTaskDefPath(t *testing.T, taskDefinitionFilename string) string {
	workingDirectory, err := os.Getwd()
	require.NoError(t, err, "error getting working directory")
	fullPath := path.Join(workingDirectory, taskDefinitionFilename)
	_, err = os.Stat(fullPath)
	require.NoError(t, err, "Could not resolve full task definition path. File does not exist at the provided location")
	return fullPath
}

func checkDockerComposeContents(t *testing.T, composeFileName string, overrideFileName string, taskDefinitionFile string) {
	taskDefinitionFullPath := getFullTaskDefPath(t, taskDefinitionFile)
	contentsCompose := fmt.Sprintf(
		`version: "3.4"
services:
  simple-test-app:
    entrypoint:
    - sh
    - -c
    environment:
      AWS_CONTAINER_CREDENTIALS_RELATIVE_URI: /creds
      ECS_CONTAINER_METADATA_URI: http://169.254.170.2/v3
    image: httpd:2.4
    labels:
      ecs-local.task-definition-input.type: local
      ecs-local.task-definition-input.value: %s
    networks:
      ecs-local-network: null
    ports:
    - target: 80
      published: 80
      protocol: tcp
networks:
  ecs-local-network:
    external: true
`, taskDefinitionFullPath)
	contentsComposeOverride := `version: "3.4"
services:
  simple-test-app:
    environment:
      AWS_CONTAINER_CREDENTIALS_RELATIVE_URI: /creds
    logging:
      driver: json-file
`
	checkFileContents(t, composeFileName, contentsCompose)
	checkFileContents(t, overrideFileName, contentsComposeOverride)
}

func checkFileContents(t *testing.T, fileName string, testData string) {
	t.Helper()
	file, err := os.Open(fileName)
	require.NoError(t, err, "could not open file to check its contents")
	defer file.Close()
	data, err := ioutil.ReadAll(file)

	require.Equal(t, string(data), testData, "Docker compose contents do not match desired test data!")
}

func getContainerNames(t *testing.T, containers []types.Container) []string {
	t.Helper()
	var names []string
	for _, ctr := range containers {
		names = append(names, ctr.Image)
	}
	return names
}

func getDockerClient(t *testing.T) *client.Client {
	t.Helper()
	dockerCli, err := client.NewClientWithOpts(
		client.WithVersion("1.39"),
	)
	require.NoError(t, err)
	return dockerCli
}

func checkNumberRunningContainers(t *testing.T, cli *client.Client, expectedContainers int) []types.Container {
	t.Helper()
	containers, err := cli.ContainerList(
		context.Background(),
		types.ContainerListOptions{All: true},
	)
	require.NoError(t, err)
	require.Equal(
		t,
		len(containers),
		expectedContainers,
		"Expected %d containers to be found by Docker SDK.",
		expectedContainers,
	)
	return containers
}
