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
	"flag"
	"os"
	"testing"

	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/codegangsta/cli"
)

// mockReadWriter implements ReadWriter interface to return just the cluster
// field whenperforming read.
type mockReadWriter struct {
	isKeyPresentValue bool
}

func (rdwr *mockReadWriter) GetConfig() (*CliConfig, error) {
	return NewCliConfig(clusterName), nil
}

func (rdwr *mockReadWriter) ReadFrom(ecsConfig *CliConfig) error {
	return nil
}

func (rdwr *mockReadWriter) IsInitialized() (bool, error) {
	return true, nil
}

func (rdwr *mockReadWriter) IsKeyPresent(section, key string) bool {
	return rdwr.isKeyPresentValue
}

func (rdwr *mockReadWriter) Save(dest *Destination) error {
	return nil
}

func TestNewCliParamsFromEnvVarsWithRegionNotSpecified(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}
	_, err := NewCliParams(context, rdwr)
	if err == nil {
		t.Errorf("Expected error when region not specified")
	}
}

func TestNewCliParamsFromEnvVarsWithRegionSpecifiedAsEnvVariable(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}

	os.Setenv("AWS_REGION", "us-west-1")
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	params, err := NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error when region is specified using environment variable AWS_REGION: ", err)
	}

	paramsRegion := aws.StringValue(params.Config.Region)
	if "us-west-1" != paramsRegion {
		t.Errorf("Unexpected region set, expected: us-west-1, got: %s", paramsRegion)
	}
}

func TestNewCliParamsFromEnvVarsWithRegionSpecifiedinAwsDefaultEnvVariable(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}

	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_DEFAULT_REGION")
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	params, err := NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error when region is specified using environment variable AWS_DEFAULT_REGION: ", err)
	}
	paramsRegion := aws.StringValue(params.Config.Region)
	if "us-west-2" != paramsRegion {
		t.Errorf("Unexpected region set, expected: us-west-2, got: %s", paramsRegion)
	}
}

func TestNewCliParamsFromConfig(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	globalSet.String("region", "us-east-1", "")
	globalContext = cli.NewContext(nil, globalSet, nil)
	context = cli.NewContext(nil, nil, globalContext)
	params, err := NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error when region is specified using environment variable AWS_DEFAULT_REGION: ", err)
	}
	paramsRegion := aws.StringValue(params.Config.Region)
	if "us-east-1" != paramsRegion {
		t.Errorf("Unexpected region set, expected: us-east-1, got: %s", paramsRegion)
	}
}

func defaultConfig() *cli.Context {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	globalSet.String("region", "us-east-1", "")
	globalContext = cli.NewContext(nil, globalSet, nil)
	return cli.NewContext(nil, nil, globalContext)
}

func TestNewCliParamsWhenPrefixesPresent(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	context := defaultConfig()

	// Prefixes are present, and values are defaulted to empty
	rdwr := &mockReadWriter{isKeyPresentValue: true}
	params, err := NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error when getting new cli params", err)
	}

	if "" != params.ComposeProjectNamePrefix {
		t.Errorf("Compose project name prefix mismatch. Expected empty string got [%s]", params.ComposeProjectNamePrefix)
	}
	if "" != params.ComposeServiceNamePrefix {
		t.Errorf("Compose service name prefix mismatch. Expected empty string got [%s]", params.ComposeServiceNamePrefix)
	}
	if "" != params.CFNStackNamePrefix {
		t.Errorf("stack name name prefix mismatch. Expected empty string got [%s]", params.CFNStackNamePrefix)
	}

}

func TestNewCliParamsWhenPrefixKeysAreNotPresent(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	context := defaultConfig()

	// Prefixes are present, and values should be set to defaults
	rdwr := &mockReadWriter{isKeyPresentValue: false}
	params, err := NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error when getting new cli params", err)
	}

	if ecscli.ComposeProjectNamePrefixDefaultValue != params.ComposeProjectNamePrefix {
		t.Errorf("Compose project name prefix mismatch. Expected [%s] got [%s]", ecscli.ComposeProjectNamePrefixDefaultValue, params.ComposeProjectNamePrefix)
	}
	if ecscli.ComposeServiceNamePrefixDefaultValue != params.ComposeServiceNamePrefix {
		t.Errorf("Compose service name prefix mismatch. Expected [%s] got [%s]", ecscli.ComposeServiceNamePrefixDefaultValue, params.ComposeServiceNamePrefix)
	}
	if ecscli.CFNStackNamePrefixDefaultValue != params.CFNStackNamePrefix {
		t.Errorf("stack name name prefix mismatch. Expected [%s] got [%s]", ecscli.CFNStackNamePrefixDefaultValue, params.CFNStackNamePrefix)
	}

}
