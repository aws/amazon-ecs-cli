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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

func fieldEmpty(field string, flagName string) error {
	if field == "" {
		return fmt.Errorf("%s can not be empty.", flagName)
	}
	return nil
}

// Migrate is the callback for the Migrate command
func Migrate(context *cli.Context) error {
	oldConfig := &config.CLIConfig{}
	dest, err := config.NewDefaultDestination()
	if err != nil {
		return errors.Wrap(err, "Error reading old config file.")
	}
	iniReadWriter, err := config.NewINIReadWriter(dest)
	if err != nil {
		return errors.Wrap(err, "Error reading old config file.")
	}
	err = iniReadWriter.GetConfig(oldConfig)
	if err != nil {
		return errors.Wrap(err, "Error reading old config file.")
	}

	if oldConfig.AWSProfile != "" {
		logrus.Warnf("Storing AWS Profile in the config is no longer supported. Please use the %s flag inline in commands instead.", command.ProfileFlag)
	}

	if context.Bool(command.ForceFlag) {
		migrateWarning(oldConfig)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		if !strings.HasPrefix(input, "y") && !strings.HasPrefix(input, "Y") {
			logrus.Info("Aborting Migration.")
			return nil
		}
	}

	cluster := &config.Cluster{Cluster: oldConfig.Cluster, Region: oldConfig.Region, CFNStackName: oldConfig.CFNStackName, ComposeServiceNamePrefix: oldConfig.ComposeServiceNamePrefix}
	profile := &config.Profile{AWSAccessKey: oldConfig.AWSAccessKey, AWSSecretKey: oldConfig.AWSProfile}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}
	if err = rdwr.SaveCluster("default", cluster); err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}
	if err = rdwr.SaveProfile("default", profile); err != nil {
		return errors.Wrap(err, "Error saving profile")
	}

	return nil
}

// Cluster is the callback for ConfigureCommand (cluster).
func Cluster(context *cli.Context) error {
	region := context.String(command.RegionFlag)
	if err := fieldEmpty(region, command.RegionFlag); err != nil {
		return err
	}
	clusterProfileName := context.String(command.ConfigNameFlag)
	if err := fieldEmpty(clusterProfileName, command.ConfigNameFlag); err != nil {
		return err
	}
	cluster := context.String(command.ClusterFlag)
	if err := fieldEmpty(cluster, command.ClusterFlag); err != nil {
		return err
	}

	cfnStackName := context.String(command.CFNStackNameFlag)
	composeServiceNamePrefix := context.String(command.ComposeServiceNamePrefixFlag)

	clusterConfig := &config.Cluster{Cluster: cluster, Region: region, CFNStackName: cfnStackName, ComposeServiceNamePrefix: composeServiceNamePrefix}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}
	if err = rdwr.SaveCluster(clusterProfileName, clusterConfig); err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}

	return nil
}

// Profile is the callback for Configure Profile subcommand.
func Profile(context *cli.Context) error {
	secretKey := context.String(command.SecretKeyFlag)
	if err := fieldEmpty(secretKey, command.SecretKeyFlag); err != nil {
		return err
	}
	accessKey := context.String(command.AccessKeyFlag)
	if err := fieldEmpty(accessKey, command.AccessKeyFlag); err != nil {
		return err
	}
	profileName := context.String(command.ProfileNameFlag)
	if err := fieldEmpty(profileName, command.ProfileNameFlag); err != nil {
		return err
	}
	profile := &config.Profile{AWSAccessKey: accessKey, AWSSecretKey: secretKey}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error saving profile")
	}
	if err = rdwr.SaveProfile(profileName, profile); err != nil {
		return errors.Wrap(err, "Error saving profile")
	}

	return nil
}

// DefaultProfile is the callback for Configure Profile Default subcommand.
func DefaultProfile(context *cli.Context) error {
	profileName := context.String(command.ProfileNameFlag)
	if err := fieldEmpty(profileName, command.ProfileNameFlag); err != nil {
		return err
	}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error setting default config")
	}
	if err = rdwr.SetDefaultProfile(profileName); err != nil {
		return errors.Wrap(err, "Error setting default config")
	}

	return nil
}

// DefaultCluster is the callback for Configure Cluster Default subcommand.
func DefaultCluster(context *cli.Context) error {
	clusterName := context.String(command.ConfigNameFlag)
	if err := fieldEmpty(clusterName, command.ConfigNameFlag); err != nil {
		return err
	}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error setting default config")
	}
	if err = rdwr.SetDefaultCluster(clusterName); err != nil {
		return errors.Wrap(err, "Error setting default config")
	}

	return nil
}
