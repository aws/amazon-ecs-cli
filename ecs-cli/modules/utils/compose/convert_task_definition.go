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

package utils

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMemLimit = 512
)

// TaskDefParams contains basic fields to build an ECS task definition
type TaskDefParams struct {
	networkMode      string
	taskRoleArn      string
	cpu              string
	memory           string
	containerDefs    ContainerDefs
	executionRoleArn string
}

// ConvertTaskDefParams contains the inputs required to convert compose & ECS inputs into an ECS task definition
type ConvertTaskDefParams struct {
	TaskDefName            string
	TaskRoleArn            string
	RequiredCompatibilites string
	Volumes                *adapter.Volumes
	ContainerConfigs       []adapter.ContainerConfig
	ECSParams              *ECSParams
}

// ConvertToTaskDefinition transforms the yaml configs to its ecs equivalent (task definition)
func ConvertToTaskDefinition(params ConvertTaskDefParams) (*ecs.TaskDefinition, error) {
	if len(params.ContainerConfigs) == 0 {
		return nil, errors.New("cannot create a task definition with no containers; invalid service config")
	}

	// Instantiates zero values for fields on task def specified by ecs-params
	taskDefParams, err := convertTaskDefParams(params.ECSParams)
	if err != nil {
		return nil, err
	}

	// The task-role-arn flag takes precedence over a taskRoleArn value specified in ecs-params file.
	if params.TaskRoleArn == "" {
		params.TaskRoleArn = taskDefParams.taskRoleArn
	}

	// Create containerDefinitions
	containerDefinitions := []*ecs.ContainerDefinition{}

	for _, containerConfig := range params.ContainerConfigs {
		name := containerConfig.Name
		// Check if there are ecs-params specified for the container
		ecsContainerDef := &ContainerDef{Essential: true}
		if cd, ok := taskDefParams.containerDefs[name]; ok {
			ecsContainerDef = &cd
		}

		// Validate essential containers
		count := len(params.ContainerConfigs)
		if !hasEssential(taskDefParams.containerDefs, count) {
			return nil, errors.New("Task definition does not have any essential containers")
		}

		containerDef, err := convertToContainerDef(&containerConfig, ecsContainerDef)
		if err != nil {
			return nil, err
		}

		containerDefinitions = append(containerDefinitions, containerDef)
	}

	ecsVolumes, err := convertToECSVolumes(params.Volumes, params.ECSParams)
	if err != nil {
		return nil, err
	}

	taskDefinition := &ecs.TaskDefinition{
		Family:               aws.String(params.TaskDefName),
		ContainerDefinitions: containerDefinitions,
		Volumes:              ecsVolumes,
		TaskRoleArn:          aws.String(params.TaskRoleArn),
		NetworkMode:          aws.String(taskDefParams.networkMode),
		Cpu:                  aws.String(taskDefParams.cpu),
		Memory:               aws.String(taskDefParams.memory),
		ExecutionRoleArn:     aws.String(taskDefParams.executionRoleArn),
	}

	// Set launch type
	if params.RequiredCompatibilites != "" {
		taskDefinition.RequiresCompatibilities = []*string{aws.String(params.RequiredCompatibilites)}
	}

	return taskDefinition, nil
}

// convertToContainerDef transforms each service in docker-compose.yml and
// ecs-params.yml to an equivalent ECS container definition
func convertToContainerDef(inputCfg *adapter.ContainerConfig, ecsContainerDef *ContainerDef) (*ecs.ContainerDefinition, error) {
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

	// Only set shmSize if specified. Otherwise we expect this sharedMemorySize for the
	// containerDefinition to be null; Docker will by default allocate 64M for shared memory if
	// shmSize is null.
	if inputCfg.ShmSize != 0 {
		outputContDef.LinuxParameters.SetSharedMemorySize(inputCfg.ShmSize)
	}

	// Only set tmpfs if tmpfs mounts are specified.
	if inputCfg.Tmpfs != nil { // will never be nil?
		outputContDef.LinuxParameters.SetTmpfs(inputCfg.Tmpfs)
	}

	// initialize container resources from inputCfg
	cpu := inputCfg.CPU
	mem := inputCfg.Memory
	memRes := inputCfg.MemoryReservation
	healthCheck := inputCfg.HealthCheck

	// Set essential & resource fields from ecs-params file, if present
	if ecsContainerDef != nil {
		outputContDef.Essential = aws.Bool(ecsContainerDef.Essential)

		// CPU and Memory are expected to be set here if compose v3 was used
		cpu = resolveIntResourceOverride(inputCfg.Name, cpu, ecsContainerDef.Cpu, "CPU")

		ecsMemInMB := adapter.ConvertToMemoryInMB(int64(ecsContainerDef.Memory))
		mem = resolveIntResourceOverride(inputCfg.Name, mem, ecsMemInMB, "MemoryLimit")

		ecsMemResInMB := adapter.ConvertToMemoryInMB(int64(ecsContainerDef.MemoryReservation))

		memRes = resolveIntResourceOverride(inputCfg.Name, memRes, ecsMemResInMB, "MemoryReservation")

		credParam := ecsContainerDef.RepositoryCredentials.CredentialsParameter

		if credParam != "" {
			outputContDef.RepositoryCredentials = &ecs.RepositoryCredentials{}
			outputContDef.RepositoryCredentials.SetCredentialsParameter(credParam)
		}

		var err error
		healthCheck, err = resolveHealthCheck(inputCfg.Name, healthCheck, ecsContainerDef.HealthCheck)
		if err != nil {
			return nil, err
		}
	}

	// One or both of memory and memoryReservation is required to register a task definition with ECS
	// If neither is provided by 1) compose file or 2) ecs-params, set default
	// NOTE: Docker does not set a memory limit for containers
	if mem == 0 && memRes == 0 {
		mem = defaultMemLimit
	}

	// Docker compose allows specifying memory reservation with memory, so
	// we should default to minimum allowable value for memory hard limit
	if mem == 0 && memRes != 0 {
		mem = memRes
	}

	if mem < memRes {
		return nil, errors.New("mem_limit must be greater than mem_reservation")
	}

	outputContDef.SetCpu(cpu)
	if mem != 0 {
		outputContDef.SetMemory(mem)
	}
	if memRes != 0 {
		outputContDef.SetMemoryReservation(memRes)
	}

	if healthCheck != nil {
		outputContDef.SetHealthCheck(healthCheck)
	}

	return outputContDef, nil
}

func resolveHealthCheck(serviceName string, healthCheck *ecs.HealthCheck, ecsParamsHealthCheck *HealthCheck) (*ecs.HealthCheck, error) {
	if ecsParamsHealthCheck != nil {
		healthCheckOverride, err := ecsParamsHealthCheck.ConvertToECSHealthCheck()
		if err != nil {
			return nil, err
		}

		if healthCheck != nil {
			healthCheck.Command = resolveStringSliceResourceOverride(serviceName, healthCheck.Command, healthCheckOverride.Command, "healthcheck command")
			healthCheck.Interval = resolveIntPointerResourceOverride(serviceName, healthCheck.Interval, healthCheckOverride.Interval, "healthcheck interval")
			healthCheck.Retries = resolveIntPointerResourceOverride(serviceName, healthCheck.Retries, healthCheckOverride.Retries, "healthcheck retries")
			healthCheck.Timeout = resolveIntPointerResourceOverride(serviceName, healthCheck.Timeout, healthCheckOverride.Timeout, "healthcheck timeout")
			healthCheck.StartPeriod = resolveIntPointerResourceOverride(serviceName, healthCheck.StartPeriod, healthCheckOverride.StartPeriod, "healthcheck start_period")
		} else {
			healthCheck = healthCheckOverride
		}
	}
	// validate healthcheck
	if healthCheck != nil && healthCheck.Validate() != nil {
		return healthCheck, fmt.Errorf("%s: test/command is a required field for container healthcheck", serviceName)
	}
	return healthCheck, nil
}

func resolveIntResourceOverride(serviceName string, composeVal, ecsParamsVal int64, option string) int64 {
	if composeVal > 0 && ecsParamsVal > 0 {
		showResourceOverrideMsg(serviceName, composeVal, ecsParamsVal, option)
	}
	if ecsParamsVal > 0 {
		return ecsParamsVal
	}
	return composeVal
}

func resolveIntPointerResourceOverride(serviceName string, composeVal, ecsParamsVal *int64, option string) *int64 {
	if composeVal != nil && ecsParamsVal != nil {
		showResourceOverrideMsg(serviceName, aws.Int64Value(composeVal), aws.Int64Value(ecsParamsVal), option)
	}
	if ecsParamsVal != nil {
		return ecsParamsVal
	}
	return composeVal
}

func resolveStringSliceResourceOverride(serviceName string, composeVal, ecsParamsVal []*string, option string) []*string {
	if len(composeVal) > 0 && len(ecsParamsVal) > 0 {
		log.WithFields(log.Fields{
			"option name":  option,
			"service name": serviceName,
		}).Infof("Using ecs-params value as override")
	}
	if len(ecsParamsVal) > 0 {
		return ecsParamsVal
	}
	return composeVal
}

func showResourceOverrideMsg(serviceName string, val int64, override int64, option string) {
	overrideMsg := "Using ecs-params value as override (was %v but is now %v)"

	log.WithFields(log.Fields{
		"option name":  option,
		"service name": serviceName,
	}).Infof(overrideMsg, val, override)
}

// convertToECSVolumes transforms the map of hostPaths to the format of ecs.Volume
func convertToECSVolumes(hostPaths *adapter.Volumes, ecsParams *ECSParams) ([]*ecs.Volume, error) {
	output := []*ecs.Volume{}
	// volumes with a host path
	for hostPath, volName := range hostPaths.VolumeWithHost {
		ecsVolume := &ecs.Volume{
			Name: aws.String(volName),
			Host: &ecs.HostVolumeProperties{
				SourcePath: aws.String(hostPath),
			}}
		output = append(output, ecsVolume)
	}

	// volumes without host path (allowed to have Docker Volume Configuration)
	volumesWithoutHost, err := mergeVolumesWithoutHost(hostPaths.VolumeEmptyHost, ecsParams)
	if err != nil {
		return nil, err
	}
	output = append(output, volumesWithoutHost...)
	return output, nil
}

func mergeVolumesWithoutHost(composeVolumes []string, ecsParams *ECSParams) ([]*ecs.Volume, error) {
	volumesWithoutHost := make(map[string]DockerVolume)
	output := []*ecs.Volume{}

	for _, volName := range composeVolumes {
		volumesWithoutHost[volName] = DockerVolume{}
	}

	if ecsParams != nil {
		for _, dockerVol := range ecsParams.TaskDefinition.DockerVolumes {
			if dockerVol.Name != "" {
				volumesWithoutHost[dockerVol.Name] = dockerVol
			} else {
				return nil, fmt.Errorf("Name is required when specifying a docker volume")
			}
		}
	}

	for volName, dVol := range volumesWithoutHost {
		ecsVolume := &ecs.Volume{
			Name: aws.String(volName),
		}
		if dVol.Name != "" {
			ecsVolume.DockerVolumeConfiguration = &ecs.DockerVolumeConfiguration{
				Autoprovision: dVol.Autoprovision,
				Driver:        aws.String(dVol.Driver),
				Scope:         aws.String(dVol.Scope),
			}
			if dVol.DriverOptions != nil {
				ecsVolume.DockerVolumeConfiguration.DriverOpts = aws.StringMap(dVol.DriverOptions)
			}
			if dVol.Labels != nil {
				ecsVolume.DockerVolumeConfiguration.Labels = aws.StringMap(dVol.Labels)
			}
		}
		output = append(output, ecsVolume)
	}
	return output, nil
}

func hasEssential(ecsParamsContainerDefs ContainerDefs, count int) bool {
	// If the customer does not set the "essential" field on any container
	// definition, ECS will mark all containers in a TaskDefinition as
	// essential. Previously, since the customer could not pass in the
	// essential field, Task Definitions created through the CLI marked all
	// containers as essential.

	// Now that the essential field can  be set by the customer via the
	// the ecs-params.yml file, we want to make sure that there is still at
	// least one essential container, i.e. that the customer does not
	// explicitly set all containers to be non-essential.

	nonEssentialCount := 0

	for _, containerDef := range ecsParamsContainerDefs {
		if !containerDef.Essential {
			nonEssentialCount++
		}
	}

	// 'count' is the total number of containers specified in the service config
	return nonEssentialCount != count
}

// Converts fields from ecsParams into the appropriate types for fields on an
// ECS Task Definition
func convertTaskDefParams(ecsParams *ECSParams) (params TaskDefParams, e error) {
	if ecsParams == nil {
		return params, nil
	}

	taskDef := ecsParams.TaskDefinition
	params.networkMode = taskDef.NetworkMode
	params.taskRoleArn = taskDef.TaskRoleArn
	params.containerDefs = taskDef.ContainerDefinitions
	params.cpu = taskDef.TaskSize.Cpu
	params.memory = taskDef.TaskSize.Memory
	params.executionRoleArn = taskDef.ExecutionRole

	return params, nil
}
