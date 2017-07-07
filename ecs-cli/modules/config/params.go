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

	"github.com/Sirupsen/logrus"
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/urfave/cli"
)

// CliParams saves config to create an aws service clients
type CliParams struct {
	Cluster                  string
	Session                  *session.Session
	ComposeProjectNamePrefix string
	ComposeServiceNamePrefix string
	CFNStackNamePrefix       string
}

// GetCfnStackName <cfn_stack_name_prefix> + <cluster_name>
func (p *CliParams) GetCfnStackName() string {
	return fmt.Sprintf("%s%s", p.CFNStackNamePrefix, p.Cluster)
}

// NewCliParams creates a new ECSParams object from the config file.
func NewCliParams(context *cli.Context, rdwr ReadWriter) (*CliParams, error) {
	ecsConfig, err := rdwr.GetConfig()
	if err != nil {
		logrus.Error("Error loading config: ", err)
		return nil, err
	}

	// If Prefixes not found, set to defaults.
	if !rdwr.IsKeyPresent(ecsSectionKey, composeProjectNamePrefixKey) {
		ecsConfig.ComposeProjectNamePrefix = ecscli.ComposeProjectNamePrefixDefaultValue
	}
	if !rdwr.IsKeyPresent(ecsSectionKey, composeServiceNamePrefixKey) {
		ecsConfig.ComposeServiceNamePrefix = ecscli.ComposeServiceNamePrefixDefaultValue
	}
	if !rdwr.IsKeyPresent(ecsSectionKey, cfnStackNamePrefixKey) {
		ecsConfig.CFNStackNamePrefix = ecscli.CFNStackNamePrefixDefaultValue
	}

	// Order of cluster resolution
	//  1) Inline flag
	//  2) Environment Variable
	//  3) ECS Config
	if clusterFromEnv := os.Getenv(ecscli.ClusterEnvVar); clusterFromEnv != "" {
		ecsConfig.Cluster = clusterFromEnv
	}
	// First try to find the flag in the global string, then try to find the flag locally
	if clusterFromFlag := context.GlobalString(ecscli.ClusterFlag); clusterFromFlag != "" {
		ecsConfig.Cluster = clusterFromFlag
	} else if clusterFromFlag := context.String(ecscli.ClusterFlag); clusterFromFlag != "" {
		ecsConfig.Cluster = clusterFromFlag
	}

	//--region flag has the highest precedence to set ecs-cli region config.
	// First try to find the flag in the global string, then try to find the flag locally
	if regionFromFlag := context.GlobalString(ecscli.RegionFlag); regionFromFlag != "" {
		ecsConfig.Region = regionFromFlag
	} else if regionFromFlag := context.String(ecscli.RegionFlag); regionFromFlag != "" {
		ecsConfig.Region = regionFromFlag
	}

	svcSession, err := ecsConfig.ToAWSSession()
	if err != nil {
		return nil, err
	}

	return &CliParams{
		Cluster:                  ecsConfig.Cluster,
		Session:                  svcSession,
		ComposeProjectNamePrefix: ecsConfig.ComposeProjectNamePrefix,
		ComposeServiceNamePrefix: ecsConfig.ComposeServiceNamePrefix,
		CFNStackNamePrefix:       ecsConfig.CFNStackNamePrefix,
	}, nil
}
