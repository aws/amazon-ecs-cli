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
	app "github.com/aws/amazon-ecs-cli/ecs-cli/modules"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local"
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
		Usage:  "Create a Compose file from an ECS task definition and run it.",
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
				Name:  flagName(flags.TaskDefinitionCompose),
				Usage: flagDescription(flags.TaskDefinitionCompose, psCmdName),
			},
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
		Usage:  "Stop and remove a running ECS task.",
		Action: local.Down,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  flagName(flags.TaskDefinitionCompose),
				Usage: flagDescription(flags.TaskDefinitionCompose, downCmdName),
			},
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
		flags.TaskDefinitionRemote:  flags.TaskDefinitionRemote + ",r",
		flags.Output:                flags.Output + ",o",
		flags.JSON:                  flags.JSON,
		flags.All:                   flags.All,
	}
	return m[longName]
}

func flagDescription(longName, cmdName string) string {
	m := map[string]map[string]string{
		flags.TaskDefinitionCompose: {
			upCmdName:   "The Compose file `name` of a task definition to run.",
			psCmdName:   "List containers created from the Compose file `name`.",
			downCmdName: "Stop and remove containers from the Compose file `name`.",
		},
		flags.TaskDefinitionFile: {
			createCmdName: "The file `name` of a task definition json to convert. If not specified, defaults to task-definition.json.",
			upCmdName:     "The file `name` of a task definition json to convert and run. If not specified, defaults to task-definition.json.",
			psCmdName:     "List all running containers matching the task definition file `name`.",
			downCmdName:   "Stop and remove all running containers matching the task definition file `name`.",
		},
		flags.TaskDefinitionRemote: {
			createCmdName: "The `arnOrFamily` of a task definition to convert.",
			upCmdName:     "The `arnOrFamily` of a task definition to convert and run.",
			psCmdName:     "List all running containers matching the task definition `arnOrFamily`.",
			downCmdName:   "Stop and remove all running containers matching the task definition `arnOrFamily`.",
		},
		flags.Output: {
			createCmdName: "The Compose file `name` to write to. If not specified, defaults to docker-compose.local.yml.",
			upCmdName:     "The Compose file `name` to write to. If not specified, defaults to docker-compose.local.yml.",
		},
		flags.JSON: {
			psCmdName: "Output in JSON format.",
		},
		flags.All: {
			psCmdName:   "List all running local ECS task containers.",
			downCmdName: "Stop and remove all running containers.",
		},
	}
	return m[longName][cmdName]
}
