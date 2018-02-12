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

	"github.com/sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/configure"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

type configureAction func(*cli.Context) error

func errorLogger(action configureAction) func(context *cli.Context) {
	return func(context *cli.Context) {
		err := action(context)
		if err != nil {
			logrus.Fatal(err)
		}
	}
}

func configureProfileCommand() cli.Command {
	return cli.Command{
		Name:         "profile",
		Usage:        "Stores a single profile.",
		Action:       errorLogger(configure.Profile),
		Flags:        configureProfileFlags(),
		OnUsageError: flags.UsageErrorFactory("profile"),
		Subcommands: []cli.Command{
			defaultProfileCommand(),
		},
	}
}

func defaultProfileCommand() cli.Command {
	return cli.Command{
		Name:         "default",
		Usage:        "Sets the default profile.",
		Action:       errorLogger(configure.DefaultProfile),
		Flags:        configureDefaultProfileFlags(),
		OnUsageError: flags.UsageErrorFactory("default"),
	}
}

func defaultClusterCommand() cli.Command {
	return cli.Command{
		Name:         "default",
		Usage:        "Sets the default cluster config.",
		Action:       errorLogger(configure.DefaultCluster),
		Flags:        configureDefaultClusterFlags(),
		OnUsageError: flags.UsageErrorFactory("default"),
	}
}

func migrateCommand() cli.Command {
	return cli.Command{
		Name:   "migrate",
		Usage:  "Migrates a legacy ECS CLI configuration file to the current YAML format.",
		Action: errorLogger(configure.Migrate),
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name: flags.ForceFlag,
				Usage: fmt.Sprintf(
					"[Optional] Omits the interactive description and confirmation step that normally occurs during the configuration file migration process.",
				),
			},
		},
		OnUsageError: flags.UsageErrorFactory("migrate"),
	}
}

// ConfigureCommand configure command help
func ConfigureCommand() cli.Command {
	return cli.Command{
		Name:   "configure",
		Usage:  "Stores a single cluster configuration.",
		Action: errorLogger(configure.Cluster),
		Flags:  configureFlags(),
		Subcommands: []cli.Command{
			configureProfileCommand(),
			defaultClusterCommand(),
			migrateCommand(),
		},
		OnUsageError: flags.UsageErrorFactory("configure"),
	}
}

func configureDefaultClusterFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: flags.ConfigNameFlag,
			Usage: fmt.Sprintf(
				"Specifies the name of the cluster configuration to use by default.",
			),
		},
	}
}

func configureDefaultProfileFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: flags.ProfileNameFlag,
			Usage: fmt.Sprintf(
				"Specifies the name of the profile to use by default.",
			),
		},
	}
}

func configureProfileFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: flags.AccessKeyFlag,
			Usage: fmt.Sprintf(
				"Specifies the AWS access key to use. The ECS CLI uses the value of your $AWS_ACCESS_KEY_ID environment variable if it is set.",
			),
			EnvVar: "AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name: flags.SecretKeyFlag,
			Usage: fmt.Sprintf(
				"Specifies the AWS secret key to use. The ECS CLI uses the value of your $AWS_SECRET_ACCESS_KEY environment variable if it is set.",
			),
			EnvVar: "AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:  flags.ProfileNameFlag,
			Value: "default",
			Usage: fmt.Sprintf(
				"Specifies the profile name to use for this configuration.",
			),
		},
		cli.StringFlag{
			Name: flags.SessionTokenFlag,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the AWS session token to use. The ECS CLI uses the value of your $AWS_SESSION_TOKEN environment variable if it is set.",
			),
			EnvVar: "AWS_SESSION_TOKEN",
		},
	}
}

func configureFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: flags.ClusterFlag + ", c",
			Usage: fmt.Sprintf(
				"Specifies the ECS cluster name to use. If the cluster does not exist, it is created when you try to add resources to it with the ecs-cli up command.",
			),
			EnvVar: "ECS_CLUSTER",
		},
		cli.StringFlag{
			Name: flags.RegionFlag + ", r",
			Usage: fmt.Sprintf(
				"Specifies the AWS region to use. If the " + flags.AwsRegionEnvVar + " environment variable is set when ecs-cli configure is run, then the AWS region is set to the value of that environment variable.",
			),
			EnvVar: flags.AwsRegionEnvVar,
		},
		cli.StringFlag{
			Name:  flags.ConfigNameFlag,
			Value: "default",
			Usage: fmt.Sprintf(
				"Specifies the cluster configuration name to use for this configuration.",
			),
		},
		cli.StringFlag{
			Name: flags.ComposeServiceNamePrefixFlag,
			Usage: fmt.Sprintf(
				"[Deprecated] Specifies the prefix added to an ECS service created from a compose file. Format <prefix><project-name>. (defaults to empty)",
			),
		},
		cli.StringFlag{
			Name: flags.CFNStackNameFlag,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the name of AWS CloudFormation stack created on ecs-cli up. (default: \"amazon-ecs-cli-setup-<cluster-name>\")",
			),
		},
		cli.StringFlag{
			Name: flags.DefaultLaunchTypeFlag,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the type of tasks that you would like to run. Options: EC2 or FARGATE. Defaults to empty string if none provided.",
			),
		},
	}
}
