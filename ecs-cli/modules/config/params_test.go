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
	"flag"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/codegangsta/cli"
)

// mockReadWriter implements ReadWriter interface to return just the cluster
// field whenperforming read.
type mockReadWriter struct{}

func (rdwr *mockReadWriter) GetConfig() (*CliConfig, error) {
	return NewCliConfig(clusterName), nil
}

func (rdwr *mockReadWriter) ReadFrom(ecsConfig *CliConfig) error {
	return nil
}

func (rdwr *mockReadWriter) IsInitialized() (bool, error) {
	return true, nil
}

func (rdwr *mockReadWriter) Save(dest *Destination) error {
	return nil
}

func TestNewCliParamsFromEnvVars(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")
	_, err := NewCliParams(context, rdwr)
	if err == nil {
		t.Errorf("Expected error when region not specified")
	}

	os.Setenv("AWS_REGION", "us-west-1")
	params, err := NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error when region is specified using environment variable AWS_REGION: ", err)
	}

	paramsRegion := aws.StringValue(params.Config.Region)
	if "us-west-1" != paramsRegion {
		t.Errorf("Unexpected region set, expected: us-west-1, got: %s", paramsRegion)
	}
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")

	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")
	params, err = NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error when region is specified using environment variable AWS_DEFAULT_REGION: ", err)
	}
	paramsRegion = aws.StringValue(params.Config.Region)
	if "us-west-2" != paramsRegion {
		t.Errorf("Unexpected region set, expected: us-west-2, got: %s", paramsRegion)
	}
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_DEFAULT_REGION")
}

func TestNewCliParamsFromConfig(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}
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
