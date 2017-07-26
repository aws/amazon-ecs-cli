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

package configure

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/urfave/cli"
)

// Configure is the callback for ConfigureCommand.
func Configure(context *cli.Context) {
	ecsConfig, err := createECSConfigFromCLI(context)
	if err != nil {
		logrus.Fatal("Error initializing: ", err)
	}
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error initializing: ", err)
	}
	err = saveConfig(ecsConfig, rdwr)
	if err != nil {
		logrus.Fatal("Error initializing: ", err)
	}
}

// createECSConfigFromCLI creates a new CliConfig object from the CLI context.
// It reads CLI flags to validate the ecs-cli config fields.
func createECSConfigFromCLI(context *cli.Context) (*config.CLIConfig, error) {
	accessKey := context.String(command.AccessKeyFlag)
	secretKey := context.String(command.SecretKeyFlag)
	region := context.String(command.RegionFlag)
	profile := context.String(command.ProfileFlag)
	cluster := context.String(command.ClusterFlag)

	if cluster == "" {
		return nil, fmt.Errorf("Missing required argument '%s'", command.ClusterFlag)
	}

	// ONLY allow for profile OR access keys to be specified
	isProfileSpecified := profile != ""
	isAccessKeySpecified := accessKey != "" || secretKey != ""
	if isProfileSpecified && isAccessKeySpecified {
		return nil, fmt.Errorf("Both AWS Access/Secret Keys and Profile were provided; only one of the two can be specified")
	}

	ecsConfig := config.NewCLIConfig(cluster)
	ecsConfig.AWSProfile = profile
	ecsConfig.AWSAccessKey = accessKey
	ecsConfig.AWSSecretKey = secretKey
	ecsConfig.Region = region

	ecsConfig.ComposeProjectNamePrefix = context.String(command.ComposeProjectNamePrefixFlag)
	ecsConfig.ComposeServiceNamePrefix = context.String(command.ComposeServiceNamePrefixFlag)
	ecsConfig.CFNStackNamePrefix = context.String(command.CFNStackNamePrefixFlag)

	return ecsConfig, nil
}

// saveConfig does the actual configuration setup. This isolated method is useful for testing.
func saveConfig(ecsConfig *config.CLIConfig, rdwr config.ReadWriter) error {

	err := rdwr.Save(ecsConfig)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"cluster": ecsConfig.Cluster,
	}).Info("Saved ECS CLI configuration for")
	return nil
}
