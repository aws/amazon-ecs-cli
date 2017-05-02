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
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
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
	context, rdwr := setupTest(t)

	_, err := NewCliParams(context, rdwr)
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

	params, err := NewCliParams(context, rdwr)
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

	params, err := NewCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when region is specified using environment variable AWS_DEFAULT_REGION")

	paramsRegion := aws.StringValue(params.Session.Config.Region)
	assert.Equal(t, region, paramsRegion, "Region should match")
}

func TestNewCliParamsFromConfig(t *testing.T) {
	region := "us-east-1"

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", region, "")
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "SECRET")
	defer os.Clearenv()

	params, err := NewCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error when region is specified")

	paramsRegion := aws.StringValue(params.Session.Config.Region)
	assert.Equal(t, region, paramsRegion, "Region should match")
}

func setupTest(t *testing.T) (*cli.Context, *mockReadWriter) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}
	return context, rdwr
}
