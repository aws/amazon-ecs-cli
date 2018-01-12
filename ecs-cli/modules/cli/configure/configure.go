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

	"github.com/sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const defaultConfigName = "default"

func fieldEmpty(field string, flagName string) error {
	if field == "" {
		return fmt.Errorf("%s can not be empty", flagName)
	}
	return nil
}

// Migrate converts the old ini formatted config to a new YAML formatted config
func Migrate(context *cli.Context) error {
	oldConfig := &config.CLIConfig{}
	dest, err := config.NewDefaultDestination()
	if err != nil {
		return errors.Wrap(err, "Error reading old configuration file.")
	}

	// Check if config file is YAML; if it is, no need to migrate.
	// This is needed because the ini library does not throw parse errors
	clusterConfig, err := config.ReadClusterFile(config.ConfigFilePath(dest))
	if err == nil && clusterConfig.Version != "" {
		// Is a new style config file
		logrus.Errorf("No need to migrate; found a YAML formatted configuration file at %s", config.ConfigFilePath(dest))
		return nil
	}

	iniReadWriter, err := config.NewINIReadWriter(dest)
	if err != nil {
		return errors.Wrap(err, "Error reading old configuration file.")
	}
	if err = iniReadWriter.GetConfig(oldConfig); err != nil {
		return errors.Wrap(err, "Error reading old configuration file.")
	}

	if oldConfig.CFNStackNamePrefix == flags.CFNStackNamePrefixDefaultValue {
		// if CFNStackName is default; don't store it.
		oldConfig.CFNStackName = ""
	} else {
		oldConfig.CFNStackName = oldConfig.CFNStackNamePrefix + oldConfig.Cluster
	}

	if !context.Bool(flags.ForceFlag) {
		if err = migrateWarning(*oldConfig); err != nil {
			return err
		}
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		input := scanner.Text()
		if !strings.HasPrefix(input, "y") && !strings.HasPrefix(input, "Y") {
			logrus.Info("Aborting Migration.")
			return nil
		}
	}

	cluster := &config.Cluster{
		Cluster:                  oldConfig.Cluster,
		Region:                   oldConfig.Region,
		CFNStackName:             oldConfig.CFNStackName,
		ComposeServiceNamePrefix: oldConfig.ComposeServiceNamePrefix,
	}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}
	if err = rdwr.SaveCluster(defaultConfigName, cluster); err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}
	if oldConfig.AWSSecretKey != "" {
		profile := &config.Profile{
			AWSAccessKey: oldConfig.AWSAccessKey,
			AWSSecretKey: oldConfig.AWSSecretKey,
		}
		if err = rdwr.SaveProfile(defaultConfigName, profile); err != nil {
			return errors.Wrap(err, "Error saving profile")
		}
	}

	logrus.Info("Migrated ECS CLI configuration.")
	return nil
}

// Cluster is the callback for ConfigureCommand (cluster).
func Cluster(context *cli.Context) error {
	region := context.String(flags.RegionFlag)
	if err := fieldEmpty(region, flags.RegionFlag); err != nil {
		return err
	}
	clusterProfileName := context.String(flags.ConfigNameFlag)
	if err := fieldEmpty(clusterProfileName, flags.ConfigNameFlag); err != nil {
		return err
	}
	cluster := context.String(flags.ClusterFlag)
	if err := fieldEmpty(cluster, flags.ClusterFlag); err != nil {
		return err
	}

	launchType := context.String(flags.DefaultLaunchTypeFlag)
	if err := config.ValidateLaunchType(launchType); err != nil {
		return err
	}

	cfnStackName := context.String(flags.CFNStackNameFlag)
	composeServiceNamePrefix := context.String(flags.ComposeServiceNamePrefixFlag)

	clusterConfig := &config.Cluster{
		Cluster:                  cluster,
		Region:                   region,
		CFNStackName:             cfnStackName,
		ComposeServiceNamePrefix: composeServiceNamePrefix,
		DefaultLaunchType:        launchType,
	}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}
	if err = rdwr.SaveCluster(clusterProfileName, clusterConfig); err != nil {
		return errors.Wrap(err, "Error saving cluster configuration")
	}

	logrus.Infof("Saved ECS CLI cluster configuration %s.", clusterProfileName)
	return nil
}

// Profile is the callback for Configure Profile subcommands.
func Profile(context *cli.Context) error {
	secretKey := context.String(flags.SecretKeyFlag)
	if err := fieldEmpty(secretKey, flags.SecretKeyFlag); err != nil {
		return err
	}
	accessKey := context.String(flags.AccessKeyFlag)
	if err := fieldEmpty(accessKey, flags.AccessKeyFlag); err != nil {
		return err
	}
	sessionToken := context.String(flags.SessionTokenFlag)
	profileName := context.String(flags.ProfileNameFlag)
	if err := fieldEmpty(profileName, flags.ProfileNameFlag); err != nil {
		return err
	}
	profile := &config.Profile{
		AWSAccessKey:    accessKey,
		AWSSecretKey:    secretKey,
		AWSSessionToken: sessionToken,
	}

	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error saving profile")
	}
	if err = rdwr.SaveProfile(profileName, profile); err != nil {
		return errors.Wrap(err, "Error saving profile")
	}

	logrus.Infof("Saved ECS CLI profile configuration %s.", profileName)
	return nil
}

// DefaultProfile is the callback for Configure Profile Default subcommand.
func DefaultProfile(context *cli.Context) error {
	profileName := context.String(flags.ProfileNameFlag)
	if err := fieldEmpty(profileName, flags.ProfileNameFlag); err != nil {
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
	clusterName := context.String(flags.ConfigNameFlag)
	if err := fieldEmpty(clusterName, flags.ConfigNameFlag); err != nil {
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
