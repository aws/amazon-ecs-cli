package project

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/sirupsen/logrus"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
)

func (p *ecsProject) parseV3() (*[]containerconfig.ContainerConfig, error) {
	logrus.Debug("Parsing v3 project...")

	// TODO: parse v3, convert ServiceConfigs to ContainerConfigs
	configDetails := types.ConfigDetails{}
	config, err := loader.Load(configDetails)
	if err != nil {
		return nil, err
	}

	for _, service := range config.Services {
		convertDockerToContainerConfig(service)
	}

	return &[]containerconfig.ContainerConfig{}, nil
}

func convertDockerToContainerConfig(serviceConfig types.ServiceConfig) {
	return
}
