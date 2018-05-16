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
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"testing"

	containers "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
)

const (
	portNumber     = 8000
	portMapping    = "8000:8000"
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
var testContainerConfig = &containers.ContainerConfig{
	Name:             "mysql",
	Command:          []string{"cmd"},
	CPU:              int64(131072),
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
	taskDefinition := convertToTaskDefinitionInTest(t, nil, testContainerConfig, taskRoleArn, "")
	fmt.Printf("TASK DEF: %+v\n\n", taskDefinition)
	containerDef := *taskDefinition.ContainerDefinitions[0]

	// verify task def fields
	assert.Equal(t, taskRoleArn, aws.StringValue(taskDefinition.TaskRoleArn), "Expected taskRoleArn to match")
	assert.Empty(t, taskDefinition.RequiresCompatibilities, "Did not expect RequiresCompatibilities to be set")

	// verify container def fields
	assert.Equal(t, aws.String(name), containerDef.Name, "Expected container def name to match")
	assert.Equal(t, aws.StringSlice(command), containerDef.Command, "Expected container def command to match")
	assert.Equal(t, aws.Int64(cpu), containerDef.Cpu, "Expected container def cpu to match")
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
func TestConvertToTaskDefinitionWithNoSharedMemorySize(t *testing.T) {
	containerConfig := &containers.ContainerConfig{
		ShmSize: int64(0),
	}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.SharedMemorySize, "Expected sharedMemorySize to be null")
}

func TestConvertToTaskDefinitionWithNoTmpfs(t *testing.T) {
	containerConfig := &containers.ContainerConfig{
		Tmpfs: nil,
	}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.Tmpfs, "Expected Tmpfs to be null")
}

// TODO add test for nil cap add/cap drop

// Test Launch Types
func TestConvertToTaskDefinitionLaunchTypeEmpty(t *testing.T) {
	containerConfig := &containers.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
	if len(taskDefinition.RequiresCompatibilities) > 0 {
		t.Error("Did not expect RequiresCompatibilities to be set")
	}
}

func TestConvertToTaskDefinitionLaunchTypeEC2(t *testing.T) {
	containerConfig := &containers.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "EC2")
	if len(taskDefinition.RequiresCompatibilities) != 1 {
		t.Error("Expected exactly one required compatibility to be set.")
	}
	assert.Equal(t, "EC2", aws.StringValue(taskDefinition.RequiresCompatibilities[0]))
}

func TestConvertToTaskDefinitionLaunchTypeFargate(t *testing.T) {
	containerConfig := &containers.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "FARGATE")
	if len(taskDefinition.RequiresCompatibilities) != 1 {
		t.Error("Expected exactly one required compatibility to be set.")
	}
	assert.Equal(t, "FARGATE", aws.StringValue(taskDefinition.RequiresCompatibilities[0]))
}

// Test Conversion with ECS Params
func testContainerConfigs(names []string) []containers.ContainerConfig {
	containerConfigs := []containers.ContainerConfig{}
	for _, name := range names {
		config := containers.ContainerConfig{Name: name}
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	_, err = convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	_, err = convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

	assert.Error(t, err, "Expected error if no containers are marked essential")
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, taskRoleArn, ecsParams)

	if assert.NoError(t, err) {
		assert.Equal(t, "host", aws.StringValue(taskDefinition.NetworkMode), "Expected network mode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/tweedledum", aws.StringValue(taskDefinition.TaskRoleArn), "Expected task role arn to match")
	}
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

	if assert.NoError(t, err) {
		assert.Equal(t, "200", aws.StringValue(taskDefinition.Cpu), "Expected CPU to match")
		assert.Equal(t, "10MB", aws.StringValue(taskDefinition.Memory), "Expected CPU to match")
	}
}

func TestConvertVolumesFrom_V1_NoOptions(t *testing.T) {
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(false),
			SourceContainer: aws.String("container_name"),
		},
	}
	v1VolumesInput := []string{"container_name"}
	actual, err := ConvertToVolumesFrom(v1VolumesInput)
	assert.NoError(t, err, "Unexpected error converting Volumes From")
	assert.Equal(t, volumesFrom, actual, "Expected VolumesFrom to match")
}

func TestConvertVolumesFrom_V1_Ro(t *testing.T) {
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(true),
			SourceContainer: aws.String("container_name"),
		},
	}
	v1VolumesInput := []string{"container_name:ro"}
	actual, err := ConvertToVolumesFrom(v1VolumesInput)
	assert.NoError(t, err, "Unexpected error converting Volumes From")
	assert.Equal(t, volumesFrom, actual, "Expected VolumesFrom to match")
}

func TestConvertVolumesFrom_V1_Rw(t *testing.T) {
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(false),
			SourceContainer: aws.String("container_name"),
		},
	}
	v1VolumesInput := []string{"container_name:rw"}
	actual, err := ConvertToVolumesFrom(v1VolumesInput)
	assert.NoError(t, err, "Unexpected error converting Volumes From")
	assert.Equal(t, volumesFrom, actual, "Expected VolumesFrom to match")
}

func TestConvertVolumesFrom_V2_NoOptions(t *testing.T) {
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(false),
			SourceContainer: aws.String("container_name"),
		},
	}
	v2VolumesInput := []string{"container:container_name"}
	actual, err := ConvertToVolumesFrom(v2VolumesInput)
	assert.NoError(t, err, "Unexpected error converting Volumes From")
	assert.Equal(t, volumesFrom, actual, "Expected VolumesFrom to match")
}

func TestConvertVolumesFrom_V2_Ro(t *testing.T) {
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(true),
			SourceContainer: aws.String("container_name"),
		},
	}
	v2VolumesInput := []string{"container:container_name:ro"}
	actual, err := ConvertToVolumesFrom(v2VolumesInput)
	assert.NoError(t, err, "Unexpected error converting Volumes From")
	assert.Equal(t, volumesFrom, actual, "Expected VolumesFrom to match")
}

func TestConvertVolumesFrom_V2_Rw(t *testing.T) {
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(false),
			SourceContainer: aws.String("container_name"),
		},
	}
	v2VolumesInput := []string{"container:container_name:rw"}
	actual, err := ConvertToVolumesFrom(v2VolumesInput)
	assert.NoError(t, err, "Unexpected error converting Volumes From")
	assert.Equal(t, volumesFrom, actual, "Expected VolumesFrom to match")
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

	containerConfig := containers.ContainerConfig{
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

	volumeConfigs := make(map[string]*config.VolumeConfig)

	containerConfigs := []containers.ContainerConfig{containerConfig}

	envLookup, err := GetDefaultEnvironmentLookup()
	assert.NoError(t, err, "Unexpected error setting up environment lookup")
	resourceLookup, err := GetDefaultResourceLookup()
	assert.NoError(t, err, "Unexpected error setting up resource lookup")
	context := &project.Context{
		ProjectName:       "ProjectName",
		Project:           &project.Project{},
		EnvironmentLookup: envLookup,
		ResourceLookup:    resourceLookup,
	}
	_, err = ConvertToTaskDefinition(context, volumeConfigs, containerConfigs, "", "", nil)
	assert.EqualError(t, err, "mem_limit must be greater than mem_reservation")
}

// TODO Modify test when top-level Volumes added
// func TestConvertToTaskDefinitionWithVolumes(t *testing.T) {
// 	volume := yaml.Volume{Source: hostPath, Destination: containerPath}
// 	volumesFrom := []string{"container1"}

// 	containerConfig := &config.ServiceConfig{
// 		Volumes:     &yaml.Volumes{Volumes: []*yaml.Volume{&volume}},
// 		VolumesFrom: volumesFrom,
// 	}

// 	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
// 	containerDef := *taskDefinition.ContainerDefinitions[0]

// 	if len(volumesFrom) != len(containerDef.VolumesFrom) ||
// 		volumesFrom[0] != aws.StringValue(containerDef.VolumesFrom[0].SourceContainer) {
// 		t.Errorf("Expected volumesFrom [%v] But was [%v]", volumesFrom, containerDef.VolumesFrom)
// 	}
// 	volumeDef := *taskDefinition.Volumes[0]
// 	mountPoint := *containerDef.MountPoints[0]

// 	if hostPath != aws.StringValue(volumeDef.Host.SourcePath) {
// 		t.Errorf("Expected HostSourcePath [%s] But was [%s]", hostPath, aws.StringValue(volumeDef.Host.SourcePath))
// 	}
// 	if containerPath != aws.StringValue(mountPoint.ContainerPath) {
// 		t.Errorf("Expected containerPath [%s] But was [%s]", containerPath, aws.StringValue(mountPoint.ContainerPath))
// 	}
// 	if aws.StringValue(volumeDef.Name) != aws.StringValue(mountPoint.SourceVolume) {
// 		t.Errorf("Expected volume name to match. "+
// 			"Got Volume.Name=[%s] And MountPoint.SourceVolume=[%s]",
// 			aws.StringValue(volumeDef.Name), aws.StringValue(mountPoint.SourceVolume))
// 	}
// }

// TODO Modify test when top-level Volumes added
// func TestConvertToTaskDefinitionWithNamedVolume(t *testing.T) {
// 	volume := yaml.Volume{Source: namedVolume, Destination: containerPath}

// 	containerConfig := &config.ServiceConfig{
// 		Volumes:  &yaml.Volumes{Volumes: []*yaml.Volume{&volume}},
// 		Networks: &yaml.Networks{Networks: []*yaml.Network{defaultNetwork}},
// 	}

// 	taskDefinition := convertToTaskDefinitionInTest(t, "name", nil, containerConfig, "", "")
// 	containerDef := *taskDefinition.ContainerDefinitions[0]

// 	volumeDef := *taskDefinition.Volumes[0]
// 	mountPoint := *containerDef.MountPoints[0]
// 	if volumeDef.Host != nil {
// 		t.Errorf("Expected volume host to be nil But was [%s]", volumeDef.Host)
// 	}
// 	if containerPath != aws.StringValue(mountPoint.ContainerPath) {
// 		t.Errorf("Expected containerPath [%s] But was [%s]", containerPath, aws.StringValue(mountPoint.ContainerPath))
// 	}
// 	if aws.StringValue(volumeDef.Name) != aws.StringValue(mountPoint.SourceVolume) {
// 		t.Errorf("Expected volume name to match. "+
// 			"Got Volume.Name=[%s] And MountPoint.SourceVolume=[%s]",
// 			aws.StringValue(volumeDef.Name), aws.StringValue(mountPoint.SourceVolume))
// 	}
// }

////////////////////////////////
// Convert individual fields //
//////////////////////////////
func TestConvertToTmpfs(t *testing.T) {
	tmpfs := []string{"/run:rw,noexec,nosuid,size=65536k", "/foo:size=1gb", "/bar:size=1gb,rw,runbindable"}

	tmpfsMounts, err := ConvertToTmpfs(tmpfs)
	assert.NoError(t, err, "Unexpected error converting tmpfs")
	mount1 := tmpfsMounts[0]
	mount2 := tmpfsMounts[1]
	mount3 := tmpfsMounts[2]

	assert.Equal(t, "/run", aws.StringValue(mount1.ContainerPath))
	assert.Equal(t, []string{"rw", "noexec", "nosuid"}, aws.StringValueSlice(mount1.MountOptions))
	assert.Equal(t, int64(64), aws.Int64Value(mount1.Size))

	assert.Equal(t, "/foo", aws.StringValue(mount2.ContainerPath))
	assert.Equal(t, []string{}, aws.StringValueSlice(mount2.MountOptions))
	assert.Equal(t, int64(1024), aws.Int64Value(mount2.Size))

	assert.Equal(t, "/bar", aws.StringValue(mount3.ContainerPath))
	assert.Equal(t, []string{"rw", "runbindable"}, aws.StringValueSlice(mount3.MountOptions))
	assert.Equal(t, int64(1024), aws.Int64Value(mount3.Size))
}

func TestConvertToTmpfs_NoPath(t *testing.T) {
	tmpfs := []string{"size=65536k"}
	_, err := ConvertToTmpfs(tmpfs)

	assert.Error(t, err)
}

func TestConvertToTmpfs_BadOptionFormat(t *testing.T) {
	tmpfs := []string{"/run,size=65536k"}
	_, err := ConvertToTmpfs(tmpfs)

	assert.Error(t, err)
}

func TestConvertToTmpfs_NoSize(t *testing.T) {
	tmpfs := []string{"/run"}
	_, err := ConvertToTmpfs(tmpfs)

	assert.Error(t, err)
}

func TestConvertToTmpfs_WithOptionsNoSize(t *testing.T) {
	tmpfs := []string{"/run:rw"}
	_, err := ConvertToTmpfs(tmpfs)

	assert.Error(t, err)
}

func TestConvertToTmpfs_WithMalformedSize(t *testing.T) {
	tmpfs := []string{"/run:1gb"}
	_, err := ConvertToTmpfs(tmpfs)

	assert.Error(t, err)
}

func TestConvertToPortMappings(t *testing.T) {
	implicitTcp := portMapping                      // 8000:8000
	explicitTcp := portMapping + "/tcp"             // "8000:8000/tcp"
	udpPort := portMapping + "/udp"                 // "8000:8000/udp"
	containerPortOnly := strconv.Itoa(portNumber)   // "8000"
	portWithIpAddress := "127.0.0.1:" + portMapping // "127.0.0.1:8000:8000"

	portMappingsIn := []string{implicitTcp, explicitTcp, udpPort, containerPortOnly, portWithIpAddress}
	portMappingsOut, err := ConvertToPortMappings("test", portMappingsIn)

	assert.NoError(t, err, "Unexpected error converting port mapping")

	if len(portMappingsIn) != len(portMappingsOut) {
		t.Errorf("Incorrect conversion. Input [%v] Output [%v]", portMappingsIn, portMappingsOut)
	}
	verifyPortMapping(t, portMappingsOut[0], portNumber, portNumber, ecs.TransportProtocolTcp)
	verifyPortMapping(t, portMappingsOut[1], portNumber, portNumber, ecs.TransportProtocolTcp)
	verifyPortMapping(t, portMappingsOut[2], portNumber, portNumber, ecs.TransportProtocolUdp)
	verifyPortMapping(t, portMappingsOut[3], 0, portNumber, ecs.TransportProtocolTcp)
	verifyPortMapping(t, portMappingsOut[4], portNumber, portNumber, ecs.TransportProtocolTcp)
}

func verifyPortMapping(t *testing.T, output *ecs.PortMapping, hostPort, containerPort int64, protocol string) {
	if protocol != *output.Protocol {
		t.Errorf("Expected protocol [%s] But was [%s]", protocol, *output.Protocol)
	}
	if hostPort != *output.HostPort {
		t.Errorf("Expected hostPort [%d] But was [%d]", hostPort, *output.HostPort)
	}
	if containerPort != *output.ContainerPort {
		t.Errorf("Expected containerPort [%d] But was [%d]", containerPort, *output.ContainerPort)
	}
}

func TestConvertToMountPoints(t *testing.T) {
	onlyContainerPath := yaml.Volume{Destination: containerPath}
	onlyContainerPath2 := yaml.Volume{Destination: containerPath2}
	hostAndContainerPath := yaml.Volume{Source: hostPath, Destination: containerPath} // "./cache:/tmp/cache"
	onlyContainerPathWithRO := yaml.Volume{Destination: containerPath, AccessMode: "ro"}
	hostAndContainerPathWithRO := yaml.Volume{Source: hostPath, Destination: containerPath, AccessMode: "ro"} // "./cache:/tmp/cache:ro"
	hostAndContainerPathWithRW := yaml.Volume{Source: hostPath, Destination: containerPath, AccessMode: "rw"}
	namedVolumeAndContainerPath := yaml.Volume{Source: namedVolume, Destination: containerPath}

	volumes := &Volumes{
		volumeWithHost:  make(map[string]string), // map with key:=hostSourcePath value:=VolumeName
		volumeEmptyHost: []string{namedVolume},   // Declare one volume with an empty host
	}

	// Valid inputs with host and container paths set
	mountPointsIn := yaml.Volumes{
		Volumes: []*yaml.Volume{
			&onlyContainerPath,
			&onlyContainerPath2,
			&hostAndContainerPath,
			&onlyContainerPathWithRO,
			&hostAndContainerPathWithRO,
			&hostAndContainerPathWithRW,
			&namedVolumeAndContainerPath,
		},
	}

	mountPointsOut, err := ConvertToMountPoints(&mountPointsIn, volumes)
	if err != nil {
		t.Fatalf("Expected to convert [%v] mountPoints without errors. But got [%v]", mountPointsIn, err)
	}
	if len(mountPointsIn.Volumes) != len(mountPointsOut) {
		t.Errorf("Incorrect conversion. Input [%v] Output [%v]", mountPointsIn, mountPointsOut)
	}

	verifyMountPoint(t, mountPointsOut[0], volumes, "", containerPath, false, 1)  // 1 is the counter for the first volume with an empty host path
	verifyMountPoint(t, mountPointsOut[1], volumes, "", containerPath2, false, 2) // 2 is the counter for the second volume with an empty host path
	verifyMountPoint(t, mountPointsOut[2], volumes, hostPath, containerPath, false, 2)
	verifyMountPoint(t, mountPointsOut[3], volumes, "", containerPath, true, 3) // 3 is the counter for the third volume with an empty host path
	verifyMountPoint(t, mountPointsOut[4], volumes, hostPath, containerPath, true, 3)
	verifyMountPoint(t, mountPointsOut[5], volumes, hostPath, containerPath, false, 3)
	verifyMountPoint(t, mountPointsOut[6], volumes, namedVolume, containerPath, false, 3)

	if mountPointsOut[0].SourceVolume == mountPointsOut[1].SourceVolume {
		t.Errorf("Expected volume %v (onlyContainerPath) and %v (onlyContainerPath2) to be different", mountPointsOut[0].SourceVolume, mountPointsOut[1].SourceVolume)
	}

	if mountPointsOut[1].SourceVolume == mountPointsOut[3].SourceVolume {
		t.Errorf("Expected volume %v (onlyContainerPath2) and %v (onlyContainerPathWithRO) to be different", mountPointsOut[0].SourceVolume, mountPointsOut[1].SourceVolume)
	}
}

func TestConvertToMountPointsWithInvalidAccessMode(t *testing.T) {
	volumes := &Volumes{
		volumeWithHost:  make(map[string]string),
		volumeEmptyHost: []string{namedVolume},
	}

	hostAndContainerPathWithIncorrectAccess := yaml.Volume{
		Source:      hostPath,
		Destination: containerPath,
		AccessMode:  "readonly",
	}

	mountPointsIn := yaml.Volumes{
		Volumes: []*yaml.Volume{&hostAndContainerPathWithIncorrectAccess},
	}

	_, err := ConvertToMountPoints(&mountPointsIn, volumes)

	if err == nil {
		t.Errorf("Expected to get error for mountPoint[%s] but didn't.", hostAndContainerPathWithIncorrectAccess)
	}
}

func TestConvertToMountPointsNullContainerVolumes(t *testing.T) {
	volumes := &Volumes{
		volumeWithHost:  make(map[string]string),
		volumeEmptyHost: []string{namedVolume},
	}
	mountPointsOut, err := ConvertToMountPoints(nil, volumes)
	if err != nil {
		t.Fatalf("Expected to convert nil mountPoints without errors. But got [%v]", err)
	}
	if len(mountPointsOut) != 0 {
		t.Errorf("Incorrect conversion. Input nil Output [%v]", mountPointsOut)
	}
}

func verifyMountPoint(t *testing.T, output *ecs.MountPoint, volumes *Volumes,
	source, containerPath string, readonly bool, EmptyHostCtr int) {
	sourceVolume := ""
	if containerPath != *output.ContainerPath {
		t.Errorf("Expected containerPath [%s] But was [%s]", containerPath, *output.ContainerPath)
	}
	if source == "" {
		sourceVolume = volumes.volumeEmptyHost[EmptyHostCtr]
	} else if project.IsNamedVolume(source) {
		sourceVolume = source
	} else {
		sourceVolume = volumes.volumeWithHost[source]
	}
	if sourceVolume != *output.SourceVolume {
		t.Errorf("Expected sourceVolume [%s] But was [%s]", sourceVolume, *output.SourceVolume)
	}
	if readonly != *output.ReadOnly {
		t.Errorf("Expected readonly [%v] But was [%v]", readonly, *output.ReadOnly)
	}
}

func TestConvertToExtraHosts(t *testing.T) {
	hostname := "test.local"
	ipAddress := "127.10.10.10"

	extraHost := hostname + ":" + ipAddress

	extraHostsIn := []string{extraHost}
	extraHostsOut, err := ConvertToExtraHosts(extraHostsIn)
	if err != nil {
		t.Errorf("Expected to convert [%v] extra hosts without errors. But got [%v]", extraHostsIn, err)
	}
	if len(extraHostsIn) != len(extraHostsOut) {
		t.Errorf("Incorrect conversion. Input [%v] Output [%v]", extraHostsIn, extraHostsOut)
	}
	verifyExtraHost(t, extraHostsOut[0], hostname, ipAddress)

	incorrectHost := hostname + "=" + ipAddress
	_, err = ConvertToExtraHosts([]string{incorrectHost})
	if err == nil {
		t.Errorf("Expected to get formatting error for extraHost=[%s], but got none", incorrectHost)
	}

	extraHostWithPort := fmt.Sprintf("%s:%s:%d", hostname, ipAddress, portNumber)
	_, err = ConvertToExtraHosts([]string{extraHostWithPort})
	if err == nil {
		t.Errorf("Expected to get formatting error for extraHost=[%s], but got none", extraHostWithPort)
	}

}

func verifyExtraHost(t *testing.T, output *ecs.HostEntry, hostname, ipAddress string) {
	if hostname != aws.StringValue(output.Hostname) {
		t.Errorf("Expected hostname [%s] But was [%s]", hostname, aws.StringValue(output.Hostname))
	}
	if ipAddress != aws.StringValue(output.IpAddress) {
		t.Errorf("Expected ipAddress [%s] But was [%s]", ipAddress, aws.StringValue(output.IpAddress))
	}
}

func TestConvertToUlimits(t *testing.T) {
	softLimit := int64(1024)
	hardLimit := int64(2048)
	typeName := "nofile"
	basicType := yaml.NewUlimit(typeName, softLimit, softLimit)         // "nofile=1024"
	typeWithHardLimit := yaml.NewUlimit(typeName, softLimit, hardLimit) // "nofile=1024:2048"

	ulimitsIn := yaml.Ulimits{
		Elements: []yaml.Ulimit{basicType, typeWithHardLimit},
	}
	ulimitsOut, err := ConvertToULimits(ulimitsIn)
	if err != nil {
		t.Errorf("Expected to convert [%v] ulimits without errors. But got [%v]", ulimitsIn, err)
	}
	if len(ulimitsIn.Elements) != len(ulimitsOut) {
		t.Errorf("Incorrect conversion. Input [%v] Output [%v]", ulimitsIn, ulimitsOut)
	}
	verifyUlimit(t, ulimitsOut[0], typeName, softLimit, softLimit)
	verifyUlimit(t, ulimitsOut[1], typeName, softLimit, hardLimit)
}

func verifyUlimit(t *testing.T, output *ecs.Ulimit, name string, softLimit, hardLimit int64) {
	if name != *output.Name {
		t.Errorf("Expected name [%s] But was [%s]", name, *output.Name)
	}
	if softLimit != *output.SoftLimit {
		t.Errorf("Expected softLimit [%d] But was [%d]", softLimit, *output.SoftLimit)
	}
	if hardLimit != *output.HardLimit {
		t.Errorf("Expected hardLimit [%d] But was [%d]", hardLimit, *output.HardLimit)
	}
}

func TestConvertToVolumes(t *testing.T) {
	libcomposeVolumeConfigs := map[string]*config.VolumeConfig{
		namedVolume: nil,
	}

	expected := &Volumes{
		volumeWithHost:  make(map[string]string), // map with key:=hostSourcePath value:=VolumeName
		volumeEmptyHost: []string{namedVolume},   // Declare one volume with an empty host
	}

	actual, err := ConvertToVolumes(libcomposeVolumeConfigs)

	assert.NoError(t, err, "Unexpected error converting libcompose volume configs")
	assert.Equal(t, expected, actual, "Named volumes should match")
}

func TestConvertToVolumes_ErrorsWithDriverSubfield(t *testing.T) {
	libcomposeVolumeConfigs := map[string]*config.VolumeConfig{
		namedVolume: &config.VolumeConfig{
			Driver: "noodles",
		},
	}

	_, err := ConvertToVolumes(libcomposeVolumeConfigs)

	assert.Error(t, err, "Expected error converting libcompose volume configs when driver is specified")
}

func TestConvertToVolumes_ErrorsWithDriverOptsSubfield(t *testing.T) {
	driverOpts := map[string]string{"foo": "bar"}

	libcomposeVolumeConfigs := map[string]*config.VolumeConfig{
		namedVolume: &config.VolumeConfig{
			DriverOpts: driverOpts,
		},
	}

	_, err := ConvertToVolumes(libcomposeVolumeConfigs)

	assert.Error(t, err, "Expected error converting libcompose volume configs when driver options are specified")
}

func TestConvertToVolumes_ErrorsWithExternalSubfield(t *testing.T) {
	external := yaml.External{External: false}

	libcomposeVolumeConfigs := map[string]*config.VolumeConfig{
		namedVolume: &config.VolumeConfig{
			External: external,
		},
	}

	_, err := ConvertToVolumes(libcomposeVolumeConfigs)

	assert.Error(t, err, "Expected error converting libcompose volume configs when external is specified")

	external = yaml.External{External: true}
	libcomposeVolumeConfigs = map[string]*config.VolumeConfig{
		namedVolume: &config.VolumeConfig{
			External: external,
		},
	}

	_, err = ConvertToVolumes(libcomposeVolumeConfigs)

	assert.Error(t, err, "Expected error converting libcompose volume configs when external is specified")
}

func convertToTaskDefinitionInTest(t *testing.T, volumeConfig *config.VolumeConfig, containerConfig *containers.ContainerConfig, taskRoleArn string, launchType string) *ecs.TaskDefinition {
	volumeConfigs := make(map[string]*config.VolumeConfig)
	volumeConfigs[namedVolume] = volumeConfig

	containerConfigs := []containers.ContainerConfig{}
	containerConfigs = append(containerConfigs, *containerConfig)

	envLookup, err := GetDefaultEnvironmentLookup()
	if err != nil {
		t.Fatal("Unexpected error setting up environment lookup")
	}
	resourceLookup, err := GetDefaultResourceLookup()
	if err != nil {
		t.Fatal("Unexpected error setting up resource lookup")
	}
	context := &project.Context{
		ProjectName:       "ProjectName",
		Project:           &project.Project{},
		EnvironmentLookup: envLookup,
		ResourceLookup:    resourceLookup,
	}
	taskDefinition, err := ConvertToTaskDefinition(context, volumeConfigs, containerConfigs, taskRoleArn, launchType, nil)
	if err != nil {
		t.Errorf("Expected to convert [%v] containerConfigs without errors. But got [%v]", containerConfig, err)
	}
	return taskDefinition
}

func convertToTaskDefWithEcsParamsInTest(t *testing.T, volumeConfig *config.VolumeConfig, containerConfigs []containers.ContainerConfig, taskRoleArn string, ecsParams *ECSParams) (*ecs.TaskDefinition, error) {
	volumeConfigs := make(map[string]*config.VolumeConfig)
	if volumeConfig != nil {
		volumeConfigs[namedVolume] = volumeConfig
	}

	envLookup, err := GetDefaultEnvironmentLookup()
	assert.NoError(t, err, "Unexpected error setting up environment lookup")

	resourceLookup, err := GetDefaultResourceLookup()
	assert.NoError(t, err, "Unexpected error setting up resource lookup")

	context := &project.Context{
		ProjectName:       "ProjectName",
		Project:           &project.Project{},
		EnvironmentLookup: envLookup,
		ResourceLookup:    resourceLookup,
	}
	taskDefinition, err := ConvertToTaskDefinition(context, volumeConfigs, containerConfigs, taskRoleArn, "", ecsParams)
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

func TestIsZeroForEmptyConfig(t *testing.T) {
	containerConfig := &containers.ContainerConfig{}

	configValue := reflect.ValueOf(containerConfig).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		f := configValue.Field(i)
		ft := configType.Field(i)
		isZero := isZero(f)
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

	containerConfig := &containers.ContainerConfig{
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

		zeroValue := isZero(f)
		_, hasValue := hasValues[fieldName]
		assert.NotEqual(t, zeroValue, hasValue)
	}
}

func TestSortedGoString(t *testing.T) {
	family := aws.String("family1")
	name := aws.String("foo")
	command := aws.StringSlice([]string{"dark", "side", "of", "the", "moon"})
	dockerLabels := map[string]string{
		"label1":         "",
		"com.foo.label2": "value",
	}

	inputA := ecs.RegisterTaskDefinitionInput{
		Family: family,
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name:         name,
				Command:      command,
				DockerLabels: aws.StringMap(dockerLabels),
			},
		},
	}
	inputB := ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Command:      command,
				Name:         name,
				DockerLabels: aws.StringMap(dockerLabels),
			},
		},
		Family: family,
	}

	strA, err := SortedGoString(inputA)
	assert.NoError(t, err, "Unexpected error generating sorted map string")
	strB, err := SortedGoString(inputB)
	assert.NoError(t, err, "Unexpected error generating sorted map string")

	assert.Equal(t, strA, strB, "Sorted inputs should match")
}
