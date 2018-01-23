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
	"flag"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	clusterName             = "defaultCluster"
	region                  = "us-east-1"
	credentialProviderCount = 2
	awsToken                = "session-token"

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
//------------------------------------------------------------------------------

func TestRegionOrderOfResolutionECSConfig(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey

	// set variable for test
	ecsConfig.Region = region // takes precedence
	ecsConfig.AWSProfile = customProfileName
	os.Setenv("AWS_REGION", envAwsRegion)
	os.Setenv("AWS_PROFILE", customProfileName)
	os.Setenv("AWS_DEFAULT_PROFILE", assumeRoleName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, region)
}

func TestRegionOrderOfResolutionEnvVar(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey

	// set variable for test
	ecsConfig.AWSProfile = customProfileName
	os.Setenv("AWS_REGION", envAwsRegion) // takes precedence
	os.Setenv("AWS_PROFILE", customProfileName)
	os.Setenv("AWS_DEFAULT_PROFILE", assumeRoleName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, envAwsRegion)
}

func TestRegionOrderOfResolutionDefaultEnvVar(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey

	// set variable for test
	ecsConfig.AWSProfile = customProfileName
	os.Setenv("AWS_DEFAULT_REGION", envAwsRegion) // takes precedence
	os.Setenv("AWS_PROFILE", customProfileName)
	os.Setenv("AWS_DEFAULT_PROFILE", assumeRoleName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, envAwsRegion)
}

// 1a) Use AWS_REGION env variable
func TestRegionWhenUsingEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey

	// set variable for test
	os.Setenv("AWS_REGION", envAwsRegion)
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, envAwsRegion)
}

// 1b) Use AWS_DEFAULT_REGION env variable
func TestRegionWhenUsingDefaultEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey

	// set variable for test
	os.Setenv("AWS_DEFAULT_REGION", envAwsRegion)
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, envAwsRegion)
}

// 2) Use Region in ECS Config
func TestRegionWhenUsingECSConfigRegion(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey

	// set variable for test
	ecsConfig.Region = region

	// invoke test and verify
	testRegionInSession(t, ecsConfig, region)
}

// 3a) Use Region from profile in ECS Config
func TestRegionWhenUsingECSConfigProfile(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	ecsConfig.AWSProfile = customProfileName
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, customAwsRegion)
}

// 3b) Use Region from AWS_PROFILE
func TestRegionWhenUsingAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_PROFILE", customProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, customAwsRegion)
}

// 3c) Use Region from AWS_DEFAULT_PROFILE
func TestRegionWhenUsingDefaultAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testRegionInSession(t, ecsConfig, defaultAwsRegion)
}

func TestRegionWhenNoneSpecified(t *testing.T) {
	// defaults
	os.Clearenv()
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey

	// NOTE: no region set

	// invoke test and verify
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, nil)
	_, err := ecsConfig.ToAWSSession(context)
	assert.Error(t, err, "Expected error when region is not specified or resolved")
}

func testRegionInSession(t *testing.T, inputConfig *CLIConfig, expectedRegion string) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, nil)
	awsSession, err := inputConfig.ToAWSSession(context)
	if err != nil {
		t.Fatal("Error generating a new session")
	}
	awsConfig := awsSession.Config

	assert.Equal(t, expectedRegion, aws.StringValue(awsConfig.Region), "Expected region to match")
}

//-------------------------------END OF REGION TESTS----------------------------

// 1a) Use AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY env variables
func TestCredentialsWhenUsingEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.Region = region

	// set variables for test
	os.Setenv("AWS_ACCESS_KEY_ID", envAwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", envAwsSecretKey)
	defer os.Clearenv()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, envAwsAccessKey, envAwsSecretKey)
}

// 1b) Use AWS_ACCESS_KEY and AWS_SECRET_KEY env variables
func TestCredentialsWhenUsingDefaultEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.Region = region

	// set variables for test
	os.Setenv("AWS_ACCESS_KEY", envAwsAccessKey)
	os.Setenv("AWS_SECRET_KEY", envAwsSecretKey)
	defer os.Clearenv()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, envAwsAccessKey, envAwsSecretKey)
}

// 2) Use access and secrets keys from ECS Config
func TestCredentialsWhenUsingECSConfig(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.Region = region

	// set variables for test
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey
	ecsConfig.AWSSessionToken = awsToken

	// invoke test and verify
	testCredentialsInSessionWithToken(t, ecsConfig, awsAccessKey, awsSecretKey, awsToken)
}

// 3a) Use credentials from profile in ECS Config
func TestCredentialsWhenUsingECSConfigProfile(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.AWSProfile = customProfileName

	// set variables for test
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, customAwsAccessKey, customAwsSecretKey)
}

// 3b) Use credentials from AWS_PROFILE
func TestCredentialsWhenUsingAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_PROFILE", customProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, customAwsAccessKey, customAwsSecretKey)
}

// 3c) Use Region from AWS_DEFAULT_PROFILE
func TestCredentialsWhenUsingDefaultAWSProfileEnvVariable(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, defaultAwsAccessKey, defaultAwsSecretKey)
}

// 3d) Use credentials from assume role profile
func TestCredentialsWhenUsingAssumeRoleAWSProfileFlag(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.AWSProfileFlag, customProfileName, "")
	context := cli.NewContext(nil, flagSet, nil)
	ecsConfig.AWSProfile = assumeRoleName
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	startingConfig := assumeRoleTestHelper()

	// invoke test and verify
	testCredentialsInSessionWithConfig(t, context, ecsConfig, startingConfig, assumeRoleAccessKey, assumeRoleSecretKey)
}

func TestCredentialsWhenUsingAssumeRoleEnvVar(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, nil)
	os.Setenv("AWS_DEFAULT_PROFILE", assumeRoleName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	startingConfig := assumeRoleTestHelper()

	// invoke test and verify
	testCredentialsInSessionWithConfig(t, context, ecsConfig, startingConfig, assumeRoleAccessKey, assumeRoleSecretKey)
}

func assumeRoleTestHelper() *aws.Config {
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

	return &startingConfig
}

//4) Use credentials from EC2 Instance Role
func TestCredentialsWhenUsingEC2InstanceRole(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, nil)
	os.Setenv("AWS_DEFAULT_PROFILE", ec2InstanceRoleName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

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
	testCredentialsInSessionWithConfig(t, context, ecsConfig, &startingConfig, ec2InstanceRoleAccessKey, ec2InstanceRoleSecretKey)
}

// Error if Session.Credentials are nil
func TestCredentialsWhenNoneSpecified(t *testing.T) {
	// defaults
	os.Clearenv()
	os.Setenv("AWS_CONTAINER_CREDENTIALS_RELATIVE_URI", "somevalue")
	ecsConfig := NewCLIConfig(clusterName)
	ecsConfig.Region = region

	// NOTE: no credentials set

	// invoke test and verify
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, nil)
	awsSession, err := ecsConfig.ToAWSSession(context)
	assert.NoError(t, err, "Unexpected error generating a new session")

	awsConfig := awsSession.Config
	_, err = awsConfig.Credentials.Get()
	assert.Error(t, err, "Expected error getting credentials")
}

func TestCredentialOrderOfResolutionECSProfileFlag(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	os.Setenv("AWS_ACCESS_KEY_ID", envAwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", envAwsSecretKey)
	defer os.Clearenv()

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.ECSProfileFlag, "default", "")
	context := cli.NewContext(nil, flagSet, nil)

	// invoke test and verify
	testCredentialsInSessionWithContext(t, context, ecsConfig, awsAccessKey, awsSecretKey)
}

func TestCredentialOrderOfResolutionAWSProfileFlag(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	ecsConfig.AWSProfile = customProfileName
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	os.Setenv("AWS_ACCESS_KEY_ID", envAwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", envAwsSecretKey)
	defer os.Clearenv()

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.AWSProfileFlag, customProfileName, "")
	context := cli.NewContext(nil, flagSet, nil)

	// invoke test and verify
	testCredentialsInSessionWithContext(t, context, ecsConfig, customAwsAccessKey, customAwsSecretKey)
}

func TestCredentialOrderOfResolutionEnvVar(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	os.Setenv("AWS_ACCESS_KEY_ID", envAwsAccessKey)
	os.Setenv("AWS_SECRET_ACCESS_KEY", envAwsSecretKey)
	defer os.Clearenv()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, envAwsAccessKey, envAwsSecretKey)
}

func TestCredentialOrderOfResolutionECSConfig(t *testing.T) {
	// defaults
	ecsConfig := NewCLIConfig(clusterName)

	// set variables for test
	ecsConfig.AWSAccessKey = awsAccessKey
	ecsConfig.AWSSecretKey = awsSecretKey
	os.Setenv("AWS_DEFAULT_PROFILE", defaultProfileName)
	os.Setenv("AWS_CONFIG_FILE", "aws_config_example.ini")
	os.Setenv("AWS_SHARED_CREDENTIALS_FILE", "aws_credentials_example.ini")
	defer os.Clearenv()

	// invoke test and verify
	testCredentialsInSession(t, ecsConfig, awsAccessKey, awsSecretKey)
}

func testCredentialsInSessionWithToken(t *testing.T, inputConfig *CLIConfig, expectedAccessKey, expectedSecretKey, expectedSessionToken string) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, nil)
	awsSession, err := inputConfig.ToAWSSession(context)
	assert.NoError(t, err, "Unexpected error generating a new session")

	verifyCredentialsInSessionWithToken(t, awsSession, expectedAccessKey, expectedSecretKey, expectedSessionToken)
}

func testCredentialsInSession(t *testing.T, inputConfig *CLIConfig, expectedAccessKey, expectedSecretKey string) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, nil)
	awsSession, err := inputConfig.ToAWSSession(context)
	assert.NoError(t, err, "Unexpected error generating a new session")

	verifyCredentialsInSession(t, awsSession, expectedAccessKey, expectedSecretKey)
}

func testCredentialsInSessionWithContext(t *testing.T, context *cli.Context, inputConfig *CLIConfig, expectedAccessKey, expectedSecretKey string) {
	awsSession, err := inputConfig.ToAWSSession(context)
	assert.NoError(t, err, "Unexpected error generating a new session")

	verifyCredentialsInSession(t, awsSession, expectedAccessKey, expectedSecretKey)
}

func testCredentialsInSessionWithConfig(t *testing.T, context *cli.Context, inputConfig *CLIConfig, ecsConfig *aws.Config,
	expectedAccessKey, expectedSecretKey string) {
	awsSession, err := inputConfig.toAWSSessionWithConfig(context, ecsConfig)
	assert.NoError(t, err, "Unexpected error generating a new session")

	verifyCredentialsInSession(t, awsSession, expectedAccessKey, expectedSecretKey)
}

func verifyCredentialsInSession(t *testing.T, awsSession *session.Session, expectedAccessKey, expectedSecretKey string) {
	awsConfig := awsSession.Config
	resolvedCredentials, err := awsConfig.Credentials.Get()
	assert.NoError(t, err, "Unexpected error fetching credentials from the chain provider")
	assert.Equal(t, expectedAccessKey, resolvedCredentials.AccessKeyID, "Expected access key to match")
	assert.Equal(t, expectedSecretKey, resolvedCredentials.SecretAccessKey, "Expected secret key to match")
}

func verifyCredentialsInSessionWithToken(t *testing.T, awsSession *session.Session, expectedAccessKey, expectedSecretKey, expectedToken string) {
	awsConfig := awsSession.Config
	resolvedCredentials, err := awsConfig.Credentials.Get()
	assert.NoError(t, err, "Unexpected error fetching credentials from the chain provider")
	assert.Equal(t, expectedAccessKey, resolvedCredentials.AccessKeyID, "Expected access key to match")
	assert.Equal(t, expectedSecretKey, resolvedCredentials.SecretAccessKey, "Expected secret key to match")
	assert.Equal(t, expectedToken, resolvedCredentials.SessionToken, "Expected session token to match")
}
