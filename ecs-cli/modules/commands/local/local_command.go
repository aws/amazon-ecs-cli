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

// Package localCommand defines the subcommands for local workflows
package localCommand

import (
	"fmt"

	app "github.com/aws/amazon-ecs-cli/ecs-cli/modules"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local"
	project "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/localproject"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

const (
	createCmdName = "create"
	upCmdName     = "up"
	psCmdName     = "ps"
	downCmdName   = "down"
)

// LocalCommand provides a list of commands that operate on a task-definition
// file (accepted formats: JSON, YAML, CloudFormation).
func LocalCommand() cli.Command {
	return cli.Command{
		Name:   "local",
		Usage:  "Run your ECS tasks locally.",
		Before: app.BeforeApp,
		Flags:  flags.AppendFlags(flags.OptECSProfileFlag(), flags.OptAWSProfileFlag(), flags.OptRegionFlag()),
		Subcommands: []cli.Command{
			createCommand(),
			upCommand(),
			downCommand(),
			psCommand(),
		},
	}
}

func createCommand() cli.Command {
	return cli.Command{
		Name:   createCmdName,
		Usage:  "Create a Compose file from an ECS task definition.",
		Before: app.BeforeApp,
		Action: local.Create,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionFile),
				Usage: flagDescription(flags.TaskDefinitionFile, createCmdName),
			},
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionRemote),
				Usage: flagDescription(flags.TaskDefinitionRemote, createCmdName),
			},
			cli.StringFlag{
				Name:  flagName(flags.Output),
				Usage: flagDescription(flags.Output, createCmdName),
			},
		},
	}
}

func upCommand() cli.Command {
	return cli.Command{
		Name:   upCmdName,
		Usage:  fmt.Sprintf("Run containers locally from an ECS Task Definition. NOTE: Creates a docker-compose file in current directory and a %s if one doesn't exist. ", network.EcsLocalNetworkName),
		Action: local.Up,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionCompose),
				Usage: flagDescription(flags.TaskDefinitionCompose, upCmdName),
			},
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionFile),
				Usage: flagDescription(flags.TaskDefinitionFile, upCmdName),
			},
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionRemote),
				Usage: flagDescription(flags.TaskDefinitionRemote, upCmdName),
			},
			cli.StringFlag{
				Name:  flagName(flags.Output),
				Usage: flagDescription(flags.Output, upCmdName),
			},
			cli.StringSliceFlag{
				Name:  flagName(flags.ComposeOverride),
				Usage: flagDescription(flags.ComposeOverride, upCmdName),
			},
		},
	}
}

func psCommand() cli.Command {
	return cli.Command{
		Name:   psCmdName,
		Usage:  "List locally running ECS task containers.",
		Action: local.Ps,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionFile),
				Usage: flagDescription(flags.TaskDefinitionFile, psCmdName),
			},
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionRemote),
				Usage: flagDescription(flags.TaskDefinitionRemote, psCmdName),
			},
			cli.BoolFlag{
				Name:  flagName(flags.All),
				Usage: flagDescription(flags.All, psCmdName),
			},
			cli.BoolFlag{
				Name:  flagName(flags.JSON),
				Usage: flagDescription(flags.JSON, psCmdName),
			},
		},
	}
}

func downCommand() cli.Command {
	return cli.Command{
		Name:   downCmdName,
		Usage:  fmt.Sprintf("Stop and remove a running ECS task. NOTE: Removes the %s if it has no more running tasks. ", network.EcsLocalNetworkName),
		Action: local.Down,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionFile),
				Usage: flagDescription(flags.TaskDefinitionFile, downCmdName),
			},
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionRemote),
				Usage: flagDescription(flags.TaskDefinitionRemote, downCmdName),
			},
			cli.BoolFlag{
				Name:  flagName(flags.All),
				Usage: flagDescription(flags.All, downCmdName),
			},
		},
	}
}

func flagName(longName string) string {
	m := map[string]string{
		flags.TaskDefinitionCompose: flags.TaskDefinitionCompose + ",c",
		flags.TaskDefinitionFile:    flags.TaskDefinitionFile + ",f",
		flags.TaskDefinitionRemote:  flags.TaskDefinitionRemote + ",t",
		flags.Output:                flags.Output + ",o",
		flags.ComposeOverride:       flags.ComposeOverride,
		flags.JSON:                  flags.JSON,
		flags.All:                   flags.All,
	}
	return m[longName]
}

func flagDescription(longName, cmdName string) string {
	m := map[string]map[string]string{
		flags.TaskDefinitionCompose: {
			upCmdName: "Specifies the filename `value` that contains the Docker Compose content to run locally.",
		},
		flags.TaskDefinitionFile: {
			createCmdName: fmt.Sprintf("Specifies the filename `value` that contains the task definition JSON to convert to a Docker Compose file. If one is not specified, the ECS CLI will look for %s.", project.LocalInFileName),
			upCmdName:     fmt.Sprintf("Specifies the filename `value` containing the task definition JSON to convert and run locally.  If one is not specified, the ECS CLI will look for %s.", project.LocalInFileName),
			psCmdName:     fmt.Sprintf("Lists all running containers matching the task definition filename `value`. If one is not specified, the ECS CLI will list containers started with the task definition filename %s.", project.LocalInFileName),
			downCmdName:   fmt.Sprintf("Stops and removes all running containers matching the task definition filename `value`. If one is not specified, the ECS CLI removes all running containers matching the task definition filename %s.", project.LocalInFileName),
		},
		flags.TaskDefinitionRemote: {
			createCmdName: "Specifies the full Amazon Resource Name (ARN) or family:revision `value` of the task definition to convert to a Docker Compose file. If you specify a task definition family without a revision, the latest revision is used.",
			upCmdName:     "Specifies the full Amazon Resource Name (ARN) or family:revision `value` of the task definition to convert and run locally. If you specify a task definition family without a revision, the latest revision is used.",
			psCmdName:     "Lists all running containers matching the task definition Amazon Resource Name (ARN) or family:revision `value`. If you specify a task definition family without a revision, the latest revision is used.",
			downCmdName:   "Stops and removes all running containers matching the task definition Amazon Resource Name (ARN) or family:revision `value`. If you specify a task definition family without a revision, the latest revision is used.",
		},
		flags.ComposeOverride: {
			upCmdName: "Specifies the local Docker Compose override filename `value` to use.",
		},
		flags.Output: {
			createCmdName: fmt.Sprintf("Specifies the local filename `value` to write the Docker Compose file to. If one is not specified, the default is %s.", project.LocalOutDefaultFileName),
			upCmdName:     fmt.Sprintf("Specifies the local filename `value` to write the Docker Compose file to. If one is not specified, the default is %s.", project.LocalOutDefaultFileName),
		},
		flags.JSON: {
			psCmdName: "Sets the output to JSON format.",
		},
		flags.All: {
			psCmdName:   "Lists all locally running containers.",
			downCmdName: "Stops and removes all locally running containers.",
		},
	}
	return m[longName][cmdName]
}
