package project

import (
	"github.com/docker/libcompose/project"
	"github.com/sirupsen/logrus"
)

func (p *ecsProject) parseV1V2() {
	libProject := project.NewProject(&p.ecsContext.Context, nil, nil)
	libProject.Parse()

	logrus.Debug("Parsing v1/2 project...")
	//TODO: convert parsed project.ServiceConfigs to ContainerConfigs
}
