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

package composeCommand

import (
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose"
	composeFactory "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/factory"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/compose/service"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
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
//
// List containers in or view details of the project:
//   ecs-cli compose ps          : calls ECS.ListTasks (running and stopped) filtered with Task group: this project
//
// Modify containers
//   ecs-cli compose scale       : calls ECS.RunTask/StopTask based on the count
//   ecs-cli compose run         : calls ECS.RunTask with overrides
//
// Stop and delete the project
//   ecs-cli compose stop        : calls ECS.StopTask and ECS deletes them (rm)
//* --------------------------------------------------- */

const (
	composeFileNameDefaultValue         = "docker-compose.yml"
	composeOverrideFileNameDefaultValue = "docker-compose.override.yml"
	ecsParamsFileNameDefaultValue       = "ecs-params.yml"
	containerNameFlag                   = "name"
)

// ComposeCommand provides a list of commands that operate on docker-compose.yml file and are integrated to run on ECS.
// This list encompasses docker-compose commands
func ComposeCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "compose",
		Usage:  "Executes docker-compose-style commands on an ECS cluster.",
		Before: ecscli.BeforeApp,
		Flags:  append(composeFlags(), flags.OptionalConfigFlags()...),
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
			serviceCommand.ServiceCommand(factory),
		},
	}
}

// commonComposeFlags lists the flags used by the compose subcommand
func composeFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.VerboseFlag + ",debug",
			Usage: "Increase the verbosity of command output to aid in diagnostics.",
		},
		cli.StringSliceFlag{
			Name:   flags.ComposeFileNameFlag + ",f",
			Usage:  "Specifies one or more Docker compose files to use. Defaults to " + composeFileNameDefaultValue + " file, and an optional " + composeOverrideFileNameDefaultValue + " file.",
			Value:  &cli.StringSlice{},
			EnvVar: "COMPOSE_FILE",
		},
		cli.StringFlag{
			Name:   flags.ProjectNameFlag + ",p",
			Usage:  "Specifies the project name to use. Defaults to the current directory name.",
			EnvVar: "COMPOSE_PROJECT_NAME",
		},
		cli.StringFlag{
			Name:  flags.TaskRoleArnFlag,
			Usage: "[Optional] Specifies the short name or full Amazon Resource Name (ARN) of the IAM role that containers in this task can assume. All containers in this task are granted the permissions that are specified in this role.",
		},
		cli.StringFlag{
			Name:  flags.ECSParamsFileNameFlag,
			Usage: "[Optional] Specifies ecs-params file to use. Defaults to " + ecsParamsFileNameDefaultValue + " file, if one exists.",
		},
	}
}

func createCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "create",
		Usage:        "Creates an ECS task definition from your compose file. Note that we do not recommend using plain text environment variables for sensitive information, such as credential data.",
		Action:       compose.WithProject(factory, compose.ProjectCreate, false),
		Flags:        append(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag()),
		OnUsageError: flags.UsageErrorFactory("create"),
	}
}

func psCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "ps",
		Aliases:      []string{"list"},
		Usage:        "Lists all the containers in your cluster that were started by the compose project.",
		Action:       compose.WithProject(factory, compose.ProjectPs, false),
		Flags:        flags.OptionalConfigFlags(),
		OnUsageError: flags.UsageErrorFactory("ps"),
	}
}

func upCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "up",
		Usage:        "Creates an ECS task definition from your compose file (if it does not already exist) and runs one instance of that task on your cluster (a combination of create and start).",
		Action:       compose.WithProject(factory, compose.ProjectUp, false),
		Flags:        append(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag()),
		OnUsageError: flags.UsageErrorFactory("up"),
	}
}

func startCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "start",
		Usage:        "Starts a single task from the task definition created from your compose file.",
		Action:       compose.WithProject(factory, compose.ProjectStart, false),
		Flags:        append(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag()),
		OnUsageError: flags.UsageErrorFactory("start"),
	}
}

func runCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "run",
		Usage:        "Starts all containers overriding commands with the supplied one-off commands for the containers.",
		ArgsUsage:    "[CONTAINER_NAME] [\"COMMAND ...\"] [CONTAINER_NAME] [\"COMMAND ...\"] ...",
		Action:       compose.WithProject(factory, compose.ProjectRun, false),
		Flags:        flags.OptionalConfigFlags(),
		OnUsageError: flags.UsageErrorFactory("run"),
	}
}

func stopCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "stop",
		Aliases:      []string{"down"},
		Usage:        "Stops all the running tasks created by the compose project.",
		Action:       compose.WithProject(factory, compose.ProjectStop, false),
		Flags:        flags.OptionalConfigFlags(),
		OnUsageError: flags.UsageErrorFactory("stop"),
	}
}

func scaleCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "scale",
		Usage:        "ecs-cli compose scale [count] - scales the number of running tasks to the specified count.",
		Action:       compose.WithProject(factory, compose.ProjectScale, false),
		Flags:        append(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag()),
		OnUsageError: flags.UsageErrorFactory("scale"),
	}
}
