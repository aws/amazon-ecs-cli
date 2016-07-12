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

package app

import (
	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/codegangsta/cli"
)

// ProjectFactory is an interface that surfaces a function to create ECS Compose Project (intended to make mocking easy in tests)
type ProjectFactory interface {
	Create(cliContext *cli.Context, isService bool) (ecscompose.Project, error)
}

// projectFactory implements ProjectFactory interface
type projectFactory struct {
}

// NewProjectFactory returns an instance of ProjectFactory implementation
func NewProjectFactory() ProjectFactory {
	return projectFactory{}
}

// Create is a factory function that creates and configures ECS Compose project using the supplied command line arguments
func (projectFactory projectFactory) Create(cliContext *cli.Context, isService bool) (ecscompose.Project, error) {
	// creates and populates the ecs context
	ecsContext := &ecscompose.Context{}
	if err := projectFactory.populateContext(ecsContext, cliContext); err != nil {
		return nil, err
	}
	ecsContext.IsService = isService

	// creates and initializes project using the context
	project := ecscompose.NewProject(ecsContext)

	// load the configs
	if err := projectFactory.loadProject(project); err != nil {
		return nil, err
	}
	return project, nil
}

// populateContext sets the required CLI arguments to the context
func (projectFactory projectFactory) populateContext(ecsContext *ecscompose.Context, cliContext *cli.Context) error {
	// populate CLI context
	populate(ecsContext, cliContext)
	ecsContext.CLIContext = cliContext

	// reads and sets the parameters (required to create ECS Service Client) from the cli context to ecs context
	rdwr, err := config.NewReadWriter()
	if err != nil {
		utils.LogError(err, "Error loading config")
		return err
	}
	params, err := config.NewCliParams(cliContext, rdwr)
	if err != nil {
		utils.LogError(err, "Unable to create an instance of ECSParams given the cli context")
		return err
	}
	ecsContext.ECSParams = params

	// populate libcompose context
	if err = projectFactory.populateLibcomposeContext(ecsContext); err != nil {
		return err
	}

	return nil
}

// populateLibcomposeContext sets the required Libcompose lookup utilities to the context
func (projectFactory projectFactory) populateLibcomposeContext(ecsContext *ecscompose.Context) error {
	envLookup, err := utils.GetDefaultEnvironmentLookup()
	if err != nil {
		return err
	}
	ecsContext.EnvironmentLookup = envLookup

	resourceLookup, err := utils.GetDefaultResourceLookup()
	if err != nil {
		return err
	}
	ecsContext.ResourceLookup = resourceLookup
	return nil
}

// loadProject opens the project by loading configs
func (projectFactory projectFactory) loadProject(project ecscompose.Project) error {
	err := project.Parse()
	if err != nil {
		utils.LogError(err, "Unable to open ECS Compose Project")
	}
	return err
}
