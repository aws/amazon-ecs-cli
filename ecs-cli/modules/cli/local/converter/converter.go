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
	"fmt"
	"strings"
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/docker/go-units"
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

	data, err := yaml.Marshal(&composeV3.Config{
		Filename: "docker-compose.local.yml",
		Version: "3.0",
		Services: services,
		// Volumes: taskDefinition.Volumes,
	})

	if err != nil {
		return nil, err
	}

	return data, nil
}

func convertToComposeService(containerDefinition *ecs.ContainerDefinition) (composeV3.ServiceConfig, error) {
	linuxParams := convertLinuxParameters(containerDefinition.LinuxParameters)
	tmpfs := linuxParams.Tmpfs
	init := linuxParams.Init
	devices := linuxParams.Devices
	shmSize := linuxParams.ShmSize
	capAdd := linuxParams.CapAdd
	capDrop := linuxParams.CapDrop

	ulimits, _ := convertUlimits(containerDefinition.Ulimits)
	environment := convertEnvironment(containerDefinition.Environment)
	extraHosts := convertExtraHosts(containerDefinition.ExtraHosts)

	service := composeV3.ServiceConfig{
		Name: aws.StringValue(containerDefinition.Name),
		Image: aws.StringValue(containerDefinition.Image),
		DNS: aws.StringValueSlice(containerDefinition.DnsServers),
		DNSSearch: aws.StringValueSlice(containerDefinition.DnsSearchDomains),
		Command: aws.StringValueSlice(containerDefinition.Command),
		Entrypoint: aws.StringValueSlice(containerDefinition.EntryPoint),
		Links: aws.StringValueSlice(containerDefinition.Links),
		Hostname: aws.StringValue(containerDefinition.Hostname),
		SecurityOpt: aws.StringValueSlice(containerDefinition.DockerSecurityOptions),
		WorkingDir: aws.StringValue(containerDefinition.WorkingDirectory),
		User: aws.StringValue(containerDefinition.User),
		Tty: aws.BoolValue(containerDefinition.PseudoTerminal),
		Privileged: aws.BoolValue(containerDefinition.Privileged),
		ReadOnly: aws.BoolValue(containerDefinition.ReadonlyRootFilesystem),
		Ulimits: ulimits,
		Tmpfs: tmpfs,
		Init: init,
		Devices: devices,
		ShmSize: shmSize,
		CapAdd: capAdd,
		CapDrop: capDrop,
		Environment: environment,
		ExtraHosts: extraHosts,

		// HealthCheck     *HealthCheckConfig               `yaml:",omitempty"`
		// Labels          Labels                           `yaml:",omitempty"`
		// Logging         *LoggingConfig                   `yaml:",omitempty"`
		// Volumes         []ServiceVolumeConfig            `yaml:",omitempty"`
	}


	// fmt.Printf("\nCOMPOSE SERVICE: %+v\n\n", service)
	return service, nil
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

func convertEnvironment(env []*ecs.KeyValuePair) map[string]*string {
	out := make(map[string]*string)
	for _, kv := range env {
		out[aws.StringValue(kv.Name)] = kv.Value
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

	return LinuxParams {
		Tmpfs: tmpfs,
		Init: init,
		Devices: devices,
		ShmSize: shmSize,
		CapAdd: capAdd,
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
		"read": "r",
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
