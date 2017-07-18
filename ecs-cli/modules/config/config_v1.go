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

	cli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	configVersion               = "v1"
	composeProjectNamePrefixKey = "compose-project-name-prefix"
	composeServiceNamePrefixKey = "compose-service-name-prefix"
	cfnStackNamePrefixKey       = "cfn-stack-name-prefix"
)

// CLIConfig is the top level struct used to map to the yaml config.
// Version: v1 = YAML formatted configs split profile and cluster
type CLIConfig struct {
	Version                  string `yaml:"version"`
	Cluster                  string `yaml:"cluster"`
	AWSProfile               string `yaml:"aws_profile"`
	Region                   string `yaml:"region"`
	AWSAccessKey             string `yaml:"aws_access_key_id"`
	AWSSecretKey             string `yaml:"aws_secret_access_key"`
	ComposeProjectNamePrefix string `yaml:"compose-project-name-prefix"`
	ComposeServiceNamePrefix string `yaml:"compose-service-name-prefix"`
	CFNStackNamePrefix       string `yaml:"cfn-stack-name-prefix"`
}

// NewCLIConfig creates a new instance of CliConfig from the cluster name.
func NewCLIConfig(cluster string) *CLIConfig {
	return &CLIConfig{Cluster: cluster}
}

// ToAWSSession creates a new Session object from the CliConfig object.
//
// Region: Order of resolution
//  1) Environment Variable - attempts to fetch the region from environment variables:
//    a) AWS_REGION (OR)
//    b) AWS_DEFAULT_REGION
//  2) ECS Config - attempts to fetch the region from the ECS config file
//  3) AWS Profile - attempts to use region from AWS profile name
//    a) profile name from ECS config file (OR)
//    b) AWS_PROFILE environment variable (OR)
//    c) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')
//
// Credentials: Order of resolution
//  1) Environment Variable - attempts to fetch the credentials from environment variables:
//   a) AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY (OR)
//   b) AWS_ACCESS_KEY and AWS_SECRET_KEY
//  2) ECS Config - attempts to fetch the credentials from the ECS config file
//  3) AWS Profile - attempts to use credentials (aws_access_key_id, aws_secret_access_key) or assume_role (role_arn, source_profile) from AWS profile name
//    a) profile name from ECS config file (OR)
//    b) AWS_PROFILE environment variable (OR)
//    c) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')
//  4) EC2 Instance role
func (cfg *CLIConfig) ToAWSSession() (*session.Session, error) {
	svcConfig := aws.Config{}
	return cfg.toAWSSessionWithConfig(svcConfig)
}

func (cfg *CLIConfig) toAWSSessionWithConfig(svcConfig aws.Config) (*session.Session, error) {
	credentialProviders := cfg.getInitialCredentialProviders()
	chainCredentials := credentials.NewChainCredentials(credentialProviders)
	if _, err := chainCredentials.Get(); err == nil {
		svcConfig.Credentials = chainCredentials
	}

	svcConfig.Region = aws.String(cfg.getRegion())

	svcSession, err := session.NewSessionWithOptions(session.Options{
		Config:            svcConfig,
		Profile:           cfg.AWSProfile,
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, err
	}
	region := *svcSession.Config.Region
	if region == "" {
		return nil, fmt.Errorf("Set a region using ecs-cli configure command with the --%s flag or %s environment variable or --%s flag", cli.RegionFlag, cli.AwsRegionEnvVar, cli.ProfileFlag)
	}

	return svcSession, nil
}

// getInitialCredentialProviders gets the starting chain of credential providers to use when creating service clients.
func (cfg *CLIConfig) getInitialCredentialProviders() []credentials.Provider {
	// Append providers in the default credential providers chain to the chain.
	// Order of credential resolution
	//  1) Environment Variable
	//  2) ECS Config
	// the rest are handled by session.NewSessionWithOptions invoked in ToAWSSession()
	credentialProviders := []credentials.Provider{
		&credentials.EnvProvider{},
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     cfg.AWSAccessKey,
				SecretAccessKey: cfg.AWSSecretKey,
			},
		},
	}

	return credentialProviders
}

// getRegion gets the region to use from environment variables or ecs-cli's config file..
func (cfg *CLIConfig) getRegion() string {
	// Order of region resolution
	//  1) Environment Variable
	//  2) ECS Config
	// the rest are handled by session.NewSessionWithOptions invoked in ToAWSSession()
	region := ""
	// Search the chain of environment variables for region.
	for _, envVar := range []string{cli.AwsRegionEnvVar, cli.AwsDefaultRegionEnvVar} {
		region = os.Getenv(envVar)
		if region != "" {
			break
		}
	}
	if region == "" {
		region = cfg.Region
	}
	return region
}
