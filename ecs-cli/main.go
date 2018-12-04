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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/factory"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/cluster"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/compose"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/configure"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/image"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/license"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/log"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/regcreds"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/logger"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	// Setup logrus for amazon-ecr-credential-helper
	logger.SetupLogger()

	app := cli.NewApp()
	app.Name = version.AppName
	app.Usage = "Command line interface for Amazon ECS"
	app.Version = version.String()
	app.Author = "Amazon Web Services"

	composeFactory := factory.NewProjectFactory()

	app.Commands = []cli.Command{
		configureCommand.ConfigureCommand(),
		clusterCommand.UpCommand(),
		clusterCommand.DownCommand(),
		clusterCommand.ScaleCommand(),
		clusterCommand.PsCommand(),
		imageCommand.PushCommand(),
		imageCommand.PullCommand(),
		imageCommand.ImagesCommand(),
		licenseCommand.LicenseCommand(),
		composeCommand.ComposeCommand(composeFactory),
		logsCommand.LogCommand(),
		regcredsCommand.RegistryCredsCommand(),
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  flags.EndpointFlag,
			Usage: "Use a custom endpoint with the ECS CLI",
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		logrus.Fatal(err)
	}
}
