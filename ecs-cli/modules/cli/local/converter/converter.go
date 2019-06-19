// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Package converter converts an ecs.TaskDefinition or a yaml file to a docker compose schema.
package converter

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/docker/go-units"
	yaml "gopkg.in/yaml.v2"
)

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

// SecretLabelPrefix is the prefix of Docker label keys
// whose value is an ARN of a secret to expose to the container.
// See https://github.com/aws/amazon-ecs-cli/issues/797
const SecretLabelPrefix = "ecs-local.secret"

const (
	// taskDefinitionLabelType represents the type of option used to
	// transform a task definition to a compose file e.g. remoteFile, localFile.
	// taskDefinitionLabelValue represents the value of the option
	// e.g. file path, arn, family.
	taskDefinitionLabelType  = "ecs-local.task.type"
	taskDefinitionLabelValue = "ecs-local.task.value"
)

// ConvertToDockerCompose creates the payload from an ECS Task Definition to be written as a docker compose file
func ConvertToDockerCompose(taskDefinition *ecs.TaskDefinition, localTaskType, localTaskValue string) ([]byte, error) {
	var services []composeV3.ServiceConfig
	for _, containerDefinition := range taskDefinition.ContainerDefinitions {
		service, err := convertToComposeService(containerDefinition)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	for _, service := range services {
		service.Labels[taskDefinitionLabelType] = localTaskType
		service.Labels[taskDefinitionLabelValue] = localTaskValue
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
	environment := convertEnvironment(containerDefinition)
	extraHosts := convertExtraHosts(containerDefinition.ExtraHosts)
	healthCheck := convertHealthCheck(containerDefinition.HealthCheck)
	labels := convertDockerLabelsWithSecrets(containerDefinition.DockerLabels, containerDefinition.Secrets)
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

func convertDockerLabelsWithSecrets(labels map[string]*string, secrets []*ecs.Secret) composeV3.Labels {
	out := make(map[string]string)

	for k, v := range labels {
		out[k] = aws.StringValue(v)
	}

	for _, secret := range secrets {
		name := aws.StringValue(secret.Name)
		key := fmt.Sprintf("%s.%s", SecretLabelPrefix, name)
		out[key] = aws.StringValue(secret.ValueFrom)
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

func convertEnvironment(def *ecs.ContainerDefinition) map[string]*string {
	out := make(map[string]*string)
	for _, kv := range def.Environment {
		name := aws.StringValue(kv.Name)
		out[name] = kv.Value
	}

	for _, secret := range def.Secrets {
		secretName := aws.StringValue(secret.Name)

		// We prefix the secret with the container name to disambiguate between
		// containers with the same secretName but different secretValue
		shellEnv := fmt.Sprintf("${%s_%s}", *def.Name, secretName)
		out[secretName] = &shellEnv
	}
	return out
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
