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
	awsSessionToken             = "aws_session_token"
	clusterKey                  = "cluster"
	clustersKey                 = "clusters"
	regionKey                   = "region"
	iniConfigVersion            = 0
	yamlConfigVersion           = 1
)

// CLIConfig is the top level struct representing the configuration information
type CLIConfig struct {
	Version                  int // which format version was the config file that was read. 1 == yaml, 0 == old ini
	Cluster                  string
	AWSProfile               string
	Region                   string
	AWSAccessKey             string
	AWSSecretKey             string
	AWSSessionToken          string
	ComposeServiceNamePrefix string
	ComposeProjectNamePrefix string // Deprecated; remains for backwards compatibility
	CFNStackName             string
	CFNStackNamePrefix       string // Deprecated; remains for backwards compatibility
	DefaultLaunchType        string
}

// Profile is a simple struct for storing a single profile config
type Profile struct {
	AWSAccessKey    string `yaml:"aws_access_key_id"`
	AWSSecretKey    string `yaml:"aws_secret_access_key"`
	AWSSessionToken string `yaml:"aws_session_token,omitempty"`
}

// Cluster is a simple struct for storing a single cluster config
type Cluster struct {
	Cluster                  string `yaml:"cluster"`
	Region                   string `yaml:"region"`
	ComposeServiceNamePrefix string `yaml:"compose-service-name-prefix,omitempty"`
	CFNStackName             string `yaml:"cfn-stack-name,omitempty"`
	DefaultLaunchType        string `yaml:"default_launch_type"`
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
//  4) AWS Profile - attempts to use region from AWS profile name
//    a) --aws-profile flag
//    b) AWS_PROFILE environment variable
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
	return cfg.toAWSSessionWithConfig(context, &aws.Config{})
}

// ToAWSSessionWithConfig processes credential order of precedence
// The argument svcConfig is needed to allow important unit tests to work
// (for example: assume role)
func (cfg *CLIConfig) toAWSSessionWithConfig(context *cli.Context, svcConfig *aws.Config) (*session.Session, error) {

	region, err := cfg.getRegion()

	if err != nil || region == "" {
		return nil, fmt.Errorf("Set a region using ecs-cli configure command with the --%s flag or %s environment variable or --%s flag", flags.RegionFlag, flags.AwsRegionEnvVar, flags.ProfileFlag)
	}

	if hasProfileFlags(context) {
		// The AWS SDK Go lets Env Vars override sourcing a profile from the shared credential file
		// This means that if the Env Vars are present, it will ignore the profile.
		// So we unset them and then reset them, to allow our order of precedence to be correct.
		keyID, secretKey := unsetEnvVars()
		defer resetEnvVars(keyID, secretKey)
		return sessionFromECSConfig(cfg, region, svcConfig)
	} else if hasEnvVars(context) {
		return sessionFromProfile("", region, svcConfig)
	} else if isDefaultECSProfileCase(cfg) {
		return sessionFromECSConfig(cfg, region, svcConfig)
	}

	return sessionFromProfile("", region, svcConfig)
}

func hasProfileFlags(context *cli.Context) bool {
	return (recursiveFlagSearch(context, flags.ECSProfileFlag) != "" || recursiveFlagSearch(context, flags.AWSProfileFlag) != "")
}

func hasEnvVars(context *cli.Context) bool {
	return (os.Getenv(flags.AWSSecretKeyEnvVar) != "" && os.Getenv(flags.AWSAccessKeyEnvVar) != "")
}

func isDefaultECSProfileCase(cfg *CLIConfig) bool {
	return (cfg.AWSAccessKey != "" || cfg.AWSSecretKey != "" || cfg.AWSProfile != "")
}

func sessionFromECSConfig(cfg *CLIConfig, region string, svcConfig *aws.Config) (*session.Session, error) {
	if cfg.AWSSecretKey != "" {
		return sessionFromKeys(region, cfg.AWSAccessKey, cfg.AWSSecretKey, cfg.AWSSessionToken, svcConfig)
	}

	return sessionFromProfile(cfg.AWSProfile, region, svcConfig)
}

func unsetEnvVars() (keyID string, secretKey string) {
	keyID = os.Getenv(flags.AWSAccessKeyEnvVar)
	secretKey = os.Getenv(flags.AWSSecretKeyEnvVar)
	os.Unsetenv(flags.AWSAccessKeyEnvVar)
	os.Unsetenv(flags.AWSSecretKeyEnvVar)

	return keyID, secretKey
}

func resetEnvVars(keyID string, secretKey string) {
	os.Setenv(flags.AWSAccessKeyEnvVar, keyID)
	os.Setenv(flags.AWSSecretKeyEnvVar, secretKey)
}

func sessionFromProfile(profile string, region string, svcConfig *aws.Config) (*session.Session, error) {
	svcConfig.Region = aws.String(region)
	return session.NewSessionWithOptions(session.Options{
		Config:            *svcConfig,
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
	})
}

func sessionFromKeys(region string, awsAccess string, awsSecret string, sessionToken string, svcConfig *aws.Config) (*session.Session, error) {
	svcConfig.Region = aws.String(region)
	svcConfig.Credentials = credentials.NewStaticCredentials(awsAccess, awsSecret, sessionToken)
	return session.NewSession(svcConfig)
}

// Region: Order of resolution
//  1) ECS CLI Flags
//   a) Region Flag --region
//   b) Cluster Config Flag (--cluster-config)
//  2) ECS Config - attempts to fetch the region from the default ECS Profile
//  3) Environment Variable - attempts to fetch the region from environment variables:
//    a) AWS_REGION (OR)
//    b) AWS_DEFAULT_REGION
//  4) AWS Profile - attempts to use region from AWS profile name
//    a) --aws-profile flag
//    b) AWS_PROFILE environment variable
//    c) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')
func (cfg *CLIConfig) getRegion() (string, error) {
	region := cfg.Region

	if region == "" {
		// Search the chain of environment variables for region.
		for _, envVar := range []string{flags.AwsRegionEnvVar, flags.AwsDefaultRegionEnvVar} {
			region = os.Getenv(envVar)
			if region != "" {
				break
			}
		}
	}

	var err error
	if region == "" {
		region, err = cfg.getRegionFromAWSProfile()
	}
	return region, err
}

func (cfg *CLIConfig) getRegionFromAWSProfile() (string, error) {
	awsProfile := ""
	if cfg.AWSProfile != "" {
		awsProfile = cfg.AWSProfile
	} else {
		awsProfile = os.Getenv(flags.AwsDefaultProfileEnvVar)
	}

	s, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Profile:           awsProfile,
	})
	if err != nil {
		return "", err
	}
	return aws.StringValue(s.Config.Region), nil

}
