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

package factory

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/project"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	libcomposecommand "github.com/docker/libcompose/cli/command"
	"github.com/urfave/cli"
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
	ecsContext := &context.Context{}
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

// populateContext sets the required CLI arguments to the ECS context
func (projectFactory projectFactory) populateContext(ecsContext *context.Context, cliContext *cli.Context) error {
	/*
		Populate the following libcompose fields on the ECS context:
		 - ComposeFiles: reads from `--file` or `-f` flags. Defaults to
		 `docker-compose.yml` and `docker-compose.override.yml` if no flags are
		 specified.
		 - ProjectName: reads from `--project-name` or `-p` flags.
	*/
	libcomposecommand.Populate(&ecsContext.Context, cliContext)
	ecsContext.CLIContext = cliContext

	// reads and sets the parameters (required to create ECS Service
	// Client) from the cli context to ECS context
	rdwr, err := config.NewReadWriter()
	if err != nil {
		utils.LogError(err, "Error loading config")
		return err
	}
	params, err := config.NewCLIParams(cliContext, rdwr)
	if err != nil {
		utils.LogError(err, "Unable to create an instance of CLIParams given the cli context")
		return err
	}
	ecsContext.CLIParams = params

	// populate libcompose context
	if err = projectFactory.populateLibcomposeContext(ecsContext); err != nil {
		return err
	}

	return nil
}

// populateLibcomposeContext sets the required Libcompose lookup utilities on the ECS context
func (projectFactory projectFactory) populateLibcomposeContext(ecsContext *context.Context) error {
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
