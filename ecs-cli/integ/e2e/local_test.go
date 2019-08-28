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
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
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
						defer os.Remove(tempFileName)
						defer os.Remove("docker-compose.ecs-local.yml")
						defer os.Remove("docker-compose.ecs-local.override.yml")
						args[3] = tempFileName
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"Successfully wrote docker-compose",
						})
					},
				},
				{
					args: []string{"local", "up", "-f", ""},
					execute: func(t *testing.T, args []string) {
						tempFileName := createLocalTaskDefFile(t)
						defer os.Remove(tempFileName)
						defer os.Remove("docker-compose.ecs-local.yml")
						defer os.Remove("docker-compose.ecs-local.override.yml")

						args[3] = tempFileName
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"Started container with ID",
						})
						downArgs := []string{"local", "down", "-f", tempFileName}
						stdout, err = integ.RunCmd(t, downArgs)

					},
				},
				{
					args: []string{"local", "up", "-c", "docker-compose.ecs-local.yml"},
					execute: func(t *testing.T, args []string) {
						tempFileName := createLocalTaskDefFile(t)
						defer os.Remove(tempFileName)
						defer os.Remove("docker-compose.ecs-local.yml")
						defer os.Remove("docker-compose.ecs-local.override.yml")

						createArgs := []string{"local", "create", "-f", tempFileName}
						stdout, err := integ.RunCmd(t, createArgs)
						stdout, err = integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"Created the amazon-ecs-local-container-endpoints container",
						})
						downArgs := []string{"local", "down", "-f", tempFileName}
						stdout, err = integ.RunCmd(t, downArgs)
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
	require.NoError(t, err, "Failed to create task-definition.json")

	_, err = tmpfile.Write([]byte(content))
	require.NoError(t, err, "Failed to write to %s", tmpfile.Name())

	err = tmpfile.Close()
	require.NoError(t, err, "Failed to close %s", tmpfile.Name())

	return tmpfile.Name()
}
