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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMemLimit = 512
)

// TaskDefParams contains basic fields to build an
// ECS task definition
type TaskDefParams struct {
	networkMode      string
	taskRoleArn      string
	cpu              string
	memory           string
	containerDefs    ContainerDefs
	executionRoleArn string
}

// ConvertToTaskDefinition transforms the yaml configs to its ecs equivalent (task definition)
func ConvertToTaskDefinition(taskDefinitionName string, volumes *adapter.Volumes,
	containerConfigs []adapter.ContainerConfig, taskRoleArn string, requiredCompatibilites string, ecsParams *ECSParams) (*ecs.TaskDefinition, error) {
	if len(containerConfigs) == 0 {
		return nil, errors.New("cannot create a task definition with no containers; invalid service config")
	}

	// Instantiates zero values for fields on task def specified by ecs-params
	taskDefParams, err := convertTaskDefParams(ecsParams)
	if err != nil {
		return nil, err
	}

	// The task-role-arn flag takes precedence over a taskRoleArn value specified in ecs-params file.
	if taskRoleArn == "" {
		taskRoleArn = taskDefParams.taskRoleArn
	}

	// Create containerDefinitions
	containerDefinitions := []*ecs.ContainerDefinition{}

	for _, containerConfig := range containerConfigs {
		// logUnsupportedServiceConfigFields(name, serviceConfig) // TODO switch this to use ContainerConfig
		name := containerConfig.Name
		// Check if there are ecs-params specified for the container
		ecsContainerDef := &ContainerDef{Essential: true}
		if cd, ok := taskDefParams.containerDefs[name]; ok {
			ecsContainerDef = &cd
		}

		// Validate essential containers TODO: merge other ecs params fields
		count := len(containerConfigs)
		if !hasEssential(taskDefParams.containerDefs, count) {
			return nil, errors.New("Task definition does not have any essential containers")
		}

		containerDef, err := convertToContainerDef(&containerConfig, ecsContainerDef)
		if err != nil {
			return nil, err
		}

		containerDefinitions = append(containerDefinitions, containerDef)
	}

	taskDefinition := &ecs.TaskDefinition{
		Family:               aws.String(taskDefinitionName),
		ContainerDefinitions: containerDefinitions,
		Volumes:              convertToECSVolumes(volumes),
		TaskRoleArn:          aws.String(taskRoleArn),
		NetworkMode:          aws.String(taskDefParams.networkMode),
		Cpu:                  aws.String(taskDefParams.cpu),
		Memory:               aws.String(taskDefParams.memory),
		ExecutionRoleArn:     aws.String(taskDefParams.executionRoleArn),
	}

	// Set launch type
	if requiredCompatibilites != "" {
		taskDefinition.RequiresCompatibilities = []*string{aws.String(requiredCompatibilites)}
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

	// Set essential & resource fields from ecs-params file, if present
	if ecsContainerDef != nil {
		outputContDef.Essential = aws.Bool(ecsContainerDef.Essential)

		cpu = getResourceValue(inputCfg.Name, cpu, ecsContainerDef.Cpu)

		ecsMemInMB := adapter.ConvertToMemoryInMB(int64(ecsContainerDef.Memory))
		mem = getResourceValue(inputCfg.Name, mem, ecsMemInMB)

		ecsMemResInMB := adapter.ConvertToMemoryInMB(int64(ecsContainerDef.MemoryReservation))
		memRes = getResourceValue(inputCfg.Name, memRes, ecsMemResInMB)
	}

	if mem < memRes {
		return nil, errors.New("mem_limit must be greater than mem_reservation")
	}

	// One or both of memory and memoryReservation is required to register a task definition with ECS
	// If neither is provided by 1) compose file or 2) ecs-params, set default
	// NOTE: Docker does not set a memory limit for containers
	if mem == 0 && memRes == 0 {
		mem = defaultMemLimit
	}

	outputContDef.SetCpu(cpu)
	if mem != 0 {
		outputContDef.SetMemory(mem)
	}
	if memRes != 0 {
		outputContDef.SetMemoryReservation(memRes)
	}

	return outputContDef, nil
}

func getResourceValue(serviceName string, inputVal, ecsVal int64) int64 {
	if ecsVal == 0 {
		return inputVal
	}
	if inputVal > 0 {
		showResourceOverrideMsg(serviceName, inputVal, ecsVal)
	}
	return ecsVal
}

func showResourceOverrideMsg(serviceName string, oldValue int64, newValue int64) {
	overrideMsg := "Using ecs-params value as override (was %v but is now %v)"

	log.WithFields(log.Fields{
		"option name":  "MemoryReservation",
		"service name": serviceName,
	}).Infof(overrideMsg, oldValue, newValue)
}

// convertToECSVolumes transforms the map of hostPaths to the format of ecs.Volume
func convertToECSVolumes(hostPaths *adapter.Volumes) []*ecs.Volume {
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
	// volumes with an empty host path
	for _, volName := range hostPaths.VolumeEmptyHost {
		ecsVolume := &ecs.Volume{
			Name: aws.String(volName),
		}
		output = append(output, ecsVolume)
	}
	return output
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
