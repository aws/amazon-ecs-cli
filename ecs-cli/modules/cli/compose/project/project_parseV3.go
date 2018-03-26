package project

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/sirupsen/logrus"
)

func (p *ecsProject) parseV3() ([]containerconfig.ContainerConfig, error) {
	logrus.Debug("Parsing v3 project...")

	// TODO: parse v3, convert ServiceConfigs to ContainerConfigs
	return []containerconfig.ContainerConfig{}, nil
}
