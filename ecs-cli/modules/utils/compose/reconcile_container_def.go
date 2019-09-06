// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"errors"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	defaultMemLimit = 512
)

type taskLevelValues struct {
	MemLimit string
}

// reconcileContainerDef transforms each service in docker-compose.yml and
// ecs-params.yml to an equivalent ECS container definition
func reconcileContainerDef(inputCfg *adapter.ContainerConfig, ecsConDef *ContainerDef, taskVals taskLevelValues) (*ecs.ContainerDefinition, error) {
	outputContDef := &ecs.ContainerDefinition{}

	// Populate ECS container definition, offloading the validation to aws-sdk
	outputContDef.SetCommand(aws.StringSlice(inputCfg.Command))
	outputContDef.SetDnsSearchDomains(aws.StringSlice(inputCfg.DNSSearchDomains))
	outputContDef.SetDnsServers(aws.StringSlice(inputCfg.DNSServers))
	outputContDef.SetDockerLabels(inputCfg.DockerLabels)
	outputContDef.SetDockerSecurityOptions(aws.StringSlice(inputCfg.DockerSecurityOptions))
	outputContDef.SetEntryPoint(aws.StringSlice(inputCfg.Entrypoint))
	outputContDef.SetEnvironment(inputCfg.Environment)
	outputContDef.SetExtraHosts(inputCfg.ExtraHosts)
	if inputCfg.Hostname != "" {
		outputContDef.SetHostname(inputCfg.Hostname)
	}
	outputContDef.SetImage(inputCfg.Image)
	outputContDef.SetLinks(aws.StringSlice(inputCfg.Links)) // TODO, read from external links
	outputContDef.SetLogConfiguration(inputCfg.LogConfiguration)
	outputContDef.SetMountPoints(inputCfg.MountPoints)
	outputContDef.SetName(inputCfg.Name)
	outputContDef.SetPrivileged(inputCfg.Privileged)
	outputContDef.SetPortMappings(inputCfg.PortMappings)
	outputContDef.SetPseudoTerminal(inputCfg.PseudoTerminal)
	outputContDef.SetReadonlyRootFilesystem(inputCfg.ReadOnly)
	outputContDef.SetUlimits(inputCfg.Ulimits)

	if inputCfg.User != "" {
		outputContDef.SetUser(inputCfg.User)
	}
	outputContDef.SetVolumesFrom(inputCfg.VolumesFrom)
	if inputCfg.WorkingDirectory != "" {
		outputContDef.SetWorkingDirectory(inputCfg.WorkingDirectory)
	}

	// Set Linux Parameters
	outputContDef.LinuxParameters = &ecs.LinuxParameters{Capabilities: &ecs.KernelCapabilities{}}
	if inputCfg.CapAdd != nil {
		outputContDef.LinuxParameters.Capabilities.SetAdd(aws.StringSlice(inputCfg.CapAdd))
	}
	if inputCfg.CapDrop != nil {
		outputContDef.LinuxParameters.Capabilities.SetDrop(aws.StringSlice(inputCfg.CapDrop))
	}
	if inputCfg.Devices != nil {
		outputContDef.LinuxParameters.SetDevices(inputCfg.Devices)
	}
	if ecsConDef.InitProcessEnabled != false {
		outputContDef.LinuxParameters.SetInitProcessEnabled(ecsConDef.InitProcessEnabled)
	}

	// Only set shmSize if specified. Docker will by default allocate 64M
	// for shared memory if shmSize is null.
	if inputCfg.ShmSize != 0 {
		outputContDef.LinuxParameters.SetSharedMemorySize(inputCfg.ShmSize)
	}

	if inputCfg.StopTimeout != nil {
		outputContDef.SetStopTimeout(*inputCfg.StopTimeout)
	}

	// Only set tmpfs if tmpfs mounts are specified.
	if inputCfg.Tmpfs != nil { // TODO: will never be nil?
		outputContDef.LinuxParameters.SetTmpfs(inputCfg.Tmpfs)
	}

	// initialize container resources from inputCfg
	cpu := inputCfg.CPU
	memLimit := inputCfg.Memory
	memRes := inputCfg.MemoryReservation
	healthCheck := inputCfg.HealthCheck
	var resourceRequirements []*ecs.ResourceRequirement

	if ecsConDef != nil {
		outputContDef.Essential = aws.Bool(ecsConDef.Essential)

		// CPU and Memory are expected to be set here if compose v3 was used
		cpu = resolveIntResourceOverride(inputCfg.Name, cpu, ecsConDef.Cpu, "CPU")

		ecsMemInMB := adapter.ConvertToMemoryInMB(int64(ecsConDef.Memory))
		memLimit = resolveIntResourceOverride(inputCfg.Name, memLimit, ecsMemInMB, "MemoryLimit")

		ecsMemResInMB := adapter.ConvertToMemoryInMB(int64(ecsConDef.MemoryReservation))

		memRes = resolveIntResourceOverride(inputCfg.Name, memRes, ecsMemResInMB, "MemoryReservation")

		credParam := ecsConDef.RepositoryCredentials.CredentialsParameter

		if ecsConDef.FirelensConfiguration.Type != "" {
			outputContDef.SetFirelensConfiguration(convertToECSFirelensConfiguration(ecsConDef.FirelensConfiguration))
		}

		if credParam != "" {
			outputContDef.RepositoryCredentials = &ecs.RepositoryCredentials{}
			outputContDef.RepositoryCredentials.SetCredentialsParameter(credParam)
		}

		if len(ecsConDef.Secrets) > 0 {
			outputContDef.SetSecrets(convertToECSSecrets(ecsConDef.Secrets))
		}

		if len(ecsConDef.Logging.SecretOptions) > 0 {
			convertedSecrets := convertToECSSecrets(ecsConDef.Logging.SecretOptions)
			outputContDef.LogConfiguration.SetSecretOptions(convertedSecrets)
		}

		var err error
		healthCheck, err = resolveHealthCheck(inputCfg.Name, healthCheck, ecsConDef.HealthCheck)
		if err != nil {
			return nil, err
		}

		if ecsConDef.GPU != "" {
			resourceType := ecs.ResourceTypeGpu
			resourceRequirement := ecs.ResourceRequirement{
				Type:  &resourceType,
				Value: &ecsConDef.GPU,
			}
			resourceRequirements = append(resourceRequirements, &resourceRequirement)
		}
	}

	// At least one memory value is required to register a task definition.
	// If no memory value is set, set default limit
	if memLimit == 0 && memRes == 0 && taskVals.MemLimit == "" {
		memLimit = defaultMemLimit
	}

	// if memLimit is set and less than memRes, show error
	if memLimit != 0 && memLimit < memRes {
		return nil, errors.New("mem_limit must be greater than mem_reservation")
	}

	if memLimit != 0 {
		outputContDef.SetMemory(memLimit)
	}

	if memRes != 0 {
		outputContDef.SetMemoryReservation(memRes)
	}

	outputContDef.SetCpu(cpu)

	if healthCheck != nil {
		outputContDef.SetHealthCheck(healthCheck)
	}

	if len(resourceRequirements) > 0 {
		outputContDef.SetResourceRequirements(resourceRequirements)
	}

	return outputContDef, nil
}
