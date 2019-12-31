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

// Package converter translates entities to a docker compose schema and vice versa.
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
	log "github.com/sirupsen/logrus"
	arnParser "github.com/aws/aws-sdk-go/aws/arn"
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

// LocalCreateMetadata houses information about what kind of task definition
// was used to create the docker compose schema
type LocalCreateMetadata struct {
	InputType string
	Value     string
}

// CommonContainerValues contains values for top-level Task Definition fields
// that apply to all containers within the task definition
type CommonContainerValues struct {
	Ipc         string
	Pid         string
	Volumes     []*ecs.Volume
	TaskRoleArn string
}

const (
	// TaskDefinitionLabelType represents the type of option used to
	// transform a task definition to a compose file.
	// Valid types are "remote" (for registered Task Definitions retrieved
	// by arn or family name) or "local", for a local file containing a
	// task definition (default task-definition.json).
	TaskDefinitionLabelType = "ecs-local.task-definition-input.type"

	// TaskDefinitionLabelValue represents the value of the option.
	// For "local", the value should be the path of the task definition file.
	// For "remote", the value should be either the full arn or the family name.
	TaskDefinitionLabelValue = "ecs-local.task-definition-input.value"
)

// Environment variables used by the AWS SDK to communicate with the Endpoints container for credentials.
// See https://github.com/awslabs/amazon-ecs-local-container-endpoints/blob/master/docs/features.md#vend-credentials-to-containers
const (
	ecsCredsProviderEnvName = "AWS_CONTAINER_CREDENTIALS_RELATIVE_URI"
	endpointsTempCredsPath  = "/creds"
)

// Environment variables used by the AWS SDK to communicate with the Endpoints container for container metadata information.
// See https://github.com/awslabs/amazon-ecs-local-container-endpoints/blob/master/docs/features.md#task-metadata-v3
const (
	ecsMetadataURIEnvName  = "ECS_CONTAINER_METADATA_URI"
	endpointsMetadataV3URI = "http://169.254.170.2/v3"
)

// composeVersion is the minimum Compose file version supporting task definition fields.
const composeVersion = "3.4"

// SecretLabelPrefix is the prefix of Docker label keys
// whose value is an ARN of a secret to expose to the container.
// See https://github.com/aws/amazon-ecs-cli/issues/797
const SecretLabelPrefix = "ecs-local.secret"

// ConvertToComposeConfig translates an ECS Task Definition to a the Docker
// Compose config and returns it.
// NOTE: Top-level Volumes are not converted since these translate as named
// volumes in docker compose. Since ECS only supports bind mounts locally, each
// container would need to know the soucePath of the host machine, so the
// sourcePath of the top level volumes are resolved when converting mountPoints
// on a container definition.
func ConvertToComposeConfig(taskDefinition *ecs.TaskDefinition, metadata *LocalCreateMetadata) (*composeV3.Config, error) {
	services, err := createComposeServices(taskDefinition, metadata)
	if err != nil {
		return nil, err
	}

	networks := make(map[string]composeV3.NetworkConfig)
	networks[network.EcsLocalNetworkName] = composeV3.NetworkConfig{
		External: composeV3.External{
			External: true,
		},
	}

	return &composeV3.Config{
		Version:  composeVersion,
		Networks: networks,
		Services: services,
	}, nil
}

func createComposeServices(taskDefinition *ecs.TaskDefinition, metadata *LocalCreateMetadata) ([]composeV3.ServiceConfig, error) {
	networkMode := aws.StringValue(taskDefinition.NetworkMode)
	if networkMode != "" {
		log.WithFields(log.Fields{
			"networkMode": networkMode,
		}).Info("Task Definition network mode is ignored when running containers locally. Tasks will be run in the ecs-local-network.")
	}

	pid := aws.StringValue(taskDefinition.PidMode)
	ipc := aws.StringValue(taskDefinition.IpcMode)

	if pid != "" && pid != ecs.PidModeHost {
		log.WithFields(log.Fields{
			"pid": pid,
		}).Info("PID mode can only be set to 'host' when running tasks locally.")
		pid = "" // set to empty
	}

	if ipc != "" && ipc != ecs.IpcModeHost && ipc != ecs.IpcModeNone {
		log.WithFields(log.Fields{
			"ipc": ipc,
		}).Info("IPC mode can only be set to 'host' or 'none' when running tasks locally.")
		ipc = "" // set to empty
	}

	commonValues := &CommonContainerValues{
		Pid:         pid,
		Ipc:         ipc,
		Volumes:     taskDefinition.Volumes,
		TaskRoleArn: aws.StringValue(taskDefinition.TaskRoleArn),
	}

	if len(taskDefinition.ContainerDefinitions) < 1 {
		return nil, fmt.Errorf("task definition must include at least one container definition")
	}

	var services []composeV3.ServiceConfig

	for _, containerDefinition := range taskDefinition.ContainerDefinitions {
		service, err := convertToComposeService(containerDefinition, commonValues)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	// NOTE metadata should always be set on project when task definition is read
	if metadata == nil {
		return nil, fmt.Errorf("Unable to set service labels")
	}

	for _, service := range services {
		service.Labels[TaskDefinitionLabelType] = metadata.InputType
		service.Labels[TaskDefinitionLabelValue] = metadata.Value
	}

	return services, nil
}

func convertToComposeService(containerDefinition *ecs.ContainerDefinition, commonValues *CommonContainerValues) (composeV3.ServiceConfig, error) {
	linuxParams := convertLinuxParameters(containerDefinition.LinuxParameters)
	tmpfs := linuxParams.Tmpfs
	init := linuxParams.Init
	devices := linuxParams.Devices
	shmSize := linuxParams.ShmSize
	capAdd := linuxParams.CapAdd
	capDrop := linuxParams.CapDrop

	ulimits, _ := convertUlimits(containerDefinition.Ulimits)
	environment := convertEnvironment(containerDefinition, commonValues.TaskRoleArn)
	extraHosts := convertExtraHosts(containerDefinition.ExtraHosts)
	healthCheck := convertHealthCheck(containerDefinition.HealthCheck)
	labels := convertDockerLabelsWithSecrets(containerDefinition.DockerLabels, containerDefinition.Secrets)
	logging := convertLogging(containerDefinition.LogConfiguration)
	volumes := convertToVolumes(containerDefinition.MountPoints, commonValues.Volumes)
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
		Pid:         commonValues.Pid,
		Ipc:         commonValues.Ipc,
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
		// NOTE this uses docker compose's long syntax for ports, supported in v3.2+
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

// namedVolumesMap maps a named top-level volume to the host source path. Used
// to resolve container mount points to correct host for creating bind mounts
func namedVolumesMap(volumes []*ecs.Volume) map[string]string {
	hostPaths := make(map[string]string)
	for _, vol := range volumes {
		name := aws.StringValue(vol.Name)
		host := vol.Host
		// NOTE: Host *shouldn't* ever be nil, as ECS should return an
		// empty host as an empty object {}. In this case, the
		// sourcePath will be an empty string.
		if host != nil {
			sourcePath := aws.StringValue(host.SourcePath)
			hostPaths[name] = sourcePath
		}
	}

	return hostPaths
}

// Resolves any named volumes to a bind mount source path
func convertToVolumes(mountPoints []*ecs.MountPoint, volumes []*ecs.Volume) []composeV3.ServiceVolumeConfig {
	out := []composeV3.ServiceVolumeConfig{}
	mapping := namedVolumesMap(volumes)

	for _, mountPoint := range mountPoints {
		volumeName := aws.StringValue(mountPoint.SourceVolume)
		// We expect sourcePath to be set as an empty string to
		//allow Docker to assign a path on the host automatically if no
		//sourcePath is specified in the ECS Task Definition
		sourcePath := mapping[volumeName]

		volume := composeV3.ServiceVolumeConfig{
			Source:   sourcePath,
			Target:   aws.StringValue(mountPoint.ContainerPath),
			ReadOnly: aws.BoolValue(mountPoint.ReadOnly),
			Type:     "bind",
		}
		out = append(out, volume)
	}

	return out
}

func convertLogging(logConfig *ecs.LogConfiguration) *composeV3.LoggingConfig {
	if logConfig == nil {
		return nil
	}
	unsupportedDrivers := []string{ecs.LogDriverAwslogs}
	driver := aws.StringValue(logConfig.LogDriver)
	for _, unsupported := range unsupportedDrivers {
		if driver == unsupported {
			log.Warnf("%s log driver is ignored when running locally. Tasks will default to %s instead. This can be changed in your compose override file.", unsupported, jsonFileLogDriver)
		}
	}
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

func convertEnvironment(def *ecs.ContainerDefinition, taskRoleARN string) map[string]*string {
	out := make(map[string]*string)
	for _, kv := range def.Environment {
		name := aws.StringValue(kv.Name)
		out[name] = kv.Value
	}

	for _, secret := range def.Secrets {
		secretName := aws.StringValue(secret.Name)

		// We prefix the secret with the container name to disambiguate between
		// containers with the same secretName but different secretValue.
		shellEnv := fmt.Sprintf("${%s_%s}", *def.Name, secretName)
		out[secretName] = &shellEnv
	}

    credsName := endpointsTempCredsPath
    if parsedRoleARN, err := arnParser.Parse(taskRoleARN); taskRoleARN != "" && err == nil {
        credsName = "/" + parsedRoleARN.Resource
    }

	out[ecsCredsProviderEnvName] = aws.String(credsName)
	out[ecsMetadataURIEnvName] = aws.String(endpointsMetadataV3URI)
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

			return nil, errors.New("host path for a device must be specified")
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
			return nil, errors.New("path and size for tmpfs mounts must be specified")
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
			return "", fmt.Errorf("invalid device permission: %s", permission)
		}
		out += opt
	}
	return out, nil
}
