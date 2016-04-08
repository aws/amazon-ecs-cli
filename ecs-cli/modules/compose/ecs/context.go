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

package ecs

import (
	ec2client "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ec2"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/codegangsta/cli"
)

// Context is a wrapper around libcompose.project.Context
type Context struct {
	libcompose.Context
	CLIContext *cli.Context
	ECSParams  *config.CliParams

	// AWS Service Clients
	ECSClient ecsclient.ECSClient
	EC2Client ec2client.EC2Client

	// IsService would decide if the resource created by this compose project would be ECS Tasks directly or through ECS Services
	IsService bool

	// keeps track of the current instance of project
	project Project
}

// open populates the context fields and creates an ECS Client
func (context *Context) open() error {
	// perform libcomposeContext open which loads the compose file and determines projectname
	err := context.Context.Open()
	if err != nil {
		utils.LogError(err, "Unable to open ECS context")
		return err
	}
	context.ECSClient = ecsclient.NewECSClient()
	context.ECSClient.Initialize(context.ECSParams)

	context.EC2Client = ec2client.NewEC2Client(context.ECSParams)

	context.Context.EnvironmentLookup = &libcompose.OsEnvLookup{}
	context.Context.ConfigLookup = &libcompose.FileConfigLookup{}
	return nil
}

// setProject sets the project instance to this context
func (context *Context) setProject(project Project) {
	context.project = project
}
