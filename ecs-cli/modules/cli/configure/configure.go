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

	clusterConfig := &config.Cluster{Cluster: cluster, Region: region}

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
