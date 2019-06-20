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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"

	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/docker/go-units"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// type Config struct {
// 	Filename string `yaml:"-"`
// 	Version  string
// 	Services Services
// 	Networks map[string]NetworkConfig   `yaml:",omitempty"`
// 	Volumes  map[string]VolumeConfig    `yaml:",omitempty"`
// 	Secrets  map[string]SecretConfig    `yaml:",omitempty"`
// 	Configs  map[string]ConfigObjConfig `yaml:",omitempty"`
// 	Extras   map[string]interface{}     `yaml:",inline"`
// }

// LinuxParams is a shim between members of ecs.LinuxParamters and their
// corresponding fields in the Docker Compose V3 ServiceConfig
type LinuxParams struct {
	CapAdd  []string
	CapDrop []string
	Devices []string
	Init    *bool
	Tmpfs   []string
	ShmSize string
}

// ConvertToDockerCompose creates the payload from an ECS Task Definition to be written as a docker compose file
func ConvertToDockerCompose(taskDefinition *ecs.TaskDefinition) ([]byte, error) {
	services := []composeV3.ServiceConfig{}
	for _, containerDefinition := range taskDefinition.ContainerDefinitions {
		service, err := convertToComposeService(containerDefinition)
		if err == nil {
			services = append(services, service)
		}
	}

	networks := make(map[string]composeV3.NetworkConfig)
	networks[network.EcsLocalNetworkName] = composeV3.NetworkConfig{
		External: composeV3.External{
			External: true,
		},
	}

	data, err := yaml.Marshal(&composeV3.Config{
		Filename: "docker-compose.local.yml",
		Version:  "3.0",
		Networks: networks,
		Services: services,
		// Volumes: taskDefinition.Volumes,
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

// TODO convert top level volumes
// TODO convert top level Neworks

func convertToComposeService(containerDefinition *ecs.ContainerDefinition) (composeV3.ServiceConfig, error) {
	linuxParams := convertLinuxParameters(containerDefinition.LinuxParameters)
	tmpfs := linuxParams.Tmpfs
	init := linuxParams.Init
	devices := linuxParams.Devices
	shmSize := linuxParams.ShmSize
	capAdd := linuxParams.CapAdd
	capDrop := linuxParams.CapDrop

	ulimits, _ := convertUlimits(containerDefinition.Ulimits)
	environment := convertEnvironment(containerDefinition.Environment, containerDefinition.Secrets)
	extraHosts := convertExtraHosts(containerDefinition.ExtraHosts)
	healthCheck := convertHealthCheck(containerDefinition.HealthCheck)
	labels := convertDockerLabels(containerDefinition.DockerLabels)
	logging := convertLogging(containerDefinition.LogConfiguration)
	volumes := convertToVolumes(containerDefinition.MountPoints)
	ports := convertToPorts(containerDefinition.PortMappings)
	sysctls := convertToSysctls(containerDefinition.SystemControls)
	networks := map[string]*composeV3.ServiceNetworkConfig{
		network.EcsLocalNetworkName: nil,
	}

	service := composeV3.ServiceConfig{
		Name:        aws.StringValue(containerDefinition.Name),
		Image:       aws.StringValue(containerDefinition.Image),
		DNS:         aws.StringValueSlice(containerDefinition.DnsServers),
		DNSSearch:   aws.StringValueSlice(containerDefinition.DnsSearchDomains),
		Command:     aws.StringValueSlice(containerDefinition.Command),
		Entrypoint:  aws.StringValueSlice(containerDefinition.EntryPoint),
		Links:       aws.StringValueSlice(containerDefinition.Links),
		Hostname:    aws.StringValue(containerDefinition.Hostname),
		SecurityOpt: aws.StringValueSlice(containerDefinition.DockerSecurityOptions),
		WorkingDir:  aws.StringValue(containerDefinition.WorkingDirectory),
		User:        aws.StringValue(containerDefinition.User),
		StdinOpen:   aws.BoolValue(containerDefinition.Interactive),
		Tty:         aws.BoolValue(containerDefinition.PseudoTerminal),
		Privileged:  aws.BoolValue(containerDefinition.Privileged),
		ReadOnly:    aws.BoolValue(containerDefinition.ReadonlyRootFilesystem),
		Ulimits:     ulimits,
		Tmpfs:       tmpfs,
		Init:        init,
		Devices:     devices,
		ShmSize:     shmSize,
		CapAdd:      capAdd,
		CapDrop:     capDrop,
		Environment: environment,
		ExtraHosts:  extraHosts,
		HealthCheck: healthCheck,
		Labels:      labels,
		Logging:     logging,
		Volumes:     volumes,
		Ports:       ports,
		Networks:    networks,
		Sysctls:     sysctls,
	}

	return service, nil
}

func convertToSysctls(systemControls []*ecs.SystemControl) []string {
	out := []string{}

	for _, sc := range systemControls {
		namespace := aws.StringValue(sc.Namespace)
		value := aws.StringValue(sc.Value)
		sysctl := fmt.Sprintf("%s=%s", namespace, value)
		out = append(out, sysctl)
	}

	return out
}

func convertToPorts(portMappings []*ecs.PortMapping) []composeV3.ServicePortConfig {
	out := []composeV3.ServicePortConfig{}

	for _, portMapping := range portMappings {
		port := composeV3.ServicePortConfig{
			Published: uint32(aws.Int64Value(portMapping.HostPort)),
			Target:    uint32(aws.Int64Value(portMapping.ContainerPort)),
			Protocol:  aws.StringValue(portMapping.Protocol),
			// Mode: "host"
		}
		out = append(out, port)
	}

	return out
}

func convertToVolumes(mountPoints []*ecs.MountPoint) []composeV3.ServiceVolumeConfig {
	out := []composeV3.ServiceVolumeConfig{}

	for _, mountPoint := range mountPoints {
		volume := composeV3.ServiceVolumeConfig{
			Source:   aws.StringValue(mountPoint.SourceVolume),
			Target:   aws.StringValue(mountPoint.ContainerPath),
			ReadOnly: aws.BoolValue(mountPoint.ReadOnly),
		}
		out = append(out, volume)
	}

	return out
}

func convertLogging(logConfig *ecs.LogConfiguration) *composeV3.LoggingConfig {
	if logConfig == nil {
		return nil
	}
	driver := aws.StringValue(logConfig.LogDriver)
	opts := make(map[string]string)
	for k, v := range logConfig.Options {
		opts[k] = aws.StringValue(v)
	}

	out := &composeV3.LoggingConfig{
		Driver:  driver,
		Options: opts,
	}
	return out
}

func convertDockerLabels(labels map[string]*string) composeV3.Labels {
	out := make(map[string]string)

	for k, v := range labels {
		out[k] = aws.StringValue(v)
	}

	return out
}

func convertHealthCheck(healthCheck *ecs.HealthCheck) *composeV3.HealthCheckConfig {
	if healthCheck == nil {
		return nil
	}
	command := aws.StringValueSlice(healthCheck.Command)

	out := &composeV3.HealthCheckConfig{
		Test: command,
	}
	if healthCheck.Interval != nil {
		interval := time.Duration(aws.Int64Value(healthCheck.Interval)) * time.Second
		out.Interval = &interval
	}
	if healthCheck.Timeout != nil {
		timeout := time.Duration(aws.Int64Value(healthCheck.Timeout)) * time.Second
		out.Timeout = &timeout
	}
	if healthCheck.Retries != nil {
		retries := uint64(aws.Int64Value(healthCheck.Retries))
		out.Retries = &retries
	}
	if healthCheck.StartPeriod != nil {
		startPeriod := time.Duration(aws.Int64Value(healthCheck.StartPeriod)) * time.Second
		out.StartPeriod = &startPeriod
	}

	return out
}

func convertExtraHosts(hosts []*ecs.HostEntry) []string {
	out := []string{}

	for _, hostEntry := range hosts {
		host := aws.StringValue(hostEntry.Hostname)
		ip := aws.StringValue(hostEntry.IpAddress)
		extraHost := strings.Join([]string{host, ip}, ":")
		out = append(out, extraHost)
	}

	return out
}

func convertEnvironment(env []*ecs.KeyValuePair, secrets []*ecs.Secret) map[string]*string {
	out := make(map[string]*string)
	for _, kv := range env {
		name := aws.StringValue(kv.Name)
		out[name] = kv.Value
	}

	for _, secret := range secrets {
		secretArn := aws.StringValue(secret.ValueFrom)
		secretVal, err := getContainerSecret(secretArn)
		if err != nil {
			logrus.Warnf("error retrieving value for secret: %s", secretArn)
		} else {
			name := aws.StringValue(secret.Name)
			out[name] = aws.String(secretVal)
		}
	}

	return out
}

// FIXME WIP
func getContainerSecret(secretArn string) (string, error) {
	arn, err := arn.Parse(secretArn)
	if err != nil {
		return "", err
	}

	switch service := arn.Service; service {
	case ssm.ServiceName:
		// call SSM
		return ssm.ServiceName, nil
	case secretsmanager.ServiceName:
		// call SecretsManager
		return secretsmanager.ServiceName, nil
	}
	return "", nil
}

func convertLinuxParameters(params *ecs.LinuxParameters) LinuxParams {
	if params == nil {
		return LinuxParams{}
	}

	init := params.InitProcessEnabled
	devices, _ := convertDevices(params.Devices)
	shmSize := convertShmSize(params.SharedMemorySize)
	tmpfs, _ := convertToTmpfs(params.Tmpfs)
	capAdd := convertCapAdd(params.Capabilities)
	capDrop := convertCapDrop(params.Capabilities)

	return LinuxParams{
		Tmpfs:   tmpfs,
		Init:    init,
		Devices: devices,
		ShmSize: shmSize,
		CapAdd:  capAdd,
		CapDrop: capDrop,
	}
}

func convertCapAdd(capabilities *ecs.KernelCapabilities) []string {
	if capabilities == nil {
		return nil
	}
	addCapabilities := capabilities.Add

	return aws.StringValueSlice(addCapabilities)
}

func convertCapDrop(capabilities *ecs.KernelCapabilities) []string {
	if capabilities == nil {
		return nil
	}
	dropCapabilities := capabilities.Drop

	return aws.StringValueSlice(dropCapabilities)
}

func convertShmSize(size *int64) string {
	if size == nil {
		return ""
	}
	sizeInMiB := aws.Int64Value(size) * units.MiB
	return units.BytesSize(float64(sizeInMiB))
}

// Note: This option is ignored when deploying a stack in swarm mode with a (version 3) Compose file.
func convertDevices(devices []*ecs.Device) ([]string, error) {
	out := []string{}

	for _, device := range devices {
		if device.HostPath == nil {
			return nil, errors.New("You must specify the host path for a device")
		}

		hostPath := aws.StringValue(device.HostPath)
		composeDevice := hostPath

		if device.ContainerPath != nil {
			containerPath := aws.StringValue(device.ContainerPath)
			composeDevice = strings.Join([]string{composeDevice, containerPath}, ":")
		}

		if device.Permissions != nil {
			permissions := aws.StringValueSlice(device.Permissions)
			composeOpts, err := convertDevicePermissions(permissions)
			if err != nil {
				return nil, err
			}
			composeDevice = strings.Join([]string{composeDevice, composeOpts}, ":")
		}

		out = append(out, composeDevice)
	}

	return out, nil
}

func convertToTmpfs(mounts []*ecs.Tmpfs) ([]string, error) {
	out := []string{}

	for _, mount := range mounts {

		if mount.ContainerPath == nil || mount.Size == nil {
			return nil, errors.New("You must specify the path and size for tmpfs mounts")
		}

		path := aws.StringValue(mount.ContainerPath)
		size := aws.Int64Value(mount.Size) * units.MiB

		composeSize := fmt.Sprintf("size=%s", units.BytesSize(float64(size)))

		tmpfs := strings.Join([]string{path, composeSize}, ":")

		if mount.MountOptions != nil {
			opts := aws.StringValueSlice(mount.MountOptions)
			composeOpts := strings.Join(opts, ",")
			tmpfs = strings.Join([]string{tmpfs, composeOpts}, ",")
		}

		out = append(out, tmpfs)
	}

	return out, nil
}

func convertUlimits(ulimits []*ecs.Ulimit) (map[string]*composeV3.UlimitsConfig, error) {
	out := make(map[string]*composeV3.UlimitsConfig)

	for _, ulimit := range ulimits {
		out[aws.StringValue(ulimit.Name)] = &composeV3.UlimitsConfig{
			Soft: int(aws.Int64Value(ulimit.SoftLimit)),
			Hard: int(aws.Int64Value(ulimit.HardLimit)),
		}
	}

	return out, nil
}

func convertDevicePermissions(permissions []string) (string, error) {
	devicePermissions := map[string]string{
		"read":  "r",
		"write": "w",
		"mknod": "m",
	}

	out := ""
	for _, permission := range permissions {
		opt, ok := devicePermissions[permission]
		if !ok {
			return "", fmt.Errorf("Invalid Device Permission: %s", permission)
		}
		out += opt
	}
	return out, nil
}
