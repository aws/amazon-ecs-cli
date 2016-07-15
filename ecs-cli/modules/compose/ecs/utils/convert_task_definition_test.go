// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"os"
	"reflect"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
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
	hostname := "foobarbaz"
	image := "testimage"
	links := []string{"container1"}
	memory := int64(100) // 1 MiB = 1048576B
	privileged := true
	readOnly := true
	securityOpts := []string{"label:type:test_virt"}
	user := "user"
	workingDir := "/var"

	serviceConfig := &config.ServiceConfig{
		CPUShares:   cpu,
		Command:     []string{command},
		Hostname:    hostname,
		Image:       image,
		Links:       links,
		MemLimit:    int64(1048576) * memory,
		Privileged:  privileged,
		ReadOnly:    readOnly,
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

	serviceConfig := &config.ServiceConfig{DNSSearch: dnsSearchDomains}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	if !reflect.DeepEqual(dnsSearchDomains, aws.StringValueSlice(containerDef.DnsSearchDomains)) {
		t.Errorf("Expected dnsSearchDomains [%v] But was [%v]", dnsSearchDomains,
			aws.StringValueSlice(containerDef.DnsSearchDomains))
	}
}

func TestConvertToTaskDefinitionWithDnsServers(t *testing.T) {
	dnsServer := "1.2.3.4"

	serviceConfig := &config.ServiceConfig{DNS: []string{dnsServer}}

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

	serviceConfig := &config.ServiceConfig{Labels: dockerLabels}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	if !reflect.DeepEqual(dockerLabels, aws.StringValueMap(containerDef.DockerLabels)) {
		t.Errorf("Expected dockerLabels [%v] But was [%v]", dockerLabels, aws.StringValueMap(containerDef.DockerLabels))
	}
}

func TestConvertToTaskDefinitionWithEnv(t *testing.T) {
	envKey := "rails_env"
	envValue := "development"
	env := envKey + "=" + envValue
	serviceConfig := &config.ServiceConfig{
		Environment: []string{env},
	}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]

	if envKey != aws.StringValue(containerDef.Environment[0].Name) ||
		envValue != aws.StringValue(containerDef.Environment[0].Value) {
		t.Errorf("Expected env [%s] But was [%v]", env, containerDef.Environment)
	}
}

func TestConvertToTaskDefinitionWithEnvFromShell(t *testing.T) {
	envKey1 := "rails_env"
	envValue1 := "development"
	env := envKey1 + "=" + envValue1
	envKey2 := "port"

	serviceConfig := &config.ServiceConfig{
		Environment: []string{envKey1, envKey2 + "="},
	}

	os.Setenv(envKey1, envValue1)
	defer func() {
		os.Unsetenv(envKey1)
	}()

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]

	// skips the second one if envKey2
	if containerDef.Environment == nil || len(containerDef.Environment) != 1 {
		t.Fatalf("Expected non empty Environment, but was [%v]", containerDef.Environment)
	}

	if envKey1 != aws.StringValue(containerDef.Environment[0].Name) ||
		envValue1 != aws.StringValue(containerDef.Environment[0].Value) {
		t.Errorf("Expected env [%s] But was [%v]", env, containerDef.Environment)
	}
}

func TestConvertToTaskDefinitionWithPortMappings(t *testing.T) {
	serviceConfig := &config.ServiceConfig{Ports: []string{portMapping}}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	verifyPortMapping(t, containerDef.PortMappings[0], portNumber, portNumber, ecs.TransportProtocolTcp)
}

func TestConvertToTaskDefinitionWithVolumesFrom(t *testing.T) {
	// compose file format v2
	setupAndTestVolumesFrom(t, "service_name", "service_name", false)
	setupAndTestVolumesFrom(t, "service_name:ro", "service_name", true)
	setupAndTestVolumesFrom(t, "service_name:rw", "service_name", false)

	setupAndTestVolumesFrom(t, "container:container_name", "container_name", false)
	setupAndTestVolumesFrom(t, "container:container_name:ro", "container_name", true)
	setupAndTestVolumesFrom(t, "container:container_name:rw", "container_name", false)

	// compose file format v1
	setupAndTestVolumesFrom(t, "container_name", "container_name", false)
	setupAndTestVolumesFrom(t, "container_name:ro", "container_name", true)
	setupAndTestVolumesFrom(t, "container_name:rw", "container_name", false)
}

func setupAndTestVolumesFrom(t *testing.T, volume, sourceContainer string, readOnly bool) {
	serviceConfig := &config.ServiceConfig{VolumesFrom: []string{volume}}
	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	verifyVolumeFrom(t, containerDef.VolumesFrom[0], sourceContainer, readOnly)
}

func TestConvertToTaskDefinitionWithExtraHosts(t *testing.T) {
	hostname := "test.local"
	ipAddress := "127.10.10.10"

	extraHost := hostname + ":" + ipAddress
	serviceConfig := &config.ServiceConfig{ExtraHosts: []string{extraHost}}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	verifyExtraHost(t, containerDef.ExtraHosts[0], hostname, ipAddress)
}

func TestConvertToTaskDefinitionWithLogConfiguration(t *testing.T) {
	taskDefinition := convertToTaskDefinitionInTest(t, "name", &config.ServiceConfig{})
	containerDef := *taskDefinition.ContainerDefinitions[0]

	if containerDef.LogConfiguration != nil {
		t.Errorf("Expected empty log configuration. But was [%v]", containerDef.LogConfiguration)
	}

	logDriver := "json-file"
	logOpts := map[string]string{
		"max-file": "50",
		"max-size": "50k",
	}
	serviceConfig := &config.ServiceConfig{
		Logging: config.Log{
			Driver:  logDriver,
			Options: logOpts,
		},
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
	basicType := yaml.NewUlimit(typeName, softLimit, softLimit) // "nofile=1024"
	serviceConfig := &config.ServiceConfig{
		Ulimits: yaml.Ulimits{Elements: []yaml.Ulimit{basicType}},
	}

	taskDefinition := convertToTaskDefinitionInTest(t, "name", serviceConfig)
	containerDef := *taskDefinition.ContainerDefinitions[0]
	verifyUlimit(t, containerDef.Ulimits[0], typeName, softLimit, softLimit)
}

func TestConvertToTaskDefinitionWithVolumes(t *testing.T) {
	volumes := []string{hostPath + ":" + containerPath}
	volumesFrom := []string{"container1"}

	serviceConfig := &config.ServiceConfig{
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

func verifyVolumeFrom(t *testing.T, output *ecs.VolumeFrom, containerName string, readOnly bool) {
	if containerName != aws.StringValue(output.SourceContainer) {
		t.Errorf("Expected SourceContainer [%s] But was [%s]", containerName, aws.StringValue(output.SourceContainer))
	}
	if readOnly != aws.BoolValue(output.ReadOnly) {
		t.Errorf("Expected ReadOnly [%t] But was [%t]", readOnly, aws.BoolValue(output.ReadOnly))
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
	ulimitsOut, err := convertToULimits(ulimitsIn)
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
		t.Errorf("Expected softLimit [%s] But was [%s]", softLimit, *output.SoftLimit)
	}
	if hardLimit != *output.HardLimit {
		t.Errorf("Expected hardLimit [%s] But was [%s]", hardLimit, *output.HardLimit)
	}
}

func convertToTaskDefinitionInTest(t *testing.T, name string, serviceConfig *config.ServiceConfig) *ecs.TaskDefinition {
	serviceConfigs := config.NewServiceConfigs()
	serviceConfigs.Add(name, serviceConfig)

	taskDefName := "ProjectName"
	envLookup, err := GetDefaultEnvironmentLookup()
	if err != nil {
		t.Fatal("Unexpected error setting up environment lookup")
	}
	resourceLookup, err := GetDefaultResourceLookup()
	if err != nil {
		t.Fatal("Unexpected error setting up resource lookup")
	}
	context := &project.Context{
		Project:           &project.Project{},
		EnvironmentLookup: envLookup,
		ResourceLookup:    resourceLookup,
	}
	taskDefinition, err := ConvertToTaskDefinition(taskDefName, context, serviceConfigs)
	if err != nil {
		t.Errorf("Expected to convert [%v] serviceConfigs without errors. But got [%v]", serviceConfig, err)
	}
	return taskDefinition
}

func TestIsZeroForEmptyConfig(t *testing.T) {
	serviceConfig := &config.ServiceConfig{}

	configValue := reflect.ValueOf(serviceConfig).Elem()
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
		"CPUShares":   true,
		"Command":     true,
		"Hostname":    true,
		"Image":       true,
		"Links":       true,
		"MemLimit":    true,
		"Privileged":  true,
		"ReadOnly":    true,
		"SecurityOpt": true,
		"User":        true,
		"WorkingDir":  true,
	}

	serviceConfig := &config.ServiceConfig{
		CPUShares:   int64(10),
		Command:     []string{"cmd"},
		Hostname:    "foobarbaz",
		Image:       "testimage",
		Links:       []string{"container1"},
		MemLimit:    int64(104857600),
		Privileged:  true,
		ReadOnly:    true,
		SecurityOpt: []string{"label:type:test_virt"},
		User:        "user",
		WorkingDir:  "/var",
	}

	configValue := reflect.ValueOf(serviceConfig).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		f := configValue.Field(i)
		ft := configType.Field(i)
		fieldName := ft.Name

		zeroValue := isZero(f)
		_, hasValue := hasValues[fieldName]
		if zeroValue == hasValue {
			t.Errorf("Expected field [%s]: hasValues[%t] but found[%t]", ft.Name, hasValues, !zeroValue)
		}
	}
}
