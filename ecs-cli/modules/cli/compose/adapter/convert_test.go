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
	"strconv"
	"testing"

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
		VolumeWithHost:  make(map[string]string), // map with key:=hostSourcePath value:=VolumeName
		VolumeEmptyHost: []string{namedVolume},   // Declare one volume with an empty host
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

	if err == nil {
		t.Errorf("Expected to get error for mountPoint[%s] but didn't.", hostAndContainerPathWithIncorrectAccess)
	}
}

func TestConvertToMountPointsNullContainerVolumes(t *testing.T) {
	volumes := &Volumes{
		VolumeWithHost:  make(map[string]string),
		VolumeEmptyHost: []string{namedVolume},
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
		sourceVolume = volumes.VolumeEmptyHost[EmptyHostCtr]
	} else if project.IsNamedVolume(source) {
		sourceVolume = source
	} else {
		sourceVolume = volumes.VolumeWithHost[source]
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
