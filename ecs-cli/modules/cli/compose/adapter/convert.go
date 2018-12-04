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

package adapter

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/go-units"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	kiB = 1024
	miB = kiB * kiB // 1048576 bytes

	// access mode with which the volume is mounted
	readOnlyVolumeAccessMode  = "ro"
	readWriteVolumeAccessMode = "rw"
	volumeFromContainerKey    = "container"
)

// ConvertToDevices transforms a slice of device strings into a slice of ECS Device structs
func ConvertToDevices(cfgDevices []string) ([]*ecs.Device, error) {
	devices := []*ecs.Device{}
	for _, devString := range cfgDevices {
		var device ecs.Device

		parts := strings.Split(devString, ":")
		numOfParts := len(parts)

		hostPath := parts[0]
		device.SetHostPath(hostPath)

		if numOfParts > 1 {
			containerPath := parts[1]
			device.SetContainerPath(containerPath)
		}
		if numOfParts > 2 {
			permissions, err := getDevicePermissions(parts[2])
			if err != nil {
				return nil, err
			}
			device.SetPermissions(aws.StringSlice(permissions))
		}
		if numOfParts > 3 {
			return nil, fmt.Errorf(
				"Invalid number of arguments in device %s", devString)
		}

		devices = append(devices, &device)
	}
	return devices, nil
}

func getDevicePermissions(perms string) ([]string, error) {
	// store in map to prevent duplicates, which will fail on RegisterTaskDefinition
	seenPerms := map[string]bool{}

	if len(perms) > 3 {
		return nil, fmt.Errorf(
			"Invalid number of device options: found %d, max is 3", len(perms))
	}
	for _, char := range perms {
		switch char {
		case 'r':
			seenPerms["read"] = true
		case 'w':
			seenPerms["write"] = true
		case 'm':
			seenPerms["mknod"] = true
		default:
			return nil, fmt.Errorf(
				"Invalid device option: found '%s', but only 'r', 'w' or 'm' are valid", string(char))
		}
	}

	permissions := []string{}
	for key := range seenPerms {
		permissions = append(permissions, key)
	}

	return permissions, nil
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

// ConvertToHealthCheck converts a compose healthcheck to ECS healthcheck
func ConvertToHealthCheck(healthCheckConfig *types.HealthCheckConfig) *ecs.HealthCheck {
	ecsHealthcheck := &ecs.HealthCheck{
		Command: aws.StringSlice(healthCheckConfig.Test),
	}
	// optional fields with defaults provided by ECS
	if healthCheckConfig.Interval != nil {
		ecsHealthcheck.Interval = ConvertToTimeInSeconds(healthCheckConfig.Interval)
	}
	if healthCheckConfig.Retries != nil {
		ecsHealthcheck.Retries = aws.Int64(int64(*healthCheckConfig.Retries))
	}
	if healthCheckConfig.Timeout != nil {
		ecsHealthcheck.Timeout = ConvertToTimeInSeconds(healthCheckConfig.Timeout)
	}
	if healthCheckConfig.StartPeriod != nil {
		ecsHealthcheck.StartPeriod = ConvertToTimeInSeconds(healthCheckConfig.StartPeriod)
	}

	return ecsHealthcheck
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

// ConvertToTimeInSeconds converts a duration to an int64 number of seconds
func ConvertToTimeInSeconds(d *time.Duration) *int64 {
	val := d.Nanoseconds() / 1E9
	return &val
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

		volumeName, err := GetSourcePathAndUpdateVolumes(source, volumes)
		if err != nil {
			return nil, err
		}

		mountPoints = append(mountPoints, &ecs.MountPoint{
			ContainerPath: aws.String(containerPath),
			SourceVolume:  aws.String(volumeName),
			ReadOnly:      aws.Bool(readOnly),
		})
	}
	return mountPoints, nil
}

// GetSourcePathAndUpdateVolumes checks for & creates an ECS Volume for a mount point without
// a source volume and returns the appropriate source volume name
func GetSourcePathAndUpdateVolumes(source string, volumes *Volumes) (string, error) {
	var volumeName string
	numVol := len(volumes.VolumeWithHost) + len(volumes.VolumeEmptyHost)
	if source == "" {
		// add mount point for volumes with an empty source path
		volumeName = getVolumeName(numVol)
		volumes.VolumeEmptyHost = append(volumes.VolumeEmptyHost, volumeName)
	} else if project.IsNamedVolume(source) {
		if !utils.InSlice(source, volumes.VolumeEmptyHost) {
			return "", fmt.Errorf(
				"named volume [%s] is used but no declaration was found in the volumes section", source)
		}
		volumeName = source
	} else {
		// add mount point for volumes with a source path
		volumeName = volumes.VolumeWithHost[source]

		if volumeName == "" {
			volumeName = getVolumeName(numVol)
			volumes.VolumeWithHost[source] = volumeName
		}
	}

	return volumeName, nil
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

// ConvertToVolumes converts the VolumeConfigs map on a libcompose project into
// a Volumes struct and populates the VolumeEmptyHost field with any named volumes
func ConvertToVolumes(volumeConfigs map[string]*config.VolumeConfig) (*Volumes, error) {
	volumes := NewVolumes()

	// Add named volume configs:
	if volumeConfigs != nil {
		for name, config := range volumeConfigs {
			if config != nil {
				return nil, logOutUnsupportedVolumeFields(config.Driver, config.DriverOpts, nil)
			}
			volumes.VolumeEmptyHost = append(volumes.VolumeEmptyHost, name)
		}
	}
	return volumes, nil
}

// ConvertToV3Volumes converts the VolumesConfig map in a docker/cli config into a
// Volumes struct and populates the VolumesEmptyHost field with any names volumes
func ConvertToV3Volumes(volConfig map[string]types.VolumeConfig) (*Volumes, error) {
	volumes := NewVolumes()

	// Add named volume configs:
	if len(volConfig) != 0 {
		for name, config := range volConfig {
			if !reflect.DeepEqual(config, types.VolumeConfig{}) {
				return nil, logOutUnsupportedVolumeFields(config.Driver, config.DriverOpts, config.Labels)
			}
			volumes.VolumeEmptyHost = append(volumes.VolumeEmptyHost, name)
		}
	}
	return volumes, nil
}

func logOutUnsupportedVolumeFields(driver string, driverOpts map[string]string, labels map[string]string) error {
	// NOTE: If Driver field is not empty, this
	// will add a prefix to the named volume on the container
	if driver != "" {
		return errors.New("Volume driver is not supported")
	}
	// Driver Options must relate to a specific volume driver
	if len(driverOpts) != 0 {
		return errors.New("Volume driver options is not supported")
	}
	if labels != nil && len(labels) != 0 {
		return errors.New("Volume labels are not supported")
	}
	return errors.New("External option is not supported")
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

// SortedGoString returns deterministic string representation
// json Marshal sorts map keys, making it deterministic
func SortedGoString(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Necessary in conjunction with SortedGoString to avoid spurious tdcache misses
func SortedContainerDefinitionsByName(request *ecs.RegisterTaskDefinitionInput) ecs.RegisterTaskDefinitionInput {
	cdefs := request.ContainerDefinitions
	sort.Slice(cdefs, func(i, j int) bool {
		return *cdefs[i].Name < *cdefs[j].Name
	})
	sorted := *request
	sorted.ContainerDefinitions = cdefs
	return sorted
}

// ConvertCamelCaseToUnderScore returns an underscore-separated name for a given camelcased string
// e.g., "NetworkMode" -> "network_mode"
func ConvertCamelCaseToUnderScore(s string) string {
	camelRegex := regexp.MustCompile("(^[^A-Z]*|[A-Z]*)([A-Z][^A-Z]+|$)")

	var chars []string
	for _, sub := range camelRegex.FindAllStringSubmatch(s, -1) {
		if sub[1] != "" {
			chars = append(chars, sub[1])
		}
		if sub[2] != "" {
			chars = append(chars, sub[2])
		}
	}
	return strings.ToLower(strings.Join(chars, "_"))
}
