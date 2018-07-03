package project

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/sirupsen/logrus"
)

func (p *ecsProject) parseV1V2() (*[]adapter.ContainerConfig, error) {
	logrus.Debug("Parsing v1/2 project...")

	libcomposeProject := project.NewProject(&p.ecsContext.Context, nil, nil)
	// libcompose.Project#Parse populates project information based on its
	// context. It sets up the name, the composefile and the composebytes
	// (the composefile content). This is where libcompose ServiceConfigs
	// and VolumeConfigs gets loaded.
	if err := libcomposeProject.Parse(); err != nil {
		return nil, err
	}
	logger.LogUnsupportedProjectFields(libcomposeProject)

	volumeConfigs := libcomposeProject.VolumeConfigs
	volumes, err := adapter.ConvertToVolumes(volumeConfigs)
	if err != nil {
		return nil, err
	}
	p.volumes = volumes

	context := &p.Context().Context
	serviceConfigs := libcomposeProject.ServiceConfigs

	// convert ServiceConfigs to ContainerConfigs
	containerConfigs := []adapter.ContainerConfig{}
	for _, serviceName := range serviceConfigs.Keys() {
		serviceConfig, ok := serviceConfigs.Get(serviceName)
		if !ok {
			return nil, fmt.Errorf("Couldn't get service with name=[%s]", serviceName)
		}

		containerConfig, err := convertV1V2ToContainerConfig(context, serviceName, volumes, serviceConfig)
		if err != nil {
			return nil, err
		}
		containerConfigs = append(containerConfigs, *containerConfig)
	}

	return &containerConfigs, nil
}

func convertV1V2ToContainerConfig(context *project.Context, serviceName string, volumes *adapter.Volumes, service *config.ServiceConfig) (*adapter.ContainerConfig, error) {
	logger.LogUnsupportedV1V2ServiceConfigFields(serviceName, service)

	devices, err := adapter.ConvertToDevices(service.Devices)
	if err != nil {
		return nil, err
	}

	environment := adapter.ConvertToKeyValuePairs(context, service.Environment, serviceName)

	extraHosts, err := adapter.ConvertToExtraHosts(service.ExtraHosts)
	if err != nil {
		return nil, err
	}

	logConfiguration, err := adapter.ConvertToLogConfiguration(service)
	if err != nil {
		return nil, err
	}

	memory := adapter.ConvertToMemoryInMB(int64(service.MemLimit))
	memoryReservation := adapter.ConvertToMemoryInMB(int64(service.MemReservation))

	// Validate memory and memory reservation
	if memory == 0 && memoryReservation != 0 {
		memory = memoryReservation
	}

	mountPoints, err := adapter.ConvertToMountPoints(service.Volumes, volumes)
	if err != nil {
		return nil, err
	}

	portMappings, err := adapter.ConvertToPortMappings(serviceName, service.Ports)
	if err != nil {
		return nil, err
	}

	shmSize := adapter.ConvertToMemoryInMB(int64(service.ShmSize))

	tmpfs, err := adapter.ConvertToTmpfs(service.Tmpfs)
	if err != nil {
		return nil, err
	}

	ulimits, err := adapter.ConvertToULimits(service.Ulimits)
	if err != nil {
		return nil, err
	}

	volumesFrom, err := adapter.ConvertToVolumesFrom(service.VolumesFrom)
	if err != nil {
		return nil, err
	}

	outputConfig := &adapter.ContainerConfig{
		Name:                  serviceName,
		CapAdd:                service.CapAdd,
		CapDrop:               service.CapDrop,
		Command:               service.Command,
		CPU:                   int64(service.CPUShares),
		Devices:               devices,
		DNSSearchDomains:      service.DNSSearch,
		DNSServers:            service.DNS,
		DockerLabels:          aws.StringMap(service.Labels),
		DockerSecurityOptions: service.SecurityOpt,
		Entrypoint:            service.Entrypoint,
		Environment:           environment,
		ExtraHosts:            extraHosts,
		Hostname:              service.Hostname, // only set if not blank?
		Image:                 service.Image,
		Links:                 service.Links,
		LogConfiguration:      logConfiguration,
		Memory:                memory,
		MemoryReservation:     memoryReservation,
		MountPoints:           mountPoints,
		PortMappings:          portMappings,
		Privileged:            service.Privileged,
		ReadOnly:              service.ReadOnly,
		ShmSize:               shmSize,
		Tmpfs:                 tmpfs,
		Ulimits:               ulimits,
		VolumesFrom:           volumesFrom,
		User:                  service.User,
		WorkingDirectory:      service.WorkingDir,
	}

	return outputConfig, nil
}
