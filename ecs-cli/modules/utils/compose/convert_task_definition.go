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
	"reflect"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMemLimit = 512
)

// supported fields/options from compose YAML file
var supportedComposeYamlOptions = []string{
	"cap_add",
	"cap_drop",
	"command",
	"cpu_shares",
	"dns",
	"dns_search",
	"entrypoint",
	"env_file",
	"environment",
	"extra_hosts",
	"hostname",
	"image",
	"labels",
	"links",
	"logging",
	"log_driver", // v1 only
	"log_opt",    // v1 only
	"mem_limit",
	"mem_reservation",
	"ports",
	"privileged",
	"read_only",
	"security_opt",
	"shm_size",
	"tmpfs",
	"ulimits",
	"user",
	"volumes", // v2
	"volumes_from",
	"working_dir",
}

var supportedComposeYamlOptionsMap = getSupportedComposeYamlOptionsMap()

func getSupportedComposeYamlOptionsMap() map[string]bool {
	optionsMap := make(map[string]bool)
	for _, value := range supportedComposeYamlOptions {
		optionsMap[value] = true
	}
	return optionsMap
}

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
// TODO container config a pointer to slice?
func ConvertToTaskDefinition(context *project.Context, volumeConfigs map[string]*config.VolumeConfig,
	containerConfigs []adapter.ContainerConfig, taskRoleArn string, requiredCompatibilites string, ecsParams *ECSParams) (*ecs.TaskDefinition, error) {
	if len(containerConfigs) == 0 {
		return nil, errors.New("cannot create a task definition with no containers; invalid service config")
	}

	logUnsupportedConfigFields(context.Project) // TODO refactor? networks only thing not supproted

	taskDefinitionName := context.ProjectName

	// Instantiates zero values for fields on task def specified by ecs-params
	taskDefParams, err := convertTaskDefParams(ecsParams)
	if err != nil {
		return nil, err
	}

	// The task-role-arn flag takes precedence over a taskRoleArn value specified in ecs-params file.
	if taskRoleArn == "" {
		taskRoleArn = taskDefParams.taskRoleArn
	}

	// TODO: Refactor when Volumes added to top level project
	volumes, err := adapter.ConvertToVolumes(volumeConfigs)
	if err != nil {
		return nil, err
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

// logUnsupportedConfigFields adds a WARNING to the customer about the fields that are unused.
func logUnsupportedConfigFields(project *project.Project) {
	// ecsProject#parseCompose, which calls the underlying libcompose.Project#Parse(),
	// always populates the project.NetworkConfig with one entry ("default").
	// See: https://github.com/docker/libcompose/blob/master/project/project.go#L277
	if project.NetworkConfigs != nil && len(project.NetworkConfigs) > 1 {
		log.WithFields(log.Fields{"option name": "networks"}).Warn("Skipping unsupported YAML option...")
	}
}

// logUnsupportedServiceConfigFields TODO move into parser logic?
func logUnsupportedServiceConfigFields(serviceName string, config *config.ServiceConfig) {
	configValue := reflect.ValueOf(config).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := configType.Field(i)
		// get the tag name (if any), defaults to fieldName
		tagName := fieldType.Name
		yamlTag := fieldType.Tag.Get("yaml") // Expected format `yaml:"tagName,omitempty"` // TODO, handle omitempty
		if yamlTag != "" {
			tags := strings.Split(yamlTag, ",")
			if len(tags) > 0 {
				tagName = tags[0]
			}
		}

		if tagName == "networks" && !validNetworksForService(config) {
			log.WithFields(log.Fields{
				"option name":  tagName,
				"service name": serviceName,
			}).Warn("Skipping unsupported YAML option for service...")
		}

		zeroValue := isZero(field)
		// if value is present for the field that is not in supportedYamlTags map, log a warning
		if tagName != "networks" && !zeroValue && !supportedComposeYamlOptionsMap[tagName] {
			log.WithFields(log.Fields{
				"option name":  tagName,
				"service name": serviceName,
			}).Warn("Skipping unsupported YAML option for service...")
		}
	}
}

func validNetworksForService(config *config.ServiceConfig) bool {
	if config.Networks == nil {
		return false
	}
	if config.Networks.Networks == nil {
		return false
	}
	if len(config.Networks.Networks) != 1 {
		return false
	}

	return true
}

// isZero checks if the value is nil or empty or zero
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		zero := true
		for i := 0; i < v.Len(); i++ {
			zero = zero && isZero(v.Index(i))
		}
		return zero
	case reflect.Struct:
		zero := true
		for i := 0; i < v.NumField(); i++ {
			zero = zero && isZero(v.Field(i))
		}
		return zero
	}
	// Compare other types directly:
	zero := reflect.Zero(v.Type())
	return v.Interface() == zero.Interface()
}

// convertToContainerDef transforms each service in docker-compose.yml and
// ecs-params.yml to an equivalent ECS container definition
func convertToContainerDef(inputCfg *adapter.ContainerConfig, ecsContainerDef *ContainerDef) (*ecs.ContainerDefinition, error) {
	outputContDef := &ecs.ContainerDefinition{}

	mem := inputCfg.Memory
	memoryReservation := inputCfg.MemoryReservation

	if mem != 0 && memoryReservation != 0 && mem < memoryReservation {
		return nil, errors.New("mem_limit must be greater than mem_reservation")
	}

	// One or both of memory and memoryReservation is required to register a task definition with ECS
	// NOTE: Docker does not set a memory limit for containers
	if mem == 0 && memoryReservation == 0 {
		mem = defaultMemLimit
	}

	// Populate ECS container definition, offloading the validation to aws-sdk
	outputContDef.SetCpu(inputCfg.CPU)
	outputContDef.SetCommand(aws.StringSlice(inputCfg.Command))
	outputContDef.SetDnsSearchDomains(aws.StringSlice(inputCfg.DNSSearchDomains))
	outputContDef.SetDnsServers(aws.StringSlice(inputCfg.DNSServers))
	outputContDef.SetDockerLabels(inputCfg.DockerLabels)
	outputContDef.SetDockerSecurityOptions(aws.StringSlice(inputCfg.DockerSecurityOptions))
	outputContDef.SetEntryPoint(aws.StringSlice(inputCfg.Entrypoint))
	outputContDef.SetEnvironment(inputCfg.Environment)
	outputContDef.SetExtraHosts(inputCfg.ExtraHosts)
	outputContDef.SetHostname(inputCfg.Hostname)
	outputContDef.SetImage(inputCfg.Image)
	outputContDef.SetLinks(aws.StringSlice(inputCfg.Links)) // TODO, read from external links
	outputContDef.SetLogConfiguration(inputCfg.LogConfiguration)

	if mem != 0 {
		outputContDef.SetMemory(mem)
	}
	if memoryReservation != 0 {
		outputContDef.SetMemoryReservation(memoryReservation)
	}

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

	// Set fields from ecs-params file
	if ecsContainerDef != nil {
		outputContDef.Essential = aws.Bool(ecsContainerDef.Essential)
	}

	return outputContDef, nil
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
