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

	"github.com/Sirupsen/logrus"
	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	defaultMemLimit = 512
	kiB             = 1024

	// access mode with which the volume is mounted
	readOnlyVolumeAccessMode  = "ro"
	readWriteVolumeAccessMode = "rw"
)

// ConvertToTaskDefinition transforms the yaml configs to its ecs equivalent (task definition)
func ConvertToTaskDefinition(context libcompose.Context,
	serviceConfigs map[string]*libcompose.ServiceConfig) (*ecs.TaskDefinition, error) {

	if len(serviceConfigs) == 0 {
		return nil, errors.New("cannot create a task definition with no containers; invalid service config")
	}

	taskDefinitionName := getTaskDefinitionName(context.ProjectName)
	containerDefinitions := []*ecs.ContainerDefinition{}
	volumes := make(map[string]string) // map with key:=hostSourcePath value:=VolumeName
	for name, config := range serviceConfigs {
		containerDef := &ecs.ContainerDefinition{
			Name: aws.String(name),
		}
		if err := convertToContainerDef(config, volumes, containerDef); err != nil {
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
func convertToContainerDef(inputCfg *libcompose.ServiceConfig,
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
	// TODO, read env file
	environment := []*ecs.KeyValuePair{}
	for _, env := range inputCfg.Environment.Slice() {
		parts := strings.SplitN(env, "=", 2)
		name := &parts[0]
		var value *string
		if len(parts) > 1 {
			value = &parts[1]
		}
		environment = append(environment, &ecs.KeyValuePair{
			Name:  name,
			Value: value,
		})
	}

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
	// populating container definition, offloading the validation to aws-sdk
	outputContDef.Cpu = aws.Int64(inputCfg.CpuShares)
	outputContDef.Memory = aws.Int64(mem)
	outputContDef.EntryPoint = aws.StringSlice(inputCfg.Entrypoint.Slice())
	outputContDef.Command = aws.StringSlice(inputCfg.Command.Slice())
	outputContDef.Environment = environment
	outputContDef.Image = aws.String(inputCfg.Image)
	outputContDef.Links = aws.StringSlice(inputCfg.Links.Slice()) //TODO, read from external links
	outputContDef.MountPoints = mountPoints
	outputContDef.PortMappings = portMappings
	outputContDef.VolumesFrom = volumesFrom
	return nil
}

// convertToECSVolumes transforms the map of hostPaths to the format of ecs.Volume

func convertToECSVolumes(hostPaths map[string]string) []*ecs.Volume {
	output := []*ecs.Volume{}
	for hostPath, volName := range hostPaths {
		if hostPath == "" {
			ecsVolume := &ecs.Volume{
				Name: aws.String(volName),
			}
			output = append(output, ecsVolume)
		} else {
			ecsVolume := &ecs.Volume{
				Name: aws.String(volName),
				Host: &ecs.HostVolumeProperties{
					SourcePath: aws.String(hostPath),
				},
			}
			output = append(output, ecsVolume)
		}
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
			logrus.WithFields(logrus.Fields{
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
