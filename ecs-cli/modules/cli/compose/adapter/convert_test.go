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

package adapter

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/libcompose/config"
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
	implicitTCP := portMapping                      // 8000:8000
	explicitTCP := portMapping + "/tcp"             // "8000:8000/tcp"
	udpPort := portMapping + "/udp"                 // "8000:8000/udp"
	containerPortOnly := strconv.Itoa(portNumber)   // "8000"
	portWithIPAddress := "127.0.0.1:" + portMapping // "127.0.0.1:8000:8000"

	portMappingsIn := []string{implicitTCP, explicitTCP, udpPort, containerPortOnly, portWithIPAddress}
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
	// Valid inputs with host and container paths set:
	// /tmp/cache
	// /tmp/cache2
	// ./cache:/tmp/cache
	// /tmp/cache:ro
	// ./cache:/tmp/cache:ro
	// ./cache:/tmp/cache:rw
	// named_volume:/tmp/cache
	onlyContainerPath := yaml.Volume{Destination: containerPath}
	onlyContainerPath2 := yaml.Volume{Destination: containerPath2}
	hostAndContainerPath := yaml.Volume{Source: hostPath, Destination: containerPath}
	onlyContainerPathWithRO := yaml.Volume{Destination: containerPath, AccessMode: "ro"}
	hostAndContainerPathWithRO := yaml.Volume{Source: hostPath, Destination: containerPath, AccessMode: "ro"}
	hostAndContainerPathWithRW := yaml.Volume{Source: hostPath, Destination: containerPath, AccessMode: "rw"}
	namedVolumeAndContainerPath := yaml.Volume{Source: namedVolume, Destination: containerPath}

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

	expectedVolumeWithHost := map[string]string{hostPath: "volume-3"}
	expectedVolumeEmptyHost := []string{namedVolume, "volume-1", "volume-2", "volume-4"}
	expectedMountPoints := []*ecs.MountPoint{
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
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-3"),
		},
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(true),
			SourceVolume:  aws.String("volume-4"),
		},
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(true),
			SourceVolume:  aws.String("volume-3"),
		},
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

	volumes := &Volumes{
		VolumeWithHost:  make(map[string]string), // This field should be empty before ConvertToMountPoints is called
		VolumeEmptyHost: []string{namedVolume},   // We expect ConvertToVolumes to already have been called, so any named volumes should have been added to VolumeEmptyHost
	}

	mountPointsOut, err := ConvertToMountPoints(&mountPointsIn, volumes)
	assert.NoError(t, err, "Unexpected error converting MountPoints")

	// Expect top-level volumes fields to be populated
	assert.Equal(t, expectedVolumeWithHost, volumes.VolumeWithHost, "Expected volumeWithHost to match")
	assert.Equal(t, expectedVolumeEmptyHost, volumes.VolumeEmptyHost, "Expected volumeEmptyHost to match")

	assert.ElementsMatch(t, expectedMountPoints, mountPointsOut, "Expected Mount Points to match")

	if mountPointsOut[0].SourceVolume == mountPointsOut[1].SourceVolume {
		t.Errorf("Expected volume %v (onlyContainerPath) and %v (onlyContainerPath2) to be different", mountPointsOut[0].SourceVolume, mountPointsOut[1].SourceVolume)
	}

	if mountPointsOut[1].SourceVolume == mountPointsOut[3].SourceVolume {
		t.Errorf("Expected volume %v (onlyContainerPath2) and %v (onlyContainerPathWithRO) to be different", mountPointsOut[0].SourceVolume, mountPointsOut[1].SourceVolume)
	}
}

func TestConvertToMountPointsWithInvalidAccessMode(t *testing.T) {
	volumes := &Volumes{
		VolumeWithHost:  make(map[string]string),
		VolumeEmptyHost: []string{namedVolume},
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
	assert.Error(t, err, "Expected error converting MountPoints")
}

func TestConvertToMountPointsNullContainerVolumes(t *testing.T) {
	volumes := &Volumes{
		VolumeWithHost:  make(map[string]string),
		VolumeEmptyHost: []string{namedVolume},
	}
	mountPointsOut, err := ConvertToMountPoints(nil, volumes)
	assert.NoError(t, err, "Unexpected error converting MountPoints")
	assert.Empty(t, mountPointsOut, "Expected mount points to be empty")
}

func TestConvertToMountPointsWithNoCorrespondingNamedVolume(t *testing.T) {
	volumes := &Volumes{
		VolumeWithHost:  make(map[string]string),
		VolumeEmptyHost: []string{}, // Top-level named volumes is empty
	}

	namedVolume := yaml.Volume{
		Source:      namedVolume,
		Destination: containerPath,
	}

	mountPointsIn := yaml.Volumes{
		Volumes: []*yaml.Volume{&namedVolume},
	}

	_, err := ConvertToMountPoints(&mountPointsIn, volumes)
	assert.Error(t, err, "Expected error converting MountPoints")
}

func TestGetSourcePathAndUpdateVolumesWithEmptySourcePath(t *testing.T) {
	expectedSourcePath := "volume-0"
	volumes := &Volumes{
		VolumeWithHost:  make(map[string]string),
		VolumeEmptyHost: []string{}, // Top-level named volumes is empty
	}

	observedSourcePath, err := GetSourcePathAndUpdateVolumes("", volumes)

	assert.NoError(t, err, "Unexpected error getting Mount Point source path")
	assert.Equal(t, expectedSourcePath, observedSourcePath)
	assert.Equal(t, expectedSourcePath, volumes.VolumeEmptyHost[0])
}

func TestGetSourcePathAndUpdateVolumesWithNamedVol(t *testing.T) {
	namedSourcePath := "logging"
	volumes := &Volumes{
		VolumeWithHost:  make(map[string]string),
		VolumeEmptyHost: []string{namedSourcePath},
	}

	observedSourcePath, err := GetSourcePathAndUpdateVolumes(namedSourcePath, volumes)

	assert.NoError(t, err, "Unexpected error getting Mount Point source path")
	assert.Equal(t, namedSourcePath, observedSourcePath)
	assert.Equal(t, namedSourcePath, volumes.VolumeEmptyHost[0])
}

func TestConvertToDevices(t *testing.T) {
	testCases := []struct {
		input                 string
		expectedHostPath      string
		expectedContainerPath string
		expectedPermissions   []string
	}{
		{"/dev/sda", "/dev/sda", "", []string{}},
		{"/dev/sda:/dev/xvdc", "/dev/sda", "/dev/xvdc", []string{}},
		{"/dev/sda:/dev/xvdc:w", "/dev/sda", "/dev/xvdc", []string{"write"}},
		{"/dev/nvid:/dev/xvdc:rw", "/dev/nvid", "/dev/xvdc", []string{"read", "write"}},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("Convert %s", test.input), func(t *testing.T) {
			inputSlice := []string{test.input}
			outputDevices, err := ConvertToDevices(inputSlice)
			assert.NoError(t, err, "Unexpected error converting Devices")
			assert.Equal(t, 1, len(outputDevices), "Expected Devices length to be 1")

			var expectedContainer *string
			var expectedPerms []*string

			if test.expectedContainerPath != "" {
				expectedContainer = aws.String(test.expectedContainerPath)
			}

			if len(test.expectedPermissions) != 0 {
				expectedPerms = aws.StringSlice(test.expectedPermissions)
			}

			outputDev := *outputDevices[0]
			assert.Equal(t, test.expectedHostPath, *outputDev.HostPath, "Expected HostPath to match")
			assert.Equal(t, expectedContainer, outputDev.ContainerPath, "Expected ContainerPath to match")
			assert.ElementsMatch(t, expectedPerms, outputDev.Permissions, "Expected Permissions to match")
		})
	}
}

func TestConvertToDevices_ErrorOnInvalidOptions(t *testing.T) {
	testCases := []string{
		"/dev/xf:/dex/gru:rw:m", // too many args
		"/dev/xx:/dex/gru:ytr",  // invalid option flags (y, t)
		"/dev/xx:/dex/sda:rrmw", // too many options (max is 3)
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("Expect error for %s", test), func(t *testing.T) {
			_, err := ConvertToDevices([]string{test})
			assert.Error(t, err, "Expected error for invalid device option")
		})
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
		VolumeWithHost:  make(map[string]string), // map with key:=hostSourcePath value:=VolumeName
		VolumeEmptyHost: []string{namedVolume},   // Declare one volume with an empty host
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

func TestRegisterTaskDefinitionInputEquivalence(t *testing.T) {
	family := aws.String("family1")
	dockerLabels := map[string]string{
		"label1":         "",
		"com.foo.label2": "value",
	}
	cdefs := []*ecs.ContainerDefinition{}
	N := 10
	for i := 0; i < N; i++ {
		command := make([]string, i+1)
		for j := 0; j < i+1; j++ {
			command[j] = strings.Repeat(string(rune(65+j)), i+1)
		}
		cdefs = append(cdefs, &ecs.ContainerDefinition{
			Name:         aws.String(strings.Repeat(string(rune(65+i)), i+1)),
			Command:      aws.StringSlice(command),
			DockerLabels: aws.StringMap(dockerLabels),
		})
	}
	inputA := ecs.RegisterTaskDefinitionInput{
		Family:               family,
		ContainerDefinitions: cdefs,
	}

	shuffle_cdefs := make([]*ecs.ContainerDefinition, len(cdefs))
	for i, v := range rand.Perm(len(cdefs)) {
		shuffle_cdefs[v] = cdefs[i]
	}

	inputB := ecs.RegisterTaskDefinitionInput{
		ContainerDefinitions: shuffle_cdefs,
		Family:               family,
	}

	strA, err := SortedGoString(SortedContainerDefinitionsByName(&inputA))
	assert.NoError(t, err, "Unexpected error generating sorted map string")
	strB, err := SortedGoString(SortedContainerDefinitionsByName(&inputB))
	assert.NoError(t, err, "Unexpected error generating sorted map string")

	assert.Equal(t, strA, strB, "Sorted inputs should match")
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

func TestConvertCamelCaseToUnderScore(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"CamelCaseString", "camel_case_string"},
		{"lowercase", "lowercase"},
		{"UPPERCASE", "uppercase"},
		{"CamelWithACRONYM", "camel_with_acronym"},
		{"Uppercase", "uppercase"},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("%s should be %s", test.input, test.expected), func(t *testing.T) {
			output := ConvertCamelCaseToUnderScore(test.input)
			assert.Equal(t, test.expected, output, "Expected output to match")
		})
	}
}

func TestConvertToHealthCheck(t *testing.T) {
	timeout := 10 * time.Second
	interval := time.Minute
	retries := uint64(3)
	startPeriod := 2 * time.Minute
	input := &types.HealthCheckConfig{
		Test:        []string{"CMD", "echo 'echo is not a good health check test command'"},
		Timeout:     &timeout,
		Interval:    &interval,
		Retries:     &retries,
		StartPeriod: &startPeriod,
	}
	output := ConvertToHealthCheck(input)
	assert.ElementsMatch(t, input.Test, aws.StringValueSlice(output.Command))
	assert.Equal(t, aws.Int64(10), output.Timeout)
	assert.Equal(t, aws.Int64(60), output.Interval)
	assert.Equal(t, aws.Int64(3), output.Retries)
	assert.Equal(t, aws.Int64(120), output.StartPeriod)
}
