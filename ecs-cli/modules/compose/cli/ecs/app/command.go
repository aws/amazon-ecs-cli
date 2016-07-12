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
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs"
	"github.com/codegangsta/cli"
)

// Flag and command names used by the cli.
const (
	composeFileNameFlag         = "file"
	composeFileNameDefaultValue = "docker-compose.yml"
	containerNameFlag           = "name"
	projectNameFlag             = "project-name"
)

//* ----------------- COMPOSE PROJECT ----------------- */
// Note: A project is scoped to a single compose yaml with multiple containers defined
// and today, 1 compose.yml has 1:1 mapping with a task definition.
// TODO: Split single compose to disjoint task definitions, so they can be run/scaled independently
//
// ---- LIFECYCLE ----
// Create and Start a project:
//   ecs-cli compose create      : creates ECS.TaskDefinition or gets from FS cache
//   ecs-cli compose start       : invokes ECS.RunTask if count(running tasks) == 0
//   ecs-cli compose up          : compose create ; compose start and does a deployment of new compose yml if changes were found
// List containers in or view details of the project:
//   ecs-cli compose ps          : calls ECS.ListTasks (running and stopped) with startedBy: this project
// Modify containers
//   ecs-cli compose scale       : calls ECS.RunTask/StopTask based on the count
//   ecs-cli compose run         : calls ECS.RunTask with overrides
// Stop and delete the project
//   ecs-cli compose stop        : calls ECS.StopTask and ECS deletes them (rm)
//* --------------------------------------------------- */

// ComposeCommand provides a list of commands that operate on docker-compose.yml file and are integrated to run on ECS.
// This list encompasses docker-compose commands
func ComposeCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "compose",
		Usage:  "Executes docker-compose-style commands on an ECS cluster.",
		Before: ecscli.BeforeApp,
		Subcommands: []cli.Command{
			createCommand(factory),
			psCommand(factory),
			runCommand(factory),
			scaleCommand(factory),
			startCommand(factory),
			stopCommand(factory),
			upCommand(factory),
			// ----- Unsupported/Unimplemented COMMANDS -----
			// build, pull, logs, port, restart, rm, kill

			// ECS Service sub command
			// TODO, should honor restart policy in the compose yaml and create ECS Services accordingly
			serviceCommand(factory),
		},
		Flags: commonComposeFlags(),
	}
}

// commonComposeFlags lists the flags used by the compose subcommand
func commonComposeFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  ecscli.VerboseFlag + ",debug",
			Usage: "Increase the verbosity of command output to aid in diagnostics.",
		},
		cli.StringFlag{
			Name:   composeFileNameFlag + ",f",
			Usage:  "Specifies the Docker compose file to use. Defaults to " + composeFileNameDefaultValue + " file.",
			Value:  composeFileNameDefaultValue,
			EnvVar: "COMPOSE_FILE",
		},
		cli.StringFlag{
			Name:   ecscompose.ProjectNameFlag + ",p",
			Usage:  "Specifies the project name to use. Defaults to the current directory name.",
			EnvVar: "COMPOSE_PROJECT_NAME",
		},
	}
}

// populate updates the specified ecs context based on command line arguments and subcommands.
func populate(ecsContext *ecscompose.Context, cliContext *cli.Context) {
	// TODO: Support multiple compose files
	ecsContext.ComposeFiles = []string{cliContext.GlobalString(composeFileNameFlag)}
	ecsContext.ProjectName = cliContext.GlobalString(ecscompose.ProjectNameFlag)
}

func createCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "Creates an ECS task definition from your compose file. Note that we do not recommend using plain text environment variables for sensitive information, such as credential data.",
		Action: WithProject(factory, ProjectCreate, false),
	}
}

func psCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:    "ps",
		Aliases: []string{"list"},
		Usage:   "Lists all the containers in your cluster that were started by the compose project.",
		Action:  WithProject(factory, ProjectPs, false),
	}
}

func upCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Creates an ECS task definition from your compose file (if it does not already exist) and runs one instance of that task on your cluster (a combination of create and start).",
		Action: WithProject(factory, ProjectUp, false),
	}
}

func startCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "start",
		Usage:  "Starts a single task from the task definition created from your compose file.",
		Action: WithProject(factory, ProjectStart, false),
	}
}

func runCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name: "run",
		Usage: "ecs-cli compose run [containerName] [command] [containerName] [command] ..." +
			"- starts all containers overriding commands with the supplied one-off commands for the containers.",
		Action: WithProject(factory, ProjectRun, false),
	}
}

func stopCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:    "stop",
		Aliases: []string{"down"},
		Usage:   "Stops all the running tasks created by the compose project.",
		Action:  WithProject(factory, ProjectStop, false),
	}
}

func scaleCommand(factory ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "scale",
		Usage:  "ecs-cli compose scale [count] - scales the number of running tasks to the specified count.",
		Action: WithProject(factory, ProjectScale, false),
	}
}
