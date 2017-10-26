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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/urfave/cli"
)

const (
	configVersion               = "v1"
	composeProjectNamePrefixKey = "compose-project-name-prefix"
	composeServiceNamePrefixKey = "compose-service-name-prefix"
	cfnStackNamePrefixKey       = "cfn-stack-name-prefix"
	awsAccessKey                = "aws_access_key_id"
	awsSecretKey                = "aws_secret_access_key"
	clusterKey                  = "cluster"
	clustersKey                 = "clusters"
	regionKey                   = "region"
)

// CLIConfig is the top level struct representing the configuration information
type CLIConfig struct {
	Cluster                  string
	AWSProfile               string
	Region                   string
	AWSAccessKey             string
	AWSSecretKey             string
	ComposeServiceNamePrefix string
	ComposeProjectNamePrefix string // Deprecated; remains for backwards compatibility
	CFNStackName             string
	CFNStackNamePrefix       string // Deprecated; remains for backwards compatibility
}

// Profile is a simple struct for storing a single profile config
type Profile struct {
	AWSAccessKey string `yaml:"aws_access_key_id"`
	AWSSecretKey string `yaml:"aws_secret_access_key"`
}

// Cluster is a simple struct for storing a single cluster config
type Cluster struct {
	Cluster                  string `yaml:"cluster"`
	Region                   string `yaml:"region"`
	ComposeServiceNamePrefix string `yaml:"compose-service-name-prefix,omitempty"`
	CFNStackName             string `yaml:"cfn-stack-name,omitempty"`
}

// ClusterConfig is the top level struct representing the cluster config file
type ClusterConfig struct {
	Version  string
	Default  string             `yaml:"default"`
	Clusters map[string]Cluster `yaml:"clusters"`
}

// ProfileConfig is the top level struct representing the Credentials file
type ProfileConfig struct {
	Version  string
	Default  string             `yaml:"default"`
	Profiles map[string]Profile `yaml:"ecs_profiles"`
}

// NewCLIConfig creates a new instance of CliConfig from the cluster name.
func NewCLIConfig(cluster string) *CLIConfig {
	return &CLIConfig{Cluster: cluster}
}

// ToAWSSession creates a new Session object from the CliConfig object.
//
// Region: Order of resolution
//  1) ECS CLI Flags
//   a) Region Flag --region
//   b) Cluster Config Flag (--cluster-config)
//  2) ECS Config - attempts to fetch the region from the default ECS Profile
//  3) Environment Variable - attempts to fetch the region from environment variables:
//    a) AWS_REGION (OR)
//    b) AWS_DEFAULT_REGION
//  4) AWS Profile - attempts to use region from AWS profile name from Env Vars
//    b) AWS_PROFILE environment variable (OR)
//    c) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')
//
// Credentials: Order of resolution
//  1) ECS CLI Profile Flags
//   a) ECS Profile (--ecs-profile)
//   b) AWS Profile (--aws-profile)
//  2) Environment Variables - attempts to fetch the credentials from environment variables:
//   a) ECS_PROFILE
//   b) AWS_PROFILE
//   c) AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY, Optional: AWS_SESSION_TOKEN
//  3) ECS Config - attempts to fetch the credentials from the default ECS Profile
//  4) Default AWS Profile - attempts to use credentials (aws_access_key_id, aws_secret_access_key) or assume_role (role_arn, source_profile) from AWS profile name
//    a) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')
//  5) EC2 Instance role
func (cfg *CLIConfig) ToAWSSession(context *cli.Context) (*session.Session, error) {

	region, err := cfg.getRegion()

	if err != nil || region == "" {
		return nil, fmt.Errorf("Set a region using ecs-cli configure command with the --%s flag or %s environment variable or --%s flag", command.RegionFlag, command.AwsRegionEnvVar, command.ProfileFlag)
	}

	if isProfileFlagsCase(context) {
		return credsFromECSConfig(cfg, region)
	} else if isEnvVarCase(context) {
		return defaultProvider(region)
	} else if isDefaultECSProfileCase(cfg) {
		return credsFromECSConfig(cfg, region)
	} else if profile := os.Getenv(command.AwsDefaultProfileEnvVar); profile != "" {
		// Currently the Go SDK does not pull creds from the AWS profile
		// defined by AWS_DEFAULT_PROFILE.
		return customProviderFromProfile(region, profile)
	} else {
		return defaultProvider(region)
	}

}

func isProfileFlagsCase(context *cli.Context) bool {
	return (recursiveFlagSearch(context, command.ECSProfileFlag) != "" || recursiveFlagSearch(context, command.AWSProfileNameFlag) != "")
}

func isEnvVarCase(context *cli.Context) bool {
	return (os.Getenv(command.AWSSecretKeyEnvVar) != "" && os.Getenv(command.AWSAccessKeyEnvVar) != "")
}

func isDefaultECSProfileCase(cfg *CLIConfig) bool {
	return (cfg.AWSAccessKey != "" || cfg.AWSSecretKey != "" || cfg.AWSProfile != "")
}

func credsFromECSConfig(cfg *CLIConfig, region string) (*session.Session, error) {
	if cfg.AWSSecretKey != "" {
		return customProviderFromKeys(region, cfg.AWSAccessKey, cfg.AWSSecretKey)
	} else {
		return customProviderFromProfile(region, cfg.AWSProfile)
	}
}

func defaultProvider(region string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
}

func customProviderFromProfile(region string, awsProfile string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", awsProfile),
	})
}

func customProviderFromKeys(region string, awsAccess string, awsSecret string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(awsAccess, awsSecret, ""),
	})
}

// Region: Order of resolution
//  1) ECS CLI Flags
//   a) Region Flag --region
//   b) Cluster Config Flag (--cluster-config)
//  2) ECS Config - attempts to fetch the region from the default ECS Profile
//  3) Environment Variable - attempts to fetch the region from environment variables:
//    a) AWS_REGION (OR)
//    b) AWS_DEFAULT_REGION
//  4) AWS Profile - attempts to use region from AWS profile name from Env Vars
//    b) AWS_PROFILE environment variable (OR)
//    c) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')
func (cfg *CLIConfig) getRegion() (string, error) {
	region := cfg.Region

	if region == "" {
		// Search the chain of environment variables for region.
		for _, envVar := range []string{command.AwsRegionEnvVar, command.AwsDefaultRegionEnvVar} {
			region = os.Getenv(envVar)
			if region != "" {
				break
			}
		}
	}

	var err error = nil
	if region == "" {
		region, err = getRegionFromDefaultProvider()
	}
	return region, err
}

func getRegionFromDefaultProvider() (string, error) {
	s, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(s.Config.Region), nil
}
