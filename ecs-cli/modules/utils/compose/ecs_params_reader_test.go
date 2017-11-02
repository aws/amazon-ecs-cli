// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadECSParams(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		assert.Equal(t, "1", ecsParams.Version, "Expected version to match")
		taskDef := ecsParams.TaskDefinition
		assert.Equal(t, "host", taskDef.NetworkMode, "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/my_role", taskDef.TaskRoleArn, "Expected task role ARN to match")
	}
}

func TestReadECSParams_FileDoesNotExist(t *testing.T) {
	_, err := ReadECSParams("nonexistant.yml")
	assert.Error(t, err)
}

func TestReadECSParams_WithServices(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role
  services:
    mysql:
      essential: false
    wordpress:
      essential: true`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		containerDefs := ecsParams.TaskDefinition.ContainerDefinitions
		assert.Equal(t, 2, len(containerDefs), "Expected 2 containers")

		mysql := containerDefs["mysql"]
		wordpress := containerDefs["wordpress"]

		assert.False(t, mysql.Essential, "Expected container to not be essential")
		assert.True(t, wordpress.Essential, "Expected container to be essential")
	}
}

func TestReadECSParams_WithRunParams(t *testing.T) {
	ecsParamsString := `version: 1
run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: [subnet-feedface, subnet-deadbeef]
      security_groups:
        - sg-bafff1ed
        - sg-c0ffeefe
      assign_public_ip: enabled`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		awsvpcConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration
		assert.Equal(t, 2, len(awsvpcConfig.Subnets), "Expected 2 subnets")
		assert.Equal(t, []string{"subnet-feedface", "subnet-deadbeef"}, awsvpcConfig.Subnets, "Expected subnets to match")
		assert.Equal(t, 2, len(awsvpcConfig.SecurityGroups), "Expected 2 securityGroups")
		assert.Equal(t, []string{"sg-bafff1ed", "sg-c0ffeefe"}, awsvpcConfig.SecurityGroups, "Expected security groups to match")
		assert.Equal(t, Enabled, awsvpcConfig.AssignPublicIp, "Expected AssignPublicIp to be enabled")
	}
}
