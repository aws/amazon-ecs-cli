// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"strconv"
	"testing"

	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	portNumber    = 8000
	portMapping   = "8000:8000"
	containerPath = "/tmp/cache"
	hostPath      = "./cache"
)

func TestConvertToTaskDefinition(t *testing.T) {
	name := "mysql"
	memory := int64(100) // 1 MiB = 1048576B
	envKey := "username"
	envValue := "root"
	env := envKey + "=" + envValue
	cpu := int64(10)
	image := "testimage"
	projectName := "ProjectName"
	volumes := []string{hostPath + ":" + containerPath}
	volumesFrom := []string{"container1"}
	links := []string{"container1"}
	command := "cmd"

	serviceConfigs := make(map[string]*libcompose.ServiceConfig)
	serviceConfigs[name] = &libcompose.ServiceConfig{
		MemLimit:    int64(1048576) * memory,
		Environment: libcompose.NewMaporEqualSlice([]string{env}),
		CpuShares:   cpu,
		Image:       image,
		Command:     libcompose.NewCommand(command),
		Volumes:     volumes,
		VolumesFrom: volumesFrom,
		Links:       libcompose.NewMaporColonSlice(links),
	}

	context := libcompose.Context{
		ProjectName: projectName,
	}
	taskDefinition, err := ConvertToTaskDefinition(context, serviceConfigs)
	if err != nil {
		t.Errorf("Expected to convert [%v] serviceConfigs without errors. But got [%v]", serviceConfigs, err)
	}
	containerDef := *taskDefinition.ContainerDefinitions[0]
	if name != *containerDef.Name {
		t.Errorf("Expected Name [%s] But was [%s]", name, *containerDef.Name)
	}
	if memory != *containerDef.Memory {
		t.Errorf("Expected memory [%s] But was [%s]", memory, *containerDef.Memory)
	}
	if cpu != *containerDef.Cpu {
		t.Errorf("Expected cpu [%s] But was [%s]", cpu, *containerDef.Name)
	}
	if image != *containerDef.Image {
		t.Errorf("Expected Image [%s] But was [%s]", image, *containerDef.Image)
	}
	if len(containerDef.Command) != 1 || command != *containerDef.Command[0] {
		t.Errorf("Expected command [%s] But was [%v]", command, containerDef.Command)
	}
	if len(volumesFrom) != len(containerDef.VolumesFrom) || volumesFrom[0] != *containerDef.VolumesFrom[0].SourceContainer {
		t.Errorf("Expected volumesFrom [%v] But was [%v]", volumesFrom, containerDef.VolumesFrom)
	}
	if len(links) != len(containerDef.Links) || links[0] != *containerDef.Links[0] {
		t.Errorf("Expected links [%v] But was [%v]", links, containerDef.Links)
	}

	volumeDef := *taskDefinition.Volumes[0]
	mountPoint := *containerDef.MountPoints[0]
	if hostPath != *volumeDef.Host.SourcePath {
		t.Errorf("Expected HostSourcePath [%s] But was [%s]", hostPath, *volumeDef.Host.SourcePath)
	}
	if containerPath != *mountPoint.ContainerPath {
		t.Errorf("Expected containerPath [%s] But was [%s]", containerPath, *mountPoint.ContainerPath)
	}
	if *volumeDef.Name != *mountPoint.SourceVolume {
		t.Errorf("Expected volume name to match. "+
			"Got Volume.Name=[%s] And MountPoint.SourceVolume=[%s]", *volumeDef.Name, *mountPoint.SourceVolume)
	}
}

func TestConvertToPortMappings(t *testing.T) {
	implicitTcp := portMapping                      // 8000:8000
	explicitTcp := portMapping + "/tcp"             // "8000:8000/tcp"
	udpPort := portMapping + "/udp"                 // "8000:8000/udp"
	containerPortOnly := strconv.Itoa(portNumber)   // "8000"
	portWithIpAddress := "127.0.0.1:" + portMapping // "127.0.0.1:8000:8000"

	portMappingsIn := []string{implicitTcp, explicitTcp, udpPort, containerPortOnly, portWithIpAddress}

	portMappingsOut, err := convertToPortMappings("test", portMappingsIn)
	if err != nil {
		t.Errorf("Expected to convert [%v] portMappings without errors. But got [%v]", portMappingsIn, err)
	}
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
		t.Errorf("Expected hostPort [%s] But was [%s]", hostPort, *output.HostPort)
	}
	if containerPort != *output.ContainerPort {
		t.Errorf("Expected containerPort [%s] But was [%s]", containerPort, *output.ContainerPort)
	}
}

func TestConvertToMountPoints(t *testing.T) {
	hostAndContainerPath := hostPath + ":" + containerPath     // "./cache:/tmp/cache"
	hostAndContainerPathWithRO := hostAndContainerPath + ":ro" // "./cache:/tmp/cache:ro"
	hostAndContainerPathWithRW := hostAndContainerPath + ":rw"

	volumes := make(map[string]string)

	mountPointsIn := []string{containerPath, hostAndContainerPath,
		hostAndContainerPathWithRO, hostAndContainerPathWithRW}

	mountPointsOut, err := convertToMountPoints(mountPointsIn, volumes)
	if err != nil {
		t.Fatalf("Expected to convert [%v] mountPoints without errors. But got [%v]", mountPointsIn, err)
	}
	if len(mountPointsIn) != len(mountPointsOut) {
		t.Errorf("Incorrect conversion. Input [%v] Output [%v]", mountPointsIn, mountPointsOut)
	}
	verifyMountPoint(t, mountPointsOut[0], volumes, "", containerPath, false)
	verifyMountPoint(t, mountPointsOut[1], volumes, hostPath, containerPath, false)
	verifyMountPoint(t, mountPointsOut[2], volumes, hostPath, containerPath, true)
	verifyMountPoint(t, mountPointsOut[3], volumes, hostPath, containerPath, false)

	hostAndContainerPathWithIncorrectAccess := hostAndContainerPath + ":readonly"
	mountPointsOut, err = convertToMountPoints([]string{hostAndContainerPathWithIncorrectAccess}, volumes)
	if err == nil {
		t.Errorf("Expected to get error for mountPoint[%s] but didn't.", hostAndContainerPathWithIncorrectAccess)
	}

	incorrectPath := ":::"
	mountPointsOut, err = convertToMountPoints([]string{incorrectPath}, volumes)
	if err == nil {
		t.Errorf("Expected to get error for mountPoint[%s] but didn't.", incorrectPath)
	}
}

func verifyMountPoint(t *testing.T, output *ecs.MountPoint, volumes map[string]string,
	hostPath, containerPath string, readonly bool) {

	if containerPath != *output.ContainerPath {
		t.Errorf("Expected containerPath [%s] But was [%s]", containerPath, *output.ContainerPath)
	}
	sourceVolume := volumes[hostPath]
	if sourceVolume != *output.SourceVolume {
		t.Errorf("Expected sourceVolume [%s] But was [%s]", sourceVolume, *output.SourceVolume)
	}
	if readonly != *output.ReadOnly {
		t.Errorf("Expected readonly [%s] But was [%s]", readonly, *output.ReadOnly)
	}
}
