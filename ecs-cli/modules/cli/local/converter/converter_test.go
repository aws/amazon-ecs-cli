// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package converter implements the logic to translate an ecs.TaskDefinition
// structure to a docker compose schema, which will be written to a
// docker-compose.local.yml file.

package converter

import (
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	composeV3 "github.com/docker/cli/cli/compose/types"

	"github.com/stretchr/testify/assert"
)

func TestConvertToComposeService(t *testing.T) {
	// GIVEN
	expectedImage := "nginx"
	expectedName := "web"
	expectedCommand := []string{"CMD-SHELL", "curl -f http://localhost"}
	expectedEntrypoint := []string{"sh", "-c"}
	expectedWorkingDir := "./app"
	expectedHostname := "myHost"
	expectedLinks := []string{"container1"}
	expectedDNS := []string{"1.2.3.4"}
	expectedDNSSearch := []string{"search.example.com"}
	expectedUser := "admin"
	expectedSecurityOpt := []string{"label:type:test_virt"}
	expectedTty := true
	expectedPrivileged := true
	expectedReadOnly := true
	expectedStdinOpen := true
	expectedUlimits := map[string]*composeV3.UlimitsConfig{
		"nofile": &composeV3.UlimitsConfig{
			Soft: 2000,
			Hard: 4000,
		},
	}
	expectedInit := true
	expectedDevices := []string{"/dev/sda:/dev/xvdc:r"}
	expectedTmpfs := []string{"/run:size=64MiB,rw,noexec,nosuid"}
	expectedShmSize := "128MiB"
	expectedCapAdd := []string{"NET_ADMIN", "MKNOD"}
	expectedCapDrop := []string{"KILL"}
	expectedEnvironment := map[string]*string{
		"rails_env":   aws.String("development"),
		"DB_PASSWORD": aws.String("${web_DB_PASSWORD}"),
		"API_KEY":     aws.String("${web_API_KEY}"),
	}
	expectedExtraHosts := []string{"somehost:162.242.195.82", "otherhost:50.31.209.229"}
	expectedHealthCheck := &composeV3.HealthCheckConfig{
		Test: []string{"CMD-SHELL", "echo hello"},
	}
	expectedLabels := composeV3.Labels{
		"foo":                          "bar",
		"ecs-local.secret.DB_PASSWORD": "arn:aws:secretsmanager:us-west-2:01234567:secret:mySecretSecret",
		"ecs-local.secret.API_KEY":     "arn:aws:ssm:us-west-2:01234567:parameter/mySecretParameter",
	}
	expectedLogging := &composeV3.LoggingConfig{
		Driver: "awslogs",
		Options: map[string]string{
			"awslogs-group":         "/ecs/fargate-task-definition",
			"awslogs-region":        "us-east-1",
			"awslogs-stream-prefix": "ecs",
		},
	}
	expectedVolumes := []composeV3.ServiceVolumeConfig{
		{
			Target:   "/tmp/cache",
			Source:   "volume-1",
			ReadOnly: true,
		},
	}
	expectedNetworkMode := ecs.NetworkModeBridge
	expectedIpc := ecs.IpcModeHost
	expectedPid := ecs.PidModeTask
	expectedNetworks := map[string]*composeV3.ServiceNetworkConfig{
		network.EcsLocalNetworkName: nil,
	}
	expectedPorts := []composeV3.ServicePortConfig{
		{
			Target:    uint32(3000),
			Published: uint32(80),
			Protocol:  "tcp",
		},
	}
	expectedSysctls := []string{
		"net.core.somaxconn=1024",
		"net.ipv4.tcp_syncookies=0",
	}

	taskDefinition := &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Image:                  aws.String(expectedImage),
				Name:                   aws.String(expectedName),
				Command:                aws.StringSlice(expectedCommand),
				EntryPoint:             aws.StringSlice(expectedEntrypoint),
				WorkingDirectory:       aws.String(expectedWorkingDir),
				Hostname:               aws.String(expectedHostname),
				Links:                  aws.StringSlice(expectedLinks),
				DnsServers:             aws.StringSlice(expectedDNS),
				DnsSearchDomains:       aws.StringSlice(expectedDNSSearch),
				User:                   aws.String(expectedUser),
				DockerSecurityOptions:  aws.StringSlice(expectedSecurityOpt),
				PseudoTerminal:         aws.Bool(expectedTty),
				Privileged:             aws.Bool(expectedPrivileged),
				Interactive:            aws.Bool(expectedStdinOpen),
				ReadonlyRootFilesystem: aws.Bool(expectedReadOnly),
				Ulimits: []*ecs.Ulimit{
					{
						Name:      aws.String("nofile"),
						SoftLimit: aws.Int64(2000),
						HardLimit: aws.Int64(4000),
					},
				},
				Environment: []*ecs.KeyValuePair{
					{
						Name:  aws.String("rails_env"),
						Value: aws.String("development"),
					},
				},
				Secrets: []*ecs.Secret{
					{
						Name:      aws.String("DB_PASSWORD"),
						ValueFrom: aws.String("arn:aws:secretsmanager:us-west-2:01234567:secret:mySecretSecret"),
					},
					{
						Name:      aws.String("API_KEY"),
						ValueFrom: aws.String("arn:aws:ssm:us-west-2:01234567:parameter/mySecretParameter"),
					},
				},
				ExtraHosts: []*ecs.HostEntry{
					{
						Hostname:  aws.String("somehost"),
						IpAddress: aws.String("162.242.195.82"),
					},
					{
						Hostname:  aws.String("otherhost"),
						IpAddress: aws.String("50.31.209.229"),
					},
				},
				HealthCheck: &ecs.HealthCheck{
					Command: aws.StringSlice([]string{"CMD-SHELL", "echo hello"}),
				},
				DockerLabels: map[string]*string{"foo": aws.String("bar")},
				LogConfiguration: &ecs.LogConfiguration{
					LogDriver: aws.String("awslogs"),
					Options: map[string]*string{
						"awslogs-group":         aws.String("/ecs/fargate-task-definition"),
						"awslogs-region":        aws.String("us-east-1"),
						"awslogs-stream-prefix": aws.String("ecs"),
					},
				},
				MountPoints: []*ecs.MountPoint{
					{
						ContainerPath: aws.String("/tmp/cache"),
						ReadOnly:      aws.Bool(true),
						SourceVolume:  aws.String("volume-1"),
					},
				},
				PortMappings: []*ecs.PortMapping{
					{
						ContainerPort: aws.Int64(3000),
						HostPort:      aws.Int64(80),
						Protocol:      aws.String("tcp"),
					},
				},
				SystemControls: []*ecs.SystemControl{
					{
						Namespace: aws.String("net.core.somaxconn"),
						Value:     aws.String("1024"),
					},
					{
						Namespace: aws.String("net.ipv4.tcp_syncookies"),
						Value:     aws.String("0"),
					},
				},
				LinuxParameters: &ecs.LinuxParameters{
					InitProcessEnabled: aws.Bool(true),
					SharedMemorySize:   aws.Int64(128),
					Capabilities: &ecs.KernelCapabilities{
						Add:  aws.StringSlice(expectedCapAdd),
						Drop: aws.StringSlice(expectedCapDrop),
					},
					Devices: []*ecs.Device{
						{
							HostPath:      aws.String("/dev/sda"),
							ContainerPath: aws.String("/dev/xvdc"),
							Permissions:   aws.StringSlice([]string{"read"}),
						},
					},
					Tmpfs: []*ecs.Tmpfs{
						{
							ContainerPath: aws.String("/run"),
							MountOptions:  aws.StringSlice([]string{"rw", "noexec", "nosuid"}),
							Size:          aws.Int64(64),
						},
					},
				},
			},
		},
	}

	containerDef := taskDefinition.ContainerDefinitions[0]

	commonValues := &CommonContainerValues{
		NetworkMode: expectedNetworkMode,
		Ipc:         expectedIpc,
		Pid:         expectedPid,
	}

	// WHEN
	service, err := convertToComposeService(containerDef, commonValues)

	// THEN
	assert.NoError(t, err, "Unexpected error converting Container Definition")
	assert.Equal(t, expectedName, service.Name, "Expected Name to match")
	assert.Equal(t, expectedImage, service.Image, "Expected Image to match")
	assert.Equal(t, composeV3.ShellCommand(expectedCommand), service.Command, "Expected Command to match")
	assert.Equal(t, composeV3.ShellCommand(expectedEntrypoint), service.Entrypoint, "Expected Entry point to match")
	assert.Equal(t, expectedWorkingDir, service.WorkingDir, "Expected WorkingDir to match")
	assert.Equal(t, expectedHostname, service.Hostname, "Expected Hostname to match")
	assert.Equal(t, expectedLinks, service.Links, "Expected Links to match")
	assert.Equal(t, composeV3.StringList(expectedDNS), service.DNS, "Expected DNS to match")
	assert.Equal(t, composeV3.StringList(expectedDNSSearch), service.DNSSearch, "Expected DNSSearch to match")
	assert.Equal(t, expectedUser, service.User, "Expected User to match")
	assert.Equal(t, expectedSecurityOpt, service.SecurityOpt, "Expected SecurityOpt to match")
	assert.Equal(t, expectedTty, service.Tty, "Expected Tty to match")
	assert.Equal(t, expectedPrivileged, service.Privileged, "Expected Privileged to match")
	assert.Equal(t, expectedStdinOpen, service.StdinOpen, "Expected StdinOpen to match")
	assert.Equal(t, expectedReadOnly, service.ReadOnly, "Expected ReadOnly to match")
	assert.Equal(t, expectedUlimits, service.Ulimits, "Expected Ulimits to match")
	assert.Equal(t, composeV3.MappingWithEquals(expectedEnvironment), service.Environment, "Expected Environment to match")
	assert.Equal(t, composeV3.HostsList(expectedExtraHosts), service.ExtraHosts, "Expected ExtraHosts to match")
	assert.Equal(t, expectedHealthCheck, service.HealthCheck, "Expected HealthCheck to match")
	assert.Equal(t, expectedLabels, service.Labels, "Expected Labels to match")
	assert.Equal(t, expectedLogging, service.Logging, "Expected Logging to match")
	assert.Equal(t, expectedVolumes, service.Volumes, "Expected Volumes to match")
	assert.Equal(t, expectedNetworks, service.Networks, "Expected Networks to match")
	assert.Equal(t, expectedNetworkMode, service.NetworkMode, "Expected NetworkMode to match")
	assert.Equal(t, expectedPid, service.Pid, "Expected Pid to match")
	assert.Equal(t, expectedIpc, service.Ipc, "Expected Ipc to match")
	assert.Equal(t, expectedPorts, service.Ports, "Expected Ports to match")
	assert.Equal(t, composeV3.StringList(expectedSysctls), service.Sysctls, "Expected Sysctls to match")

	// Fields from LinuxParameters
	assert.Equal(t, composeV3.StringList(expectedTmpfs), service.Tmpfs, "Expected Tmpfs to match")
	assert.Equal(t, aws.Bool(expectedInit), service.Init, "Expected Init to match")
	assert.Equal(t, expectedDevices, service.Devices, "Expected Devices to match")
	assert.Equal(t, expectedShmSize, service.ShmSize, "Expected ShmSize to match")
	assert.Equal(t, expectedCapAdd, service.CapAdd, "Expected CapAdd to match")
	assert.Equal(t, expectedCapDrop, service.CapDrop, "Expected CapDrop to match")
}

func TestCreateComposeService_SetsNetworkMode(t *testing.T) {
	// GIVEN
	expectedNetworkMode := ecs.NetworkModeBridge

	taskDefinition := &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Image: aws.String("myApp"),
			},
		},
		NetworkMode: aws.String(expectedNetworkMode),
	}

	// WHEN
	services, err := createComposeServices(taskDefinition, &LocalCreateMetadata{})
	service := services[0]

	// THEN
	assert.NoError(t, err, "Unexpected error creating Compose services")
	assert.Equal(t, expectedNetworkMode, service.NetworkMode, "Expected NetworkMode to match")
}

func TestCreateComposeService_SetsLabels(t *testing.T) {
	// GIVEN
	taskDefinition := &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Image: aws.String("myApp"),
			},
		},
	}

	expectedInputType := "remote"
	expectedValue := "arn:aws:ecs:us-west-2:123412341234:task-definition/myTaskDef:1"
	expectedMetadata := &LocalCreateMetadata{
		InputType: expectedInputType,
		Value:     expectedValue,
	}

	// WHEN
	services, err := createComposeServices(taskDefinition, expectedMetadata)
	service := services[0]

	// THEN
	assert.NoError(t, err, "Unexpected error creating Compose services")
	assert.Equal(t, expectedInputType, service.Labels[taskDefinitionLabelType], "Expected Metadata Type label to match")
	assert.Equal(t, expectedValue, service.Labels[taskDefinitionLabelValue], "Expected Metadata Value label to match")
}

func TestConvertToComposeService_ErrorsWithAwsVpcNetworkMode(t *testing.T) {
	// GIVEN
	taskDefinition := &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Image: aws.String("myApp"),
			},
		},
		NetworkMode: aws.String(ecs.NetworkModeAwsvpc),
	}

	// WHEN
	_, err := ConvertToDockerCompose(taskDefinition, &LocalCreateMetadata{}) // FIXME

	// THEN
	assert.Error(t, err)
}

func TestConvertToTmpfs(t *testing.T) {
	expectedTmpfs := []string{
		"/run:size=64MiB,rw,noexec,nosuid",
		"/foo:size=1GiB",
	}

	input := []*ecs.Tmpfs{
		{
			ContainerPath: aws.String("/run"),
			MountOptions:  aws.StringSlice([]string{"rw", "noexec", "nosuid"}),
			Size:          aws.Int64(64),
		},
		{
			ContainerPath: aws.String("/foo"),
			Size:          aws.Int64(1024),
		},
	}

	actual, err := convertToTmpfs(input)
	assert.NoError(t, err, "Unexpected error converting Tmpfs")
	assert.ElementsMatch(t, expectedTmpfs, actual)
}

func TestConvertToTmpfs_ErrorsIfNoSize(t *testing.T) {
	input := []*ecs.Tmpfs{
		{
			ContainerPath: aws.String("/run"),
			MountOptions:  aws.StringSlice([]string{"rw", "noexec", "nosuid"}),
		},
	}

	_, err := convertToTmpfs(input)
	assert.Error(t, err)
}

func TestConvertToTmpfs_ErrorsIfNoPath(t *testing.T) {
	input := []*ecs.Tmpfs{
		{
			MountOptions: aws.StringSlice([]string{"rw", "noexec", "nosuid"}),
			Size:         aws.Int64(1024),
		},
	}

	_, err := convertToTmpfs(input)
	assert.Error(t, err)
}

func TestConvertUlimits(t *testing.T) {
	expected := map[string]*composeV3.UlimitsConfig{
		"nofile": &composeV3.UlimitsConfig{
			Soft: 2000,
			Hard: 4000,
		},
		// Ignoring "Single" field - hack
		"rss": &composeV3.UlimitsConfig{
			Soft: 65535,
			Hard: 65535,
		},
	}

	input := []*ecs.Ulimit{
		{
			Name:      aws.String("nofile"),
			HardLimit: aws.Int64(4000),
			SoftLimit: aws.Int64(2000),
		},
		{
			Name:      aws.String("rss"),
			HardLimit: aws.Int64(65535),
			SoftLimit: aws.Int64(65535),
		},
	}

	actual, err := convertUlimits(input)

	assert.NoError(t, err, "Unexpected error converting Ulimits")
	assert.Equal(t, expected["rss"], actual["rss"])
	assert.Equal(t, expected["nofile"], actual["nofile"])
}

func TestConvertDevices(t *testing.T) {
	expected := []string{
		"/dev/sda",
		"/dev/sda:/dev/xvdc",
		"/dev/sda:/dev/xvdc:r",
		"/dev/nvid:/dev/xvdc:rw",
	}

	input := []*ecs.Device{
		{
			HostPath: aws.String("/dev/sda"),
		},
		{
			HostPath:      aws.String("/dev/sda"),
			ContainerPath: aws.String("/dev/xvdc"),
		},
		{
			HostPath:      aws.String("/dev/sda"),
			ContainerPath: aws.String("/dev/xvdc"),
			Permissions:   aws.StringSlice([]string{"read"}),
		},
		{
			HostPath:      aws.String("/dev/nvid"),
			ContainerPath: aws.String("/dev/xvdc"),
			Permissions:   aws.StringSlice([]string{"read", "write"}),
		},
	}

	actual, err := convertDevices(input)

	assert.NoError(t, err, "Unexpected error converting Devices")
	assert.ElementsMatch(t, expected, actual)
}

func TestConvertShmSize(t *testing.T) {
	input := aws.Int64(1024)
	expected := "1GiB"
	actual := convertShmSize(input)

	assert.Equal(t, expected, actual)
}

func TestConvertShmSize_Nil(t *testing.T) {
	expected := ""
	actual := convertShmSize(nil)

	assert.Equal(t, expected, actual)
}

func TestConvertCapAddCapDrop(t *testing.T) {
	addCapabilities := []string{"NET_ADMIN", "MKNOD"}
	dropCapabilities := []string{"KILL"}

	input := &ecs.KernelCapabilities{
		Add:  aws.StringSlice(addCapabilities),
		Drop: aws.StringSlice(dropCapabilities),
	}
	actualCapAdd := convertCapAdd(input)
	actualCapDrop := convertCapDrop(input)

	assert.ElementsMatch(t, addCapabilities, actualCapAdd)
	assert.ElementsMatch(t, dropCapabilities, actualCapDrop)
}

func TestConvertExtraHosts(t *testing.T) {
	input := []*ecs.HostEntry{
		{
			Hostname:  aws.String("somehost"),
			IpAddress: aws.String("162.242.195.82"),
		},
		{
			Hostname:  aws.String("otherhost"),
			IpAddress: aws.String("50.31.209.229"),
		},
	}

	expected := []string{"somehost:162.242.195.82", "otherhost:50.31.209.229"}
	actual := convertExtraHosts(input)

	assert.Equal(t, expected, actual)
}

func TestConvertHealthCheck(t *testing.T) {
	command := []string{"CMD", "curl", "-f", "http://localhost"}
	input := &ecs.HealthCheck{
		Command:     aws.StringSlice(command),
		Retries:     aws.Int64(3),
		Interval:    aws.Int64(90),
		Timeout:     aws.Int64(10),
		StartPeriod: aws.Int64(40),
	}

	interval := time.Duration(90) * time.Second
	timeout := time.Duration(10) * time.Second
	startPeriod := time.Duration(40) * time.Second
	retries := uint64(3)

	expected := &composeV3.HealthCheckConfig{
		Test:        command,
		Retries:     &retries,
		Interval:    &interval,
		Timeout:     &timeout,
		StartPeriod: &startPeriod,
	}
	actual := convertHealthCheck(input)

	assert.Equal(t, expected, actual)
}

func TestConvertLogging(t *testing.T) {
	input := &ecs.LogConfiguration{
		LogDriver: aws.String("awslogs"),
		Options: map[string]*string{
			"awslogs-group":         aws.String("/ecs/fargate-task-definition"),
			"awslogs-region":        aws.String("us-east-1"),
			"awslogs-stream-prefix": aws.String("ecs"),
		},
	}

	expected := &composeV3.LoggingConfig{
		Driver: "awslogs",
		Options: map[string]string{
			"awslogs-group":         "/ecs/fargate-task-definition",
			"awslogs-region":        "us-east-1",
			"awslogs-stream-prefix": "ecs",
		},
	}

	actual := convertLogging(input)

	assert.Equal(t, expected, actual)
}

func TestConvertToVolumes(t *testing.T) {
	input := []*ecs.MountPoint{
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-1"),
		},
		{
			ContainerPath: aws.String("/tmp/cache2"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-2"),
		},
	}

	expected := []composeV3.ServiceVolumeConfig{
		{
			Target:   "/tmp/cache",
			ReadOnly: false,
			Source:   "volume-1",
		},
		{
			Target:   "/tmp/cache2",
			ReadOnly: false,
			Source:   "volume-2",
		},
	}

	actual := convertToVolumes(input)

	assert.Equal(t, expected, actual)
}

func TestConvertToPorts(t *testing.T) {
	input := []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(3000),
			Protocol:      aws.String("tcp"),
			HostPort:      aws.Int64(80),
		},
	}

	expected := []composeV3.ServicePortConfig{
		{
			Target:    uint32(3000),
			Published: uint32(80),
			Protocol:  "tcp",
		},
	}

	actual := convertToPorts(input)

	assert.Equal(t, expected, actual)
}

func TestConvertToSysctls(t *testing.T) {
	input := []*ecs.SystemControl{
		{
			Namespace: aws.String("net.core.somaxconn"),
			Value:     aws.String("1024"),
		},
		{
			Namespace: aws.String("net.ipv4.tcp_syncookies"),
			Value:     aws.String("0"),
		},
	}

	expected := []string{
		"net.core.somaxconn=1024",
		"net.ipv4.tcp_syncookies=0",
	}

	actual := convertToSysctls(input)

	assert.Equal(t, expected, actual)
}
