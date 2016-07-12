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
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	clusterName             = "defaultCluster"
	profileName             = "customProfile"
	region                  = "us-west-1"
	awsAccessKey            = "AKID"
	awsSecretKey            = "SKID"
	defaultAwsAccessKey     = "defaultAwsAccessKey"
	defaultAwsSecretKey     = "defaultAwsSecretKey"
	envAwsAccessKey         = "envAKID"
	envAwsSecretKey         = "envSKID"
	credentialProviderCount = 4
)

func TestGetCredentialProvidersVerifyProviderCountHasNotChanged(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region
	credentialProviders := ecsConfig.getCredentialProviders(&ec2metadata.EC2Metadata{})
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

func TestToServiceConfigWhenAWSProfileSpecified(t *testing.T) {
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	ecsConfig.AwsProfile = profileName

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
func TestToServiceConfigWhenAWSProfileIsNotSpecified(t *testing.T) {
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region
	// not setting the profileName: it should use the "default" profile

	awsConfig, _ := ecsConfig.ToServiceConfig()
	resolvedCredentials, err := awsConfig.Credentials.Get()
	if err != nil {
		t.Error("Error fetching credentials from the chain provider")
	}

	if defaultAwsAccessKey != resolvedCredentials.AccessKeyID {
		t.Errorf("Invalid access key set. Expected [%s]. Got [%s]", defaultAwsAccessKey, resolvedCredentials.AccessKeyID)
	}
	if defaultAwsSecretKey != resolvedCredentials.SecretAccessKey {
		t.Errorf("Invalid secret key set. Expected [%s]. Got [%s]", defaultAwsSecretKey, resolvedCredentials.SecretAccessKey)
	}

}

func TestToServiceConfigWhenRegionIsNotSpecified(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)

	_, err := ecsConfig.ToServiceConfig()
	if err == nil {
		t.Error("There should always be an error when region is not specified in the ecsConfig.")
	}
}

// Code excerpt to start a test server for ec2MetadataClient is taken from
// github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds/ec2_role_provider_test.go
const credsRespTmpl = `{
  "Code": "Success",
  "Type": "AWS-HMAC",
  "AccessKeyId" : "AKID",
  "SecretAccessKey" : "SKID",
  "Token" : "token",
  "Expiration" : "%s",
  "LastUpdated" : "2009-11-23T0:00:00Z"
}`

const credsFailRespTmpl = `{
  "Code": "ErrorCode",
  "Message": "ErrorMsg",
  "LastUpdated": "2009-11-23T0:00:00Z"
}`

func initTestServer(expireOn string, failAssume bool) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest/meta-data/iam/security-credentials" {
			fmt.Fprintln(w, "RoleName")
		} else if r.URL.Path == "/latest/meta-data/iam/security-credentials/RoleName" {
			if failAssume {
				fmt.Fprintf(w, credsFailRespTmpl)
			} else {
				fmt.Fprintf(w, credsRespTmpl, expireOn)
			}
		} else {
			http.Error(w, "bad request", http.StatusBadRequest)
		}
	}))

	return server
}

func TestGetCredentialProvidersWhenUsingEC2InstanceRole(t *testing.T) {
	server := initTestServer("2016-06-19T00:00:00Z", false)
	defer server.Close()

	metadataClient := ec2metadata.New(session.New(), &aws.Config{Endpoint: aws.String(server.URL + "/latest")})

	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region
	credentialProviders := ecsConfig.getCredentialProviders(metadataClient)
	chainCredentials := credentials.NewChainCredentials(credentialProviders)
	creds, err := chainCredentials.Get()
	if err != nil {
		t.Error("Unexpected error occured when retrieving credentials from EC2 metadata service")
	}

	if awsAccessKey != creds.AccessKeyID {
		t.Errorf("Invalid access key set. Expected [%s]. Got [%s]", awsAccessKey, creds.AccessKeyID)
	}
	if awsSecretKey != creds.SecretAccessKey {
		t.Errorf("Invalid secret key set. Expected [%s]. Got [%s]", awsSecretKey, creds.SecretAccessKey)
	}
}

func TestGetCredentialProvidersWhenEC2MetadataServiceReturnsFailure(t *testing.T) {
	server := initTestServer("2016-06-19T00:00:00Z", true)
	defer server.Close()

	metadataClient := ec2metadata.New(session.New(), &aws.Config{Endpoint: aws.String(server.URL + "/latest")})

	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region
	credentialProviders := ecsConfig.getCredentialProviders(metadataClient)
	chainCredentials := credentials.NewChainCredentials(credentialProviders)
	_, err := chainCredentials.Get()
	if err == nil {
		t.Error("Expected an error while retrieving credentials from EC2 metadata service")
	}
}

func TestToServiceConfigWhenNoCredentialsAreAvailable(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	_, err := ecsConfig.ToServiceConfig()
	if err == nil {
		t.Error("Should get an error for no credentials available")
	}
}
