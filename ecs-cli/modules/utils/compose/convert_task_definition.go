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
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/go-units"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	defaultMemLimit = 512
	kiB             = 1024
	miB             = kiB * kiB // 1048576 bytes

	// access mode with which the volume is mounted
	readOnlyVolumeAccessMode  = "ro"
	readWriteVolumeAccessMode = "rw"
	volumeFromContainerKey    = "container"
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

type Volumes struct {
	volumeWithHost  map[string]string
	volumeEmptyHost []string
}

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
	volumes, err := ConvertToVolumes(volumeConfigs)
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

// ConvertToKeyValuePairs transforms the map of environment variables into list of ecs.KeyValuePair.
// Environment variables with only a key are resolved by reading the variable from the shell where ecscli is executed from.
// TODO: use this logic to generate RunTask overrides for ecscli compose commands (instead of always creating a new task def)
func ConvertToKeyValuePairs(context *project.Context, envVars yaml.MaporEqualSlice, serviceName string) []*ecs.KeyValuePair {
	environment := []*ecs.KeyValuePair{}
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		key := parts[0]

		// format: key=value
		if len(parts) > 1 && parts[1] != "" {
			environment = append(environment, createKeyValuePair(key, parts[1]))
			continue
		}

		// format: key
		// format: key=
		if context.EnvironmentLookup != nil {
			resolvedEnvVars := context.EnvironmentLookup.Lookup(key, nil)

			// If the environment variable couldn't be resolved, set the value to an empty string
			// Reference: https://github.com/docker/libcompose/blob/3c40e1001a2646ec6f7a6613873cf5a30122a417/config/interpolation.go#L148
			if len(resolvedEnvVars) == 0 {
				log.WithFields(log.Fields{"key name": key}).Warn("Environment variable is unresolved. Setting it to a blank value...")
				environment = append(environment, createKeyValuePair(key, ""))
				continue
			}

			// Use first result if many are given
			value := resolvedEnvVars[0]
			lookupParts := strings.SplitN(value, "=", 2)
			environment = append(environment, createKeyValuePair(key, lookupParts[1]))
		}
	}
	return environment
}

// createKeyValuePair generates an ecs.KeyValuePair object
func createKeyValuePair(key, value string) *ecs.KeyValuePair {
	return &ecs.KeyValuePair{
		Name:  aws.String(key),
		Value: aws.String(value),
	}
}

// convertToECSVolumes transforms the map of hostPaths to the format of ecs.Volume
func convertToECSVolumes(hostPaths *Volumes) []*ecs.Volume {
	output := []*ecs.Volume{}
	// volumes with a host path
	for hostPath, volName := range hostPaths.volumeWithHost {
		ecsVolume := &ecs.Volume{
			Name: aws.String(volName),
			Host: &ecs.HostVolumeProperties{
				SourcePath: aws.String(hostPath),
			}}
		output = append(output, ecsVolume)
	}
	// volumes with an empty host path
	for _, volName := range hostPaths.volumeEmptyHost {
		ecsVolume := &ecs.Volume{
			Name: aws.String(volName),
		}
		output = append(output, ecsVolume)
	}
	return output
}

// ConvertToPortMappings transforms the yml ports string slice to ecs compatible PortMappings slice
func ConvertToPortMappings(serviceName string, cfgPorts []string) ([]*ecs.PortMapping, error) {
	portMappings := []*ecs.PortMapping{}
	for _, portMapping := range cfgPorts {
		// TODO: suffix-check case insensitive?

		// Example format "8000:8000/udp"
		protocol := ecs.TransportProtocolTcp // default protocol:tcp
		tcp := strings.HasSuffix(portMapping, "/"+ecs.TransportProtocolTcp)
		udp := strings.HasSuffix(portMapping, "/"+ecs.TransportProtocolUdp)
		if tcp || udp {
			protocol = portMapping[len(portMapping)-3:] // slice protocol name from portMapping, 3=len(ecs.TransportProtocolTcp)
			portMapping = portMapping[0 : len(portMapping)-4]
		}

		// either has 1 part (just the containerPort) or has 2 parts (hostPort:containerPort)
		parts := strings.Split(portMapping, ":")
		var containerPort, hostPort int
		var portErr error
		switch len(parts) {
		case 1: // Format "containerPort" Example "8000"
			containerPort, portErr = strconv.Atoi(parts[0])
		case 2: // Format "hostPort:containerPort" Example "8000:8000"
			hostPort, portErr = strconv.Atoi(parts[0])
			containerPort, portErr = strconv.Atoi(parts[1])
		case 3: // Format "ipAddr:hostPort:containerPort" Example "127.0.0.0.1:8000:8000"
			log.WithFields(log.Fields{
				"container":   serviceName,
				"portMapping": portMapping,
			}).Warn("Ignoring the ip address while transforming it to task definition")
			hostPort, portErr = strconv.Atoi(parts[1])
			containerPort, portErr = strconv.Atoi(parts[2])
		default:
			return nil, fmt.Errorf(
				"expected format [hostPort]:containerPort. Could not parse portmappings: %s", portMapping)
		}
		if portErr != nil {
			return nil, fmt.Errorf("Could not convert port into integer in portmappings: %v", portErr)
		}

		portMappings = append(portMappings, &ecs.PortMapping{
			Protocol:      aws.String(protocol),
			ContainerPort: aws.Int64(int64(containerPort)),
			HostPort:      aws.Int64(int64(hostPort)),
		})
	}
	return portMappings, nil
}

// ConvertToVolumesFrom transforms the yml volumes from to ecs compatible VolumesFrom slice
// Examples for compose format v2:
// volumes_from:
// - service_name
// - service_name:ro
// - container:container_name
// - container:container_name:rw
// Examples for compose format v1:
// volumes_from:
// - service_name
// - service_name:ro
// - container_name
// - container_name:rw
func ConvertToVolumesFrom(cfgVolumesFrom []string) ([]*ecs.VolumeFrom, error) {
	volumesFrom := []*ecs.VolumeFrom{}

	for _, cfgVolumeFrom := range cfgVolumesFrom {
		parts := strings.Split(cfgVolumeFrom, ":")

		var containerName, accessModeStr string

		parseErr := fmt.Errorf(
			"expected format [container:]SERVICE|CONTAINER[:ro|rw]. could not parse cfgVolumeFrom: %s", cfgVolumeFrom)

		switch len(parts) {
		// for the following volumes_from formats (supported by compose file formats v1 and v2),
		// name: refers to either service_name or container_name
		// container: is a keyword thats introduced in v2 to differentiate between service_name and container:container_name
		// ro|rw: read-only or read-write access
		case 1: // Format: name
			containerName = parts[0]
		case 2: // Format: name:ro|rw (OR) container:name
			if parts[0] == volumeFromContainerKey {
				containerName = parts[1]
			} else {
				containerName = parts[0]
				accessModeStr = parts[1]
			}
		case 3: // Format: container:name:ro|rw
			if parts[0] != volumeFromContainerKey {
				return nil, parseErr
			}
			containerName = parts[1]
			accessModeStr = parts[2]
		default:
			return nil, parseErr
		}

		// parse accessModeStr
		var readOnly bool
		if accessModeStr != "" {
			if accessModeStr == readOnlyVolumeAccessMode {
				readOnly = true
			} else if accessModeStr == readWriteVolumeAccessMode {
				readOnly = false
			} else {
				return nil, fmt.Errorf("Could not parse access mode %s", accessModeStr)
			}
		}
		volumesFrom = append(volumesFrom, &ecs.VolumeFrom{
			SourceContainer: aws.String(containerName),
			ReadOnly:        aws.Bool(readOnly),
		})
	}
	return volumesFrom, nil
}

// ConvertToMountPoints transforms the yml volumes slice to ecs compatible MountPoints slice
// It also uses the hostPath from volumes if present, else adds one to it
func ConvertToMountPoints(cfgVolumes *yaml.Volumes, volumes *Volumes) ([]*ecs.MountPoint, error) {
	mountPoints := []*ecs.MountPoint{}
	if cfgVolumes == nil {
		return mountPoints, nil
	}
	for _, cfgVolume := range cfgVolumes.Volumes {
		source := cfgVolume.Source
		containerPath := cfgVolume.Destination

		accessMode := cfgVolume.AccessMode
		var readOnly bool
		if accessMode != "" {
			if accessMode == readOnlyVolumeAccessMode {
				readOnly = true
			} else if accessMode == readWriteVolumeAccessMode {
				readOnly = false
			} else {
				return nil, fmt.Errorf(
					"expected format [HOST:]CONTAINER[:ro|rw]. could not parse volume: %s", cfgVolume)
			}
		}

		var volumeName string
		numVol := len(volumes.volumeWithHost) + len(volumes.volumeEmptyHost)
		if source == "" {
			// add mount point for volumes with an empty host path
			volumeName = getVolumeName(numVol)
			volumes.volumeEmptyHost = append(volumes.volumeEmptyHost, volumeName)
		} else if project.IsNamedVolume(source) {
			if !utils.InSlice(source, volumes.volumeEmptyHost) {
				return nil, fmt.Errorf(
					"named volume [%s] is used but no declaration was found in the volumes section", cfgVolume)
			}
			volumeName = source
		} else {
			// add mount point for volumes with a host path
			volumeName = volumes.volumeWithHost[source]

			if volumeName == "" {
				volumeName = getVolumeName(numVol)
				volumes.volumeWithHost[source] = volumeName
			}
		}

		mountPoints = append(mountPoints, &ecs.MountPoint{
			ContainerPath: aws.String(containerPath),
			SourceVolume:  aws.String(volumeName),
			ReadOnly:      aws.Bool(readOnly),
		})
	}
	return mountPoints, nil
}

// ConvertToExtraHosts transforms the yml extra hosts slice to ecs compatible HostEntry slice
func ConvertToExtraHosts(cfgExtraHosts []string) ([]*ecs.HostEntry, error) {
	extraHosts := []*ecs.HostEntry{}
	for _, cfgExtraHost := range cfgExtraHosts {
		parts := strings.Split(cfgExtraHost, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"expected format HOSTNAME:IPADDRESS. could not parse ExtraHost: %s", cfgExtraHost)
		}
		extraHost := &ecs.HostEntry{
			Hostname:  aws.String(parts[0]),
			IpAddress: aws.String(parts[1]),
		}
		extraHosts = append(extraHosts, extraHost)
	}

	return extraHosts, nil
}

// ConvertToULimits transforms the yml extra hosts slice to ecs compatible Ulimit slice
func ConvertToULimits(cfgUlimits yaml.Ulimits) ([]*ecs.Ulimit, error) {
	ulimits := []*ecs.Ulimit{}
	for _, cfgUlimit := range cfgUlimits.Elements {
		ulimit := &ecs.Ulimit{
			Name:      aws.String(cfgUlimit.Name),
			SoftLimit: aws.Int64(cfgUlimit.Soft),
			HardLimit: aws.Int64(cfgUlimit.Hard),
		}
		ulimits = append(ulimits, ulimit)
	}

	return ulimits, nil
}

// ConvertToTmpfs transforms the yml Tmpfs slice of strings to slice of pointers to Tmpfs structs
func ConvertToTmpfs(tmpfsPaths yaml.Stringorslice) ([]*ecs.Tmpfs, error) {

	if len(tmpfsPaths) == 0 {
		return nil, nil
	}

	mounts := []*ecs.Tmpfs{}
	for _, mount := range tmpfsPaths {

		// mount should be of the form "<path>:<options>"
		tmpfsParams := strings.SplitN(mount, ":", 2)

		if len(tmpfsParams) < 2 {
			return nil, errors.New("Path and Size are required options for tmpfs")
		}

		path := tmpfsParams[0]
		options := strings.Split(tmpfsParams[1], ",")

		var mountOptions []string
		var size int64

		// See: https://github.com/docker/go-units/blob/master/size.go#L34
		s := regexp.MustCompile(`size=(\d+(\.\d+)*) ?([kKmMgGtTpP])?[bB]?`)

		for _, option := range options {
			if sizeOption := s.FindString(option); sizeOption != "" {
				sizeValue := strings.SplitN(sizeOption, "=", 2)[1]
				sizeInBytes, err := units.RAMInBytes(sizeValue)

				if err != nil {
					return nil, err
				}

				size = sizeInBytes / miB
			} else {
				mountOptions = append(mountOptions, option)
			}
		}

		if size == 0 {
			return nil, errors.New("You must specify the size option for tmpfs")
		}

		tmpfs := &ecs.Tmpfs{
			ContainerPath: aws.String(path),
			MountOptions:  aws.StringSlice(mountOptions),
			Size:          aws.Int64(size),
		}
		mounts = append(mounts, tmpfs)
	}
	return mounts, nil
}

// SortedGoString returns deterministic string representation
// json Marshal sorts map keys, making it deterministic
func SortedGoString(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
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

// ConvertToLogConfiguration converts a libcompose logging
// fields to an ECS LogConfiguration
func ConvertToLogConfiguration(inputCfg *config.ServiceConfig) (*ecs.LogConfiguration, error) {
	var logConfig *ecs.LogConfiguration
	if inputCfg.Logging.Driver != "" {
		logConfig = &ecs.LogConfiguration{
			LogDriver: aws.String(inputCfg.Logging.Driver),
			Options:   aws.StringMap(inputCfg.Logging.Options),
		}
	}
	return logConfig, nil
}

// ConvertToMemoryInMB converts libcompose-parsed bytes to MiB, expected by ECS
func ConvertToMemoryInMB(bytes int64) int64 {
	var memory int64
	if bytes != 0 {
		memory = int64(bytes) / miB
	}
	return memory
}

// ConvertToVolumes converts the VolumeConfigs map on a libcompose project into
// a Volumes struct and populates the volumeEmptyHost field with any named volumes
func ConvertToVolumes(volumeConfigs map[string]*config.VolumeConfig) (*Volumes, error) {
	volumes := &Volumes{
		volumeWithHost: make(map[string]string), // map with key:=hostSourcePath value:=VolumeName
	}

	// Add named volume configs:
	if volumeConfigs != nil {
		for name, config := range volumeConfigs {
			if config != nil {
				// NOTE: If Driver field is not empty, this
				// will add a prefix to the named volume on the container
				if config.Driver != "" {
					return nil, errors.New("Volume driver is not supported")
				}
				// Driver Options must relate to a specific volume driver
				if len(config.DriverOpts) != 0 {
					return nil, errors.New("Volume driver options is not supported")
				}
				return nil, errors.New("External option is not supported")
			}
			volumes.volumeEmptyHost = append(volumes.volumeEmptyHost, name)
		}
	}

	return volumes, nil
}
