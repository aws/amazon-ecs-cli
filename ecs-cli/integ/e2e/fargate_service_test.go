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

// Package e2e contains the end-to-end integration tests for the ECS CLI.
package e2e

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/cmd"

	"github.com/stretchr/testify/require"
)

// TestCreateClusterWithFargateService runs the sequence of ecs-cli commands from
// the Fargate tutorial: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-cli-tutorial-fargate.html
func TestCreateClusterWithFargateService(t *testing.T) {

	// Create the cluster
	conf := cmd.TestFargateTutorialConfig(t)
	vpc := cmd.TestUp(t, conf)

	// Create the files for a task definition
	project := cmd.NewProject("e2e-fargate-test-service", conf.ConfigName)
	project.ComposeFileName = createFargateTutorialComposeFile(t)
	project.ECSParamsFileName = createFargateTutorialECSParamsFile(t, vpc.Subnets)
	defer os.Remove(project.ComposeFileName)
	defer os.Remove(project.ECSParamsFileName)

	// Create a new service
	cmd.TestServiceUp(t, project)
	cmd.TestServicePs(t, project, 1)

	// Increase the number of running tasks
	cmd.TestServiceScale(t, project, 2)
	cmd.TestServicePs(t, project, 2)

	// Delete the service
	cmd.TestServiceDown(t, project)

	// Delete the cluster
	cmd.TestDown(t, conf)
}

func createFargateTutorialComposeFile(t *testing.T) string {
	content := `
version: '3'
services:
  wordpress:
    image: wordpress
    stop_grace_period: "1m30s"
    ports:
      - "80:80"
    logging:
      driver: awslogs
      options:
        awslogs-group: tutorial
        awslogs-region: us-east-1
        awslogs-stream-prefix: wordpress`

	tmpfile, err := ioutil.TempFile("", "docker-compose-*.yml")
	require.NoError(t, err, "Failed to create docker-compose.yml")

	_, err = tmpfile.Write([]byte(content))
	require.NoErrorf(t, err, "Failed to write to %s", tmpfile.Name())

	err = tmpfile.Close()
	require.NoErrorf(t, err, "Failed to close %s", tmpfile.Name())

	t.Logf("Created %s successfully", tmpfile.Name())
	return tmpfile.Name()
}

func createFargateTutorialECSParamsFile(t *testing.T, subnets []string) string {
	content := `
version: 1
task_definition:
  task_execution_role: ecsTaskExecutionRole
  ecs_network_mode: awsvpc
  task_size:
    mem_limit: 0.5GB
    cpu_limit: 256
run_params:
  network_configuration:
    awsvpc_configuration:
      assign_public_ip: ENABLED
      subnets:`
	for _, subnet := range subnets {
		content += `
        - "` + subnet + `"`
	}
	tmpfile, err := ioutil.TempFile("", "ecs-params-*.yml")
	require.NoError(t, err, "Failed to create ecs-params.yml")

	_, err = tmpfile.Write([]byte(content))
	require.NoErrorf(t, err, "Failed to write to %s", tmpfile.Name())

	err = tmpfile.Close()
	require.NoErrorf(t, err, "Failed to close %s", tmpfile.Name())

	t.Logf("Created %s successfully", tmpfile.Name())
	return tmpfile.Name()
}
