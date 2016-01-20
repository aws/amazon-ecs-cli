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
	"fmt"
	"reflect"
	"strconv"
	"testing"

	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
	"github.com/aws/aws-sdk-go/aws"
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
	cpu := int64(10)
	command := "cmd"
	essential := false
	hostname := "foobarbaz"
	image := "testimage"
	links := []string{"container1"}
	memory := int64(100) // 1 MiB = 1048576B
	privileged := true
	readOnly := true
	restart := "no"
	securityOpts := []string{"label:type:test_virt"}
	user := "user"
	workingDir := "/var"

	serviceConfig := &libcompose.ServiceConfig{
		CpuShares:   cpu,
		Command:     libcompose.NewCommand(command),
		Hostname:    hostname,
		Image:       image,
		Links:       libcompose.NewMaporColonSlice(links),
		MemLimit:    int64(1048576) * memory,
		Privileged:  privileged,
		ReadOnly:    readOnly,
		Restart:     restart,
		SecurityOpt: securityOpts,
		User:        user,
		WorkingDir:  workingDir,
	}

	// convert
	taskDefinition := convertToTaskDefinitionInTest(t, name, serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]

	// verify
	if name != aws.StringValue(containerDef.Name) {
		t.Errorf("Expected Name [%s] But was [%s]", name, aws.StringValue(containerDef.Name))
	}
	if cpu != aws.Int64Value(containerDef.Cpu) {
		t.Errorf("Expected cpu [%s] But was [%s]", cpu, aws.Int64Value(containerDef.Cpu))
	}
	if len(containerDef.Command) != 1 || command != aws.StringValue(containerDef.Command[0]) {
		t.Errorf("Expected command [%s] But was [%v]", command, containerDef.Command)
	}
	if essential != aws.BoolValue(containerDef.Essential) {
		t.Errorf("Expected essential [%s] But was [%s]", essential, aws.BoolValue(containerDef.Essential))
	}
	if !reflect.DeepEqual(securityOpts, aws.StringValueSlice(containerDef.DockerSecurityOptions)) {
		t.Errorf("Expected securityOpt [%v] But was [%v]", securityOpts, aws.StringValueSlice(containerDef.DockerSecurityOptions))
	}
	if hostname != aws.StringValue(containerDef.Hostname) {
		t.Errorf("Expected hostname [%s] But was [%s]", hostname, aws.StringValue(containerDef.Hostname))
	}
	if image != aws.StringValue(containerDef.Image) {
		t.Errorf("Expected Image [%s] But was [%s]", image, aws.StringValue(containerDef.Image))
	}
	if !reflect.DeepEqual(links, aws.StringValueSlice(containerDef.Links)) {
		t.Errorf("Expected links [%v] But was [%v]", links, aws.StringValueSlice(containerDef.Links))
	}
	if memory != aws.Int64Value(containerDef.Memory) {
		t.Errorf("Expected memory [%s] But was [%s]", memory, aws.Int64Value(containerDef.Memory))
	}
	if privileged != aws.BoolValue(containerDef.Privileged) {
		t.Errorf("Expected privileged [%s] But was [%s]", privileged, aws.BoolValue(containerDef.Privileged))
	}
	if readOnly != aws.BoolValue(containerDef.ReadonlyRootFilesystem) {
		t.Errorf("Expected ReadonlyRootFilesystem [%s] But was [%s]", readOnly, aws.BoolValue(containerDef.ReadonlyRootFilesystem))
	}
	if user != aws.StringValue(containerDef.User) {
		t.Errorf("Expected user [%s] But was [%s]", user, aws.StringValue(containerDef.User))
	}
	if workingDir != aws.StringValue(containerDef.WorkingDirectory) {
		t.Errorf("Expected WorkingDirectory [%s] But was [%s]", workingDir, aws.StringValue(containerDef.WorkingDirectory))
	}
}

func TestConvertToTaskDefinitionWithDnsSearch(t *testing.T) {
	dnsSearchDomains := []string{"search.example.com"}

	serviceConfig := &libcompose.ServiceConfig{DNSSearch: libcompose.NewStringorslice(dnsSearchDomains...)}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	if !reflect.DeepEqual(dnsSearchDomains, aws.StringValueSlice(containerDef.DnsSearchDomains)) {
		t.Errorf("Expected dnsSearchDomains [%v] But was [%v]", dnsSearchDomains,
			aws.StringValueSlice(containerDef.DnsSearchDomains))
	}
}

func TestConvertToTaskDefinitionWithDnsServers(t *testing.T) {
	dnsServer := "1.2.3.4"

	serviceConfig := &libcompose.ServiceConfig{DNS: libcompose.NewStringorslice(dnsServer)}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	if !reflect.DeepEqual([]string{dnsServer}, aws.StringValueSlice(containerDef.DnsServers)) {
		t.Errorf("Expected dnsServer [%s] But was [%v]", dnsServer, aws.StringValueSlice(containerDef.DnsServers))
	}
}

func TestConvertToTaskDefinitionWithDockerLabels(t *testing.T) {
	dockerLabels := map[string]string{
		"label1":         "",
		"com.foo.label2": "value",
	}

	serviceConfig := &libcompose.ServiceConfig{Labels: libcompose.NewSliceorMap(dockerLabels)}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	if !reflect.DeepEqual(dockerLabels, aws.StringValueMap(containerDef.DockerLabels)) {
		t.Errorf("Expected dockerLabels [%v] But was [%v]", dockerLabels, aws.StringValueMap(containerDef.DockerLabels))
	}
}

func TestConvertToTaskDefinitionWithEnv(t *testing.T) {
	envKey := "username"
	envValue := "root"
	env := envKey + "=" + envValue
	serviceConfig := &libcompose.ServiceConfig{
		Environment: libcompose.NewMaporEqualSlice([]string{env}),
	}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]

	if envKey != aws.StringValue(containerDef.Environment[0].Name) ||
		envValue != aws.StringValue(containerDef.Environment[0].Value) {
		t.Errorf("Expected env [%s] But was [%v]", env, containerDef.Environment)
	}
}

func TestConvertToTaskDefinitionWithPortMappings(t *testing.T) {
	serviceConfig := &libcompose.ServiceConfig{Ports: []string{portMapping}}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	verifyPortMapping(t, containerDef.PortMappings[0], portNumber, portNumber, ecs.TransportProtocolTcp)
}

func TestConvertToTaskDefinitionWithExtraHosts(t *testing.T) {
	hostname := "test.local"
	ipAddress := "127.10.10.10"

	extraHost := hostname + ":" + ipAddress
	serviceConfig := &libcompose.ServiceConfig{ExtraHosts: []string{extraHost}}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	verifyExtraHost(t, containerDef.ExtraHosts[0], hostname, ipAddress)
}

func TestConvertToTaskDefinitionWithLogConfiguration(t *testing.T) {
	taskDefinition := convertToTaskDefinitionInTest(t, "name", &libcompose.ServiceConfig{})
	containerDef := *taskDefinition.ContainerDefinitions[0]

	if containerDef.LogConfiguration != nil {
		t.Errorf("Expected empty log configuration. But was [%v]", containerDef.LogConfiguration)
	}

	logDriver := "json-file"
	logOpts := map[string]string{
		"max-file": "50",
		"max-size": "50k",
	}
	serviceConfig := &libcompose.ServiceConfig{
		LogDriver: logDriver,
		LogOpt:    logOpts,
	}

	taskDefinition = convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef = *taskDefinition.ContainerDefinitions[0]
	if logDriver != aws.StringValue(containerDef.LogConfiguration.LogDriver) {
		t.Errorf("Expected Log driver [%s]. But was [%s]", containerDef.LogConfiguration)
	}
	if !reflect.DeepEqual(logOpts, aws.StringValueMap(containerDef.LogConfiguration.Options)) {
		t.Errorf("Expected Log options [%v]. But was [%v]", logOpts, aws.StringValueMap(containerDef.LogConfiguration.Options))
	}
}

func TestConvertToTaskDefinitionWithUlimits(t *testing.T) {
	softLimit := int64(1024)
	typeName := "nofile"
	basicType := fmt.Sprintf("%s=%d", typeName, softLimit)
	serviceConfig := &libcompose.ServiceConfig{ULimits: []string{basicType}}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	verifyUlimit(t, containerDef.Ulimits[0], typeName, softLimit, softLimit)
}

func TestConvertToTaskDefinitionWithVolumes(t *testing.T) {
	volumes := []string{hostPath + ":" + containerPath}
	volumesFrom := []string{"container1"}

	serviceConfig := &libcompose.ServiceConfig{
		Volumes:     volumes,
		VolumesFrom: volumesFrom,
	}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]

	if len(volumesFrom) != len(containerDef.VolumesFrom) ||
		volumesFrom[0] != aws.StringValue(containerDef.VolumesFrom[0].SourceContainer) {
		t.Errorf("Expected volumesFrom [%v] But was [%v]", volumesFrom, containerDef.VolumesFrom)
	}
	volumeDef := *taskDefinition.Volumes[0]
	mountPoint := *containerDef.MountPoints[0]
	if hostPath != aws.StringValue(volumeDef.Host.SourcePath) {
		t.Errorf("Expected HostSourcePath [%s] But was [%s]", hostPath, aws.StringValue(volumeDef.Host.SourcePath))
	}
	if containerPath != aws.StringValue(mountPoint.ContainerPath) {
		t.Errorf("Expected containerPath [%s] But was [%s]", containerPath, aws.StringValue(mountPoint.ContainerPath))
	}
	if aws.StringValue(volumeDef.Name) != aws.StringValue(mountPoint.SourceVolume) {
		t.Errorf("Expected volume name to match. "+
			"Got Volume.Name=[%s] And MountPoint.SourceVolume=[%s]",
			aws.StringValue(volumeDef.Name), aws.StringValue(mountPoint.SourceVolume))
	}
}

func convertToTaskDefinitionInTest(t *testing.T, name string, serviceConfig *libcompose.ServiceConfig) *ecs.TaskDefinition {
	serviceConfigs := make(map[string]*libcompose.ServiceConfig)
	serviceConfigs[name] = serviceConfig

	projectName := "ProjectName"
	context := libcompose.Context{
		ProjectName: projectName,
	}
	taskDefinition, err := ConvertToTaskDefinition(context, serviceConfigs)
	if err != nil {
		t.Errorf("Expected to convert [%v] serviceConfigs without errors. But got [%v]", serviceConfigs, err)
	}
	return taskDefinition
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

func TestConvertToExtraHosts(t *testing.T) {
	hostname := "test.local"
	ipAddress := "127.10.10.10"

	extraHost := hostname + ":" + ipAddress

	extraHostsIn := []string{extraHost}
	extraHostsOut, err := convertToExtraHosts(extraHostsIn)
	if err != nil {
		t.Errorf("Expected to convert [%v] extra hosts without errors. But got [%v]", extraHostsIn, err)
	}
	if len(extraHostsIn) != len(extraHostsOut) {
		t.Errorf("Incorrect conversion. Input [%v] Output [%v]", extraHostsIn, extraHostsOut)
	}
	verifyExtraHost(t, extraHostsOut[0], hostname, ipAddress)

	incorrectHost := hostname + "=" + ipAddress
	_, err = convertToExtraHosts([]string{incorrectHost})
	if err == nil {
		t.Errorf("Expected to get formatting error for extraHost=[%s], but got none", incorrectHost)
	}

	extraHostWithPort := fmt.Sprintf("%s:%s:%d", hostname, ipAddress, portNumber)
	_, err = convertToExtraHosts([]string{extraHostWithPort})
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
	basicType := fmt.Sprintf("%s=%d", typeName, softLimit)          // "nofile=1024"
	typeWithHardLimit := fmt.Sprintf("%s:%d", basicType, hardLimit) // "nofile=1024:2048"

	ulimitsIn := []string{basicType, typeWithHardLimit}
	ulimitsOut, err := convertToULimits(ulimitsIn)
	if err != nil {
		t.Errorf("Expected to convert [%v] ulimits without errors. But got [%v]", ulimitsIn, err)
	}
	if len(ulimitsIn) != len(ulimitsOut) {
		t.Errorf("Incorrect conversion. Input [%v] Output [%v]", ulimitsIn, ulimitsOut)
	}
	verifyUlimit(t, ulimitsOut[0], typeName, softLimit, softLimit)
	verifyUlimit(t, ulimitsOut[1], typeName, softLimit, hardLimit)

	incorrectType := "incorrect"
	_, err = convertToULimits([]string{incorrectType})
	if err == nil {
		t.Errorf("Expected to get formatting error for ulimit value=[%s], but got none", incorrectType)
	}

	incorrectHardLimit := basicType + ":random"
	_, err = convertToULimits([]string{incorrectHardLimit})
	if err == nil {
		t.Errorf("Expected to get formatting error for the ulimit with random hardLimit=[%s], but got none", incorrectHardLimit)
	}
}

func verifyUlimit(t *testing.T, output *ecs.Ulimit, name string, softLimit, hardLimit int64) {
	if name != *output.Name {
		t.Errorf("Expected name [%s] But was [%s]", name, *output.Name)
	}
	if softLimit != *output.SoftLimit {
		t.Errorf("Expected softLimit [%s] But was [%s]", softLimit, *output.SoftLimit)
	}
	if hardLimit != *output.HardLimit {
		t.Errorf("Expected hardLimit [%s] But was [%s]", hardLimit, *output.HardLimit)
	}
}
