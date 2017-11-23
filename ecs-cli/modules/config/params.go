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

package config

import (
	"fmt"
	"os"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	// Launch types are case sensitive
	LaunchTypeFargate = "FARGATE"
	LaunchTypeEC2     = "EC2"
	LaunchTypeDefault = "EC2"
)

// CLIParams saves config to create an aws service clients
type CLIParams struct {
	Cluster                  string
	Session                  *session.Session
	ComposeServiceNamePrefix string
	ComposeProjectNamePrefix string // Deprecated; remains for backwards compatibility
	CFNStackName             string
	LaunchType               string
}

// Searches as far up the context as necessary. This function works no matter
// how many layers of nested subcommands there are. It is more powerful
// than merely calling context.String and context.GlobalString
func recursiveFlagSearch(context *cli.Context, flag string) string {
	if context == nil {
		return ""
	} else if value := context.String(flag); value != "" {
		return value
	} else {
		return recursiveFlagSearch(context.Parent(), flag)
	}
}

// NewCLIParams creates a new CLIParams object from the config file.
func NewCLIParams(context *cli.Context, rdwr ReadWriter) (*CLIParams, error) {
	clusterConfig := recursiveFlagSearch(context, flags.ClusterConfigFlag)
	profileConfig := recursiveFlagSearch(context, flags.ECSProfileFlag)
	ecsConfig, err := rdwr.Get(clusterConfig, profileConfig)

	if err != nil {
		return nil, errors.Wrap(err, "Error loading config")
	}

	// launch type from the flag overrides defaul launch type
	if launchTypeFromFlag := recursiveFlagSearch(context, flags.LaunchTypeFlag); launchTypeFromFlag != "" {
		ecsConfig.DefaultLaunchType = launchTypeFromFlag
	}

	if err = ValidateLaunchType(ecsConfig.DefaultLaunchType); err != nil {
		return nil, err
	}

	// Order of cluster resolution
	//  1) Inline flag
	//  2) Environment Variable
	//  3) ECS Config
	if clusterFromEnv := os.Getenv(flags.ClusterEnvVar); clusterFromEnv != "" {
		ecsConfig.Cluster = clusterFromEnv
	}
	if clusterFromFlag := recursiveFlagSearch(context, flags.ClusterFlag); clusterFromFlag != "" {
		ecsConfig.Cluster = clusterFromFlag
	}

	//--region flag has the highest precedence to set ecs-cli region config.
	if regionFromFlag := recursiveFlagSearch(context, flags.RegionFlag); regionFromFlag != "" {
		ecsConfig.Region = regionFromFlag
	}

	if awsProfileFromFlag := recursiveFlagSearch(context, flags.AWSProfileFlag); awsProfileFromFlag != "" {
		ecsConfig.AWSProfile = awsProfileFromFlag
		// unset Access Key and Secret Key, otherwise they will take precedence
		ecsConfig.AWSAccessKey = ""
		ecsConfig.AWSSecretKey = ""
	}

	svcSession, err := ecsConfig.ToAWSSession(context)
	if err != nil {
		return nil, err
	}

	if ecsConfig.Version == iniConfigVersion {
		ecsConfig.CFNStackName = ecsConfig.CFNStackNamePrefix + ecsConfig.Cluster
	}

	if ecsConfig.CFNStackName == "" {
		ecsConfig.CFNStackName = flags.CFNStackNamePrefixDefaultValue + ecsConfig.Cluster
	}

	return &CLIParams{
		Cluster:                  ecsConfig.Cluster,
		Session:                  svcSession,
		ComposeServiceNamePrefix: ecsConfig.ComposeServiceNamePrefix,
		ComposeProjectNamePrefix: ecsConfig.ComposeProjectNamePrefix, // deprecated; remains for backwards compatibility
		CFNStackName:             ecsConfig.CFNStackName,
		LaunchType:               ecsConfig.DefaultLaunchType,
	}, nil
}

// ValidateLaunchType checks that the launch type specified was an allowed value
func ValidateLaunchType(launchType string) error {
	if (launchType != "") && (launchType != LaunchTypeEC2) && (launchType != LaunchTypeFargate) {
		return fmt.Errorf("Supported launch types are '%s' and '%s'; %s is not a valid launch type.", LaunchTypeEC2, LaunchTypeFargate, launchType)
	}
	return nil
}
