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

package command

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/codegangsta/cli"
)

// ConfigureCommand defines subcommand to configure the ecs-cli.
func ConfigureCommand() cli.Command {
	return cli.Command{
		Name:   "configure",
		Usage:  "Configures your AWS credentials, the AWS region to use, and the ECS cluster name to use with the Amazon ECS CLI. The resulting configuration is stored in the ~/.ecs/config file.",
		Action: configure,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: ecscli.RegionFlag + ", r",
				Usage: fmt.Sprintf(
					"Specifies the AWS region to use. If the " + ecscli.AwsRegionEnvVar + " environment variable is set when ecs-cli configure is run, then the AWS region is set to the value of that environment variable.",
				),
				EnvVar: ecscli.AwsRegionEnvVar,
			},
			cli.StringFlag{
				Name: ecscli.AccessKeyFlag,
				Usage: fmt.Sprintf(
					"Specifies the AWS access key to use. If the AWS_ACCESS_KEY_ID environment variable is set when ecs-cli configure is run, then the AWS access key ID is set to the value of that environment variable.",
				),
				EnvVar: "AWS_ACCESS_KEY_ID",
			},
			cli.StringFlag{
				Name: ecscli.SecretKeyFlag,
				Usage: fmt.Sprintf(
					"Specifies the AWS secret key to use. If the AWS_SECRET_ACCESS_KEY environment variable is set when ecs-cli configure is run, then the AWS secret access key is set to the value of that environment variable.",
				),
				EnvVar: "AWS_SECRET_ACCESS_KEY",
			},
			cli.StringFlag{
				Name: ecscli.ProfileFlag + ", p",
				Usage: fmt.Sprintf(
					"Specifies your AWS credentials with an existing named profile from ~/.aws/credentials. If the AWS_PROFILE environment variable is set when ecs-cli configure is run, then the AWS named profile is set to the value of that environment variable.",
				),
				EnvVar: "AWS_PROFILE",
			},
			cli.StringFlag{
				Name: ecscli.ClusterFlag + ", c",
				Usage: fmt.Sprintf(
					"Specifies the ECS cluster name to use. If the cluster does not exist, it is created when you try to add resources to it with the ecs-cli up command.",
				),
				// TODO: Override behavior for all ecs-cli commands : CommandLineFlags > EnvVar > ConfigFile > Defaults
				// Commenting it now to avoid user misunderstanding the behavior of this env var with other ecs-cli commands
				// EnvVar: "ECS_CLUSTER",
			},
			cli.StringFlag{
				Name:  ecscli.ComposeProjectNamePrefixFlag,
				Value: ecscli.ComposeProjectNamePrefixDefaultValue,
				Usage: fmt.Sprintf(
					"[Optional] Specifies the prefix added to an ECS task definition created from a compose file. Format <prefix><project-name>.",
				),
			},
			cli.StringFlag{
				Name:  ecscli.ComposeServiceNamePrefixFlag,
				Value: ecscli.ComposeServiceNamePrefixDefaultValue,
				Usage: fmt.Sprintf(
					"[Optional] Specifies the prefix added to an ECS service created from a compose file. Format <prefix><project-name>.",
				),
			},
			cli.StringFlag{
				Name:  ecscli.CFNStackNamePrefixFlag,
				Value: ecscli.CFNStackNamePrefixDefaultValue,
				Usage: fmt.Sprintf(
					"[Optional] Specifies the prefix added to the AWS CloudFormation stack created on ecs-cli up. Format <prefix><cluster-name>.",
				),
			},
		},
	}
}

// configure is the callback for ConfigureCommand.
func configure(context *cli.Context) {
	ecsConfig, err := createECSConfigFromCli(context)
	if err != nil {
		logrus.Error("Error initializing: ", err)
		return
	}
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error initializing: ", err)
		return
	}
	err = saveConfig(ecsConfig, rdwr, rdwr.Destination)
	if err != nil {
		logrus.Error("Error initializing: ", err)
	}
}

// createECSConfigFromCli creates a new CliConfig object from the CLI context.
// It reads CLI flags to validate the ecs-cli config fields.
func createECSConfigFromCli(context *cli.Context) (*config.CliConfig, error) {
	accessKey := context.String(ecscli.AccessKeyFlag)
	secretKey := context.String(ecscli.SecretKeyFlag)
	region := context.String(ecscli.RegionFlag)
	profile := context.String(ecscli.ProfileFlag)
	cluster := context.String(ecscli.ClusterFlag)

	if cluster == "" {
		return nil, fmt.Errorf("Missing required argument '%s'", ecscli.ClusterFlag)
	}

	// ONLY allow for profile OR access keys to be specified
	isProfileSpecified := profile != ""
	isAccessKeySpecified := accessKey != "" || secretKey != ""
	if isProfileSpecified && isAccessKeySpecified {
		return nil, fmt.Errorf("Both AWS Access/Secret Keys and Profile were provided; only one of the two can be specified")
	}

	ecsConfig := config.NewCliConfig(cluster)
	ecsConfig.AwsProfile = profile
	ecsConfig.AwsAccessKey = accessKey
	ecsConfig.AwsSecretKey = secretKey
	ecsConfig.Region = region

	ecsConfig.ComposeProjectNamePrefix = context.String(ecscli.ComposeProjectNamePrefixFlag)
	ecsConfig.ComposeServiceNamePrefix = context.String(ecscli.ComposeServiceNamePrefixFlag)
	ecsConfig.CFNStackNamePrefix = context.String(ecscli.CFNStackNamePrefixFlag)

	return ecsConfig, nil
}

// saveConfig does the actual configuration setup. This isolated method is useful for testing.
func saveConfig(ecsConfig *config.CliConfig, rdwr config.ReadWriter, dest *config.Destination) error {
	err := rdwr.ReadFrom(ecsConfig)
	if err != nil {
		return err
	}

	err = rdwr.Save(dest)
	if err != nil {
		return err
	}
	logrus.Infof("Saved ECS CLI configuration for cluster (%s)", ecsConfig.Cluster)
	return nil
}
