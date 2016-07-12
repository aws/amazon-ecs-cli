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
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/codegangsta/cli"
)

// displayTitle flag is used to print the title for the fields
const displayTitle = true

// ProjectAction is an adapter to allow the use of ordinary functions as libcompose actions.
// Any function that has the appropriate signature can be register as an action on a codegansta/cli command.
//
// cli.Command{
//		Name:   "ps",
//		Usage:  "List containers",
//		Action: app.WithProject(factory, app.ProjectPs),
//	}
type ProjectAction func(project ecscompose.Project, c *cli.Context)

// WithProject is an helper function to create a cli.Command action with a ProjectFactory.
func WithProject(factory ProjectFactory, action ProjectAction, isService bool) func(context *cli.Context) {
	return func(context *cli.Context) {
		// TODO, instead of passing isService around, we can determine
		// the command name cliContext.Parent().Command.Name = service and set appropriate context
		// However, parentContext is not being set appropriately by cli. Investigate.
		p, err := factory.Create(context, isService)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Unable to create and read ECS Compose Project")
		}
		action(p, context)
	}
}

// ProjectCreate creates the task definition required for the containers but does not start them.
func ProjectCreate(p ecscompose.Project, c *cli.Context) {
	err := p.Create()
	if err != nil {
		log.Fatal(err)
	}
}

// ProjectStart starts containers.
func ProjectStart(p ecscompose.Project, c *cli.Context) {
	err := p.Start()
	if err != nil {
		log.Fatal(err)
	}
}

// ProjectUp brings all containers up.
func ProjectUp(p ecscompose.Project, c *cli.Context) {
	err := p.Up()
	if err != nil {
		log.Fatal(err)
	}
}

// ProjectPs lists the containers.
func ProjectPs(p ecscompose.Project, c *cli.Context) {
	allInfo, err := p.Info()
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.WriteString(allInfo.String(displayTitle))
}

// ProjectRun starts containers and executes one-time command against the container
func ProjectRun(p ecscompose.Project, c *cli.Context) {
	args := c.Args()
	if len(args)%2 != 0 {
		log.Fatal("Please pass arguments in the form: CONTAINER COMMAND [CONTAINER COMMAND]...")
	}
	commandOverrides := make(map[string]string)
	for i := 0; i < len(args); i += 2 {
		commandOverrides[args[i]] = args[i+1]
	}
	err := p.Run(commandOverrides)
	if err != nil {
		log.Fatal(err)
	}
}

// ProjectScale scales containers.
func ProjectScale(p ecscompose.Project, c *cli.Context) {
	if len(c.Args()) != 1 {
		log.Fatal("Please pass arguments in the form: ecs-cli compose scale COUNT")
	}
	count, err := strconv.Atoi(c.Args().First())
	if err != nil {
		log.Fatal("Please pass an integer value for argument COUNT")
	}
	err = p.Scale(count)
	if err != nil {
		log.Fatal(err)
	}
}

// ProjectStop brings all containers down.
func ProjectStop(p ecscompose.Project, c *cli.Context) {
	err := p.Stop()
	if err != nil {
		log.Fatal(err)
	}
}

// ProjectDown brings all containers down.
func ProjectDown(p ecscompose.Project, c *cli.Context) {
	err := p.Down()
	if err != nil {
		log.Fatal(err)
	}
}

// common function for all unsupported operations
func unsupportedOperation(c *cli.Context) {
	log.WithFields(log.Fields{
		"command": c.Command.Name,
		"Error":   utils.ErrUnsupported,
	}).Error("Unable to run command")
}
