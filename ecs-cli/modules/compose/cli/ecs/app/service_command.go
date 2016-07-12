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
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs"
	"github.com/codegangsta/cli"
)

//* ----------------- COMPOSE PROJECT with ECS Service ----------------- */
// Note: A project is scoped to a single compose yaml with multiple containers defined
// and today, 1 compose.yml has 1:1 mapping with a task definition and a ECS Service.
// TODO: Split single compose to disjoint task definitions, so they can be run/scaled independently
//
// ---- LIFECYCLE ----
// Create and Start a project with service:
//   ecs-cli compose service create      : creates ECS.CreateTaskDefinition or gets from FS cache and ECS.CreateService with desiredCount=0
//   ecs-cli compose service start       : invokes ECS.UpdateService with desiredCount=1
//   ecs-cli compose service up          : compose service create ; compose service start. If the compose yml was changed, it updates the service with new task definition
// List containers in or view details of the project:
//   ecs-cli compose service ps          : calls ECS.ListTasks of this service
// Modify containers
//   ecs-cli compose service scale       : calls ECS.UpdateService with new count
// Stop and delete the project
//   ecs-cli compose service stop        : calls ECS.UpdateService with count=0
//   ecs-cli compose service down        : calls ECS.DeleteService
//* -------------------------------------------------------------------- */

// serviceCommand provides a list of commands that operate on docker-compose.yml file
// and are integrated to run on ECS as a service
func serviceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Manage Amazon ECS services with docker-compose-style commands on an ECS cluster.",
		Subcommands: []cli.Command{
			createServiceCommand(factory),
			startServiceCommand(factory),
			upServiceCommand(factory),
			psServiceCommand(factory),
			scaleServiceCommand(factory),
			stopServiceCommand(factory),
			rmServiceCommand(factory),
		},
	}
}

func createServiceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   ecs.CreateServiceCommandName,
		Usage:  "Creates an ECS service from your compose file. The service is created with a desired count of 0, so no containers are started by this command. Note that we do not recommend using plain text environment variables for sensitive information, such as credential data.",
		Action: WithProject(factory, ProjectCreate, true),
		Flags:  deploymentConfigFlags(true),
	}
}

func startServiceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "Starts one copy of each of the containers on the created ECS service. This command updates the desired count of the service to 1.",
		Action: WithProject(factory, ProjectStart, true),
	}
}

func upServiceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Creates an ECS service from your compose file (if it does not already exist) and runs one instance of that task on your cluster (a combination of create and start). This command updates the desired count of the service to 1.",
		Action: WithProject(factory, ProjectUp, true),
		Flags:  deploymentConfigFlags(true),
	}
}

func psServiceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:    "ps",
		Aliases: []string{"list"},
		Usage:   "Lists all the containers in your cluster that belong to the service created with the compose project.",
		Action:  WithProject(factory, ProjectPs, true),
	}
}

func scaleServiceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "scale",
		Usage:  "ecs-cli compose service scale [count] - scales the desired count of the service to the specified count",
		Action: WithProject(factory, ProjectScale, true),
		Flags:  deploymentConfigFlags(false),
	}
}

func stopServiceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "stop",
		Usage:  "Stops the running tasks that belong to the service created with the compose project. This command updates the desired count of the service to 0.",
		Action: WithProject(factory, ProjectStop, true),
	}
}

func rmServiceCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:    "rm",
		Aliases: []string{"delete", "down"},
		Usage:   "Updates the desired count of the service to 0 and then deletes the service.",
		Action:  WithProject(factory, ProjectDown, true),
	}
}

func deploymentConfigFlags(specifyDefaults bool) []cli.Flag {
	maxPercentUsageString := "[Optional] Specifies the upper limit (as a percentage of the service's desiredCount) of the number of running tasks that can be running in a service during a deployment."
	minHealthyPercentUsageString := "[Optional] Specifies the lower limit (as a percentage of the service's desiredCount) of the number of running tasks that must remain running and healthy in a service during a deployment."
	if specifyDefaults {
		maxPercentUsageString += fmt.Sprintf(" Defaults to %d.", ecs.DeploymentMaxPercentDefaultValue)
		minHealthyPercentUsageString += fmt.Sprintf(" Defaults to %d.", ecs.DeploymentMinHealthyPercentDefaultValue)
	}
	return []cli.Flag{
		cli.StringFlag{
			Name:  ecs.DeploymentMaxPercentFlag,
			Usage: maxPercentUsageString,
		},
		cli.StringFlag{
			Name:  ecs.DeploymentMinHealthyPercentFlag,
			Usage: minHealthyPercentUsageString,
		},
	}
}
