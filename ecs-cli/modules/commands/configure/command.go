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

package configureCommand

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/configure"
	flags "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/urfave/cli"
)

func ConfigureCommand() cli.Command {
	return cli.Command{
		Name:   "configure",
		Usage:  "Configures your AWS credentials, the AWS region to use, and the ECS cluster name to use with the Amazon ECS CLI. The resulting configuration is stored in the ~/.ecs/config file.",
		Action: configure.Configure,
		Flags:  configureFlags(),
	}
}

func configureFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: flags.RegionFlag + ", r",
			Usage: fmt.Sprintf(
				"Specifies the AWS region to use. If the " + flags.AwsRegionEnvVar + " environment variable is set when ecs-cli configure is run, then the AWS region is set to the value of that environment variable.",
			),
			EnvVar: flags.AwsRegionEnvVar,
		},
		cli.StringFlag{
			Name: flags.AccessKeyFlag,
			Usage: fmt.Sprintf(
				"Specifies the AWS access key to use. If the AWS_ACCESS_KEY_ID environment variable is set when ecs-cli configure is run, then the AWS access key ID is set to the value of that environment variable.",
			),
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name: flags.SecretKeyFlag,
			Usage: fmt.Sprintf(
				"Specifies the AWS secret key to use. If the AWS_SECRET_ACCESS_KEY environment variable is set when ecs-cli configure is run, then the AWS secret access key is set to the value of that environment variable.",
			),
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name: flags.ProfileFlag + ", p",
			Usage: fmt.Sprintf(
				"Specifies your AWS credentials with an existing named profile from ~/.aws/credentials. If the AWS_PROFILE environment variable is set when ecs-cli configure is run, then the AWS named profile is set to the value of that environment variable.",
			),
			EnvVar: "AWS_PROFILE",
		},
		cli.StringFlag{
			Name: flags.ClusterFlag + ", c",
			Usage: fmt.Sprintf(
				"Specifies the ECS cluster name to use. If the cluster does not exist, it is created when you try to add resources to it with the ecs-cli up command.",
			),
			// TODO: Override behavior for all ecs-cli commands : CommandLineFlags > EnvVar > ConfigFile > Defaults
			// Commenting it now to avoid user misunderstanding the behavior of this env var with other ecs-cli commands
			// EnvVar: "ECS_CLUSTER",
		},
		cli.StringFlag{
			Name:  flags.ComposeProjectNamePrefixFlag,
			Value: flags.ComposeProjectNamePrefixDefaultValue,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the prefix added to an ECS task definition created from a compose file. Format <prefix><project-name>.",
			),
		},
		cli.StringFlag{
			Name:  flags.ComposeServiceNamePrefixFlag,
			Value: flags.ComposeServiceNamePrefixDefaultValue,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the prefix added to an ECS service created from a compose file. Format <prefix><project-name>.",
			),
		},
		cli.StringFlag{
			Name:  flags.CFNStackNamePrefixFlag,
			Value: flags.CFNStackNamePrefixDefaultValue,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the prefix added to the AWS CloudFormation stack created on ecs-cli up. Format <prefix><cluster-name>.",
			),
		},
	}
}
