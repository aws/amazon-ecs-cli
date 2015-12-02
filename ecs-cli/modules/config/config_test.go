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
	"os"
	"testing"
)

const (
	clusterName             = "defaultCluster"
	profileName             = "defaultProfile"
	region                  = "us-west-1"
	awsAccessKey            = "AKID"
	awsSecretKey            = "SKID"
	envAwsAccessKey         = "envAKID"
	envAwsSecretKey         = "envSKID"
	credentialProviderCount = 4
)

func TestGetCredentialProvidersVerifyProviderCountHasNotChanged(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region
	credentialProviders := ecsConfig.getCredentialProviders()
	if len(credentialProviders) != credentialProviderCount {
		t.Fatal("Unexpected number of credential providers in the chain: ", len(credentialProviders))
	}

}

func TestToServiceConfigWhenUsingEnvVariables(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	os.Setenv("AWS_ACCESS_KEY_ID", envAwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", envAwsSecretKey)

	// Clear env variables as they persist past the individual test boundary
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	awsConfig, _ := ecsConfig.ToServiceConfig()
	resolvedCredentials, err := awsConfig.Credentials.Get()
	if err != nil {
		t.Error("Error fetching credentials from the chain provider")
	}

	if envAwsAccessKey != resolvedCredentials.AccessKeyID {
		t.Errorf("Invalid access key set. Expected [%s]. Got [%s]", envAwsAccessKey, resolvedCredentials.AccessKeyID)
	}
	if envAwsSecretKey != resolvedCredentials.SecretAccessKey {
		t.Errorf("Invalid secret key set. Expected [%s]. Got [%s]", envAwsSecretKey, resolvedCredentials.SecretAccessKey)
	}
}

func TestToServiceConfigWhenUsingECSProfileCredentials(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	ecsConfig.AwsAccessKey = awsAccessKey
	ecsConfig.AwsSecretKey = awsSecretKey

	awsConfig, _ := ecsConfig.ToServiceConfig()
	resolvedCredentials, err := awsConfig.Credentials.Get()
	if err != nil {
		t.Error("Error fetching credentials from the chain provider")
	}

	if awsAccessKey != resolvedCredentials.AccessKeyID {
		t.Errorf("Invalid access key set. Expected [%s]. Got [%s]", awsAccessKey, resolvedCredentials.AccessKeyID)
	}
	if awsSecretKey != resolvedCredentials.SecretAccessKey {
		t.Errorf("Invalid secret key set. Expected [%s]. Got [%s]", awsSecretKey, resolvedCredentials.SecretAccessKey)
	}
}

// TODO Add proper tests for the shared credential provider resolution
func TestToServiceConfigWhenAWSProfileSpecified(t *testing.T) {
	t.Skip("Implement me")
}
func TestToServiceConfigWhenNoAWSProfileSpecified(t *testing.T) {
	t.Skip("Implement me")
}
func TestToServiceConfigWhenUsingEC2InstanceRole(t *testing.T) {
	t.Skip("Implement me")
}

func TestToServiceConfigWhenRegionIsNotSpecified(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)

	_, err := ecsConfig.ToServiceConfig()
	if err == nil {
		t.Error("There should always be an error when region is not specified in the ecsConfig.")
	}
}
func TestToServiceConfigWhenNoCredentialsAreAvailable(t *testing.T) {
	t.Skip("Implement me")
}
