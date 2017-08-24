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
	"errors"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/urfave/cli"
)

// fieldEmptywill logs an error message for empty fields
func fieldEmpty(flag string, context *cli.Context) bool {
	field := context.String(flag)
	if field == "" {
		logrus.Errorf("%s can not be empty.", flag)
		return true
	}
	return false
}

// Cluster is the callback for ConfigureCommand (cluster).
// Note: The error returned is only used in the unit tests
// This function is called by urfave/cli which does not check the error returned
func Cluster(context *cli.Context) error {
	if fieldEmpty(command.RegionFlag, context) || fieldEmpty(command.ConfigNameFlag, context) || fieldEmpty(command.ClusterFlag, context) {
		// fieldEmpty() will log the error; returned error is only used for unit tests
		return errors.New("a required field was empty")
	}

	region := context.String(command.RegionFlag)
	clusterProfileName := context.String(command.ConfigNameFlag)
	cluster := context.String(command.ClusterFlag)

	clusterConfig := &config.Cluster{Cluster: cluster, Region: region}

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error saving cluster configuration: ", err)
		return err
	}
	if err = rdwr.SaveCluster(clusterProfileName, clusterConfig); err != nil {
		logrus.Error("Error saving cluster configuration: ", err)
		return err
	}

	return nil
}

// Profile is the callback for Configure Profile subcommand.
// Note: The error returned is only used in the unit tests
// This function is called by urfave/cli which does not check the error returned
func Profile(context *cli.Context) error {
	// validate fields not empty
	if fieldEmpty(command.SecretKeyFlag, context) || fieldEmpty(command.AccessKeyFlag, context) || fieldEmpty(command.ProfileNameFlag, context) {
		// fieldEmpty() will log the error; returned error is only used for unit tests
		return errors.New("a required field was empty")
	}

	secretKey := context.String(command.SecretKeyFlag)
	profileName := context.String(command.ProfileNameFlag)
	accessKey := context.String(command.AccessKeyFlag)

	profile := &config.Profile{AWSAccessKey: accessKey, AWSSecretKey: secretKey}

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error saving profile: ", err)
		return err
	}
	if err = rdwr.SaveProfile(profileName, profile); err != nil {
		logrus.Error("Error saving profile: ", err)
		return err
	}

	return nil
}

// DefaultProfile is the callback for Configure Profile Default subcommand.
// Note: The error returned is only used in the unit tests
// This function is called by urfave/cli which does not check the error returned
func DefaultProfile(context *cli.Context) error {
	// validate field not empty
	if fieldEmpty(command.ProfileNameFlag, context) {
		// fieldEmpty() will log the error; returned error is only used for unit tests
		return errors.New("Profile name can not be empty")
	}

	// get relevant fields
	profileName := context.String(command.ProfileNameFlag)

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error setting default config: ", err)
		return err
	}
	if err = rdwr.SetDefaultProfile(profileName); err != nil {
		logrus.Error("Error setting default config: ", err)
		return err
	}

	return nil
}

// DefaultCluster is the callback for Configure Cluster Default subcommand.
// Note: The error returned is only used in the unit tests
// This function is called by urfave/cli which does not check the error returned
func DefaultCluster(context *cli.Context) error {
	// validate field not empty
	if fieldEmpty(command.ConfigNameFlag, context) {
		// fieldEmpty() will log the error; returned error is only used for unit tests
		return errors.New("Config name can not be empty")
	}
	// get relevant fields
	clusterName := context.String(command.ConfigNameFlag)

	// modify the profile config file
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error setting default config: ", err)
		return err
	}
	if err = rdwr.SetDefaultCluster(clusterName); err != nil {
		logrus.Error("Error setting default config: ", err)
		return err
	}

	return nil
}
