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

package factory

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/project/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	composeFileNameTest = "docker-compose-test.yml"
	projectNameTest     = "project"
)

func TestPopulateContext(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	cliContext := cli.NewContext(nil, flagSet, globalContext)
	ecsContext := &context.Context{}

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}

	setUpTempEnvironment(t, tempDirName)
	defer removeTempEnvironment(tempDirName)

	// write a dummy ecs config file
	saveDummyConfig(t, tempDirName)

	projectFactory := projectFactory{}
	err = projectFactory.populateContext(ecsContext, cliContext)

	if err != nil {
		t.Fatal("Error while populating the context")
	}

	if ecsContext.CLIParams == nil {
		t.Error("CLI Params was expected to be set for ecsContext but was nil")
	}
}

func setUpTempEnvironment(t *testing.T, tempDirName string) {
	// Create a temprorary directory for the dummy ecs config
	os.Setenv("HOME", tempDirName)
	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
}

func removeTempEnvironment(tempDirName string) {
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_ACCESS_KEY")
	os.Unsetenv("AWS_SECRET_KEY")
	os.Unsetenv("HOME")
	os.RemoveAll(tempDirName)
}

func saveDummyConfig(t *testing.T, tempDirName string) {
	configContents := `[ecs]
cluster = testCluster
aws_profile =
region = us-west-2
aws_access_key_id = ***
aws_secret_access_key = ***
compose-project-name-prefix =
compose-service-name-prefix =
cfn-stack-name-prefix =
`
	err := os.MkdirAll(tempDirName+"/.ecs", 0777)
	assert.NoError(t, err, "Could not create config directory")

	err = ioutil.WriteFile(tempDirName+"/.ecs/config", []byte(configContents), 0600)
	assert.NoError(t, err)
}

func TestPopulateContextWithGlobalFlagOverrides(t *testing.T) {
	// populate when compose file and project name flag overrides are provided
	overrides := flag.NewFlagSet("ecs-cli", 0)
	composeFiles := &cli.StringSlice{}
	composeFiles.Set(composeFileNameTest)
	// test multiple --file
	composeFiles.Set("docker-compose-test2.yml")
	overrides.Var(composeFiles, flags.ComposeFileNameFlag, "")
	overrides.String(flags.ProjectNameFlag, projectNameTest, "")
	parentContext := cli.NewContext(nil, overrides, nil)
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	cliContext := cli.NewContext(nil, flagSet, parentContext)
	ecsContext := &context.Context{}

	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	setUpTempEnvironment(t, tempDirName)
	defer removeTempEnvironment(tempDirName)

	// write a dummy ecs config file
	saveDummyConfig(t, tempDirName)

	projectFactory := projectFactory{}
	err = projectFactory.populateContext(ecsContext, cliContext)

	assert.NoError(t, err, "Unexpected error")
	assert.Len(t, ecsContext.ComposeFiles, 2, "Expected composeFiles to be set")
	assert.Equal(t, composeFileNameTest, ecsContext.ComposeFiles[0], "Expected compose file to match")
	assert.Equal(t, projectNameTest, ecsContext.ProjectName, "Expected project name to match")
}

func TestLoadProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProject := mock_project.NewMockProject(ctrl)
	var expectedErr error
	mockProject.EXPECT().Parse().Return(expectedErr)

	projectFactory := projectFactory{}
	observedErr := projectFactory.loadProject(mockProject)

	if expectedErr != observedErr {
		t.Errorf("LoadProject should mimic what Project.Parse returns. Unexpected error [%s] was thrown", observedErr)
	}
}
