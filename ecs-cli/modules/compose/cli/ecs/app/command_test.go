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
	"testing"

	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs"
	"github.com/codegangsta/cli"
)

const (
	composeFileNameTest = "docker-compose-test.yml"
	projectNameTest     = "project"
)

func TestPopulateWithGlobalFlagOverrides(t *testing.T) {
	// populate when compose file and project name flag overrides are provided
	setComposeFileAndProjectName := flag.NewFlagSet("ecs-cli", 0)
	setComposeFileAndProjectName.String(composeFileNameFlag, composeFileNameTest, "")
	setComposeFileAndProjectName.String(projectNameFlag, projectNameTest, "")
	parentContext := cli.NewContext(nil, setComposeFileAndProjectName, nil)
	cliContext := cli.NewContext(nil, nil, parentContext)
	ecsContext := &ecscompose.Context{}

	populate(ecsContext, cliContext)

	if len(ecsContext.ComposeFiles) != 1 {
		t.Fatalf("ComposeFiles not set. Expected [%s] Got empty", composeFileNameTest)
	}
	if composeFileNameTest != ecsContext.ComposeFiles[0] {
		t.Errorf("ComposeFile not overriden. Expected [%s] Got [%s]", composeFileNameTest, ecsContext.ComposeFiles[0])
	}
	if projectNameTest != ecsContext.ProjectName {
		t.Errorf("ProjectName not overriden. Expected [%s] Got [%s]", projectNameTest, ecsContext.ProjectName)
	}
}
