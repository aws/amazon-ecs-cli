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
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	composeServiceNamePrefix = "ecs-service-"
	composeProjectNamePrefix = "ecs-project-"
	cfnStackName             = "cfn-stack-ecs"
	cfnStackNamePrefix       = "cfn-stack-"
	awsAccess                = "ecs-access"
	awsSecret                = "ecs-secret"
	awsAccessAWSProfile      = "aws-access"
	awsSecretAWSProfile      = "aws-secret"
	awsProfileName           = "awsprofile"
)

// mockReadWriter implements ReadWriter interface
// field whenperforming read.
type mockReadWriter struct {
	isKeyPresentValue bool
	fargate           bool
	version           int
}

func (rdwr *mockReadWriter) Get(clusterConfig string, profileConfig string) (*CLIConfig, error) {
	config := NewCLIConfig(clusterName)
	if rdwr.isKeyPresentValue && rdwr.version == iniConfigVersion {
		config.ComposeServiceNamePrefix = composeServiceNamePrefix
		config.CFNStackNamePrefix = cfnStackNamePrefix
		config.ComposeProjectNamePrefix = composeProjectNamePrefix
	}
	if rdwr.isKeyPresentValue && rdwr.version == yamlConfigVersion {
		config.ComposeServiceNamePrefix = composeServiceNamePrefix
		config.CFNStackName = cfnStackName
		config.DefaultLaunchType = LaunchTypeEC2
		if rdwr.fargate {
			config.DefaultLaunchType = LaunchTypeFargate
		}
	}
	config.Version = rdwr.version
	return config, nil
}

func (rdwr *mockReadWriter) SaveProfile(configName string, profile *Profile) error {
	return nil
}

func (rdwr *mockReadWriter) SaveCluster(configName string, cluster *Cluster) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultProfile(configName string) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultCluster(configName string) error {
	return nil
}

func TestNewCliParamsFromEnvVarsWithRegionNotSpecified(t *testing.T) {
	context, rdwr := setupTest(t)

	_, err := NewCLIParams(context, rdwr)
	if err == nil {
		t.Errorf("Expected error when region not specified")
	}
}

func TestNewCliParamsFromEnvVarsWithRegionSpecifiedAsEnvVariable(t *testing.T) {
	region := "us-west-1"
	context, rdwr := setupTest(t)

	os.Setenv("AWS_REGION", region)
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when region is specified using environment variable AWS_REGION")

	paramsRegion := aws.StringValue(params.Session.Config.Region)
	assert.Equal(t, region, paramsRegion, "Region should match")
}

func TestNewCliParamsFromEnvVarsWithRegionSpecifiedinAwsDefaultEnvVariable(t *testing.T) {
	region := "us-west-2"
	context, rdwr := setupTest(t)

	os.Setenv("AWS_DEFAULT_REGION", region)
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when region is specified using environment variable AWS_DEFAULT_REGION")

	paramsRegion := aws.StringValue(params.Session.Config.Region)
	assert.Equal(t, region, paramsRegion, "Region should match")
}

func TestNewCliParamsFromConfig(t *testing.T) {
	region := "us-east-1"

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String("region", region, "")
	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := &mockReadWriter{}

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when region is specified")

	paramsRegion := aws.StringValue(params.Session.Config.Region)
	assert.Equal(t, region, paramsRegion, "Region should match")
}

func TestNewCliParamsWhenPrefixesPresentINIVersion(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	context := defaultConfig()

	// Prefixes are present, and values are defaulted to empty
	rdwr := &mockReadWriter{isKeyPresentValue: true, version: iniConfigVersion}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new cli params")
	assert.Equal(t, composeProjectNamePrefix, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be set")
	assert.Equal(t, composeServiceNamePrefix, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be set")
	assert.Equal(t, cfnStackNamePrefix+clusterName, params.CFNStackName, "Expected CFNStackName to be default")
	assert.Empty(t, params.LaunchType, "Expected Launch Type to be empty")
}

func TestNewCliParamsWhenPrefixKeysAreNotPresentINIVersion(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	context := defaultConfig()

	// Prefixes are present, and values should be set to defaults
	rdwr := &mockReadWriter{isKeyPresentValue: false, version: iniConfigVersion}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new CLI params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Empty(t, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be empty")
	assert.Equal(t, clusterName, params.CFNStackName, "Expected CFNStackName to equal cluster name")
	assert.Empty(t, params.LaunchType, "Expected Launch Type to be empty")
}

func TestNewCliParamsINIVersionLaunchTypeFlagEC2(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	context := configWithLaunchType(LaunchTypeEC2)

	// Prefixes are present, and values should be set to defaults
	rdwr := &mockReadWriter{isKeyPresentValue: false, version: iniConfigVersion}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new CLI params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Empty(t, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be empty")
	assert.Equal(t, clusterName, params.CFNStackName, "Expected CFNStackName to equal cluster name")
	assert.Equal(t, LaunchTypeEC2, params.LaunchType)
}

func TestNewCliParamsINIVersionLaunchTypeFlagFargate(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	context := configWithLaunchType(LaunchTypeFargate)

	// Prefixes are present, and values should be set to defaults
	rdwr := &mockReadWriter{isKeyPresentValue: false, version: iniConfigVersion}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new CLI params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Empty(t, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be empty")
	assert.Equal(t, clusterName, params.CFNStackName, "Expected CFNStackName to equal cluster name")
	assert.Equal(t, LaunchTypeFargate, params.LaunchType)
}

func TestNewCliParamsWhenPrefixesPresentYAMLVersion(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	context := defaultConfig()

	// Prefixes are present, and values are defaulted to empty
	rdwr := &mockReadWriter{isKeyPresentValue: true, version: yamlConfigVersion}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new cli params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Equal(t, composeServiceNamePrefix, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be set")
	assert.Equal(t, cfnStackName, params.CFNStackName, "Expected CFNStackName to be set")
	assert.Equal(t, LaunchTypeEC2, params.LaunchType)
}

func TestNewCliParamsWhenPrefixKeysAreNotPresentYAMLVersion(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	context := defaultConfig()

	// Prefixes are present, and values should be set to defaults
	rdwr := &mockReadWriter{isKeyPresentValue: false, version: yamlConfigVersion}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new cli params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposProjectNamePrefix to be empty")
	assert.Empty(t, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be empty")
	assert.Equal(t, flags.CFNStackNamePrefixDefaultValue+clusterName, params.CFNStackName, "Expected CFNStackName to be default")
	assert.Empty(t, params.LaunchType, "Expected Launch Type to be empty")
}

func TestNewCliParamsYAMLVersionLaunchTypeEC2(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	context := defaultConfig()

	// Prefixes are present, and values are defaulted to empty
	rdwr := &mockReadWriter{isKeyPresentValue: true, version: yamlConfigVersion, fargate: false}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new cli params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Equal(t, composeServiceNamePrefix, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be set")
	assert.Equal(t, cfnStackName, params.CFNStackName, "Expected CFNStackName to be set")
	assert.Equal(t, LaunchTypeEC2, params.LaunchType)
}

func TestNewCliParamsYAMLVersionLaunchTypeFargate(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	context := defaultConfig()

	// Prefixes are present, and values are defaulted to empty
	rdwr := &mockReadWriter{isKeyPresentValue: true, version: yamlConfigVersion, fargate: true}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new cli params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Equal(t, composeServiceNamePrefix, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be set")
	assert.Equal(t, cfnStackName, params.CFNStackName, "Expected CFNStackName to be set")
	assert.Equal(t, LaunchTypeFargate, params.LaunchType)
}

func TestNewCliParamsYAMLVersionLaunchTypeOverriddenFargate(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	context := configWithLaunchType(LaunchTypeFargate)

	// Prefixes are present, and values are defaulted to empty
	rdwr := &mockReadWriter{isKeyPresentValue: true, version: yamlConfigVersion, fargate: false}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new cli params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Equal(t, composeServiceNamePrefix, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be set")
	assert.Equal(t, cfnStackName, params.CFNStackName, "Expected CFNStackName to be set")
	assert.Equal(t, LaunchTypeFargate, params.LaunchType)
}

func TestNewCliParamsYAMLVersionLaunchTypeOverriddenEC2(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	context := configWithLaunchType(LaunchTypeEC2)

	// Prefixes are present, and values are defaulted to empty
	rdwr := &mockReadWriter{isKeyPresentValue: true, version: yamlConfigVersion, fargate: true}
	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when getting new cli params")
	assert.Empty(t, params.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Equal(t, composeServiceNamePrefix, params.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be set")
	assert.Equal(t, cfnStackName, params.CFNStackName, "Expected CFNStackName to be set")
	assert.Equal(t, LaunchTypeEC2, params.LaunchType)
}

func TestNewCliParamsWithAWSProfile(t *testing.T) {
	// Keys in env vars take highest precedence; ensure they are not set
	os.Unsetenv("AWS_ACCESS_KEY")
	os.Unsetenv("AWS_SECRET_KEY")

	configContents := `[awsprofile]
aws_access_key_id = aws-access
aws_secret_access_key = aws-secret
`
	// Create a temporary directory for the dummy aws config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	os.Setenv("AWS_DEFAULT_REGION", region)
	defer os.Clearenv()
	defer os.RemoveAll(tempDirName)

	// save the aws config
	fileInfo, err := os.Stat(tempDirName)
	assert.NoError(t, err)
	mode := fileInfo.Mode()
	err = os.MkdirAll(tempDirName+"/.aws", mode)
	assert.NoError(t, err, "Could not create aws config directory")
	err = ioutil.WriteFile(tempDirName+"/.aws/credentials", []byte(configContents), mode)
	assert.NoError(t, err)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String("aws-profile", awsProfileName, "")
	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := &mockReadWriter{}

	params, err := NewCLIParams(context, rdwr)
	assert.NoError(t, err)
	creds, err := params.Session.Config.Credentials.Get()
	assert.NoError(t, err)
	assert.Equal(t, awsAccessAWSProfile, creds.AccessKeyID, "Expected AWS Access Key to be read from the AWS Profile")
	assert.Equal(t, awsSecretAWSProfile, creds.SecretAccessKey, "Expected AWS Secret Access Key to be read from the AWS Profile")
}

func defaultConfig() *cli.Context {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String("region", "us-east-1", "")
	return cli.NewContext(nil, flagSet, globalContext)
}

func configWithLaunchType(launchType string) *cli.Context {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String("region", "us-east-1", "")
	flagSet.String(flags.LaunchTypeFlag, launchType, "")
	return cli.NewContext(nil, flagSet, globalContext)
}

func setupTest(t *testing.T) (*cli.Context, *mockReadWriter) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := &mockReadWriter{}
	return context, rdwr
}
