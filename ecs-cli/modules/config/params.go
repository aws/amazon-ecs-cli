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
	"os"

	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

// CLIParams saves config to create an aws service clients
type CLIParams struct {
	Cluster                  string
	Session                  *session.Session
	ComposeServiceNamePrefix string
	ComposeProjectNamePrefix string // Deprecated; remains for backwards compatibility
	CFNStackName             string
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

// NewCLIParams creates a new ECSParams object from the config file.
func NewCLIParams(context *cli.Context, rdwr ReadWriter) (*CLIParams, error) {
	clusterConfig := recursiveFlagSearch(context, ecscli.ClusterConfigFlag)
	profileConfig := recursiveFlagSearch(context, ecscli.ProfileConfigFlag)
	ecsConfig, err := rdwr.Get(clusterConfig, profileConfig)

	if err != nil {
		return nil, errors.Wrap(err, "Error loading config")
	}

	// Order of cluster resolution
	//  1) Inline flag
	//  2) Environment Variable
	//  3) ECS Config
	if clusterFromEnv := os.Getenv(ecscli.ClusterEnvVar); clusterFromEnv != "" {
		ecsConfig.Cluster = clusterFromEnv
	}
	if clusterFromFlag := recursiveFlagSearch(context, ecscli.ClusterFlag); clusterFromFlag != "" {
		ecsConfig.Cluster = clusterFromFlag
	}

	//--region flag has the highest precedence to set ecs-cli region config.
	if regionFromFlag := recursiveFlagSearch(context, ecscli.RegionFlag); regionFromFlag != "" {
		ecsConfig.Region = regionFromFlag
	}

	if awsProfileFromFlag := recursiveFlagSearch(context, ecscli.AWSProfileNameFlag); awsProfileFromFlag != "" {
		ecsConfig.AWSProfile = awsProfileFromFlag
		// unset Access Key and Secret Key, otherwise they will take precedence
		ecsConfig.AWSAccessKey = ""
		ecsConfig.AWSSecretKey = ""
	}

	svcSession, err := ecsConfig.ToAWSSession()
	if err != nil {
		return nil, err
	}
	if ecsConfig.CFNStackName == "" && ecsConfig.CFNStackNamePrefix != "" {
		ecsConfig.CFNStackName = ecsConfig.CFNStackNamePrefix + ecsConfig.Cluster
	} else if ecsConfig.CFNStackName == "" {
		// set default value
		ecsConfig.CFNStackName = ecscli.CFNStackNamePrefixDefaultValue + ecsConfig.Cluster
	}
	return &CLIParams{
		Cluster:                  ecsConfig.Cluster,
		Session:                  svcSession,
		ComposeServiceNamePrefix: ecsConfig.ComposeServiceNamePrefix,
		ComposeProjectNamePrefix: ecsConfig.ComposeProjectNamePrefix, // deprecated; remains for backwards compatibility
		CFNStackName:             ecsConfig.CFNStackName,
	}, nil
}
