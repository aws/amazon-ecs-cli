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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/aws"
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

// CommandConfig contains the configuration parameters and AWS Session required to run a specific command
type CommandConfig struct {
	Cluster                  string
	Session                  *session.Session
	ComposeServiceNamePrefix string
	ComposeProjectNamePrefix string // Deprecated; remains for backwards compatibility
	CFNStackName             string
	LaunchType               string
}

func (c *CommandConfig) Region() string {
	return aws.StringValue(c.Session.Config.Region)
}

// Searches as far up the context as necessary. This function works no matter
// how many layers of nested subcommands there are. It is more powerful than
// merely calling context.String and context.GlobalString
func RecursiveFlagSearch(context *cli.Context, flag string) string {
	if context == nil {
		return ""
	} else if value := context.String(flag); value != "" {
		return value
	} else {
		return RecursiveFlagSearch(context.Parent(), flag)
	}
}

// NewCommandConfig creates a new CommandConfig object from the local ECS config file and flags
func NewCommandConfig(context *cli.Context, rdwr ReadWriter) (*CommandConfig, error) {
	clusterConfig := RecursiveFlagSearch(context, flags.ClusterConfigFlag)
	profileConfig := RecursiveFlagSearch(context, flags.ECSProfileFlag)
	ecsConfig, err := rdwr.Get(clusterConfig, profileConfig)

	if err != nil {
		return nil, errors.Wrap(err, "Error loading config")
	}

	// Configuration passed in via flags take precedence over stored config
	err = ecsConfig.applyFlags(context)
	if err != nil {
		return nil, errors.Wrap(err, "Error reading flags")
	}

	// Instantiate AWS Session
	svcSession, err := ecsConfig.ToAWSSession(context)
	if err != nil {
		return nil, err
	}

	// Determine Cloudformation StackName
	if ecsConfig.Version == iniConfigVersion {
		ecsConfig.CFNStackName = ecsConfig.CFNStackNamePrefix + ecsConfig.Cluster
	}
	if ecsConfig.CFNStackName == "" {
		ecsConfig.CFNStackName = flags.CFNStackNamePrefixDefaultValue + ecsConfig.Cluster
	}

	return &CommandConfig{
		Cluster:                  ecsConfig.Cluster,
		Session:                  svcSession,
		ComposeServiceNamePrefix: ecsConfig.ComposeServiceNamePrefix,
		ComposeProjectNamePrefix: ecsConfig.ComposeProjectNamePrefix, // deprecated; remains for backwards compatibility
		CFNStackName:             ecsConfig.CFNStackName,
		LaunchType:               ecsConfig.DefaultLaunchType,
	}, nil
}
