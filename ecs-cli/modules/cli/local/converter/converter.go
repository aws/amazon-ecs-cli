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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	composeV3 "github.com/docker/cli/cli/compose/types"
	"gopkg.in/yaml.v2"
)

// TODO wrap compose type?
// type Compose struct {
// }

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
	service := composeV3.ServiceConfig{
		Name: aws.StringValue(containerDefinition.Name),
		Image: aws.StringValue(containerDefinition.Image),
		// Devices: aws.StringValueSlice(containerDefinition.Devices)
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

		// CapAdd: containerDefinition.LinuxParameters.KernalCapabilities.Add
		// CapDrop: containerDefinition.LinuxParameters.KernalCapabilities.Drop
		// =======

		// CapAdd          []string                         `mapstructure:"cap_add" yaml:"cap_add,omitempty"`
		// CapDrop         []string                         `mapstructure:"cap_drop" yaml:"cap_drop,omitempty"`

		// Devices         []string                         `yaml:",omitempty"`
		// Environment     MappingWithEquals                `yaml:",omitempty"`
		// EnvFile         StringList                       `mapstructure:"env_file" yaml:"env_file,omitempty"`
		// ExtraHosts      HostsList                        `mapstructure:"extra_hosts" yaml:"extra_hosts,omitempty"`
		// HealthCheck     *HealthCheckConfig               `yaml:",omitempty"`
		// Labels          Labels                           `yaml:",omitempty"`
		// Links           []string                         `yaml:",omitempty"`
		// Logging         *LoggingConfig                   `yaml:",omitempty"`
		// Ports           []ServicePortConfig              `yaml:",omitempty"`
		// ReadOnly        bool                             `mapstructure:"read_only" yaml:"read_only,omitempty"`
		// Tmpfs           StringList                       `yaml:",omitempty"`
		// Ulimits         map[string]*UlimitsConfig        `yaml:",omitempty"`
		// Volumes         []ServiceVolumeConfig            `yaml:",omitempty"`
	}


	return service, nil
}
