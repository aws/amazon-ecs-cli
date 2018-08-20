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
		Usage:        "Generates AWS Secrets Manager secrets and an IAM Task Execution Role for use in an ECS Task Definition.",
		Action:       regcreds.Up,
		Flags:        regcredsUpFlags(),
		OnUsageError: flags.UsageErrorFactory("up"),
	}
}

// TODO: add rest of flags as functionality implemented
func regcredsUpFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  flags.InputFileFlag,
			Usage: "Specifies the name of the file containing registry credentials.",
		},
	}
}
