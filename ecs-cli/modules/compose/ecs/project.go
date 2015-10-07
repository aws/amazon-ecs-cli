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

package ecs

import (
	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
)

// Project is the starting point for the compose app to interact with and issue commands
// It acts as a blanket for the context and entities created as a part of this compose project
type Project interface {
	Name() string
	Parse() error

	Context() *Context
	ServiceConfigs() map[string]*libcompose.ServiceConfig
	Entity() ProjectEntity

	// commands
	Create() error
	Start() error
	Up() error
	Info() (libcompose.InfoSet, error)
	Run(commandOverrides map[string]string) error
	Scale(count int) error
	Stop() error
	Down() error
}

// ecsProject struct is an implementation of Project.
type ecsProject struct {
	context *Context

	// placeholder for compose yaml configurations map [serviceName -> composeConfigurations]
	serviceConfigs map[string]*libcompose.ServiceConfig

	// TODO: track a map of entities [taskDefinition -> Entity]
	// 1 task definition for every disjoint set of containers in the compose file
	entity ProjectEntity
}

// NewProject creates a new instance of the ECS Compose Project
func NewProject(context *Context) Project {
	p := &ecsProject{
		context:        context,
		serviceConfigs: make(map[string]*libcompose.ServiceConfig),
	}

	if context.IsService {
		p.entity = NewService(context)
	} else {
		p.entity = NewTask(context)
	}

	context.setProject(p)
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
func (p *ecsProject) ServiceConfigs() map[string]*libcompose.ServiceConfig {
	return p.serviceConfigs
}

// Entity returns the project entity that operates on the compose file and integrates with ecs
func (p *ecsProject) Entity() ProjectEntity {
	return p.entity
}

// Parse reads the context and sets appropriate project fields
func (p *ecsProject) Parse() error {
	context := p.context
	if err := context.open(); err != nil {
		return err
	}

	if err := p.load(context.Context); err != nil {
		return err
	}

	return nil
}

// load parses the compose yml and transforms into task definition
func (p *ecsProject) load(context libcompose.Context) error {
	logrus.Debug("Parsing the compose yaml...")
	configs, err := utils.UnmarshalComposeConfig(context)
	if err != nil {
		return err
	}
	p.serviceConfigs = configs

	logrus.Debug("Transforming yaml to task definition...")
	taskDefinition, err := utils.ConvertToTaskDefinition(context, p.serviceConfigs)
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

func (p *ecsProject) Info() (libcompose.InfoSet, error) {
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
