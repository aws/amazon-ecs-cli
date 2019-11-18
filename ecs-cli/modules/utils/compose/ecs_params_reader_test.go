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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
)

func TestReadECSParams(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		assert.Equal(t, "1", ecsParams.Version, "Expected version to match")
		taskDef := ecsParams.TaskDefinition
		assert.Equal(t, "host", taskDef.NetworkMode, "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/my_role", taskDef.TaskRoleArn, "Expected task role ARN to match")
		// Should still populate other fields with empty values
		assert.Empty(t, taskDef.ExecutionRole)
		awsvpcConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration
		assert.Empty(t, awsvpcConfig.Subnets)
		assert.Empty(t, awsvpcConfig.SecurityGroups)
	}
}

func TestReadECSParams_FileDoesNotExist(t *testing.T) {
	_, err := ReadECSParams("nonexistant.yml")
	assert.Error(t, err)
}

func TestReadECSParams_NoFile(t *testing.T) {
	ecsParams, err := ReadECSParams("")
	if assert.NoError(t, err) {
		assert.Nil(t, ecsParams)
	}
}

func TestReadECSParams_WithServices(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role
  services:
    log_router:
      firelens_configuration:
        type: fluentbit
        options:
          enable-ecs-log-metadata: "true"
    mysql:
      essential: false
      cpu_shares: 100
      mem_limit: 524288000
      mem_reservation: 500mb
    wordpress:
      essential: true
      repository_credentials:
        credentials_parameter: arn:aws:secretsmanager:1234567890:secret:test-RT4iv
      logging:
        secret_options:
          - value_from: arn:aws:ssm:eu-west-1:111111111111:parameter/mysecrets/dbpassword
            name: DB_PASSWORD
          - value_from: /mysecrets/dbusername
            name: DB_USERNAME
      secrets:
        - value_from: arn:aws:ssm:eu-west-1:111111111111:parameter/mysecrets/dbpassword
          name: DB_PASSWORD
        - value_from: /mysecrets/dbusername
          name: DB_USERNAME`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		taskDef := ecsParams.TaskDefinition
		assert.Equal(t, "host", ecsParams.TaskDefinition.NetworkMode, "Expected NetworkMode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/my_role", taskDef.TaskRoleArn, "Expected TaskRoleArn to match")

		containerDefs := taskDef.ContainerDefinitions
		assert.Equal(t, 3, len(containerDefs), "Expected 3 containers")

		mysql := containerDefs["mysql"]
		wordpress := containerDefs["wordpress"]
		log_router := containerDefs["log_router"]

		assert.False(t, mysql.Essential, "Expected container to not be essential")
		assert.Equal(t, int64(100), mysql.Cpu)
		assert.Equal(t, yaml.MemStringorInt(524288000), mysql.Memory)
		assert.Equal(t, yaml.MemStringorInt(524288000), mysql.MemoryReservation)
		assert.True(t, wordpress.Essential, "Expected container to be essential")
		assert.Equal(t, "arn:aws:secretsmanager:1234567890:secret:test-RT4iv", wordpress.RepositoryCredentials.CredentialsParameter, "Expected CredentialsParameter to match")

		expectedSecrets := []Secret{
			Secret{
				ValueFrom: "arn:aws:ssm:eu-west-1:111111111111:parameter/mysecrets/dbpassword",
				Name:      "DB_PASSWORD",
			},
			Secret{
				ValueFrom: "/mysecrets/dbusername",
				Name:      "DB_USERNAME",
			},
		}

		assert.ElementsMatch(t, expectedSecrets, wordpress.Secrets, "Expected secrets to match")
		assert.ElementsMatch(t, expectedSecrets, wordpress.Logging.SecretOptions, "Expected secrets to match")

		assert.Equal(t, "fluentbit", log_router.FirelensConfiguration.Type, "Except firelens_configuration type to be fluentbit")
		assert.Equal(t, "true", log_router.FirelensConfiguration.Options["enable-ecs-log-metadata"], "Expected Firelens 'enable-ecs-log-metadata' to be 'true'")
	}
}

func TestReadECSParams_WithRunParams(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: awsvpc
run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: [subnet-feedface, subnet-deadbeef]
      security_groups:
        - sg-bafff1ed
        - sg-c0ffeefe`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		taskDef := ecsParams.TaskDefinition
		assert.Equal(t, "awsvpc", taskDef.NetworkMode, "Expected network mode to match")

		awsvpcConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration
		assert.Equal(t, 2, len(awsvpcConfig.Subnets), "Expected 2 subnets")
		assert.Equal(t, []string{"subnet-feedface", "subnet-deadbeef"}, awsvpcConfig.Subnets, "Expected subnets to match")
		assert.Equal(t, 2, len(awsvpcConfig.SecurityGroups), "Expected 2 securityGroups")
		assert.Equal(t, []string{"sg-bafff1ed", "sg-c0ffeefe"}, awsvpcConfig.SecurityGroups, "Expected security groups to match")
		assert.Equal(t, AssignPublicIp(""), awsvpcConfig.AssignPublicIp, "Expected AssignPublicIP to be empty")
	}
}

// Task Size, Task Execution Role, and Assign Public Ip are required for Fargate tasks
func TestReadECSParams_WithFargateRunParams(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: awsvpc
  task_execution_role: arn:aws:iam::123456789012:role/fargate_role
  task_size:
    mem_limit: 0.5GB
    cpu_limit: 256
run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: [subnet-feedface, subnet-deadbeef]
      security_groups:
        - sg-bafff1ed
        - sg-c0ffeefe
      assign_public_ip: ENABLED`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		taskDef := ecsParams.TaskDefinition
		assert.Equal(t, "awsvpc", taskDef.NetworkMode, "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/fargate_role", taskDef.ExecutionRole)
		assert.Equal(t, "0.5GB", taskDef.TaskSize.Memory)
		assert.Equal(t, "256", taskDef.TaskSize.Cpu)

		awsvpcConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration
		assert.Equal(t, 2, len(awsvpcConfig.Subnets), "Expected 2 subnets")
		assert.Equal(t, []string{"subnet-feedface", "subnet-deadbeef"}, awsvpcConfig.Subnets, "Expected subnets to match")
		assert.Equal(t, 2, len(awsvpcConfig.SecurityGroups), "Expected 2 securityGroups")
		assert.Equal(t, []string{"sg-bafff1ed", "sg-c0ffeefe"}, awsvpcConfig.SecurityGroups, "Expected security groups to match")
		assert.Equal(t, Enabled, awsvpcConfig.AssignPublicIp, "Expected AssignPublicIp to match")
	}
}

func TestReadECSParams_WithTaskPlacement(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  placement_constraints:
    - type: memberOf
      expression: attribute:ecs.os-type == linux
run_params:
  task_placement:
    strategy:
      - field: memory
        type: binpack
      - field: attribute:ecs.availability-zone
        type: spread
    constraints:
      - expression: attribute:ecs.instance-type =~ t2.*
        type: memberOf
      - type: distinctInstance`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	expectedStrategies := []Strategy{
		{
			Field: "memory",
			Type:  ecs.PlacementStrategyTypeBinpack,
		},
		{
			Field: "attribute:ecs.availability-zone",
			Type:  ecs.PlacementStrategyTypeSpread,
		},
	}

	expectedConstraints := []Constraint{
		{
			Expression: "attribute:ecs.instance-type =~ t2.*",
			Type:       ecs.PlacementConstraintTypeMemberOf,
		},
		{
			Type: ecs.PlacementConstraintTypeDistinctInstance,
		},
	}

	expectedTaskDefConstraints := []Constraint{
		{
			Expression: "attribute:ecs.os-type == linux",
			Type:       ecs.PlacementConstraintTypeMemberOf,
		},
	}

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		taskPlacement := ecsParams.RunParams.TaskPlacement
		strategies := taskPlacement.Strategies
		constraints := taskPlacement.Constraints
		assert.Len(t, strategies, 2)
		assert.Len(t, constraints, 2)
		assert.ElementsMatch(t, expectedStrategies, strategies)
		assert.ElementsMatch(t, expectedConstraints, constraints)

		taskDefConstraints := ecsParams.TaskDefinition.PlacementConstraints
		assert.ElementsMatch(t, expectedTaskDefConstraints, taskDefConstraints)
	}
}

func TestReadECSParams_MemoryWithUnits(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: awsvpc
  task_size:
    mem_limit: 0.5GB
    cpu_limit: 256`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		taskSize := ecsParams.TaskDefinition.TaskSize
		assert.Equal(t, "256", taskSize.Cpu, "Expected CPU limit to match")
		assert.Equal(t, "0.5GB", taskSize.Memory, "Expected Memory limit to match")
	}
}

// Task Size must match specific CPU/Memory buckets, but we leave validation to ECS.
func TestReadECSParams_WithTaskSize(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  task_size:
    mem_limit: 1024
    cpu_limit: 256`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		taskSize := ecsParams.TaskDefinition.TaskSize
		assert.Equal(t, "256", taskSize.Cpu, "Expected CPU limit to match")
		assert.Equal(t, "1024", taskSize.Memory, "Expected Memory limit to match")
	}
}

func TestReadECSParams_WithPIDandIPC(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  pid_mode: host
  ipc_mode: task`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		assert.Equal(t, "host", ecsParams.TaskDefinition.PIDMode, "Expected PIDMode to be set")
		assert.Equal(t, "task", ecsParams.TaskDefinition.IPCMode, "Expected IPCMode to be set")
	}
}

/** ConvertToECSNetworkConfiguration tests **/

func TestConvertToECSNetworkConfiguration(t *testing.T) {
	taskDef := EcsTaskDef{NetworkMode: "awsvpc"}
	subnets := []string{"subnet-feedface"}
	securityGroups := []string{"sg-c0ffeefe"}
	awsVpconfig := AwsVpcConfiguration{
		Subnets:        subnets,
		SecurityGroups: securityGroups,
	}

	networkConfig := NetworkConfiguration{
		AwsVpcConfiguration: awsVpconfig,
	}

	ecsParams := &ECSParams{
		TaskDefinition: taskDef,
		RunParams: RunParams{
			NetworkConfiguration: networkConfig,
		},
	}

	ecsNetworkConfig, err := ConvertToECSNetworkConfiguration(ecsParams)

	if assert.NoError(t, err) {
		ecsAwsConfig := ecsNetworkConfig.AwsvpcConfiguration
		assert.Equal(t, subnets[0], aws.StringValue(ecsAwsConfig.Subnets[0]), "Expected subnets to match")
		assert.Equal(t, securityGroups[0], aws.StringValue(ecsAwsConfig.SecurityGroups[0]), "Expected securityGroups to match")
		assert.Nil(t, ecsAwsConfig.AssignPublicIp, "Expected AssignPublicIp to be nil")
	}
}

func TestConvertToECSNetworkConfiguration_NoSecurityGroups(t *testing.T) {
	taskDef := EcsTaskDef{NetworkMode: "awsvpc"}
	subnets := []string{"subnet-feedface"}
	awsVpconfig := AwsVpcConfiguration{
		Subnets: subnets,
	}

	networkConfig := NetworkConfiguration{
		AwsVpcConfiguration: awsVpconfig,
	}

	ecsParams := &ECSParams{
		TaskDefinition: taskDef,
		RunParams: RunParams{
			NetworkConfiguration: networkConfig,
		},
	}

	ecsNetworkConfig, err := ConvertToECSNetworkConfiguration(ecsParams)

	if assert.NoError(t, err) {
		ecsAwsConfig := ecsNetworkConfig.AwsvpcConfiguration
		assert.Equal(t, subnets[0], aws.StringValue(ecsAwsConfig.Subnets[0]), "Expected subnets to match")
		assert.Nil(t, ecsAwsConfig.AssignPublicIp, "Expected AssignPublicIp to be nil")
	}
}

func TestConvertToECSNetworkConfiguration_ErrorWhenNoSubnets(t *testing.T) {
	taskDef := EcsTaskDef{NetworkMode: "awsvpc"}
	subnets := []string{}

	awsVpconfig := AwsVpcConfiguration{
		Subnets: subnets,
	}

	networkConfig := NetworkConfiguration{
		AwsVpcConfiguration: awsVpconfig,
	}

	ecsParams := &ECSParams{
		TaskDefinition: taskDef,
		RunParams: RunParams{
			NetworkConfiguration: networkConfig,
		},
	}

	_, err := ConvertToECSNetworkConfiguration(ecsParams)

	assert.Error(t, err)
}

func TestConvertToECSNetworkConfiguration_WhenNoECSParams(t *testing.T) {
	ecsParams, err := ConvertToECSNetworkConfiguration(nil)

	if assert.NoError(t, err) {
		assert.Nil(t, ecsParams)
	}
}

func TestConvertToECSNetworkConfiguration_WithAssignPublicIp(t *testing.T) {
	taskDef := EcsTaskDef{NetworkMode: "awsvpc"}
	subnets := []string{"subnet-feedface"}
	awsVpconfig := AwsVpcConfiguration{
		Subnets:        subnets,
		AssignPublicIp: Enabled,
	}

	networkConfig := NetworkConfiguration{
		AwsVpcConfiguration: awsVpconfig,
	}

	ecsParams := &ECSParams{
		TaskDefinition: taskDef,
		RunParams: RunParams{
			NetworkConfiguration: networkConfig,
		},
	}

	ecsNetworkConfig, err := ConvertToECSNetworkConfiguration(ecsParams)

	if assert.NoError(t, err) {
		ecsAwsConfig := ecsNetworkConfig.AwsvpcConfiguration
		assert.Equal(t, subnets[0], aws.StringValue(ecsAwsConfig.Subnets[0]), "Expected subnets to match")
		assert.Equal(t, "ENABLED", aws.StringValue(ecsAwsConfig.AssignPublicIp), "Expected AssignPublicIp to match")
	}
}

func TestConvertToECSNetworkConfiguration_NoNetworkConfig(t *testing.T) {
	taskDef := EcsTaskDef{NetworkMode: "bridge"}

	ecsParams := &ECSParams{
		TaskDefinition: taskDef,
	}

	ecsNetworkConfig, err := ConvertToECSNetworkConfiguration(ecsParams)

	if assert.NoError(t, err) {
		assert.Nil(t, ecsNetworkConfig, "Expected AssignPublicIp to be nil")
	}
}

func TestConvertToECSPlacementConstraints(t *testing.T) {
	constraint1 := Constraint{
		Expression: "attribute:ecs.instance-type =~ t2.*",
		Type:       ecs.PlacementConstraintTypeMemberOf,
	}
	constraint2 := Constraint{
		Type: ecs.PlacementConstraintTypeDistinctInstance,
	}
	constraints := []Constraint{constraint1, constraint2}
	taskPlacement := TaskPlacement{
		Constraints: constraints,
	}

	ecsParams := &ECSParams{
		RunParams: RunParams{
			TaskPlacement: taskPlacement,
		},
	}

	expectedConstraints := []*ecs.PlacementConstraint{
		&ecs.PlacementConstraint{
			Expression: aws.String("attribute:ecs.instance-type =~ t2.*"),
			Type:       aws.String(ecs.PlacementConstraintTypeMemberOf),
		},
		&ecs.PlacementConstraint{
			Type: aws.String(ecs.PlacementConstraintTypeDistinctInstance),
		},
	}

	ecsPlacementConstraints, err := ConvertToECSPlacementConstraints(ecsParams)

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expectedConstraints, ecsPlacementConstraints, "Expected placement constraints to match")
	}
}

func TestConvertToECSPlacementStrategy(t *testing.T) {
	strategy1 := Strategy{
		Field: "instanceId",
		Type:  ecs.PlacementStrategyTypeBinpack,
	}
	strategy2 := Strategy{
		Field: "attribute:ecs.availability-zone",
		Type:  ecs.PlacementStrategyTypeSpread,
	}
	strategy3 := Strategy{
		Type: ecs.PlacementStrategyTypeRandom,
	}
	strategy := []Strategy{strategy1, strategy2, strategy3}
	taskPlacement := TaskPlacement{
		Strategies: strategy,
	}

	ecsParams := &ECSParams{
		RunParams: RunParams{
			TaskPlacement: taskPlacement,
		},
	}

	expectedStrategy := []*ecs.PlacementStrategy{
		&ecs.PlacementStrategy{
			Field: aws.String("instanceId"),
			Type:  aws.String(ecs.PlacementStrategyTypeBinpack),
		},
		&ecs.PlacementStrategy{
			Field: aws.String("attribute:ecs.availability-zone"),
			Type:  aws.String(ecs.PlacementStrategyTypeSpread),
		},
		&ecs.PlacementStrategy{
			Type: aws.String(ecs.PlacementStrategyTypeRandom),
		},
	}

	ecsPlacementStrategy, err := ConvertToECSPlacementStrategy(ecsParams)

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expectedStrategy, ecsPlacementStrategy, "Expected placement strategy to match")
	}
}
func TestReadECSParams_WithDockerVolumes(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  docker_volumes:
    - name: my_volume
      scope: shared
      autoprovision: true
      driver: doggyromcom
      driver_opts:
        pudding: is-engaged-to-marry-Tum-Tum
        clyde: professes-his-love-at-the-ceremony
        it: does-not-go-well
        this: is-not-a-movie
      labels:
        pudding: mad
        clyde: sad
        life: sucks`

	expectedVolumes := []DockerVolume{
		DockerVolume{
			Name:          "my_volume",
			Scope:         aws.String("shared"),
			Autoprovision: aws.Bool(true),
			Driver:        aws.String("doggyromcom"),
			DriverOptions: map[string]string{
				"pudding": "is-engaged-to-marry-Tum-Tum",
				"clyde":   "professes-his-love-at-the-ceremony",
				"it":      "does-not-go-well",
				"this":    "is-not-a-movie",
			},
			Labels: map[string]string{
				"pudding": "mad",
				"clyde":   "sad",
				"life":    "sucks",
			},
		},
	}

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		volumes := ecsParams.TaskDefinition.DockerVolumes
		assert.ElementsMatch(t, expectedVolumes, volumes, "Expected volumes to match")
	}
}

func TestReadECSParams_WithHealthCheck(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  services:
    mysql:
      healthcheck:
        test: ["CMD", "curl", "-f", "http://localhost"]
        interval: 1m30s
        timeout: 10s
        retries: 3
        start_period: 40s
    wordpress:
      healthcheck:
        command: ["CMD-SHELL", "curl -f http://localhost"]
        interval: 70
        timeout: 15
        retries: 5
        start_period: 40
    logstash:
      healthcheck:
        test: curl -f http://localhost
        interval: 10m
        timeout: 15s
        retries: 5
        start_period: 50
    elasticsearch:
      healthcheck:
        command: curl http://example.com
        interval: 10
        timeout: 15
        retries: 5
        start_period: 50s`

	mysqlExpectedHealthCheck := &HealthCheck{
		Test:        []string{"CMD", "curl", "-f", "http://localhost"},
		Command:     nil,
		Interval:    "1m30s",
		Timeout:     "10s",
		Retries:     3,
		StartPeriod: "40s",
	}

	wordpressExpectedHealthCheck := &HealthCheck{
		Command:     []string{"CMD-SHELL", "curl -f http://localhost"},
		Test:        nil,
		Interval:    "70",
		Timeout:     "15",
		Retries:     int64(5),
		StartPeriod: "40",
	}

	logstashExpectedHealthCheck := &HealthCheck{
		Test:        []string{"curl -f http://localhost"},
		Command:     nil,
		Interval:    "10m",
		Timeout:     "15s",
		Retries:     5,
		StartPeriod: "50",
	}

	elasticsearchExpectedHealthCheck := &HealthCheck{
		Command:     []string{"curl http://example.com"},
		Test:        nil,
		Interval:    "10",
		Timeout:     "15",
		Retries:     5,
		StartPeriod: "50s",
	}

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		taskDef := ecsParams.TaskDefinition

		containerDefs := taskDef.ContainerDefinitions
		assert.Equal(t, 4, len(containerDefs), "Expected 4 containers")

		mysql := containerDefs["mysql"]
		wordpress := containerDefs["wordpress"]
		logstash := containerDefs["logstash"]
		elasticsearch := containerDefs["elasticsearch"]

		assert.Equal(t, mysqlExpectedHealthCheck, mysql.HealthCheck)
		assert.Equal(t, wordpressExpectedHealthCheck, wordpress.HealthCheck)
		assert.Equal(t, logstashExpectedHealthCheck, logstash.HealthCheck)
		assert.Equal(t, elasticsearchExpectedHealthCheck, elasticsearch.HealthCheck)
	}
}

func TestReadECSParams_WithServiceDiscoveryAllFields(t *testing.T) {
	ecsParamsString := `version: 1
run_params:
  service_discovery:
    container_name: nginx
    container_port: 80
    private_dns_namespace:
      vpc: vpc-8BAADF00D
      id: ns-CA15CA15CA15CA15
      name: corp
      description: This is a private namespace
    public_dns_namespace:
      id: ns-C0VF3F3
      name: amazon.com
    service_discovery_service:
      name: mysds
      description: This is an SDS
      dns_config:
        type: A
        ttl: 60
      healthcheck_custom_config:
        failure_threshold: 1`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	if assert.NoError(t, err) {
		serviceDiscovery := ecsParams.RunParams.ServiceDiscovery
		assert.Equal(t, "nginx", serviceDiscovery.ContainerName, "Expected ContainerName to match")
		assert.Equal(t, aws.Int64(80), serviceDiscovery.ContainerPort, "Expected ContainerPort to match")
		assert.Equal(t, "vpc-8BAADF00D", serviceDiscovery.PrivateDNSNamespace.VPC, "Expected VPC to match")
		assert.Equal(t, "ns-CA15CA15CA15CA15", serviceDiscovery.PrivateDNSNamespace.ID, "Expected private namespace ID to match")
		assert.Equal(t, "corp", serviceDiscovery.PrivateDNSNamespace.Name, "Expected private namespace Name to match")
		assert.Equal(t, "This is a private namespace", serviceDiscovery.PrivateDNSNamespace.Description, "Expected private namespace description to match")
		assert.Equal(t, "ns-C0VF3F3", serviceDiscovery.PublicDNSNamespace.ID, "Expected public namespace ID to match")
		assert.Equal(t, "amazon.com", serviceDiscovery.PublicDNSNamespace.Name, "Expected public namespace name to match")
		sds := serviceDiscovery.ServiceDiscoveryService
		assert.Equal(t, "mysds", sds.Name, "Expected SDS Name to match")
		assert.Equal(t, "This is an SDS", sds.Description, "Expected SDS Description to match")
		assert.Equal(t, "A", sds.DNSConfig.Type, "Expected SDS DNSConfig Type to match")
		assert.Equal(t, aws.Int64(60), sds.DNSConfig.TTL, "Expected SDS DNSConfig TTL to match")
		assert.Equal(t, aws.Int64(1), sds.HealthCheckCustomConfig.FailureThreshold, "Expected SDS HealthCheckCustomConfig FailureThreshold to match")
	}
}

func TestConvertToECSHealthCheck(t *testing.T) {
	testHealthCheck := &HealthCheck{
		Test:        []string{"CMD-SHELL", "curl -f http://localhost"},
		Command:     nil,
		Interval:    "10m",
		Timeout:     "15s",
		Retries:     5,
		StartPeriod: "50s",
	}

	expected := &ecs.HealthCheck{
		Command:     aws.StringSlice([]string{"CMD-SHELL", "curl -f http://localhost"}),
		Interval:    aws.Int64(600),
		Timeout:     aws.Int64(15),
		Retries:     aws.Int64(5),
		StartPeriod: aws.Int64(50),
	}

	actual, err := testHealthCheck.ConvertToECSHealthCheck()

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual, "Expected healthcheck to match")
	}
}

func TestConvertToECSHealthCheck_AltFormat(t *testing.T) {
	testHealthCheck := &HealthCheck{
		Command:     []string{"CMD-SHELL", "curl -f http://localhost"},
		Test:        nil,
		Interval:    "600",
		Timeout:     "15",
		Retries:     5,
		StartPeriod: "50",
	}

	expected := &ecs.HealthCheck{
		Command:     aws.StringSlice([]string{"CMD-SHELL", "curl -f http://localhost"}),
		Interval:    aws.Int64(600),
		Timeout:     aws.Int64(15),
		Retries:     aws.Int64(5),
		StartPeriod: aws.Int64(50),
	}

	actual, err := testHealthCheck.ConvertToECSHealthCheck()

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual, "Expected healthcheck to match")
	}
}

func TestConvertToECSHealthCheck_PrependForStringCommand(t *testing.T) {
	testHealthCheck := &HealthCheck{
		Command: []string{"curl -f http://localhost"},
	}

	expected := []string{"CMD-SHELL", "curl -f http://localhost"}

	actual, err := testHealthCheck.ConvertToECSHealthCheck()
	if assert.NoError(t, err) {
		assert.Equal(t, aws.StringSlice(expected), actual.Command, "Expected healthcheck command to match")
	}

	// With Test key
	testHealthCheck = &HealthCheck{
		Test: []string{"curl -f http://localhost"},
	}

	actual, err = testHealthCheck.ConvertToECSHealthCheck()
	if assert.NoError(t, err) {
		assert.Equal(t, aws.StringSlice(expected), actual.Command, "Expected healthcheck command to match")
	}
}

func TestConvertToECSHealthCheck_ErrorCase_InvalidInterval(t *testing.T) {
	testHealthCheck := &HealthCheck{
		Test:        []string{"CMD", "curl", "-f", "http://localhost"},
		Command:     nil,
		Interval:    "cat",
		Timeout:     "10s",
		Retries:     3,
		StartPeriod: "40s",
	}
	_, err := testHealthCheck.ConvertToECSHealthCheck()

	assert.Error(t, err, "Expected error parsing interval field in healthcheck")
}

func TestConvertToECSHealthCheck_ErrorCase_TestAndCommand(t *testing.T) {
	testHealthCheck := &HealthCheck{
		Test:        []string{"CMD", "curl", "-f", "http://localhost"},
		Command:     []string{"CMD", "curl", "-f", "http://localhost"},
		Interval:    "5s",
		Timeout:     "10s",
		Retries:     3,
		StartPeriod: "40s",
	}
	_, err := testHealthCheck.ConvertToECSHealthCheck()

	assert.Error(t, err, "Expected error reading ecs-params: healthcheck test and command can not both be specified")
}

func TestConvertToECSHealthCheck_IntFieldsBlank(t *testing.T) {
	testHealthCheck := &HealthCheck{
		Command: []string{"CMD", "curl", "-f", "http://localhost"},
	}

	expected := &ecs.HealthCheck{
		Command:     aws.StringSlice([]string{"CMD", "curl", "-f", "http://localhost"}),
		Interval:    nil,
		Timeout:     nil,
		Retries:     nil,
		StartPeriod: nil,
	}

	actual, err := testHealthCheck.ConvertToECSHealthCheck()
	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestConvertToECSHealthCheck_TestFieldBlank(t *testing.T) {
	testHealthCheck := &HealthCheck{
		Command:     nil,
		Test:        nil,
		Interval:    "",
		Timeout:     "",
		Retries:     0,
		StartPeriod: "10",
	}

	expected := &ecs.HealthCheck{
		Command:     nil,
		Interval:    nil,
		Timeout:     nil,
		Retries:     nil,
		StartPeriod: aws.Int64(10),
	}

	actual, err := testHealthCheck.ConvertToECSHealthCheck()

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}
