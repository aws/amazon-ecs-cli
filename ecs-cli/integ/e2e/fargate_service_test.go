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

	"github.com/stretchr/testify/assert"
)

// TestCreateClusterWithFargateService runs the sequence of ecs-cli commands from
// the Fargate tutorial: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-cli-tutorial-fargate.html
func TestCreateClusterWithFargateService(t *testing.T) {
	// Create the cluster
	conf := cmd.TestFargateConfig(t)
	vpc := cmd.TestUp(t, conf)

	// Create the files for a task definition
	createComposeFile(t)
	createECSParamsFile(t, vpc.Subnets)

	// Create a new service
	project := cmd.NewProject("e2e-fargate-test-service", conf.ConfigName)
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

func createComposeFile(t *testing.T) {
	content := `
version: '3'
services:
  wordpress:
    image: wordpress
    ports:
      - "80:80"
    logging:
      driver: awslogs
      options: 
        awslogs-group: tutorial
        awslogs-region: us-east-1
        awslogs-stream-prefix: wordpress`
	err := ioutil.WriteFile("./docker-compose.yml", []byte(content), os.ModePerm)
	assert.NoError(t, err, "Failed to create docker-compose.yml")
}

func createECSParamsFile(t *testing.T, subnets []string) {
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
	err := ioutil.WriteFile("./ecs-params.yml", []byte(content), os.ModePerm)
	assert.NoError(t, err, "Failed to create ecs-params.yml")
}
