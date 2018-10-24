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

package project

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/service"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/task"
	"github.com/sirupsen/logrus"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcredio"
	"github.com/docker/libcompose/project"
)

// Project is the starting point for the compose app to interact with and issue commands
// It acts as a blanket for the ECS context and entities created as a part of this compose project
type Project interface {
	Name() string
	Parse() error

	Context() *context.ECSContext
	ContainerConfigs() []adapter.ContainerConfig // TODO change this to a pointer to slice?
	VolumeConfigs() *adapter.Volumes
	Entity() entity.ProjectEntity

	// commands
	Create() error
	Start() error
	Up() error
	Info() (project.InfoSet, error)
	Run(commandOverrides map[string][]string) error
	Scale(count int) error
	Stop() error
	Down() error
}

// ecsProject struct is an implementation of Project.
type ecsProject struct {
	containerConfigs []adapter.ContainerConfig
	volumes          *adapter.Volumes
	ecsContext       *context.ECSContext
	ecsRegistryCreds *regcredio.ECSRegistryCredsOutput

	// TODO: track a map of entities [taskDefinition -> Entity]
	// 1 task definition for every disjoint set of containers in the compose file
	entity entity.ProjectEntity
}

// NewProject creates a new instance of the ECS Compose Project
func NewProject(ecsContext *context.ECSContext) Project {
	p := &ecsProject{
		ecsContext:       ecsContext,
		containerConfigs: []adapter.ContainerConfig{},
		volumes:          adapter.NewVolumes(),
	}

	if ecsContext.IsService {
		p.entity = service.NewService(ecsContext)
	} else {
		p.entity = task.NewTask(ecsContext)
	}

	return p
}

// Name returns the name of the project
func (p *ecsProject) Name() string {
	return p.Context().Context.ProjectName
}

// Context returns the ecsContext of the project, which encompasses the cli configurations required to setup this project
func (p *ecsProject) Context() *context.ECSContext {
	return p.ecsContext
}

// VolumeConfigs returns a map of Volume Configuration loaded from compose yaml file
func (p *ecsProject) VolumeConfigs() *adapter.Volumes {
	return p.volumes
}

func (p *ecsProject) ContainerConfigs() []adapter.ContainerConfig {
	return p.containerConfigs
}

// Entity returns the project entity that operates on the compose file and integrates with ecs
func (p *ecsProject) Entity() entity.ProjectEntity {
	return p.entity
}

// Parse reads the ecsContext and sets appropriate project fields
func (p *ecsProject) Parse() error {
	ecsContext := p.ecsContext

	// initialize the ECS context and project entity fields
	if err := ecsContext.Open(); err != nil {
		return err
	}

	if err := p.Entity().LoadContext(); err != nil {
		return err
	}

	if err := p.parseCompose(); err != nil {
		return err
	}

	// Populates ecs-params onto project ecsContext
	if err := p.parseECSParams(); err != nil {
		return err
	}

	if err := p.parseECSRegistryCreds(); err != nil {
		return err
	}

	return p.transformTaskDefinition()
}

// parseCompose sets data from the compose files on the ecsProject, including ContainerConfigs and VolumeConfigs
func (p *ecsProject) parseCompose() error {
	logrus.Debug("Parsing the compose yaml...")

	// check for Compose version and call appropriate parsing function
	version, err := p.checkComposeVersion()
	if err != nil {
		return err
	}
	switch version {
	case "", "1", "1.0", "2", "2.0":
		configs, err := p.parseV1V2()
		if err != nil {
			return err
		}
		// TODO: set this in parseV1V2 itself?
		p.containerConfigs = *configs
	case "3", "3.0":
		configs, err := p.parseV3()
		if err != nil {
			return err
		}
		p.containerConfigs = *configs
	default:
		return fmt.Errorf("Unsupported Docker Compose version found: %s", version)
	}

	// libcompose sanitizes the project name and removes any non alpha-numeric character.
	// The following undoes that and sets the project name as user defined it.
	return p.ecsContext.SetProjectName()
}

// parseECSParams sets data from the ecs-params.yml file on the ecsProject.context
func (p *ecsProject) parseECSParams() error {
	logrus.Debug("Parsing the ecs-params yaml...")
	ecsParamsFileName := p.ecsContext.CLIContext.GlobalString(flags.ECSParamsFileNameFlag)
	ecsParams, err := utils.ReadECSParams(ecsParamsFileName)

	if err != nil {
		return err
	}

	p.ecsContext.ECSParams = ecsParams

	return nil
}

func (p *ecsProject) parseECSRegistryCreds() error {
	logrus.Debug("Parsing the ecs-registry-creds yaml...")
	registryCredsFileName := p.ecsContext.CLIContext.GlobalString(flags.RegistryCredsFileNameFlag)
	regCreds, err := regcredio.ReadCredsOutput(registryCredsFileName)
	if err != nil {
		return err
	}

	p.ecsRegistryCreds = regCreds

	return nil
}

// transformTaskDefinition converts the compose yml and ecs-params yml into an
// ECS task definition and loads it onto the project entity
func (p *ecsProject) transformTaskDefinition() error {
	ecsContext := p.ecsContext

	// convert to task definition
	logrus.Debug("Transforming yaml to task definition...")

	taskRoleArn := ecsContext.CLIContext.GlobalString(flags.TaskRoleArnFlag)
	requiredCompatibilities := ecsContext.CommandConfig.LaunchType
	taskDefinitionName := ecsContext.ProjectName

	convertParams := utils.ConvertTaskDefParams{
		TaskDefName:            taskDefinitionName,
		TaskRoleArn:            taskRoleArn,
		RequiredCompatibilites: requiredCompatibilities,
		Volumes:                p.VolumeConfigs(),
		ContainerConfigs:       p.ContainerConfigs(), // TODO Change to pointer on project?
		ECSParams:              ecsContext.ECSParams,
		ECSRegistryCreds:       p.ecsRegistryCreds,
	}

	taskDefinition, err := utils.ConvertToTaskDefinition(convertParams)

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

func (p *ecsProject) Run(commandOverrides map[string][]string) error {
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
