package project

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/docker/libcompose/project"
	"github.com/sirupsen/logrus"
)

func (p *ecsProject) parseV1V2() ([]containerconfig.ContainerConfig, error) {
	libProject := project.NewProject(&p.ecsContext.Context, nil, nil)
	libProject.Parse()

	logrus.Debug("Parsing v1/2 project...")

	//TODO: convert parsed project.ServiceConfigs to ContainerConfigs
	return []containerconfig.ContainerConfig{}, nil
}
