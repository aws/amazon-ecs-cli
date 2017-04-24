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

package main

import (
	"os"

	"github.com/aws/amazon-ecs-cli/ecs-cli/license"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/command"
	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/cli/ecs/app"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/version"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils/logger"
	"github.com/cihub/seelog"
	"github.com/urfave/cli"
)

func main() {
	// Setup seelog for amazon-ecr-credential-helper
	logger.SetupLogger()
	defer seelog.Flush()

	app := cli.NewApp()
	app.Name = version.AppName
	app.Usage = "Command line interface for Amazon ECS"
	app.Version = version.String()
	app.Author = "Amazon Web Services"

	composeFactory := ecscompose.NewProjectFactory()

	app.Commands = []cli.Command{
		command.ConfigureCommand(),
		command.UpCommand(),
		command.DownCommand(),
		command.ScaleCommand(),
		command.PsCommand(),
		command.PushCommand(),
		command.PullCommand(),
		command.ImagesCommand(),
		license.LicenseCommand(),
		ecscompose.ComposeCommand(composeFactory),
	}
	app.Run(os.Args)
}
