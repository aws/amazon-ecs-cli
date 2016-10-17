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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	clusterName             = "defaultCluster"
	region                  = "us-east-1"
	awsAccessKey            = "AKID"
	awsSecretKey            = "SKID"
	credentialProviderCount = 2

	customProfileName  = "customProfile"
	customAwsAccessKey = "customAKID"
	customAwsRegion    = "us-west-1"
	customAwsSecretKey = "customSKID"

	defaultProfileName  = "default"
	defaultAwsAccessKey = "defaultAwsAccessKey"
	defaultAwsRegion    = "us-west-2"
	defaultAwsSecretKey = "defaultAwsSecretKey"

	envAwsAccessKey = "envAKID"
	envAwsRegion    = "eu-west-1"
	envAwsSecretKey = "envSKID"

	assumeRoleName      = "assumeRoleWithCreds"
	assumeRoleAccessKey = "assumeRoleAKID"
	assumeRoleRegion    = "us-east-2"
	assumeRoleSecretKey = "assumeRoleSKID"

	ec2InstanceRoleName      = "ec2InstanceRole"
	ec2InstanceRoleAccessKey = "ec2InstanceRoleAKID"
	ec2InstanceRoleRegion    = "ap-northeast-1"
	ec2InstanceRoleSecretKey = "ec2InstanceRoleSKID"
)

//------------------------------------------------------------------------------
// ToAWSSession() --> REGION TESTS
// Order of resolution:
// 1a) Use AWS_REGION env variable
// 1b) Use AWS_DEFAULT_REGION env variable
// 2) Use Region in ECS Config
// 3a) Use Region from profile in ECS Config
// 3b) Use Region from AWS_PROFILE
// 3c) Use Region from AWS_DEFAULT_PROFILE
//------------------------------------------------------------------------------

// 1a) Use AWS_REGION env variable
func TestRegionWhenUsingEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.AwsAccessKey = awsAccessKey
	ecsConfig.AwsSecretKey = awsSecretKey

	// set variable for test
	os.Setenv("AWS_REGION", envAwsRegion)
	// Clear env variables as they persist past the individual test boundary
	defer func() {
		os.Unsetenv("AWS_REGION")
	}()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, envAwsRegion)
}

// 1b) Use AWS_DEFAULT_REGION env variable
func TestRegionWhenUsingDefaultEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.AwsAccessKey = awsAccessKey
	ecsConfig.AwsSecretKey = awsSecretKey

	// set variable for test
	os.Setenv("AWS_DEFAULT_REGION", envAwsRegion)
	defer func() {
		os.Unsetenv("AWS_DEFAULT_REGION")
	}()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, envAwsRegion)
}

// 2) Use Region in ECS Config
func TestRegionWhenUsingECSConfigRegion(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.AwsAccessKey = awsAccessKey
	ecsConfig.AwsSecretKey = awsSecretKey

	// set variable for test
	ecsConfig.Region = region

	// invoke test and verify
	testRegionInSession(t, ecsConfig, region)
}

// 3a) Use Region from profile in ECS Config
func TestRegionWhenUsingECSConfigProfile(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)

	// set variables for test
	ecsConfig.AwsProfile = customProfileName
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, customAwsRegion)
}

// 3b) Use Region from AWS_PROFILE
func TestRegionWhenUsingAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_PROFILE", customProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_PROFILE")
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, customAwsRegion)
}

// 3c) Use Region from AWS_DEFAULT_PROFILE
func TestRegionWhenUsingDefaultAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_DEFAULT_PROFILE")
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, defaultAwsRegion)
}

func TestRegionWhenNoneSpecified(t *testing.T) {
	// defaults
	os.Clearenv()
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.AwsAccessKey = awsAccessKey
	ecsConfig.AwsSecretKey = awsSecretKey

	// NOTE: no region set

	// invoke test and verify
	if _, err := ecsConfig.ToAWSSession(); err == nil {
		t.Error("There should always be an error when region is not specified or resolved.")
	}
}

func testRegionInSession(t *testing.T, inputConfig *CliConfig, expectedRegion string) {
	awsSession, err := inputConfig.ToAWSSession()
	if err != nil {
		t.Fatal("Error generating a new session")
	}
	awsConfig := awsSession.Config

	if expectedRegion != *awsConfig.Region {
		t.Errorf("Invalid region. Expected [%s]. Got [%s]", expectedRegion, *awsConfig.Region)
	}
}

//-------------------------------END OF REGION TESTS----------------------------

//------------------------------------------------------------------------------
// ToAWSSession() --> CREDENTIALS TESTS
// Order of resolution:
// 1a) Use AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY env variables
// 1b) Use AWS_ACCESS_KEY and AWS_SECRET_KEY env variables
// 2) Use access and secrets keys from ECS Config
// 3a) Use credentials from profile in ECS Config
// 3b) Use credentials from AWS_PROFILE
// 3c) Use credentials from AWS_DEFAULT_PROFILE
// 3d) Use credentials from assume role profile
// 4) EC2 Instance role
//------------------------------------------------------------------------------

func TestGetInitialCredentialProvidersVerifyProviderCountHasNotChanged(t *testing.T) {
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region
	credentialProviders := ecsConfig.getInitialCredentialProviders()
	if len(credentialProviders) != credentialProviderCount {
		t.Fatal("Unexpected number of credential providers in the chain: ", len(credentialProviders))
	}
}

// 1a) Use AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY env variables
func TestCredentialsWhenUsingEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	// set variables for test
	os.Setenv("AWS_ACCESS_KEY_ID", envAwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", envAwsSecretKey)
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY_ID")
		os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	}()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, envAwsAccessKey, envAwsSecretKey)
}

// 1b) Use AWS_ACCESS_KEY and AWS_SECRET_KEY env variables
func TestCredentialsWhenUsingDefaultEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	// set variables for test
	os.Setenv("AWS_ACCESS_KEY", envAwsAccessKey)
	os.Setenv("AWS_SECRET_KEY", envAwsSecretKey)
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, envAwsAccessKey, envAwsSecretKey)
}

// 2) Use access and secrets keys from ECS Config
func TestCredentialsWhenUsingECSConfigRegion(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	// set variables for test
	ecsConfig.AwsAccessKey = awsAccessKey
	ecsConfig.AwsSecretKey = awsSecretKey

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, awsAccessKey, awsSecretKey)
}

// 3a) Use credentials from profile in ECS Config
func TestCredentialsWhenUsingECSConfigProfile(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.AwsProfile = customProfileName

	// set variables for test
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, customAwsAccessKey, customAwsSecretKey)
}

// 3b) Use credentials from AWS_PROFILE
func TestCredentialsWhenUsingAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_PROFILE", customProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_PROFILE")
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, customAwsAccessKey, customAwsSecretKey)
}

// 3c) Use Region from AWS_DEFAULT_PROFILE
func TestCredentialsWhenUsingDefaultAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_DEFAULT_PROFILE")
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, defaultAwsAccessKey, defaultAwsSecretKey)
}

// 3d) Use credentials from assume role profile
func TestCredentialsWhenUsingAssumeRoleProfile(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_DEFAULT_PROFILE", assumeRoleName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_DEFAULT_PROFILE")
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		const respMsg = `
	<AssumeRoleResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
	  <AssumeRoleResult>
	    <AssumedRoleUser>
	      <Arn>arn:aws:sts::account_id:assumed-role/role/session_name</Arn>
	      <AssumedRoleId>AKID:session_name</AssumedRoleId>
	    </AssumedRoleUser>
	    <Credentials>
	      <AccessKeyId>` + assumeRoleAccessKey + `</AccessKeyId>
	      <SecretAccessKey>` + assumeRoleSecretKey + `</SecretAccessKey>
	      <SessionToken>SESSION_TOKEN</SessionToken>
	      <Expiration>%s</Expiration>
	    </Credentials>
	  </AssumeRoleResult>
	  <ResponseMetadata>
	    <RequestId>request-id</RequestId>
	  </ResponseMetadata>
	</AssumeRoleResponse>
	`
		w.Write([]byte(fmt.Sprintf(respMsg, time.Now().Add(15*time.Minute).Format("2006-01-02T15:04:05Z"))))
	}))

	startingConfig := aws.Config{}
	startingConfig.Endpoint = aws.String(server.URL)
	startingConfig.DisableSSL = aws.Bool(true)

	// invoke test and verify
	testCredentialsInSessionWithConfig(t, ecsConfig, &startingConfig, assumeRoleAccessKey, assumeRoleSecretKey)
}

// 4) Use credentials from EC2 Instance Role
func TestCredentialsWhenUsingEC2InstanceRole(t *testing.T) {
	// defaults
	ecsConfig := NewCliConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_DEFAULT_PROFILE", ec2InstanceRoleName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer func() {
		os.Unsetenv("AWS_DEFAULT_PROFILE")
		os.Unsetenv("AWS_CONFIG_FILE")
		os.Unsetenv("AWS_SHARED_CREDENTIALS_FILE")
	}()

	ec2Creds := `{
	  "Code": "Success",
	  "Type": "AWS-HMAC",
	  "AccessKeyId" : "` + ec2InstanceRoleAccessKey + `",
	  "SecretAccessKey" : "` + ec2InstanceRoleSecretKey + `",
	  "Token" : "token",
	  "Expiration" : "%s",
	  "LastUpdated" : "2009-11-23T0:00:00Z"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/latest/meta-data/iam/security-credentials" {
			fmt.Fprintln(w, ec2InstanceRoleName)
		} else if r.URL.Path == "/latest/meta-data/iam/security-credentials/"+ec2InstanceRoleName {
			fmt.Fprintf(w, ec2Creds, "2014-12-16T01:51:37Z")
		} else {
			http.Error(w, "bad request", http.StatusBadRequest)
		}
	}))

	myCustomResolver := func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		return endpoints.ResolvedEndpoint{
			URL:           server.URL + "/latest",
			SigningRegion: ec2InstanceRoleRegion,
		}, nil
	}
	startingConfig := aws.Config{}
	startingConfig.EndpointResolver = endpoints.ResolverFunc(myCustomResolver)

	// invoke test and verify
	testCredentialsInSessionWithConfig(t, ecsConfig, &startingConfig, ec2InstanceRoleAccessKey, ec2InstanceRoleSecretKey)
}

// Error if Session.Credentials are nil
func TestCredentialsWhenNoneSpecified(t *testing.T) {
	// defaults
	os.Clearenv()
	ecsConfig := NewCliConfig(clusterName)
	ecsConfig.Region = region

	// NOTE: no credentials set

	// invoke test and verify
	awsSession, err := ecsConfig.ToAWSSession()
	if err != nil {
		t.Fatal("Error generating a new session")
	}
	awsConfig := awsSession.Config
	if _, err = awsConfig.Credentials.Get(); err == nil {
		t.Fatal("Should have error getting credentials")
	}
}

func testCredentialsInSession(t *testing.T, inputConfig *CliConfig, expectedAccessKey, expectedSecretKey string) {
	awsSession, err := inputConfig.ToAWSSession()
	if err != nil {
		t.Fatal("Error generating a new session")
	}
	verifyCredentialsInSession(t, awsSession, expectedAccessKey, expectedSecretKey)
}

func testCredentialsInSessionWithConfig(t *testing.T, inputConfig *CliConfig, ecsConfig *aws.Config,
	expectedAccessKey, expectedSecretKey string) {
	awsSession, err := inputConfig.toAWSSessionWithConfig(*ecsConfig)
	if err != nil {
		t.Fatalf("Error generating a new session", err)
	}
	verifyCredentialsInSession(t, awsSession, expectedAccessKey, expectedSecretKey)
}

func verifyCredentialsInSession(t *testing.T, awsSession *session.Session, expectedAccessKey, expectedSecretKey string) {
	awsConfig := awsSession.Config
	resolvedCredentials, err := awsConfig.Credentials.Get()
	if err != nil {
		t.Fatal("Error fetching credentials from the chain provider")
	}

	if expectedAccessKey != resolvedCredentials.AccessKeyID {
		t.Errorf("Invalid access key set. Expected [%s]. Got [%s]", expectedAccessKey, resolvedCredentials.AccessKeyID)
	}
	if expectedSecretKey != resolvedCredentials.SecretAccessKey {
		t.Errorf("Invalid secret key set. Expected [%s]. Got [%s]", expectedSecretKey, resolvedCredentials.SecretAccessKey)
	}
}
