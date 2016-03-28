// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
)

// TODO: This needs a better home.

const (
	RegionFlag = "region"
	// This time.Minute value comes from the SDK defaults package
	eC2RoleProviderExpiryWindow = 5 * time.Minute
)

// CliConfig is the top level struct used to map to the ini config.
type CliConfig struct {
	// TODO Add metadata information like version etc.
	*SectionKeys `ini:"ecs"`
}

// SectionKeys is the struct embedded in CliConfig. It groups all the keys in the 'ecs' section in the ini file.
type SectionKeys struct {
	Cluster      string `ini:"cluster"`
	AwsProfile   string `ini:"aws_profile"`
	Region       string `ini:"region"`
	AwsAccessKey string `ini:"aws_access_key_id"`
	AwsSecretKey string `ini:"aws_secret_access_key"`
}

// NewCliConfig creates a new instance of CliConfig from the cluster name.
func NewCliConfig(cluster string) *CliConfig {
	return &CliConfig{&SectionKeys{Cluster: cluster}}
}

// ToServiceConfig creates an aws Config object from the CliConfig object.
func (cfg *CliConfig) ToServiceConfig() (*aws.Config, error) {
	region := cfg.getRegion()
	if region == "" {
		// TODO: Move AWS_REGION to a const.
		return nil, fmt.Errorf("Set a region with the --%s flag or AWS_REGION environment variable", RegionFlag)
	}

	chainCredentials := credentials.NewChainCredentials(cfg.getCredentialProviders())
	creds, err := chainCredentials.Get()
	if err != nil {
		return nil, err
	}

	// This is just a fail-fast check to ensure that valid credentials are available before returning to the caller.
	if creds.AccessKeyID == "" {
		return nil, fmt.Errorf("Error getting valid credentials")
	}

	svcConfig := defaults.Get().Config
	svcConfig.Credentials = chainCredentials
	svcConfig.Region = aws.String(region)

	return svcConfig, nil
}

// getCredentialProviders gets the chain of credentail provides to use when creating service clients.
func (cfg *CliConfig) getCredentialProviders() []credentials.Provider {
	// Append providers in the default credential providers chain to the chain.
	// Order of credential resolution
	// 1) Environment Variable provider
	// 2) ECS Profile provider - attempts to fetch the credentials from the ECS config file
	// 3) AWS profile - attempts to use the AWS profile specified in the ECS config file;
	// If the AWS profile has not been specified, provider will attempt to use the 'default' profile
	// 4) EC2 Instance role
	credentialProviders := []credentials.Provider{
		&credentials.EnvProvider{},
		&credentials.StaticProvider{
			Value: credentials.Value{
				AccessKeyID:     cfg.AwsAccessKey,
				SecretAccessKey: cfg.AwsSecretKey,
			},
		},
		&credentials.SharedCredentialsProvider{
			Filename: "",
			Profile:  cfg.AwsProfile,
		},
		&ec2rolecreds.EC2RoleProvider{
			ExpiryWindow: eC2RoleProviderExpiryWindow,
		},
	}

	return credentialProviders
}

// getRegion gets the region to use from ecs-cli's config file..
func (cfg *CliConfig) getRegion() string {
	region := cfg.Region
	if region == "" {
		// Search the chain of environment variables for region.
		// TODO: Move these to const's
		for _, envVar := range []string{"AWS_REGION", "AWS_DEFAULT_REGION"} {
			region = os.Getenv(envVar)
			if region != "" {
				break
			}
		}
	}

	return region
}
