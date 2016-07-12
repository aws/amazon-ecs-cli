// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package ecs

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
)

// Project is the starting point for the compose app to interact with and issue commands
// It acts as a blanket for the context and entities created as a part of this compose project
type Project interface {
	Name() string
	Parse() error

	Context() *Context
	ServiceConfigs() *config.ServiceConfigs
	Entity() ProjectEntity

	// commands
	Create() error
	Start() error
	Up() error
	Info() (project.InfoSet, error)
	Run(commandOverrides map[string]string) error
	Scale(count int) error
	Stop() error
	Down() error
}

// ecsProject struct is an implementation of Project.
type ecsProject struct {
	project.Project

	context *Context

	// TODO: track a map of entities [taskDefinition -> Entity]
	// 1 task definition for every disjoint set of containers in the compose file
	entity ProjectEntity
}

// NewProject creates a new instance of the ECS Compose Project
func NewProject(context *Context) Project {
	libcomposeProject := project.NewProject(&context.Context, nil, nil)

	p := &ecsProject{
		context: context,
		Project: *libcomposeProject,
	}

	if context.IsService {
		p.entity = NewService(context)
	} else {
		p.entity = NewTask(context)
	}

	return p
}

// Name returns the name of the project
func (p *ecsProject) Name() string {
	return p.Context().Context.ProjectName
}

// Context returns the context of the project, which encompasses the cli configurations required to setup this project
func (p *ecsProject) Context() *Context {
	return p.context
}

// ServiceConfigs returns a map of Service Configuration loaded from compose yaml file
func (p *ecsProject) ServiceConfigs() *config.ServiceConfigs {
	return p.Project.ServiceConfigs
}

// Entity returns the project entity that operates on the compose file and integrates with ecs
func (p *ecsProject) Entity() ProjectEntity {
	return p.entity
}

// Parse reads the context and sets appropriate project fields
func (p *ecsProject) Parse() error {
	context := p.context

	// initialize the context and project entity fields
	if err := context.open(); err != nil {
		return err
	}
	if err := p.Entity().LoadContext(); err != nil {
		return err
	}

	if err := p.parseCompose(); err != nil {
		return err
	}

	return p.transformTaskDefinition()
}

// parseCompose loads and parses the compose yml files
func (p *ecsProject) parseCompose() error {
	// load the compose files using libcompose
	logrus.Debug("Parsing the compose yaml...")
	if err := p.Project.Parse(); err != nil {
		return err
	}

	// libcompose sanitizes the project name and removes any non alpha-numeric character.
	// The following undo-es that and sets the project name as user defined it.
	return p.context.SetProjectName()
}

// transformTaskDefinition converts the compose yml into task definition
func (p *ecsProject) transformTaskDefinition() error {
	context := p.context

	// convert to task definition
	logrus.Debug("Transforming yaml to task definition...")
	taskDefinitionName := utils.GetTaskDefinitionName(context.ECSParams.ComposeProjectNamePrefix, context.Context.ProjectName)
	taskDefinition, err := utils.ConvertToTaskDefinition(taskDefinitionName, &context.Context, p.ServiceConfigs())
	if err != nil {
		return err
	}
	p.entity.SetTaskDefinition(taskDefinition)
	return nil
}

//* ----------------- commands ----------------- */

func (p *ecsProject) Create() error {
	return p.entity.Create()
}

func (p *ecsProject) Start() error {
	return p.entity.Start()
}

func (p *ecsProject) Up() error {
	return p.entity.Up()
}

func (p *ecsProject) Info() (project.InfoSet, error) {
	return p.entity.Info(true)
}

func (p *ecsProject) Run(commandOverrides map[string]string) error {
	return p.entity.Run(commandOverrides)
}

func (p *ecsProject) Scale(count int) error {
	return p.entity.Scale(count)
}

func (p *ecsProject) Stop() error {
	return p.entity.Stop()
}

func (p *ecsProject) Down() error {
	return p.entity.Down()
}
