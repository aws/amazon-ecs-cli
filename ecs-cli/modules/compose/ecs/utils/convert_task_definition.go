// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"fmt"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	defaultMemLimit   = 512
	kiB               = 1024
	ulimitFormatError = "expected format TYPE=SOFT-LIMIT[:HARD-LIMIT]. could not parse ulimit "

	// access mode with which the volume is mounted
	readOnlyVolumeAccessMode  = "ro"
	readWriteVolumeAccessMode = "rw"
)

// ConvertToTaskDefinition transforms the yaml configs to its ecs equivalent (task definition)
func ConvertToTaskDefinition(taskDefinitionName string, context libcompose.Context,
	serviceConfigs map[string]*libcompose.ServiceConfig) (*ecs.TaskDefinition, error) {

	if len(serviceConfigs) == 0 {
		return nil, errors.New("cannot create a task definition with no containers; invalid service config")
	}

	containerDefinitions := []*ecs.ContainerDefinition{}
	volumes := make(map[string]string) // map with key:=hostSourcePath value:=VolumeName
	for name, config := range serviceConfigs {
		containerDef := &ecs.ContainerDefinition{
			Name: aws.String(name),
		}
		if err := convertToContainerDef(context, config, volumes, containerDef); err != nil {
			return nil, err
		}
		containerDefinitions = append(containerDefinitions, containerDef)
	}
	taskDefinition := &ecs.TaskDefinition{
		Family:               aws.String(taskDefinitionName),
		ContainerDefinitions: containerDefinitions,
		Volumes:              convertToECSVolumes(volumes),
	}
	return taskDefinition, nil
}

// convertToContainerDef transforms each service in the compose yml
// to an equivalent container definition
func convertToContainerDef(context libcompose.Context, inputCfg *libcompose.ServiceConfig,
	volumes map[string]string, outputContDef *ecs.ContainerDefinition) error {
	// setting memory
	var mem int64
	if inputCfg.MemLimit != 0 {
		mem = inputCfg.MemLimit / kiB / kiB // convert bytes to MiB
	}
	if mem == 0 {
		mem = defaultMemLimit
	}

	// convert environment variables
	envVars, err := libcompose.GetEnvVarsFromConfig(context, inputCfg)
	if err != nil {
		return err
	}
	environment := convertToKeyValuePairs(context, envVars, *outputContDef.Name, inputCfg)

	// convert port mappings
	portMappings, err := convertToPortMappings(*outputContDef.Name, inputCfg.Ports)
	if err != nil {
		return err
	}

	// convert volumes from
	volumesFrom := []*ecs.VolumeFrom{}
	for _, val := range inputCfg.VolumesFrom {
		volumeFrom := &ecs.VolumeFrom{
			SourceContainer: aws.String(val),
		}
		volumesFrom = append(volumesFrom, volumeFrom)
	}

	// convert mount points
	mountPoints, err := convertToMountPoints(inputCfg.Volumes, volumes)
	if err != nil {
		return err
	}

	// convert extra hosts
	extraHosts, err := convertToExtraHosts(inputCfg.ExtraHosts)
	if err != nil {
		return err
	}

	// convert log configuration
	var logConfig *ecs.LogConfiguration
	if inputCfg.LogDriver != "" {
		logConfig = &ecs.LogConfiguration{
			LogDriver: aws.String(inputCfg.LogDriver),
			Options:   aws.StringMap(inputCfg.LogOpt),
		}
	}

	// convert ulimits
	ulimits, err := convertToULimits(inputCfg.ULimits)
	if err != nil {
		return err
	}

	// populating container definition, offloading the validation to aws-sdk
	outputContDef.Cpu = aws.Int64(inputCfg.CpuShares)
	outputContDef.Command = aws.StringSlice(inputCfg.Command.Slice())
	outputContDef.DnsSearchDomains = aws.StringSlice(inputCfg.DNSSearch.Slice())
	outputContDef.DnsServers = aws.StringSlice(inputCfg.DNS.Slice())
	outputContDef.DockerLabels = aws.StringMap(inputCfg.Labels.MapParts())
	outputContDef.DockerSecurityOptions = aws.StringSlice(inputCfg.SecurityOpt)
	outputContDef.EntryPoint = aws.StringSlice(inputCfg.Entrypoint.Slice())
	outputContDef.Environment = environment
	outputContDef.ExtraHosts = extraHosts
	if inputCfg.Hostname != "" {
		outputContDef.Hostname = aws.String(inputCfg.Hostname)
	}
	outputContDef.Image = aws.String(inputCfg.Image)
	outputContDef.Links = aws.StringSlice(inputCfg.Links.Slice()) //TODO, read from external links
	outputContDef.LogConfiguration = logConfig
	outputContDef.Memory = aws.Int64(mem)
	outputContDef.MountPoints = mountPoints
	outputContDef.Privileged = aws.Bool(inputCfg.Privileged)
	outputContDef.PortMappings = portMappings
	outputContDef.ReadonlyRootFilesystem = aws.Bool(inputCfg.ReadOnly)
	outputContDef.Ulimits = ulimits
	if inputCfg.User != "" {
		outputContDef.User = aws.String(inputCfg.User)
	}
	outputContDef.VolumesFrom = volumesFrom
	if inputCfg.WorkingDir != "" {
		outputContDef.WorkingDirectory = aws.String(inputCfg.WorkingDir)
	}

	return nil
}

// convertToKeyValuePairs transforms the map of environment variables into list of ecs.KeyValuePair.
// Environment variables with only a key are resolved by reading the variable from the shell where ecs-cli is executed from.
// TODO: use this logic to generate RunTask overrides for ecs-cli compose commands (instead of always creating a new task def)
func convertToKeyValuePairs(context libcompose.Context, envVars []string,
	serviceName string, inputCfg *libcompose.ServiceConfig) []*ecs.KeyValuePair {

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
			resolvedEnvVars := context.EnvironmentLookup.Lookup(key, serviceName, inputCfg)

			// couldn't resolve env var from where the command is executed. Skip the key
			if len(resolvedEnvVars) == 0 {
				log.WithFields(log.Fields{"key name": key}).Warn("Skipping unresolved Environment variable...")
				continue
			}

			// found env var values from where the command is executed
			for _, value := range resolvedEnvVars {
				lookupParts := strings.SplitN(value, "=", 2)
				environment = append(environment, createKeyValuePair(key, lookupParts[1]))
			}
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
func convertToECSVolumes(hostPaths map[string]string) []*ecs.Volume {
	output := []*ecs.Volume{}
	for hostPath, volName := range hostPaths {
		ecsVolume := &ecs.Volume{
			Name: aws.String(volName),
			Host: &ecs.HostVolumeProperties{
				SourcePath: aws.String(hostPath),
			},
		}
		output = append(output, ecsVolume)
	}
	return output
}

// convertToPortMappings transforms the yml ports string slice to ecs compatible PortMappings slice
func convertToPortMappings(serviceName string, cfgPorts []string) ([]*ecs.PortMapping, error) {
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

// convertToMountPoints transforms the yml volumes slice to ecs compatible MountPoints slice
// It also uses the hostPath from volumes if present, else adds one to it
func convertToMountPoints(cfgVolumes []string, volumes map[string]string) ([]*ecs.MountPoint, error) {
	mountPoints := []*ecs.MountPoint{}
	for _, cfgVolume := range cfgVolumes {
		parts := strings.Split(cfgVolume, ":")
		var containerPath, hostPath string
		var readOnly bool
		switch len(parts) {
		case 1: // Format CONT_PATH Example- /var/lib/mysql
			containerPath = parts[0]
		case 2: // Format HOST_PATH:CONT_PATH Example - ./cache:/tmp/cache
			hostPath = parts[0]
			containerPath = parts[1]
		case 3: // Format HOST_PATH:CONT_PATH:RO Example - ~/configs:/etc/configs/:ro
			hostPath = parts[0]
			containerPath = parts[1]
			accessModeStr := parts[2]
			if accessModeStr == readOnlyVolumeAccessMode {
				readOnly = true
			} else if accessModeStr == readWriteVolumeAccessMode {
				readOnly = false
			} else {
				return nil, fmt.Errorf(
					"expected format [HOST:]CONTAINER[:ro|rw]. could not parse volume: %s", cfgVolume)
			}
		default:
			return nil, fmt.Errorf(
				"expected format [HOST:]CONTAINER[:ro]. could not parse volume: %s", cfgVolume)
		}

		var volumeName string
		if len(volumes) > 0 {
			volumeName = volumes[hostPath]
		}
		if volumeName == "" {
			volumeName = getVolumeName(len(volumes))
			volumes[hostPath] = volumeName
		}

		mountPoints = append(mountPoints, &ecs.MountPoint{
			ContainerPath: aws.String(containerPath),
			SourceVolume:  aws.String(volumeName),
			ReadOnly:      aws.Bool(readOnly),
		})
	}
	return mountPoints, nil
}

// convertToExtraHosts transforms the yml extra hosts slice to ecs compatible HostEntry slice
func convertToExtraHosts(cfgExtraHosts []string) ([]*ecs.HostEntry, error) {
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

// convertToULimits transforms the yml extra hosts slice to ecs compatible Ulimit slice
func convertToULimits(cfgUlimits []string) ([]*ecs.Ulimit, error) {
	ulimits := []*ecs.Ulimit{}
	for _, cfgUlimit := range cfgUlimits {
		parts := strings.Split(cfgUlimit, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf(
				"%s:%s", ulimitFormatError, cfgUlimit)
		}

		limits := strings.Split(parts[1], ":")
		var softLimit, hardLimit int64
		var limitErr error
		switch len(limits) {
		case 1: // Format TYPE=SOFT-LIMIT Example- nofile=1024
			softLimit, limitErr = strconv.ParseInt(limits[0], 10, 64)
			hardLimit = softLimit
		case 2: // Format TYPE=SOFT-LIMIT:HARD-LIMIT Example nofile=1024:1024
			softLimit, limitErr = strconv.ParseInt(limits[0], 10, 64)
			hardLimit, limitErr = strconv.ParseInt(limits[1], 10, 64)
		default:
			return nil, fmt.Errorf(
				"%s:%s", ulimitFormatError, cfgUlimit)
		}
		if limitErr != nil {
			return nil, fmt.Errorf("Could not convert limits into integers in ulimits[%s]. Error=[%v]", cfgUlimit, limitErr)
		}

		ulimit := &ecs.Ulimit{
			Name:      aws.String(parts[0]),
			SoftLimit: aws.Int64(softLimit),
			HardLimit: aws.Int64(hardLimit),
		}
		ulimits = append(ulimits, ulimit)
	}

	return ulimits, nil
}
