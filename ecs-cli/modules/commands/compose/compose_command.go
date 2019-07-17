// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Package composeCommand contains the subcommands for compose workflows
package composeCommand

import (
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose"
	composeFactory "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/factory"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/compose/service"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/usage"
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
)

// ComposeCommand provides a list of commands that operate on docker-compose.yml file and are integrated to run on ECS.
// This list encompasses docker-compose commands
func ComposeCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:   "compose",
		Usage:  usage.Compose,
		Before: ecscli.BeforeApp,
		Flags:  flags.AppendFlags(composeFlags(), flags.DebugFlag(), flags.OptionalConfigFlags()),
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
		cli.StringFlag{
			Name:  flags.RegistryCredsFileNameFlag,
			Usage: "[Optional] Specifies the ecs-registry-creds file to use. Defaults to latest 'ecs-registry-creds' output file, if one exists.",
		},
	}
}

func createCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "create",
		Usage:        usage.ComposeCreate,
		Action:       compose.WithProject(factory, compose.ProjectCreate, false),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag(), resourceTagsFlag(false)),
		OnUsageError: flags.UsageErrorFactory("create"),
	}
}

func psCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "ps",
		Aliases:      []string{"list"},
		Usage:        usage.ComposePs,
		Action:       compose.WithProject(factory, compose.ProjectPs, false),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), flags.OptionalDesiredStatusFlag()),
		OnUsageError: flags.UsageErrorFactory("ps"),
	}
}

func upCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "up",
		Usage:        usage.ComposeUp,
		Action:       compose.WithProject(factory, compose.ProjectUp, false),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag(), flags.OptionalForceUpdateFlag(), resourceTagsFlag(true), disableECSManagedTagsFlag()),
		OnUsageError: flags.UsageErrorFactory("up"),
	}
}

func startCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "start",
		Usage:        usage.ComposeStart,
		Action:       compose.WithProject(factory, compose.ProjectStart, false),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag(), resourceTagsFlag(true), disableECSManagedTagsFlag()),
		OnUsageError: flags.UsageErrorFactory("start"),
	}
}

func runCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "run",
		Usage:        usage.ComposeRun,
		ArgsUsage:    "[CONTAINER_NAME] [\"COMMAND ...\"] [CONTAINER_NAME] [\"COMMAND ...\"] ...",
		Action:       compose.WithProject(factory, compose.ProjectRun, false),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), resourceTagsFlag(true), disableECSManagedTagsFlag()),
		OnUsageError: flags.UsageErrorFactory("run"),
	}
}

func stopCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "stop",
		Aliases:      []string{"down"},
		Usage:        usage.ComposeStop,
		Action:       compose.WithProject(factory, compose.ProjectStop, false),
		Flags:        flags.OptionalConfigFlags(),
		OnUsageError: flags.UsageErrorFactory("stop"),
	}
}

func scaleCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "scale",
		Usage:        usage.ComposeScale,
		Action:       compose.WithProject(factory, compose.ProjectScale, false),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), resourceTagsFlag(true), disableECSManagedTagsFlag()),
		OnUsageError: flags.UsageErrorFactory("scale"),
	}
}

func resourceTagsFlag(runTasks bool) []cli.Flag {
	usage := "[Optional] Specify resource tags for your Task Definition. Specify tags in the format 'key1=value1,key2=value2,key3=value3'."
	if runTasks {
		usage = "[Optional] Specify resource tags for your ECS Tasks and Task Definition. Specify tags in the format 'key1=value1,key2=value2,key3=value3'."
	}
	return []cli.Flag{
		cli.StringFlag{
			Name:  flags.ResourceTagsFlag,
			Usage: usage,
		},
	}
}

func disableECSManagedTagsFlag() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.DisableECSManagedTagsFlag,
			Usage: "[Optional] Disable ECS Managed Tags (A Cluster name tag will not be automatically added to tasks).",
		},
	}
}
