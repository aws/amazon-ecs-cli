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
	"reflect"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcredio"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/value"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
)

const (
	projectName    = "ProjectName"
	containerPath  = "/tmp/cache"
	containerPath2 = "/tmp/cache2"
	hostPath       = "./cache"
	namedVolume    = "named_volume"
	namedVolume2   = "named_volume2"
	namedVolume3   = "named_volume3"
)

var defaultNetwork = &yaml.Network{
	Name:     "default",
	RealName: "project_default",
}

// TODO Extract test docker file and use in test (to avoid gaps between parse and conversion unit tests)
var testContainerConfig = adapter.ContainerConfig{
	Name:    "mysql",
	Command: []string{"cmd"},
	CPU:     int64(131072),
	Devices: []*ecs.Device{
		{
			HostPath:      aws.String("/dev/sda"),
			ContainerPath: aws.String("/dev/sdd"),
			Permissions:   aws.StringSlice([]string{"read"}),
		},
		{
			HostPath: aws.String("/dev/sda"),
		},
	},
	DNSSearchDomains: []string{"search.example.com"},
	DNSServers:       []string{"1.2.3.4"},
	DockerLabels: map[string]*string{
		"label1":         aws.String(""),
		"com.foo.label2": aws.String("value"),
	},
	DockerSecurityOptions: []string{"label:type:test_virt"},
	Entrypoint:            []string{"/code/entrypoint.sh"},
	Environment: []*ecs.KeyValuePair{
		{
			Name:  aws.String("rails_env"),
			Value: aws.String("development"),
		},
	},
	ExtraHosts: []*ecs.HostEntry{
		{
			Hostname:  aws.String("test.local"),
			IpAddress: aws.String("127.10.10.10"),
		},
	},
	HealthCheck: &ecs.HealthCheck{
		Command:     aws.StringSlice([]string{"CMD-SHELL", "curl -f http://localhost"}),
		Interval:    aws.Int64(int64(70)),
		Timeout:     aws.Int64(int64(15)),
		Retries:     aws.Int64(int64(5)),
		StartPeriod: aws.Int64(int64(40)),
	},
	Hostname: "foobarbaz",
	Image:    "testimage",
	Links:    []string{"container1"},
	LogConfiguration: &ecs.LogConfiguration{
		LogDriver: aws.String("json-file"),
		Options: map[string]*string{
			"max-file": aws.String("50"),
			"max-size": aws.String("50k"),
		},
	},
	Memory:            int64(131072),
	MemoryReservation: int64(65536),
	MountPoints: []*ecs.MountPoint{
		{
			ContainerPath: aws.String("./code"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-0"),
		},
	},
	PortMappings: []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(5000),
			HostPort:      aws.Int64(5000),
			Protocol:      aws.String("tcp"),
		},
	},
	Privileged:     true,
	PseudoTerminal: true,
	ReadOnly:       true,
	ShmSize:        int64(128), // Realistically, we expect customers to specify sizes larger than the default of 64M
	Tmpfs: []*ecs.Tmpfs{
		{
			ContainerPath: aws.String("/tmp"),
			MountOptions:  aws.StringSlice([]string{"ro", "rw"}),
			Size:          aws.Int64(64),
		},
	},
	Ulimits: []*ecs.Ulimit{
		{
			Name:      aws.String("nofile"),
			HardLimit: aws.Int64(40000),
			SoftLimit: aws.Int64(20000),
		},
	},
	VolumesFrom: []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(true),
			SourceContainer: aws.String("web"),
		},
	},
	User:             "user",
	WorkingDirectory: "/var",
}

func TestConvertToTaskDefinition(t *testing.T) {
	// Expected values on container
	name := "mysql"
	cpu := int64(131072) // 128 * 1024
	command := []string{"cmd"}
	devices := []*ecs.Device{
		{
			HostPath:      aws.String("/dev/sda"),
			ContainerPath: aws.String("/dev/sdd"),
			Permissions:   aws.StringSlice([]string{"read"}),
		},
		{
			HostPath: aws.String("/dev/sda"),
		},
	}
	entryPoint := []string{"/code/entrypoint.sh"}
	env := []*ecs.KeyValuePair{
		{
			Name:  aws.String("rails_env"),
			Value: aws.String("development"),
		},
	}
	extraHosts := []*ecs.HostEntry{
		{
			Hostname:  aws.String("test.local"),
			IpAddress: aws.String("127.10.10.10"),
		},
	}
	dnsSearchDomains := []string{"search.example.com"}
	dnsServers := []string{"1.2.3.4"}
	dockerLabels := map[string]string{
		"label1":         "",
		"com.foo.label2": "value",
	}
	hostname := "foobarbaz"
	image := "testimage"
	links := []string{"container1"}
	logOpts := map[string]*string{
		"max-file": aws.String("50"),
		"max-size": aws.String("50k"),
	}
	logging := &ecs.LogConfiguration{
		LogDriver: aws.String("json-file"),
		Options:   logOpts,
	}
	memory := int64(131072) // 128 GiB = 131072 MiB
	memoryReservation := int64(65536)
	mountPoints := []*ecs.MountPoint{
		{
			ContainerPath: aws.String("./code"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-0"),
		},
	}
	ports := []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(5000),
			HostPort:      aws.Int64(5000),
			Protocol:      aws.String("tcp"),
		},
	}
	privileged := true
	pseudoterminal := true
	readOnly := true
	securityOpts := []string{"label:type:test_virt"}
	shmSize := int64(128)
	tmpfs := []*ecs.Tmpfs{
		{
			ContainerPath: aws.String("/tmp"),
			MountOptions:  aws.StringSlice([]string{"ro", "rw"}),
			Size:          aws.Int64(64),
		},
	}
	user := "user"
	ulimits := []*ecs.Ulimit{
		{
			Name:      aws.String("nofile"),
			HardLimit: aws.Int64(40000),
			SoftLimit: aws.Int64(20000),
		},
	}
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(true),
			SourceContainer: aws.String("web"),
		},
	}
	workingDir := "/var"

	// Expected values on task definition
	taskRoleArn := "arn:aws:iam::123456789012:role/my_role"
	// TODO add top-level volumes

	// convert
	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{testContainerConfig}, taskRoleArn, "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	containerDef := *taskDefinition.ContainerDefinitions[0]

	// verify task def fields
	assert.Equal(t, taskRoleArn, aws.StringValue(taskDefinition.TaskRoleArn), "Expected taskRoleArn to match")
	assert.Empty(t, taskDefinition.RequiresCompatibilities, "Did not expect RequiresCompatibilities to be set")
	// PID and IPC should be unset
	assert.Nil(t, taskDefinition.IpcMode, "Expected IpcMode to be nil")
	assert.Nil(t, taskDefinition.PidMode, "Expected PidMode to be nil")

	// verify container def fields
	assert.Equal(t, aws.String(name), containerDef.Name, "Expected container def name to match")
	assert.Equal(t, aws.StringSlice(command), containerDef.Command, "Expected container def command to match")
	assert.Equal(t, aws.Int64(cpu), containerDef.Cpu, "Expected container def cpu to match")
	assert.ElementsMatch(t, devices, containerDef.LinuxParameters.Devices, "Expected container def devices to match")
	assert.Equal(t, aws.StringSlice(dnsSearchDomains), containerDef.DnsSearchDomains, "Expected container def DNS Search Domains to match")
	assert.Equal(t, aws.StringSlice(dnsServers), containerDef.DnsServers, "Expected container def DNS Servers to match")
	assert.Equal(t, aws.StringMap(dockerLabels), containerDef.DockerLabels, "Expected container def Docker labels to match")
	assert.Equal(t, aws.StringSlice(securityOpts), containerDef.DockerSecurityOptions, "Expected container def docker security options to match")
	assert.Equal(t, env, containerDef.Environment, "Expected Environment to match")
	assert.Equal(t, aws.StringSlice(entryPoint), containerDef.EntryPoint, "Expected EntryPoint to be match")
	assert.Equal(t, extraHosts, containerDef.ExtraHosts, "Expected ExtraHosts to match")
	assert.Equal(t, aws.String(hostname), containerDef.Hostname, "Expected container def hostname to match")
	assert.Equal(t, aws.String(image), containerDef.Image, "Expected container def image to match")
	assert.Equal(t, aws.StringSlice(links), containerDef.Links, "Expected container def links to match")
	assert.Equal(t, logging, containerDef.LogConfiguration, "Expected LogConfiguration to match")
	assert.Equal(t, aws.Int64(memory), containerDef.Memory, "Expected container def memory to match")
	assert.Equal(t, aws.Int64(memoryReservation), containerDef.MemoryReservation, "Expected container def memoryReservation to match")
	assert.Equal(t, mountPoints, containerDef.MountPoints, "Expected MountPoints to match")

	assert.Equal(t, ports, containerDef.PortMappings, "Expected PortMappings to match")

	assert.Equal(t, aws.Bool(privileged), containerDef.Privileged, "Expected container def privileged to match")
	assert.Equal(t, aws.Bool(pseudoterminal), containerDef.PseudoTerminal, "Expected container def pseudoterminal to match")
	assert.Equal(t, aws.Bool(readOnly), containerDef.ReadonlyRootFilesystem, "Expected container def ReadonlyRootFilesystem to match")
	assert.Equal(t, aws.Int64(shmSize), containerDef.LinuxParameters.SharedMemorySize, "Expected sharedMemorySize to match")
	assert.ElementsMatch(t, tmpfs, containerDef.LinuxParameters.Tmpfs, "Expected tmpfs to match")
	assert.Equal(t, aws.String(user), containerDef.User, "Expected container def user to match")
	assert.ElementsMatch(t, ulimits, containerDef.Ulimits, "Expected Ulimits to match")
	assert.Equal(t, volumesFrom, containerDef.VolumesFrom, "Expected VolumesFrom to match")
	assert.Equal(t, aws.String(workingDir), containerDef.WorkingDirectory, "Expected container def WorkingDirectory to match")
	assert.Equal(t, testContainerConfig.HealthCheck, containerDef.HealthCheck)

	// If no containers are specified as being essential, all containers
	// are marked "essential"
	for _, container := range taskDefinition.ContainerDefinitions {
		assert.True(t, aws.BoolValue(container.Essential), "Expected essential to be true")
	}
}

func TestConvertToTaskDefinitionWithNoSharedMemorySize(t *testing.T) {
	containerConfig := adapter.ContainerConfig{
		ShmSize: int64(0),
	}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.SharedMemorySize, "Expected sharedMemorySize to be null")
}

func TestConvertToTaskDefinitionWithNoTmpfs(t *testing.T) {
	containerConfig := adapter.ContainerConfig{
		Tmpfs: nil,
	}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.Tmpfs, "Expected Tmpfs to be null")
}

func TestConvertToTaskDefinitionWithBlankHostname(t *testing.T) {
	containerConfig := adapter.ContainerConfig{
		Hostname: "",
	}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.Hostname, "Expected Hostname to be nil")
}

// TODO add test for nil cap add/cap drop

// Test Launch Types
func TestConvertToTaskDefinitionLaunchTypeEmpty(t *testing.T) {
	containerConfig := adapter.ContainerConfig{}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")
	if len(taskDefinition.RequiresCompatibilities) > 0 {
		t.Error("Did not expect RequiresCompatibilities to be set")
	}
}

func TestConvertToTaskDefinitionLaunchTypeEC2(t *testing.T) {
	containerConfig := adapter.ContainerConfig{}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "EC2", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	if len(taskDefinition.RequiresCompatibilities) != 1 {
		t.Error("Expected exactly one required compatibility to be set.")
	}
	assert.Equal(t, "EC2", aws.StringValue(taskDefinition.RequiresCompatibilities[0]))
}

func TestConvertToTaskDefinitionLaunchTypeFargate(t *testing.T) {
	containerConfig := adapter.ContainerConfig{}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "FARGATE", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	if len(taskDefinition.RequiresCompatibilities) != 1 {
		t.Error("Expected exactly one required compatibility to be set.")
	}
	assert.Equal(t, "FARGATE", aws.StringValue(taskDefinition.RequiresCompatibilities[0]))
}

// Tests for ConvertToTaskDefinition with ECS Params
func TestConvertToTaskDefinitionWithECSParams_ComposeMemoryLessThanMemoryRes(t *testing.T) {
	// set up containerConfig w/o value for Memory
	containerConfig := &adapter.ContainerConfig{
		Name:   "web",
		Image:  "httpd",
		CPU:    int64(5),
		Memory: int64(512),
	}

	// define ecs-params value we expect to be present in final containerDefinition
	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      mem_reservation: 1g`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	_, err = convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	assert.Error(t, err, "Expected error because memory reservation is greater than memory limit")
}

func testContainerConfigs(names []string) []adapter.ContainerConfig {
	containerConfigs := []adapter.ContainerConfig{}
	for _, name := range names {
		config := adapter.ContainerConfig{Name: name}
		containerConfigs = append(containerConfigs, config)
	}

	return containerConfigs
}

func TestConvertToTaskDefinitionWithECSParams(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	if assert.NoError(t, err) {
		// PID and IPC should be unset
		assert.Nil(t, taskDefinition.IpcMode, "Expected IpcMode to be nil")
		assert.Nil(t, taskDefinition.PidMode, "Expected PidMode to be nil")

		assert.Equal(t, "host", aws.StringValue(taskDefinition.NetworkMode), "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/my_role", aws.StringValue(taskDefinition.TaskRoleArn), "Expected task role ARN to match")

		// If no containers are specified as being essential, all
		// containers are marked "essential"
		for _, container := range taskDefinition.ContainerDefinitions {
			assert.True(t, aws.BoolValue(container.Essential), "Expected essential to be true")
		}
	}
}

func TestConvertToTaskDefinitionWithECSParams_AllFields(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: awsvpc
  task_role_arn: arn:aws:iam::123456789012:role/tweedledee
  services:
    mysql:
      essential: false
      init_process_enabled: true
      repository_credentials:
        credentials_parameter: arn:aws:secretsmanager:1234567890:secret:test-secret
  task_size:
    mem_limit: 5Gb
    cpu_limit: 256`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	mysql := findContainerByName("mysql", containerDefs)
	wordpress := findContainerByName("wordpress", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, "awsvpc", aws.StringValue(taskDefinition.NetworkMode), "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/tweedledee", aws.StringValue(taskDefinition.TaskRoleArn), "Expected task role ARN to match")

		assert.False(t, aws.BoolValue(mysql.Essential), "Expected container with name: '%v' to be false", *mysql.Name)
		assert.Equal(t, "256", aws.StringValue(taskDefinition.Cpu), "Expected CPU to match")
		assert.Equal(t, "5Gb", aws.StringValue(taskDefinition.Memory), "Expected CPU to match")
		assert.True(t, aws.BoolValue(wordpress.Essential), "Expected container with name: '%v' to be true", *wordpress.Name)
		assert.Equal(t, "arn:aws:secretsmanager:1234567890:secret:test-secret", aws.StringValue(mysql.RepositoryCredentials.CredentialsParameter), "Expected CredentialsParameter to match")
		assert.Equal(t, true, aws.BoolValue(mysql.LinuxParameters.InitProcessEnabled), "Expected container def initProcessEnabled to match")
	}
}

func TestConvertToTaskDefinitionWithECSParams_Essential_OneContainer(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  services:
    mysql:
      essential: false`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	mysql := findContainerByName("mysql", containerDefs)
	wordpress := findContainerByName("wordpress", containerDefs)

	if assert.NoError(t, err) {
		assert.False(t, aws.BoolValue(mysql.Essential), "Expected container with name: '%v' to be false", *mysql.Name)
		assert.True(t, aws.BoolValue(wordpress.Essential), "Expected container with name: '%v' to be true", *wordpress.Name)
	}
}

func TestConvertToTaskDefinitionWithECSParams_EssentialExplicitlyMarkedTrue(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  services:
    mysql:
      essential: true
    wordpress:
      essential: true`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	mysql := findContainerByName("mysql", containerDefs)
	wordpress := findContainerByName("wordpress", containerDefs)

	if assert.NoError(t, err) {
		assert.True(t, aws.BoolValue(mysql.Essential), "Expected container with name: '%v' to be true", *mysql.Name)
		assert.True(t, aws.BoolValue(wordpress.Essential), "Expected container with name: '%v' to be true", *wordpress.Name)
	}
}

func TestConvertToTaskDefinitionWithECSParams_EssentialExplicitlyMarked(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  services:
    mysql:
      essential: false
    wordpress:
      essential: true`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	mysql := findContainerByName("mysql", containerDefs)
	wordpress := findContainerByName("wordpress", containerDefs)

	if assert.NoError(t, err) {
		assert.False(t, aws.BoolValue(mysql.Essential), "Expected container with name: '%v' to be false", *mysql.Name)
		assert.True(t, aws.BoolValue(wordpress.Essential), "Expected container with name: '%v' to be true", *wordpress.Name)
	}
}

func TestConvertToTaskDefinitionWithECSParams_EssentialBlankForOneService(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role
  services:
    wordpress:`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	mysql := findContainerByName("mysql", containerDefs)
	wordpress := findContainerByName("wordpress", containerDefs)

	if assert.NoError(t, err) {
		assert.True(t, aws.BoolValue(mysql.Essential), "Expected container with name: '%v' to be true", *mysql.Name)
		assert.False(t, aws.BoolValue(wordpress.Essential), "Expected container with name: '%v' to be false", *wordpress.Name)
	}
}

func TestConvertToTaskDefinitionWithECSParams_EssentialBlankForAllServices(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role
  services:
    mysql:
    wordpress:`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	_, err = convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	assert.Error(t, err, "Expected error if no containers are marked essential")
}

func TestConvertToTaskDefinitionWithECSParams_AllContainersMarkedNotEssential(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  services:
    mysql:
      essential: false
    wordpress:
      essential: false`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	_, err = convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	assert.Error(t, err, "Expected error if no containers are marked essential")
}

func TestConvertToTaskDefinitionWithECSParams_EssentialDefaultsToTrueWhenNoServicesSpecified(t *testing.T) {
	// We expect essential to be set to be true in the converter
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)
	assert.NoError(t, err, "Unexpected error when no containers are marked essential")

	mysql := findContainerByName("mysql", taskDefinition.ContainerDefinitions)
	assert.True(t, *mysql.Essential, "Expected mysql to be essential")
	wordpress := findContainerByName("wordpress", taskDefinition.ContainerDefinitions)
	assert.True(t, *wordpress.Essential, "Expected wordpressto be essential")
}

func TestConvertToTaskDefinitionWithECSParams_EssentialDefaultsToTrueWhenNotSpecified(t *testing.T) {
	// We expect essential to be set to be true in the unmarshaller
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  services:
    wordpress:
      mem_limit: 1000000
    mysql:
      mem_limit: 1000000`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	assert.NoError(t, err, "Unexpected error when no containers are marked essential")
	mysql := findContainerByName("mysql", taskDefinition.ContainerDefinitions)
	assert.True(t, *mysql.Essential, "Expected mysql to be essential")
	wordpress := findContainerByName("wordpress", taskDefinition.ContainerDefinitions)
	assert.True(t, *wordpress.Essential, "Expected wordpressto be essential")
}

func TestConvertToTaskDefinitionWithECSParamsAndTaskRoleArnFlag(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/tweedledee`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskRoleArn := "arn:aws:iam::123456789012:role/tweedledum"

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, taskRoleArn, "", ecsParams, nil)

	if assert.NoError(t, err) {
		assert.Equal(t, "host", aws.StringValue(taskDefinition.NetworkMode), "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/tweedledum", aws.StringValue(taskDefinition.TaskRoleArn), "Expected task role arn to match")
	}
}

func TestConvertToTaskDefinitionWithECSParams_ContainerResourcesPresent(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})

	mysqlCPU := int64(100)
	mysqlMem := int64(15)
	mysqlMemRes := int64(10)

	wordpressCPU := int64(4)
	wordpressMem := int64(8)
	wordpressMemRes := int64(5)

	ecsParamsString := `version: 1
task_definition:
  services:
    mysql:
      cpu_shares: 100
      mem_limit: 15m
      mem_reservation: 10m
    wordpress:
      cpu_shares: 4
      mem_limit: 8m
      mem_reservation: 5m`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	mysql := findContainerByName("mysql", containerDefs)
	wordpress := findContainerByName("wordpress", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, mysqlCPU, aws.Int64Value(mysql.Cpu), "Expected CPU to match")
		assert.Equal(t, mysqlMem, aws.Int64Value(mysql.Memory), "Expected Memory to match")
		assert.Equal(t, mysqlMemRes, aws.Int64Value(mysql.MemoryReservation), "Expected MemoryReservation to match")

		assert.Equal(t, wordpressCPU, aws.Int64Value(wordpress.Cpu), "Expected CPU to match")
		assert.Equal(t, wordpressMem, aws.Int64Value(wordpress.Memory), "Expected Memory to match")
		assert.Equal(t, wordpressMemRes, aws.Int64Value(wordpress.MemoryReservation), "Expected MemoryReservation to match")
	}
}

func TestConvertToTaskDefinitionWithECSParams_ContainerResourcesOverrideProvidedVals(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:              "web",
		Image:             "httpd",
		CPU:               int64(2),
		Memory:            int64(3),
		MemoryReservation: int64(3),
	}

	// define ecs-params values we expect to override containerConfig vals
	webCPU := int64(5)
	webMem := int64(15)
	webMemRes := int64(10)

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      cpu_shares: 5
      mem_limit: 15m
      mem_reservation: 10m`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, webCPU, aws.Int64Value(web.Cpu), "Expected CPU to match")
		assert.Equal(t, webMem, aws.Int64Value(web.Memory), "Expected Memory to match")
		assert.Equal(t, webMemRes, aws.Int64Value(web.MemoryReservation), "Expected MemoryReservation to match")
	}
}

func TestConvertToTaskDefinitionWithECSParams_NoMemoryProvided(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
	}

	// define ecs-params values we expect to override containerConfig vals
	webCPU := int64(5)

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      cpu_shares: 5`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, webCPU, aws.Int64Value(web.Cpu), "Expected CPU to match")
		assert.Equal(t, int64(defaultMemLimit), aws.Int64Value(web.Memory), "Expected Memory to match default")
		assert.Empty(t, aws.Int64Value(web.MemoryReservation), "Expected MemoryReservation to be empty")
	}
}

func TestConvertToTaskDefinitionWithECSParams_MemLimitOnlyProvided(t *testing.T) {
	// set up containerConfig w/o value for Memory
	containerConfig := &adapter.ContainerConfig{
		Name:              "web",
		Image:             "httpd",
		CPU:               int64(5),
		Memory:            int64(15),
		MemoryReservation: int64(10),
	}

	// define ecs-params value we expect to be present in final containerDefinition
	webMem := int64(20)

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      mem_limit: 20m`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	if assert.NoError(t, err) {
		// check ecs-params override value is present
		assert.Equal(t, webMem, aws.Int64Value(web.Memory), "Expected Memory to match")
		// check config values not present in ecs-params are present
		assert.Equal(t, containerConfig.CPU, aws.Int64Value(web.Cpu), "Expected CPU to match")
		assert.Equal(t, containerConfig.MemoryReservation, aws.Int64Value(web.MemoryReservation), "Expected MemoryReservation to match")
	}
}

func TestConvertToTaskDefinitionWithECSParams_MemReservationOnlyProvided(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"web"})

	// define ecs-params value we expect to be present in final containerDefinition
	webMem := int64(20)

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      mem_reservation: 20m`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, webMem, aws.Int64Value(web.MemoryReservation), "Expected MemoryReservation to match")
		// check mem_limit not set
		assert.Empty(t, web.Memory, "Expected Memory to be nil")
		// check config values not present in ecs-params are present
		assert.Empty(t, web.Cpu, "Expected CPU to be empty")
	}
}

func TestConvertToTaskDefinitionWithECSParams_MemResGreaterThanMemLimit(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{Name: "web"}
	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      mem_limit: 10m
      mem_reservation: 15m`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	_, err = convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	assert.Error(t, err, "Expected error if mem_reservation was more than mem_limit")
}

func TestConvertToTaskDefinitionWithECSParams_WithTaskSize(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  task_size:
    mem_limit: 10MB
    cpu_limit: 200`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	if assert.NoError(t, err) {
		assert.Equal(t, "200", aws.StringValue(taskDefinition.Cpu), "Expected CPU to match")
		assert.Equal(t, "10MB", aws.StringValue(taskDefinition.Memory), "Expected CPU to match")
	}
}

func TestConvertToTaskDefinition_MemLimitOnlyProvided(t *testing.T) {
	webMem := int64(1048576)
	containerConfig := adapter.ContainerConfig{
		Name:   "web",
		Memory: webMem,
	}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	assert.Equal(t, webMem, aws.Int64Value(web.Memory), "Expected Memory to match")
	assert.Empty(t, web.MemoryReservation, "Expected MemoryReservation to be empty")
	assert.Empty(t, web.Cpu, "Expected CPU to be empty")
}

func TestConvertToTaskDefinition_MemReservationOnlyProvided(t *testing.T) {
	webMem := int64(1048576)
	containerConfig := adapter.ContainerConfig{
		Name:              "web",
		MemoryReservation: webMem,
	}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	assert.Empty(t, web.Memory, "Expected Memory to be nil")
	assert.Equal(t, webMem, aws.Int64Value(web.MemoryReservation), "Expected MemoryReservation to match")
	assert.Empty(t, web.Cpu, "Expected CPU to be empty")
}

func TestConvertToTaskDefinition_NoMemoryProvided(t *testing.T) {
	containerConfig := adapter.ContainerConfig{
		Name: "web",
	}

	taskDefinition, err := convertToTaskDefinitionForTest(t, []adapter.ContainerConfig{containerConfig}, "", "", nil, nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	assert.Equal(t, aws.Int64(defaultMemLimit), web.Memory, "Expected Memory to match default")
	assert.Empty(t, web.MemoryReservation, "Expected MemoryReservation to be empty")
	assert.Empty(t, web.Cpu, "Expected CPU to be empty")
}

func TestMemReservationHigherThanMemLimit(t *testing.T) {
	containerConfig := adapter.ContainerConfig{
		Memory:            int64(524288),
		MemoryReservation: int64(1048576),
	}

	volumeConfigs := adapter.NewVolumes()
	containerConfigs := []adapter.ContainerConfig{containerConfig}

	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                volumeConfigs,
		ContainerConfigs:       containerConfigs,
		ECSParams:              nil,
	}

	_, err := ConvertToTaskDefinition(testParams)
	assert.EqualError(t, err, "mem_limit must be greater than mem_reservation")
}

func TestConvertToTaskDefinitionWithECSParams_PIDandIPC(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
	}

	ecsParamsString := `version: 1
task_definition:
  pid_mode: task
  ipc_mode: host`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	if assert.NoError(t, err) {
		assert.Equal(t, "task", aws.StringValue(taskDefinition.PidMode))
		assert.Equal(t, "host", aws.StringValue(taskDefinition.IpcMode))
	}
}

func TestConvertToTaskDefinitionWithVolumes(t *testing.T) {
	volumeConfigs := &adapter.Volumes{
		VolumeWithHost: map[string]string{
			hostPath: containerPath,
		},
		VolumeEmptyHost: []string{namedVolume},
	}

	mountPoints := []*ecs.MountPoint{
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-0"),
		},
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("named_volume"),
		},
	}
	containerConfig := adapter.ContainerConfig{
		MountPoints: mountPoints,
	}

	containerConfigs := []adapter.ContainerConfig{containerConfig}

	host := &ecs.HostVolumeProperties{SourcePath: aws.String(hostPath)}
	expectedVolumes := []*ecs.Volume{
		{
			Host: host,
			Name: aws.String(containerPath),
		},
		{
			Name: aws.String(namedVolume),
		},
	}

	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                volumeConfigs,
		ContainerConfigs:       containerConfigs,
		ECSParams:              nil,
	}

	taskDefinition, err := ConvertToTaskDefinition(testParams)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	actualVolumes := taskDefinition.Volumes
	assert.ElementsMatch(t, expectedVolumes, actualVolumes, "Expected volumes to match")
}

func TestConvertToTaskDefinitionWithVolumesWithHostOnly(t *testing.T) {
	volumeConfigs := &adapter.Volumes{
		VolumeWithHost: map[string]string{
			hostPath: containerPath,
		},
	}

	mountPoints := []*ecs.MountPoint{
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-0"),
		},
	}
	containerConfig := adapter.ContainerConfig{
		MountPoints: mountPoints,
	}

	containerConfigs := []adapter.ContainerConfig{containerConfig}

	host := &ecs.HostVolumeProperties{SourcePath: aws.String(hostPath)}
	expectedVolumes := []*ecs.Volume{
		{
			Host: host,
			Name: aws.String(containerPath),
		},
	}

	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                volumeConfigs,
		ContainerConfigs:       containerConfigs,
		ECSParams:              nil,
	}

	taskDefinition, err := ConvertToTaskDefinition(testParams)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	actualVolumes := taskDefinition.Volumes
	assert.ElementsMatch(t, expectedVolumes, actualVolumes, "Expected volumes to match")
}

func TestConvertToTaskDefinitionWithECSParamsVolumeWithoutNameError(t *testing.T) {
	volumeConfigs := &adapter.Volumes{
		VolumeEmptyHost: []string{namedVolume, namedVolume2},
	}

	mountPoints := []*ecs.MountPoint{
		{
			ContainerPath: aws.String("/var/log"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("named_volume"),
		},
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("named_volume2"),
		},
	}
	containerConfig := adapter.ContainerConfig{
		MountPoints: mountPoints,
	}

	containerConfigs := []adapter.ContainerConfig{containerConfig}
	labels := map[string]string{
		"testing.thisdoesntactuallyreallyadvancetheplot": "true",
	}
	options := map[string]string{
		"Clyde": "says Goodbye Stranger, decides to Take The Long Way Home, and enjoys some Breakfast in America",
		"He":    "is a big fan of 70s music",
	}

	ecsParams := &ECSParams{
		TaskDefinition: EcsTaskDef{
			DockerVolumes: []DockerVolume{
				DockerVolume{
					Autoprovision: aws.Bool(true),
					Scope:         aws.String("shared"),
					Driver:        nil,
					DriverOptions: options,
					Labels:        labels,
				},
			},
		},
	}

	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                volumeConfigs,
		ContainerConfigs:       containerConfigs,
		ECSParams:              ecsParams,
	}

	_, err := ConvertToTaskDefinition(testParams)
	assert.Error(t, err, "Expected error converting Task Definition with ECS Params volume without name")
}

func TestConvertToTaskDefinitionWithEFSVolume(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
		MountPoints: []*ecs.MountPoint{{
			SourceVolume:  aws.String("myEFSVolume"),
			ContainerPath: aws.String("/mount/efs"),
			ReadOnly:      aws.Bool(true),
		}},
	}

	ecsParamsString := `version: 1
task_definition:
  efs_volumes:
    - name: myEFSVolume
      filesystem_id: fs-1234
      root_directory: /
      transit_encryption: "DISABLED"
      iam: "DISABLED"`

	ecsParams, err := createTempECSParamsForTest(t, ecsParamsString)
	assert.NoError(t, err)

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	volumes := adapter.Volumes{
		VolumeWithHost: map[string]string{
			"/mount/efs": "myEFSVolume",
		},
	}
	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                &volumes,
		ContainerConfigs:       containerConfigs,
		ECSParams:              ecsParams,
		ECSRegistryCreds:       nil,
	}

	taskDefinition, err := ConvertToTaskDefinition(testParams)
	assert.NoError(t, err)
	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)
	mp := web.MountPoints[0]
	if assert.NoError(t, err) {
		assert.Equal(t, *mp.SourceVolume, "myEFSVolume")
		assert.Equal(t, mp.ContainerPath, aws.String("/mount/efs"))
		assert.Equal(t, *taskDefinition.Volumes[0].Name, "myEFSVolume")
	}
}
func TestConvertToTaskDefinitionWithEFSVolumeNoId(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
		MountPoints: []*ecs.MountPoint{{
			SourceVolume:  aws.String("myEFSVolume"),
			ContainerPath: aws.String("/mount/efs"),
			ReadOnly:      aws.Bool(true),
		}},
	}
	ecsParamsString := `version: 1
task_definition:
  efs_volumes:
    - name: myEFSVolume
      root_directory: /
      transit_encryption: "DISABLED"
      iam: "DISABLED"`

	ecsParams, err := createTempECSParamsForTest(t, ecsParamsString)
	assert.NoError(t, err)

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	volumes := adapter.Volumes{
		VolumeWithHost: map[string]string{
			"/mount/efs": "myEFSVolume",
		},
	}
	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                &volumes,
		ContainerConfigs:       containerConfigs,
		ECSParams:              ecsParams,
		ECSRegistryCreds:       nil,
	}

	_, err = ConvertToTaskDefinition(testParams)
	assert.EqualError(t, err, "file system id is required for efs volumes")
}

func TestConvertToTaskDefinitionWithEFSVolumeAuthError(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
		MountPoints: []*ecs.MountPoint{{
			SourceVolume:  aws.String("myEFSVolume"),
			ContainerPath: aws.String("/mount/efs"),
			ReadOnly:      aws.Bool(true),
		}},
	}
	ecsParamsString := `version: 1
task_definition:
  efs_volumes:
    - name: myEFSVolume
      filesystem_id: fs-1234
      root_directory: /
      transit_encryption: "DISABLED"
      iam: "ENABLED"`

	ecsParams, err := createTempECSParamsForTest(t, ecsParamsString)
	assert.NoError(t, err)

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	volumes := adapter.Volumes{
		VolumeWithHost: map[string]string{
			"/mount/efs": "myEFSVolume",
		},
	}
	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                &volumes,
		ContainerConfigs:       containerConfigs,
		ECSParams:              ecsParams,
		ECSRegistryCreds:       nil,
	}

	_, err = ConvertToTaskDefinition(testParams)
	assert.EqualError(t, err, "Transit encryption is required when using IAM access or an access point")
}

func TestConvertToTaskDefinitionWithEFSVolumeAPError(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
		MountPoints: []*ecs.MountPoint{{
			SourceVolume:  aws.String("myEFSVolume"),
			ContainerPath: aws.String("/mount/efs"),
			ReadOnly:      aws.Bool(true),
		}},
	}
	ecsParamsString := `version: 1
task_definition:
  efs_volumes:
    - name: myEFSVolume
      filesystem_id: fs-1234
      root_directory: /
      transit_encryption: "DISABLED"
      access_point: "ap-1234"`

	ecsParams, err := createTempECSParamsForTest(t, ecsParamsString)
	assert.NoError(t, err)

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	volumes := adapter.Volumes{
		VolumeWithHost: map[string]string{
			"/mount/efs": "myEFSVolume",
		},
	}
	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            "",
		RequiredCompatibilites: "",
		Volumes:                &volumes,
		ContainerConfigs:       containerConfigs,
		ECSParams:              ecsParams,
		ECSRegistryCreds:       nil,
	}

	_, err = ConvertToTaskDefinition(testParams)
	assert.EqualError(t, err, "Transit encryption is required when using IAM access or an access point")
}

func TestIsZeroForEmptyConfig(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{}

	configValue := reflect.ValueOf(containerConfig).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		f := configValue.Field(i)
		ft := configType.Field(i)
		isZero := value.IsZero(f)
		if !isZero {
			t.Errorf("Expected field [%s] to be zero but was not", ft.Name)
		}
	}
}

func TestIsZeroWhenConfigHasValues(t *testing.T) {
	hasValues := map[string]bool{
		"CPU":               true,
		"Command":           true,
		"Hostname":          true,
		"Image":             true,
		"Links":             true,
		"Memory":            true,
		"MemoryReservation": true,
		"Privileged":        true,
		"ReadOnly":          true,
		"User":              true,
		"WorkingDirectory":  true,
	}

	containerConfig := &adapter.ContainerConfig{
		CPU:               int64(10),
		Command:           []string{"cmd"},
		Hostname:          "foobarbaz",
		Image:             "testimage",
		Links:             []string{"container1"},
		Memory:            int64(104857600),
		MemoryReservation: int64(52428800),
		Privileged:        true,
		ReadOnly:          true,
		User:              "user",
		WorkingDirectory:  "/var",
	}

	configValue := reflect.ValueOf(containerConfig).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		f := configValue.Field(i)
		ft := configType.Field(i)
		fieldName := ft.Name

		zeroValue := value.IsZero(f)
		_, hasValue := hasValues[fieldName]
		assert.NotEqual(t, zeroValue, hasValue)
	}
}

func TestConvertToTaskDefinitionWithECSParams_HealthCheck(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
	}

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      healthcheck:
        test: curl -f http://localhost
        interval: 10m
        timeout: 15s
        retries: 5
        start_period: 50`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"CMD-SHELL", "curl -f http://localhost"}, aws.StringValueSlice(web.HealthCheck.Command))
		assert.Equal(t, aws.Int64(600), web.HealthCheck.Interval)
		assert.Equal(t, aws.Int64(15), web.HealthCheck.Timeout)
		assert.Equal(t, aws.Int64(5), web.HealthCheck.Retries)
		assert.Equal(t, aws.Int64(50), web.HealthCheck.StartPeriod)
	}
}

func TestConvertToTaskDefinitionWithECSParams_OnlyTaskMemProvided(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name: "web",
	}
	ecsParams := &ECSParams{
		TaskDefinition: EcsTaskDef{
			TaskSize: TaskSize{
				Memory: "1gb",
			},
		},
	}

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	assert.NoError(t, err)
	assert.Empty(t, web.Memory, "Expected Memory to be nil")
	assert.Equal(t, taskDefinition.Memory, aws.String("1gb"))
}

func TestConvertToTaskDefinitionWithECSParams_HealthCheckOverrideCompose(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
		HealthCheck: &ecs.HealthCheck{
			Command:     aws.StringSlice([]string{"CMD", "curl", "-f", "http://example.com"}),
			Interval:    aws.Int64(int64(91)),
			Timeout:     aws.Int64(int64(17)),
			Retries:     aws.Int64(int64(7)),
			StartPeriod: aws.Int64(int64(73)),
		},
	}

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      healthcheck:
        test: curl -f http://localhost
        interval: 10m
        timeout: 15s
        retries: 5
        start_period: 50`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, []string{"CMD-SHELL", "curl -f http://localhost"}, aws.StringValueSlice(web.HealthCheck.Command))
		assert.Equal(t, aws.Int64(600), web.HealthCheck.Interval)
		assert.Equal(t, aws.Int64(15), web.HealthCheck.Timeout)
		assert.Equal(t, aws.Int64(5), web.HealthCheck.Retries)
		assert.Equal(t, aws.Int64(50), web.HealthCheck.StartPeriod)
	}
}

func TestConvertToTaskDefinitionWithECSRegistryCreds(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	credsFileString := `version: "1"
registry_credential_outputs:
  task_execution_role: someTestRole
  container_credentials:
    my.example.registry.net:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net
      container_names:
      - mysql`

	content := []byte(credsFileString)

	tmpfile, err := ioutil.TempFile("", regcredio.ECSCredFileBaseName)
	assert.NoError(t, err, "Could not create reg creds tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to reg creds tempfile")

	regCreds, err := regcredio.ReadCredsOutput(tmpfile.Name())
	assert.NoError(t, err, "Could not read reg creds file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", nil, regCreds)
	assert.NoError(t, err, "Unexpected error when converting task definition")
	assert.Equal(t, "someTestRole", aws.StringValue(taskDefinition.ExecutionRoleArn))

	mysqlContainer := findContainerByName("mysql", taskDefinition.ContainerDefinitions)
	assert.NotEmpty(t, mysqlContainer)
	assert.NotEmpty(t, mysqlContainer.RepositoryCredentials)
	assert.Equal(t, "arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net", aws.StringValue(mysqlContainer.RepositoryCredentials.CredentialsParameter))
}

func TestConvertToTaskDefinitionWithECSRegistryCreds_EmptyContainerCredMap(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	credsNoContainersFileString := `version: "1"
registry_credential_outputs:
  task_execution_role: someTestRole
  container_credentials:
    my.example.registry.net:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net
      container_names:`

	content := []byte(credsNoContainersFileString)

	tmpfile, err := ioutil.TempFile("", regcredio.ECSCredFileBaseName)
	assert.NoError(t, err, "Could not create reg creds tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to reg creds tempfile")

	regCreds, err := regcredio.ReadCredsOutput(tmpfile.Name())
	assert.NoError(t, err, "Could not read reg creds file")

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", nil, regCreds)
	assert.NoError(t, err, "Unexpected error when converting task definition")
	assert.Equal(t, "someTestRole", aws.StringValue(taskDefinition.ExecutionRoleArn))

	mysqlContainer := findContainerByName("mysql", taskDefinition.ContainerDefinitions)
	assert.NotEmpty(t, mysqlContainer)
	assert.Empty(t, mysqlContainer.RepositoryCredentials)
}

func TestConvertToTaskDefinitionWithECSRegistryCreds_OverrideECSParamsValues(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})

	// set up reg cred file
	credsFileString := `version: "1"
registry_credential_outputs:
  task_execution_role: someTestRole
  container_credentials:
    my.example.registry.net:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net
      container_names:
      - mysql`

	credsContent := []byte(credsFileString)

	tmpfileCreds, err := ioutil.TempFile("", regcredio.ECSCredFileBaseName)
	assert.NoError(t, err, "Could not create reg creds tempfile")

	defer os.Remove(tmpfileCreds.Name())

	_, err = tmpfileCreds.Write(credsContent)
	assert.NoError(t, err, "Could not write data to reg creds tempfile")

	regCreds, err := regcredio.ReadCredsOutput(tmpfileCreds.Name())
	assert.NoError(t, err, "Could not read reg creds file")

	// set up ecs-params
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_execution_role: arn:aws:iam::123456789012:role/my_role
  services:
    mysql:
      essential: true
      cpu_shares: 100
      mem_limit: 524288000
      repository_credentials:
        credentials_parameter: arn:aws:secretsmanager:1234567890:secret:test-RT4iv`

	contentECSParams := []byte(ecsParamsString)

	tmpfileECSParams, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs-params tempfile")

	ecsParamsFileName := tmpfileECSParams.Name()
	defer os.Remove(tmpfileECSParams.Name())

	_, err = tmpfileECSParams.Write(contentECSParams)
	assert.NoError(t, err, "Could not write data to ecs-params tempfile")

	err = tmpfileECSParams.Close()
	assert.NoError(t, err, "Could not close ecs-params tempfile")

	ecsParams, err := ReadECSParams(ecsParamsFileName)

	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, regCreds)
	assert.NoError(t, err, "Unexpected error when converting task definition")
	assert.Equal(t, "someTestRole", aws.StringValue(taskDefinition.ExecutionRoleArn))

	mysqlContainer := findContainerByName("mysql", taskDefinition.ContainerDefinitions)
	assert.NotEmpty(t, mysqlContainer)
	assert.NotEmpty(t, mysqlContainer.RepositoryCredentials)
	assert.Equal(t, "arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net", aws.StringValue(mysqlContainer.RepositoryCredentials.CredentialsParameter))
}

func TestConvertToTaskDefinitionWithECSRegistryCreds_ErrorOnDuplicateContainers(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	credsDuplicateContainersFileString := `version: "1"
registry_credential_outputs:
  task_execution_role: someTestRole
  container_credentials:
    my.example.registry.net:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net
      container_names:
      - mysql
    my.otherregistry.net:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.otherregistry.net
      container_names:
      - mysql`

	content := []byte(credsDuplicateContainersFileString)

	tmpfile, err := ioutil.TempFile("", regcredio.ECSCredFileBaseName)
	assert.NoError(t, err, "Could not create reg creds tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to reg creds tempfile")

	regCreds, err := regcredio.ReadCredsOutput(tmpfile.Name())
	assert.NoError(t, err, "Could not read reg creds file")

	_, err = convertToTaskDefinitionForTest(t, containerConfigs, "", "", nil, regCreds)
	assert.Error(t, err, "Expected error when converting task definition")
}

func TestConvertToTaskDefinitionWithECSParams_FirelensConfiguration(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "log_router",
		Image: "amazon/aws-for-fluent-bit",
	}

	ecsParamsString := `version: 1
task_definition:
  services:
    log_router:
      firelens_configuration:
        type: fluentbit
        options:
           enable-ecs-log-metadata: "true"`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	log_router := findContainerByName("log_router", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, "fluentbit", aws.StringValue(log_router.FirelensConfiguration.Type))
		assert.Equal(t, "true", aws.StringValue(log_router.FirelensConfiguration.Options["enable-ecs-log-metadata"]))
	}
}

func TestConvertToTaskDefinitionWithECSParams_Secrets(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "wordpress",
	}

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      secrets:
        - value_from: /mysecrets/dbusername
          name: DB_USERNAME
        - value_from: arn:aws:ssm:eu-west-1:111111111111:parameter/mysecrets/dbpassword
          name: DB_PASSWORD`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	expectedSecrets := []*ecs.Secret{
		&ecs.Secret{
			ValueFrom: aws.String("arn:aws:ssm:eu-west-1:111111111111:parameter/mysecrets/dbpassword"),
			Name:      aws.String("DB_PASSWORD"),
		},
		&ecs.Secret{
			ValueFrom: aws.String("/mysecrets/dbusername"),
			Name:      aws.String("DB_USERNAME"),
		},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expectedSecrets, web.Secrets, "Expected secrets to match")
	}
}

func TestConvertToTaskDefinitionWithECSParams_LoggingSecretOptions(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "wordpress",
		LogConfiguration: &ecs.LogConfiguration{
			LogDriver: aws.String("json-file"),
		},
	}

	ecsParamsString := `version: 1
task_definition:
  services:
    web:
      logging:
        secret_options:
          - value_from: /mysecrets/dbusername
            name: DB_USERNAME
          - value_from: arn:aws:ssm:eu-west-1:111111111111:parameter/mysecrets/dbpassword
            name: DB_PASSWORD`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := tmpfile.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")

	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	expectedSecretOptions := []*ecs.Secret{
		&ecs.Secret{
			ValueFrom: aws.String("arn:aws:ssm:eu-west-1:111111111111:parameter/mysecrets/dbpassword"),
			Name:      aws.String("DB_PASSWORD"),
		},
		&ecs.Secret{
			ValueFrom: aws.String("/mysecrets/dbusername"),
			Name:      aws.String("DB_USERNAME"),
		},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t, expectedSecretOptions, web.LogConfiguration.SecretOptions, "Expected secrets to match")
	}
}

func TestConvertToTaskDefinitionWithECSParams_Gpu(t *testing.T) {
	expectedGpuValue := "2"
	content := `version: 1
task_definition:
  services:
    web:
      gpu: ` + expectedGpuValue
	ecsParams, err := createTempECSParamsForTest(t, content)

	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "wordpress",
	}
	containerConfigs := []adapter.ContainerConfig{*containerConfig}
	taskDefinition, err := convertToTaskDefinitionForTest(t, containerConfigs, "", "", ecsParams, nil)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	resourceType := ecs.ResourceTypeGpu
	expectedResourceRequirements := []*ecs.ResourceRequirement{
		{
			Type:  &resourceType,
			Value: &expectedGpuValue,
		},
	}

	if assert.NoError(t, err) {
		assert.ElementsMatch(t,
			expectedResourceRequirements,
			web.ResourceRequirements,
			"Expected resourceRequirements to match")
	}
}

///////////////////////
// helper functions //
//////////////////////

func convertToTaskDefinitionForTest(t *testing.T, containerConfigs []adapter.ContainerConfig, taskRoleArn string, launchType string, ecsParams *ECSParams, ecsRegCreds *regcredio.ECSRegistryCredsOutput) (*ecs.TaskDefinition, error) {
	volumeConfigs := &adapter.Volumes{
		VolumeEmptyHost: []string{namedVolume},
	}

	testParams := ConvertTaskDefParams{
		TaskDefName:            projectName,
		TaskRoleArn:            taskRoleArn,
		RequiredCompatibilites: launchType,
		Volumes:                volumeConfigs,
		ContainerConfigs:       containerConfigs,
		ECSParams:              ecsParams,
		ECSRegistryCreds:       ecsRegCreds,
	}

	taskDefinition, err := ConvertToTaskDefinition(testParams)
	if err != nil {
		return nil, err
	}

	return taskDefinition, nil
}

func findContainerByName(name string, containerDefs []*ecs.ContainerDefinition) *ecs.ContainerDefinition {
	for _, cd := range containerDefs {
		if aws.StringValue(cd.Name) == name {
			return cd
		}
	}
	return nil
}

func createTempECSParamsForTest(t *testing.T, content string) (*ECSParams, error) {
	b := []byte(content)

	f, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs params tempfile")

	defer os.Remove(f.Name())

	_, err = f.Write(b)
	assert.NoError(t, err, "Could not write data to ecs params tempfile")

	err = f.Close()
	assert.NoError(t, err, "Could not close tempfile")

	ecsParamsFileName := f.Name()
	ecsParams, err := ReadECSParams(ecsParamsFileName)
	assert.NoError(t, err, "Could not read ECS Params file")
	return ecsParams, err
}
