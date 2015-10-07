// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
		Usage:  "Configures your AWS credentials, the AWS region to use, and the EC2 Container Service cluster name to use with the ECS CLI.",
		Action: configure,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name: ecscli.RegionFlag + ", r",
				Usage: fmt.Sprintf(
					"Specify the AWS Region to use.",
				),
				EnvVar: "AWS_REGION",
			},
			cli.StringFlag{
				Name: ecscli.AccessKeyFlag,
				Usage: fmt.Sprintf(
					"Specify the AWS access key to use.",
				),
				EnvVar: "AWS_ACCESS_KEY_ID",
			},
			cli.StringFlag{
				Name: ecscli.SecretKeyFlag,
				Usage: fmt.Sprintf(
					"Specify the AWS secret key to use.",
				),
				EnvVar: "AWS_SECRET_ACCESS_KEY",
			},
			cli.StringFlag{
				Name: ecscli.ProfileFlag + ", p",
				Usage: fmt.Sprintf(
					"Specify your AWS credentials with an existing named profile from ~/.aws/credentials.",
				),
				EnvVar: "AWS_PROFILE",
			},
			cli.StringFlag{
				Name: ecscli.ClusterFlag + ", c",
				Usage: fmt.Sprintf(
					"Specify the ECS cluster name to use. If the cluster does not exist, it will be created.",
				),
				// TODO: Override behavior for all ecs-cli commands : CommandLineFlags > EnvVar > ConfigFile > Defaults
				// Commenting it now to avoid user misunderstanding the behavior of this env var with other ecs-cli commands
				// EnvVar: "ECS_CLUSTER",
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

	if profile != "" {
		if accessKey != "" || secretKey != "" {
			return nil, fmt.Errorf("Missing required credentials. Specify either '%s' with the name of an existing named profile in ~/.aws/credentials, or your AWS credentials with '%s' and '%s'", ecscli.ProfileFlag, ecscli.AccessKeyFlag, ecscli.SecretKeyFlag)
		}
	}

	ecsConfig := config.NewCliConfig(cluster)
	ecsConfig.AwsProfile = profile
	ecsConfig.AwsAccessKey = accessKey
	ecsConfig.AwsSecretKey = secretKey
	ecsConfig.Region = region

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
