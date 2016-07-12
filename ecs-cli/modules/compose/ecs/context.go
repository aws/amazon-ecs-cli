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

package ecs

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	ec2client "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ec2"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
)

const ProjectNameFlag = "project-name"

// Context is a wrapper around libcompose.project.Context
type Context struct {
	project.Context

	CLIContext *cli.Context
	ECSParams  *config.CliParams

	// AWS Service Clients
	ECSClient ecsclient.ECSClient
	EC2Client ec2client.EC2Client

	// IsService would decide if the resource created by this compose project would be ECS Tasks directly or through ECS Services
	IsService bool
}

// open populates the context with new ECS and EC2 Clients
func (context *Context) open() error {
	// setup aws service clients
	context.ECSClient = ecsclient.NewECSClient()
	context.ECSClient.Initialize(context.ECSParams)

	context.EC2Client = ec2client.NewEC2Client(context.ECSParams)

	return nil
}

// SetProjectName sets the project name, which is resolved in the order
// 1. Command line option
// 2. Environment variable
// 3. Current working directory
func (context *Context) SetProjectName() error {
	projectName := context.CLIContext.GlobalString(ProjectNameFlag)
	if projectName != "" {
		context.ProjectName = projectName
		return nil
	}
	projectName, err := context.lookupProjectName()
	if err != nil {
		return err
	}
	context.ProjectName = projectName
	return nil
}

// This following is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found :
// https://github.com/docker/libcompose/blob/master/project/context.go
func (c *Context) lookupProjectName() (string, error) {
	file := "."
	if len(c.ComposeFiles) > 0 {
		file = c.ComposeFiles[0]
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
