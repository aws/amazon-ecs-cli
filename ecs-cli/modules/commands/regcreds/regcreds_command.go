// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package regcredsCommand

import (
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/regcreds"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

// RegistryCredsCommand provides a list of commands that facilitate use of private registry credentials with ECS.
func RegistryCredsCommand() cli.Command {
	return cli.Command{
		Name:   "registry-creds",
		Usage:  "Facilitates the creation and use of private registry credentials within ECS.",
		Before: ecscli.BeforeApp,
		Flags:  flags.OptionalRegionAndProfileFlags(),
		Subcommands: []cli.Command{
			upCommand(),
		},
	}
}

func upCommand() cli.Command {
	return cli.Command{
		Name:         "up",
		Usage:        "Uses a YAML input file to generate AWS Secrets Manager secrets and an IAM Task Execution Role for use in an ECS Task Definition.",
		Action:       regcreds.Up,
		Flags:        flags.AppendFlags(flags.OptionalRegionAndProfileFlags(), regcredsUpFlags()),
		OnUsageError: flags.UsageErrorFactory("up"),
	}
}

func regcredsUpFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.UpdateExistingSecretsFlag,
			Usage: "[Optional] Specifies whether existing secrets should be updated with new credential values.",
		},
		cli.StringFlag{
			Name:  flags.RoleNameFlag,
			Usage: "The name to use for the new task execution role. If the role already exists, new policies will be attached to the existing role.",
		},
		cli.BoolFlag{
			Name:  flags.NoRoleFlag,
			Usage: "[Optional] If specified, no task execution role will be created.",
		},
		cli.BoolFlag{
			Name:  flags.NoOutputFileFlag,
			Usage: "[Optional] If specified, no output file for use with 'compose' will be created.",
		},
		cli.StringFlag{
			Name:  flags.OutputDirFlag,
			Usage: "[Optional] The directory where the output file should be created. If none specified, file will be created in the current working directory.",
		},
	}
}
