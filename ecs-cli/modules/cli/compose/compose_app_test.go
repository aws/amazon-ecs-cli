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

package compose

import (
	"flag"
	"strconv"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/factory/mock"
	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/project"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/project/mock"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
)

func TestWithProject(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	cliContext := cli.NewContext(nil, nil, globalContext)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProjectFactory := mock_factory.NewMockProjectFactory(ctrl)
	mockProjectFactory.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, nil)

	testFuncVisited := false
	testFunc := func(project ecscompose.Project, c *cli.Context) {
		testFuncVisited = true
	}

	function := WithProject(mockProjectFactory, testFunc, false)
	function(cliContext)

	if !testFuncVisited {
		t.Error("Expected test function to be visited but wasn't")
	}
}

func TestRun(t *testing.T) {
	containers := []string{"cont1", "cont2"}
	commands := []string{"cmd1 cmd2", "cmd3"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProject := mock_project.NewMockProject(ctrl)
	mockProject.EXPECT().Run(map[string][]string{"cont1": {"cmd1", "cmd2"}, "cont2": {"cmd3"}}).Return(nil)

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)
	// flag with 2 containers with 2 commands
	flagSet.Parse([]string{containers[0], commands[0], containers[1], commands[1]})

	ProjectRun(mockProject, cliContext)
}

func TestScale(t *testing.T) {
	expectedCount := 5

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProject := mock_project.NewMockProject(ctrl)
	mockProject.EXPECT().Scale(expectedCount).Return(nil)

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)
	flagSet.Parse([]string{strconv.Itoa(expectedCount)})

	ProjectScale(mockProject, cliContext)
}
