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

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/ecs"

	"github.com/stretchr/testify/require"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/cmd"
)

// TestCreateClusterWithEC2Task runs the sequence of ecs-cli commands from
// the EC2 tutorial: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-cli-tutorial-ec2.html
func TestCreateClusterWithEC2Task(t *testing.T) {
	t.Parallel()

	// Create the cluster
	conf := cmd.TestEC2TutorialConfig(t)
	cmd.TestUp(t,
		conf,
		cmd.WithCapabilityIAM(),
		cmd.WithInstanceType("t2.medium"),
		cmd.WithSize(2),
	)
	ecs.TestClusterSize(t, conf.ClusterName, 2)

	// Create the files for a task definition
	project := cmd.NewProject("e2e-ec2-tutorial-taskdef", conf.ConfigName)
	project.ComposeFileName = createEC2TutorialComposeFile(t)
	project.ECSParamsFileName = createEC2TutorialECSParamsFile(t)
	defer os.Remove(project.ComposeFileName)
	defer os.Remove(project.ECSParamsFileName)

	// Create a new task
	cmd.TestTaskUp(t, project)
	ecs.TestListTasks(t, conf.ClusterName, 1)
	cmd.TestPsRunning(t, project, 2)

	// Increase the number of running tasks
	cmd.TestTaskScale(t, project, 2)
	ecs.TestListTasks(t, conf.ClusterName, 2)
	cmd.TestPsRunning(t, project, 4)

	// Delete the task
	cmd.TestTaskDown(t, project)
	ecs.TestListTasks(t, conf.ClusterName, 0)

	// Delete the cluster
	cmd.TestDown(t, conf)
}

func createEC2TutorialComposeFile(t *testing.T) string {
	content := `
version: '3'
services:
  wordpress:
    image: wordpress
    ports:
      - "80:80"
    links:
      - mysql
  mysql:
    image: mysql:5.7
    environment:
      MYSQL_ROOT_PASSWORD: password`

	tmpfile, err := ioutil.TempFile("", "docker-compose-*.yml")
	require.NoError(t, err, "Failed to create docker-compose.yml")

	_, err = tmpfile.Write([]byte(content))
	require.NoErrorf(t, err, "Failed to write to %s", tmpfile.Name())

	err = tmpfile.Close()
	require.NoErrorf(t, err, "Failed to close %s", tmpfile.Name())

	t.Logf("Created %s successfully", tmpfile.Name())
	return tmpfile.Name()
}

func createEC2TutorialECSParamsFile(t *testing.T) string {
	content := `
version: 1
task_definition:
  services:
    wordpress:
      cpu_shares: 100
      mem_limit: 524288000
    mysql:
      cpu_shares: 100
      mem_limit: 524288000`
	tmpfile, err := ioutil.TempFile("", "ecs-params-*.yml")
	require.NoError(t, err, "Failed to create ecs-params.yml")

	_, err = tmpfile.Write([]byte(content))
	require.NoErrorf(t, err, "Failed to write to %s", tmpfile.Name())

	err = tmpfile.Close()
	require.NoErrorf(t, err, "Failed to close %s", tmpfile.Name())

	t.Logf("Created %s successfully", tmpfile.Name())
	return tmpfile.Name()
}
