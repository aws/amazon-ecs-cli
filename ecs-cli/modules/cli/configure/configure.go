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

func fieldEmpty(flag string, context *cli.Context) error {
	field := context.String(flag)
	if field == "" {
		return fmt.Errorf("%s can not be empty.", flag)
	}
	return nil
}

// Cluster is the callback for ConfigureCommand (cluster).
func Cluster(context *cli.Context) error {
	if err := fieldEmpty(command.RegionFlag, context); err != nil {
		return err
	}
	if err := fieldEmpty(command.ConfigNameFlag, context); err != nil {
		return err
	}
	if err := fieldEmpty(command.ClusterFlag, context); err != nil {
		return err
	}

	region := context.String(command.RegionFlag)
	clusterProfileName := context.String(command.ConfigNameFlag)
	cluster := context.String(command.ClusterFlag)

	clusterConfig := &config.Cluster{Cluster: cluster, Region: region}

	// modify the profile config file
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
	if err := fieldEmpty(command.SecretKeyFlag, context); err != nil {
		return err
	}
	if err := fieldEmpty(command.AccessKeyFlag, context); err != nil {
		return err
	}
	if err := fieldEmpty(command.ProfileNameFlag, context); err != nil {
		return err
	}

	secretKey := context.String(command.SecretKeyFlag)
	profileName := context.String(command.ProfileNameFlag)
	accessKey := context.String(command.AccessKeyFlag)

	profile := &config.Profile{AWSAccessKey: accessKey, AWSSecretKey: secretKey}

	// modify the profile config file
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
	// validate field not empty
	if err := fieldEmpty(command.ProfileNameFlag, context); err != nil {
		return err
	}

	// get relevant fields
	profileName := context.String(command.ProfileNameFlag)

	// modify the profile config file
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
	// validate field not empty
	if err := fieldEmpty(command.ConfigNameFlag, context); err != nil {
		return err
	}
	// get relevant fields
	clusterName := context.String(command.ConfigNameFlag)

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		return errors.Wrap(err, "Error setting default config")
	}
	if err = rdwr.SetDefaultCluster(clusterName); err != nil {
		return errors.Wrap(err, "Error setting default config")
	}

	return nil
}
