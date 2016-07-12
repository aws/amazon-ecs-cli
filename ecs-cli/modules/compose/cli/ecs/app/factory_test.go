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

package app

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/mocks"
	"github.com/codegangsta/cli"
	"github.com/golang/mock/gomock"
)

const testProjectName = "projectName"

func TestPopulateContext(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	cliContext := cli.NewContext(nil, nil, globalContext)
	ecsContext := &ecscompose.Context{}

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	defer os.Remove(tempDirName)
	os.Setenv("HOME", tempDirName)

	os.Setenv("AWS_REGION", "us-east-1")
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_REGION")
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
		os.Unsetenv("HOME")
	}()

	projectFactory := projectFactory{}
	err = projectFactory.populateContext(ecsContext, cliContext)

	if err != nil {
		t.Fatal("Error while populating the context")
	}

	if ecsContext.ECSParams == nil {
		t.Error("ECS Params was expected to be set for ecsContext but was nil")
	}
}

func TestLoadProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEcsProject := mock_ecs.NewMockProject(ctrl)
	var expectedErr error
	mockEcsProject.EXPECT().Parse().Return(expectedErr)

	projectFactory := projectFactory{}
	observedErr := projectFactory.loadProject(mockEcsProject)

	if expectedErr != observedErr {
		t.Errorf("LoadProject should mimic what Project.Parse returns. Unexpected error [%s] was thrown", observedErr)
	}
}
