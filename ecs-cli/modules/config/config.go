// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/private/endpoints"
)

// This time.Minute value comes from the SDK defaults package
const (
	ec2RoleProviderExpiryWindow = 5 * time.Minute
	ecsSectionKey               = "ecs"
	composeProjectNamePrefixKey = "compose-project-name-prefix"
	composeServiceNamePrefixKey = "compose-service-name-prefix"
	cfnStackNamePrefixKey       = "cfn-stack-name-prefix"
)

// CliConfig is the top level struct used to map to the ini config.
type CliConfig struct {
	// TODO Add metadata information like version etc.
	*SectionKeys `ini:"ecs"`
}

// SectionKeys is the struct embedded in CliConfig. It groups all the keys in the 'ecs' section in the ini file.
type SectionKeys struct {
	Cluster                  string `ini:"cluster"`
	AwsProfile               string `ini:"aws_profile"`
	Region                   string `ini:"region"`
	AwsAccessKey             string `ini:"aws_access_key_id"`
	AwsSecretKey             string `ini:"aws_secret_access_key"`
	ComposeProjectNamePrefix string `ini:"compose-project-name-prefix"`
	ComposeServiceNamePrefix string `ini:"compose-service-name-prefix"`
	CFNStackNamePrefix       string `ini:"cfn-stack-name-prefix"`
}

// NewCliConfig creates a new instance of CliConfig from the cluster name.
func NewCliConfig(cluster string) *CliConfig {
	return &CliConfig{&SectionKeys{Cluster: cluster}}
}

// ToServiceConfig creates an aws Config object from the CliConfig object.
func (cfg *CliConfig) ToServiceConfig() (*aws.Config, error) {
	region := cfg.getRegion()
	if region == "" {
		return nil, fmt.Errorf("Set a region with the --%s flag or %s environment variable", cli.RegionFlag, cli.AwsRegionEnvVar)
	}

	awsDefaults := defaults.Get()
	credentialProviders := cfg.getCredentialProviders(cfg.getEC2MetadataClient(&awsDefaults))
	chainCredentials := credentials.NewChainCredentials(credentialProviders)
	creds, err := chainCredentials.Get()
	if err != nil {
		return nil, err
	}

	// This is just a fail-fast check to ensure that valid credentials are available before returning to the caller.
	if creds.AccessKeyID == "" {
		return nil, fmt.Errorf("Error getting valid credentials")
	}

	svcConfig := awsDefaults.Config
	svcConfig.Region = aws.String(region)
	svcConfig.Credentials = chainCredentials

	return svcConfig, nil
}

// getCredentialProviders gets the chain of credentail provides to use when creating service clients.
func (cfg *CliConfig) getCredentialProviders(ec2MetadataClient *ec2metadata.EC2Metadata) []credentials.Provider {
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
			Client:       ec2MetadataClient,
			ExpiryWindow: ec2RoleProviderExpiryWindow,
		},
	}

	return credentialProviders
}

// getEC2MetadataClient creates a new instance of the EC2Metadata client
func (cfg *CliConfig) getEC2MetadataClient(awsDefaults *defaults.Defaults) *ec2metadata.EC2Metadata {
	endpoint, signingRegion := endpoints.EndpointForRegion(ec2metadata.ServiceName, cfg.getRegion(), true)
	return ec2metadata.NewClient(*awsDefaults.Config, awsDefaults.Handlers, endpoint, signingRegion)
}

// getRegion gets the region to use from ecs-cli's config file..
func (cfg *CliConfig) getRegion() string {
	region := cfg.Region
	if region == "" {
		// Search the chain of environment variables for region.
		for _, envVar := range []string{cli.AwsRegionEnvVar, cli.AwsDefaultRegionEnvVar} {
			region = os.Getenv(envVar)
			if region != "" {
				break
			}
		}
	}

	return region
}
