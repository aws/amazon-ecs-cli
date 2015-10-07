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
	"testing"

	"github.com/aws/aws-sdk-go/aws/credentials"
)

const (
	clusterName  = "defaultCluster"
	profileName  = "defaultProfile"
	region       = "us-west-1"
	awsAccessKey = "AKID"
	awsSecretKey = "SKID"
)

func TestGetCredentialProvidersWithEmptyProfileAndCredentials(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	credentialProviders := ecsConfig.getCredentialProviders()
	if len(credentialProviders) != 4 {
		t.Error("Unexpected number of credential providers in the chain: ", len(credentialProviders))
	}

	// credentialProviders is composed of:
	// 0 -> env provider
	// 1 -> static creds provider
	// 2 -> default profile provider
	// 3 -> role provider
	profileDefaultCredsProvider, ok := credentialProviders[2].(*credentials.SharedCredentialsProvider)
	if !ok {
		t.Fatal("Mismatch in credential provider chain. Expected to use 'default' profile creds provider")
	}
	// This would be the 'default' profile creds provider in this case.
	if "" != profileDefaultCredsProvider.Profile {
		t.Errorf("Invalid profile name set. Expected empty string. Got [%s]", profileDefaultCredsProvider.Profile)
	}
}

func TestGetCredentialProvidersWithProfile(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.AwsProfile = profileName
	credentialProviders := ecsConfig.getCredentialProviders()
	if len(credentialProviders) != 3 {
		t.Fatal("Unexpected number of credential providers in the chain: ", len(credentialProviders))
	}

	// credentialProviders is composed of:
	// 0 -> profile provider
	// 1 -> env provider
	// 2 -> role provider
	profileCredsProvider, ok := credentialProviders[0].(*credentials.SharedCredentialsProvider)
	if !ok {
		t.Fatal("Mismatch in credential provider chain. Expected to use 'default' profile creds provider")
	}
	// This would be the 'default' profile creds provider in this case.
	if profileName != profileCredsProvider.Profile {
		t.Errorf("Invalid profile name set. Expected [%s]. Got [%s]", profileName, profileCredsProvider.Profile)
	}
}

func TestGetCredentialProvidersWithCredentials(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.AwsAccessKey = awsAccessKey
	ecsConfig.AwsSecretKey = awsSecretKey

	credentialProviders := ecsConfig.getCredentialProviders()
	if len(credentialProviders) != 4 {
		t.Fatal("Unexpected number of credential providers in the chain: ", len(credentialProviders))
	}

	// credentialProviders is composed of:
	// 0 -> static creds provider
	// 1 -> env provider
	// 2 -> default profile provider
	// 3 -> role provider
	staticCredsProvider, ok := credentialProviders[0].(*credentials.StaticProvider)
	if !ok {
		t.Fatal("Mismatch in credential provider chain. Expected to use static creds provider")
	}
	if awsAccessKey != staticCredsProvider.AccessKeyID {
		t.Errorf("Invalid access key set. Expected [%s]. Got [%s]", awsAccessKey, staticCredsProvider.AccessKeyID)
	}
	if awsSecretKey != staticCredsProvider.SecretAccessKey {
		t.Errorf("Invalid secret key set. Expected [%s]. Got [%s]", awsSecretKey, staticCredsProvider.SecretAccessKey)
	}
}
