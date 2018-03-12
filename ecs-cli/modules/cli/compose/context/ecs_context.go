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

package context

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	ec2client "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ec2"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/docker/libcompose/project"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// ECSContext is a wrapper around libcompose.project.Context
type ECSContext struct {
	project.Context

	CLIContext *cli.Context

	CLIParams *config.CLIParams

	// NOTE: Ideally, would like to only store the non-TaskDef related fields here (e.g. "DeploymentConfig")
	ECSParams *utils.ECSParams

	// AWS Service Clients
	ECSClient ecsclient.ECSClient
	EC2Client ec2client.EC2Client

	// IsService would decide if the resource created by this compose project would be ECS Tasks directly or through ECS Services
	IsService bool
}

// Open populates the ECSContext with new ECS and EC2 Clients
func (ecsContext *ECSContext) Open() error {
	// setup AWS service clients
	ecsContext.ECSClient = ecsclient.NewECSClient()
	ecsContext.ECSClient.Initialize(ecsContext.CLIParams)

	ecsContext.EC2Client = ec2client.NewEC2Client(ecsContext.CLIParams)

	return nil
}

// SetProjectName sets the project name, which is resolved in the order
// 1. Command line option
// 2. Environment variable
// 3. Current working directory
func (ecsContext *ECSContext) SetProjectName() error {
	projectName := ecsContext.CLIContext.GlobalString(flags.ProjectNameFlag)
	if projectName != "" {
		ecsContext.ProjectName = projectName
		return nil
	}
	projectName, err := ecsContext.lookupProjectName()
	if err != nil {
		return err
	}
	ecsContext.ProjectName = projectName
	return nil
}

// This following is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found :
// https://github.com/docker/libcompose/blob/master/project/context.go
func (ecsContext *ECSContext) lookupProjectName() (string, error) {
	file := "."
	if len(ecsContext.ComposeFiles) > 0 {
		file = ecsContext.ComposeFiles[0]
	}

	f, err := filepath.Abs(file)
	if err != nil {
		logrus.Errorf("Failed to get absolute directory for: %s", file)
		return "", err
	}

	f = toUnixPath(f)

	parent := path.Base(path.Dir(f))
	if parent != "" && parent != "." {
		return parent, nil
	} else if wd, err := os.Getwd(); err != nil {
		return "", err
	} else {
		return path.Base(toUnixPath(wd)), nil
	}
}

func toUnixPath(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}
