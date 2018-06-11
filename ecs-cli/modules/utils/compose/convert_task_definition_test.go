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
)

var defaultNetwork = &yaml.Network{
	Name:     "default",
	RealName: "project_default",
}

// TODO Extract test docker file and use in test (to avoid gaps between parse and conversion unit tests)
var testContainerConfig = &adapter.ContainerConfig{
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
	Privileged: true,
	ReadOnly:   true,
	ShmSize:    int64(128), // Realistically, we expect customers to specify sizes larger than the default of 64M
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
	taskDefinition := convertToTaskDefinitionInTest(t, testContainerConfig, taskRoleArn, "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	// verify task def fields
	assert.Equal(t, taskRoleArn, aws.StringValue(taskDefinition.TaskRoleArn), "Expected taskRoleArn to match")
	assert.Empty(t, taskDefinition.RequiresCompatibilities, "Did not expect RequiresCompatibilities to be set")

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
	assert.Equal(t, aws.Bool(readOnly), containerDef.ReadonlyRootFilesystem, "Expected container def ReadonlyRootFilesystem to match")
	assert.Equal(t, aws.Int64(shmSize), containerDef.LinuxParameters.SharedMemorySize, "Expected sharedMemorySize to match")
	assert.ElementsMatch(t, tmpfs, containerDef.LinuxParameters.Tmpfs, "Expected tmpfs to match")
	assert.Equal(t, aws.String(user), containerDef.User, "Expected container def user to match")
	assert.ElementsMatch(t, ulimits, containerDef.Ulimits, "Expected Ulimits to match")
	assert.Equal(t, volumesFrom, containerDef.VolumesFrom, "Expected VolumesFrom to match")
	assert.Equal(t, aws.String(workingDir), containerDef.WorkingDirectory, "Expected container def WorkingDirectory to match")

	// If no containers are specified as being essential, all containers
	// are marked "essential"
	for _, container := range taskDefinition.ContainerDefinitions {
		assert.True(t, aws.BoolValue(container.Essential), "Expected essential to be true")
	}
}

// ConvertToContainerDefinition tests

func TestConvertToTaskDefinitionWithECSParams_DefaultMemoryLessThanMemoryRes(t *testing.T) {
	// set up containerConfig w/o value for Memory
	containerConfig := &adapter.ContainerConfig{
		Name:  "web",
		Image: "httpd",
		CPU:   int64(5),
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
	_, err = convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

	assert.Error(t, err, "Expected error because memory reservation is greater than memory limit")
}
func TestConvertToTaskDefinitionWithNoSharedMemorySize(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		ShmSize: int64(0),
	}

	taskDefinition := convertToTaskDefinitionInTest(t, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.SharedMemorySize, "Expected sharedMemorySize to be null")
}

func TestConvertToTaskDefinitionWithNoTmpfs(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Tmpfs: nil,
	}

	taskDefinition := convertToTaskDefinitionInTest(t, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.Tmpfs, "Expected Tmpfs to be null")
}

func TestConvertToTaskDefinitionWithBlankHostname(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Hostname: "",
	}

	taskDefinition := convertToTaskDefinitionInTest(t, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.Hostname, "Expected Hostname to be nil")
}

// TODO add test for nil cap add/cap drop

// Test Launch Types
func TestConvertToTaskDefinitionLaunchTypeEmpty(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, containerConfig, "", "")
	if len(taskDefinition.RequiresCompatibilities) > 0 {
		t.Error("Did not expect RequiresCompatibilities to be set")
	}
}

func TestConvertToTaskDefinitionLaunchTypeEC2(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, containerConfig, "", "EC2")
	if len(taskDefinition.RequiresCompatibilities) != 1 {
		t.Error("Expected exactly one required compatibility to be set.")
	}
	assert.Equal(t, "EC2", aws.StringValue(taskDefinition.RequiresCompatibilities[0]))
}

func TestConvertToTaskDefinitionLaunchTypeFargate(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, containerConfig, "", "FARGATE")
	if len(taskDefinition.RequiresCompatibilities) != 1 {
		t.Error("Expected exactly one required compatibility to be set.")
	}
	assert.Equal(t, "FARGATE", aws.StringValue(taskDefinition.RequiresCompatibilities[0]))
}

// Test Conversion with ECS Params
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

	if assert.NoError(t, err) {
		assert.Equal(t, "host", aws.StringValue(taskDefinition.NetworkMode), "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/my_role", aws.StringValue(taskDefinition.TaskRoleArn), "Expected task role ARN to match")

		// If no containers are specified as being essential, all
		// containers are marked "essential"
		for _, container := range taskDefinition.ContainerDefinitions {
			assert.True(t, aws.BoolValue(container.Essential), "Expected essential to be true")
		}
	}
}

func TestConvertToTaskDefinition_WithECSParamsAllFields(t *testing.T) {
	containerConfigs := testContainerConfigs([]string{"mysql", "wordpress"})
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: awsvpc
  task_role_arn: arn:aws:iam::123456789012:role/tweedledee
  services:
    mysql:
      essential: false
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	_, err = convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	_, err = convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, taskRoleArn, ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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
	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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
	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

	containerDefs := taskDefinition.ContainerDefinitions
	web := findContainerByName("web", containerDefs)

	if assert.NoError(t, err) {
		assert.Equal(t, webCPU, aws.Int64Value(web.Cpu), "Expected CPU to match")
		assert.Equal(t, int64(defaultMemLimit), aws.Int64Value(web.Memory), "Expected Memory to match default")
		assert.Empty(t, aws.Int64Value(web.MemoryReservation), "Expected MemoryReservation to be empty")
	}
}

func TestConvertToTaskDefinitionWithECSParams_SomeContainerResourcesProvided(t *testing.T) {
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
	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

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
	_, err = convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

	assert.Error(t, err, "Expected error if mem_reservation was more than mem_limit")
}

func TestConvertToTaskDefinition_WithTaskSize(t *testing.T) {
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, containerConfigs, "", ecsParams)

	if assert.NoError(t, err) {
		assert.Equal(t, "200", aws.StringValue(taskDefinition.Cpu), "Expected CPU to match")
		assert.Equal(t, "10MB", aws.StringValue(taskDefinition.Memory), "Expected CPU to match")
	}
}

func TestMemReservationHigherThanMemLimit(t *testing.T) {
	cpu := int64(131072) // 128 * 1024
	command := "cmd"
	hostname := "local360"
	image := "testimage"
	memory := int64(65536) // 64mb
	privileged := true
	readOnly := true
	user := "user"
	workingDir := "/var"

	containerConfig := adapter.ContainerConfig{
		CPU:               cpu,
		Command:           []string{command},
		Hostname:          hostname,
		Image:             image,
		Memory:            int64(524288) * memory,
		MemoryReservation: int64(1048576) * memory,
		Privileged:        privileged,
		ReadOnly:          readOnly,
		User:              user,
		WorkingDirectory:  workingDir,
	}

	volumeConfigs := adapter.NewVolumes()
	containerConfigs := []adapter.ContainerConfig{containerConfig}

	_, err := ConvertToTaskDefinition(projectName, volumeConfigs, containerConfigs, "", "", nil)
	assert.EqualError(t, err, "mem_limit must be greater than mem_reservation")
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
			SourceVolume:  aws.String("volume-3"),
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

	taskDefinition, err := ConvertToTaskDefinition(projectName, volumeConfigs, containerConfigs, "", "", nil)
	assert.NoError(t, err, "Unexpected error converting Task Definition")

	actualVolumes := taskDefinition.Volumes
	assert.ElementsMatch(t, expectedVolumes, actualVolumes, "Expected volumes to match")
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

///////////////////////
// helper functions //
//////////////////////
func convertToTaskDefinitionInTest(t *testing.T, containerConfig *adapter.ContainerConfig, taskRoleArn string, launchType string) *ecs.TaskDefinition {
	volumeConfigs := &adapter.Volumes{
		VolumeEmptyHost: []string{namedVolume},
	}

	containerConfigs := []adapter.ContainerConfig{}
	containerConfigs = append(containerConfigs, *containerConfig)

	taskDefinition, err := ConvertToTaskDefinition(projectName, volumeConfigs, containerConfigs, taskRoleArn, launchType, nil)
	if err != nil {
		t.Errorf("Expected to convert [%v] containerConfigs without errors. But got [%v]", containerConfig, err)
	}
	return taskDefinition
}

func convertToTaskDefWithEcsParamsInTest(t *testing.T, containerConfigs []adapter.ContainerConfig, taskRoleArn string, ecsParams *ECSParams) (*ecs.TaskDefinition, error) {
	volumeConfigs := &adapter.Volumes{
		VolumeEmptyHost: []string{namedVolume},
	}

	taskDefinition, err := ConvertToTaskDefinition(projectName, volumeConfigs, containerConfigs, taskRoleArn, "", ecsParams)
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
