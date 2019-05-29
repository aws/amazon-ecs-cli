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

// LocalCommand provides a list of commands that operate on a task-definition
// file (accepted formats: JSON, YAML, CloudFormation).
func LocalCommand() cli.Command {
	return cli.Command{
		Name:   "local",
		Usage:  "Run your ECS tasks locally.",
		Before: app.BeforeApp,
		Flags:  flags.OptionalRegionAndProfileFlags(),
		Subcommands: []cli.Command{
			createCommand(),
			upCommand(),
			stopCommand(),
			psCommand(),
		},
	}
}

func createCommand() cli.Command {
	return cli.Command{
		Name:   "create",
		Usage:  "Create a Compose file from an ECS task definition.",
		Before: app.BeforeApp,
		Action: local.Create,
		Flags:  createFlags(),
	}
}

// TODO This is a placeholder function used to test the ECS local network configuration.
func upCommand() cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Create a Compose file from an ECS task definition and run it.",
		Action: local.Up,
	}
}

// TODO This is a placeholder function used to test the teardown of the ECS local network.
func stopCommand() cli.Command {
	return cli.Command{
		Name:   "stop",
		Usage:  "Stop a running local ECS task.",
		Action: local.Stop,
	}
}

func psCommand() cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "List locally running ECS task containers.",
		Action: local.Ps,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  flags.JsonFlag,
				Usage: "Output in JSON format.",
			},
		},
	}
}

func createFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  flags.TaskDefinitionFileFlag + ",f",
			Usage: "The file name of the task definition to convert.",
		},
		cli.StringFlag{
			Name:  flags.TaskDefinitionArnFlag + ",a",
			Usage: "The ARN of the task definition to convert.",
		},
		cli.StringFlag{
			Name:  flags.LocalOutputFlag + ",o",
			Usage: "The name of the file to write to. If not specified, defaults to docker-compose.local.yml",
		},
	}
}
