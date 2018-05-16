package project

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	"github.com/sirupsen/logrus"
)

func (p *ecsProject) parseV1V2() (*[]containerconfig.ContainerConfig, error) {
	logrus.Debug("Parsing v1/2 project...")

	// libcompose.Project#Parse populates project information based on its
	// context. It sets up the name, the composefile and the composebytes
	// (the composefile content). This is where libcompose ServiceConfigs
	// and VolumeConfigs gets loaded.
	if err := p.Project.Parse(); err != nil {
		return nil, err
	}

	volumeConfigs := p.Project.VolumeConfigs
	volumes, err := utils.ConvertToVolumes(volumeConfigs)
	if err != nil {
		return nil, err
	}

	context := &p.Context().Context
	serviceConfigs := p.Project.ServiceConfigs

	// convert ServiceConfigs to ContainerConfigs
	containerConfigs := []containerconfig.ContainerConfig{}
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

func convertV1V2ToContainerConfig(context *project.Context, serviceName string, volumes *utils.Volumes, service *config.ServiceConfig) (*containerconfig.ContainerConfig, error) {

	environment := utils.ConvertToKeyValuePairs(context, service.Environment, serviceName)

	extraHosts, err := utils.ConvertToExtraHosts(service.ExtraHosts)
	if err != nil {
		return nil, err
	}

	logConfiguration, err := utils.ConvertToLogConfiguration(service)
	if err != nil {
		return nil, err
	}

	memory := utils.ConvertToMemoryInMB(int64(service.MemLimit))
	memoryReservation := utils.ConvertToMemoryInMB(int64(service.MemReservation))

	mountPoints, err := utils.ConvertToMountPoints(service.Volumes, volumes)
	if err != nil {
		return nil, err
	}

	portMappings, err := utils.ConvertToPortMappings(serviceName, service.Ports)
	if err != nil {
		return nil, err
	}

	shmSize := utils.ConvertToMemoryInMB(int64(service.ShmSize))

	tmpfs, err := utils.ConvertToTmpfs(service.Tmpfs)
	if err != nil {
		return nil, err
	}

	ulimits, err := utils.ConvertToULimits(service.Ulimits)
	if err != nil {
		return nil, err
	}

	volumesFrom, err := utils.ConvertToVolumesFrom(service.VolumesFrom)
	if err != nil {
		return nil, err
	}

	outputConfig := &containerconfig.ContainerConfig{
		Name:                  serviceName,
		CapAdd:                service.CapAdd,
		CapDrop:               service.CapDrop,
		Command:               service.Command,
		CPU:                   int64(service.CPUShares),
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
