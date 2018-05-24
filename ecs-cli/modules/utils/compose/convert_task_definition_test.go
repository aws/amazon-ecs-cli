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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
)

const (
	// TODO move when volumes tests added?
	containerPath  = "/tmp/cache"
	containerPath2 = "/tmp/cache2"
	hostPath       = "./cache"

	namedVolume = "named_volume"
)

var defaultNetwork = &yaml.Network{
	Name:     "default",
	RealName: "project_default",
}

// TODO Extract test docker file and use in test (to avoid gaps between parse and conversion unit tests)
var testContainerConfig = &adapter.ContainerConfig{
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
	containerConfig := &adapter.ContainerConfig{
		ShmSize: int64(0),
	}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.SharedMemorySize, "Expected sharedMemorySize to be null")
}

func TestConvertToTaskDefinitionWithNoTmpfs(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Tmpfs: nil,
	}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.LinuxParameters.Tmpfs, "Expected Tmpfs to be null")
}

func TestConvertToTaskDefinitionWithBlankHostname(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{
		Hostname: "",
	}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
	containerDef := *taskDefinition.ContainerDefinitions[0]

	assert.Nil(t, containerDef.Hostname, "Expected Hostname to be nil")
}

// TODO add test for nil cap add/cap drop

// Test Launch Types
func TestConvertToTaskDefinitionLaunchTypeEmpty(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "")
	if len(taskDefinition.RequiresCompatibilities) > 0 {
		t.Error("Did not expect RequiresCompatibilities to be set")
	}
}

func TestConvertToTaskDefinitionLaunchTypeEC2(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "EC2")
	if len(taskDefinition.RequiresCompatibilities) != 1 {
		t.Error("Expected exactly one required compatibility to be set.")
	}
	assert.Equal(t, "EC2", aws.StringValue(taskDefinition.RequiresCompatibilities[0]))
}

func TestConvertToTaskDefinitionLaunchTypeFargate(t *testing.T) {
	containerConfig := &adapter.ContainerConfig{}

	taskDefinition := convertToTaskDefinitionInTest(t, nil, containerConfig, "", "FARGATE")
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)
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

	taskDefinition, err := convertToTaskDefWithEcsParamsInTest(t, nil, containerConfigs, "", ecsParams)

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

	volumeConfigs := make(map[string]*config.VolumeConfig)

	containerConfigs := []adapter.ContainerConfig{containerConfig}

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

func convertToTaskDefinitionInTest(t *testing.T, volumeConfig *config.VolumeConfig, containerConfig *adapter.ContainerConfig, taskRoleArn string, launchType string) *ecs.TaskDefinition {
	volumeConfigs := make(map[string]*config.VolumeConfig)
	volumeConfigs[namedVolume] = volumeConfig

	containerConfigs := []adapter.ContainerConfig{}
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

func convertToTaskDefWithEcsParamsInTest(t *testing.T, volumeConfig *config.VolumeConfig, containerConfigs []adapter.ContainerConfig, taskRoleArn string, ecsParams *ECSParams) (*ecs.TaskDefinition, error) {
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
	containerConfig := &adapter.ContainerConfig{}

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

		zeroValue := isZero(f)
		_, hasValue := hasValues[fieldName]
		assert.NotEqual(t, zeroValue, hasValue)
	}
}
