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

package serviceCommand

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/service"
	composeFactory "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/factory"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

//* ----------------- COMPOSE PROJECT with ECS Service ----------------- */
// Note: A project is scoped to a single compose yaml with multiple containers defined
// and today, 1 compose.yml has 1:1 mapping with a task definition and a ECS Service.
// TODO: Split single compose to disjoint task definitions, so they can be run/scaled independently
//
// ---- LIFECYCLE ----
// Create and Start a project with service:
//   ecs-cli compose service create      : creates ECS.CreateTaskDefinition or gets from FS cache and ECS.CreateService with desiredCount=0
//   ecs-cli compose service start       : invokes ECS.UpdateService with desiredCount=1
//   ecs-cli compose service up          : compose service create ; compose service start. If the compose yml was changed, it updates the service with new task definition
// List containers in or view details of the project:
//   ecs-cli compose service ps          : calls ECS.ListTasks of this service
// Modify containers
//   ecs-cli compose service scale       : calls ECS.UpdateService with new count
// Stop and delete the project
//   ecs-cli compose service stop        : calls ECS.UpdateService with count=0
//   ecs-cli compose service down        : calls ECS.DeleteService
//* -------------------------------------------------------------------- */

// ServiceCommand provides a list of commands that operate on docker-compose.yml file
// and are integrated to run on ECS as a service
func ServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:  "service",
		Usage: "Manage Amazon ECS services with docker-compose-style commands on an ECS cluster.",
		Subcommands: []cli.Command{
			createServiceCommand(factory),
			startServiceCommand(factory),
			upServiceCommand(factory),
			psServiceCommand(factory),
			scaleServiceCommand(factory),
			stopServiceCommand(factory),
			rmServiceCommand(factory),
		},
		Flags: flags.OptionalConfigFlags(),
	}
}

func createServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "create",
		Usage:        "Creates an ECS service from your compose file. The service is created with a desired count of 0, so no containers are started by this command. Note that we do not recommend using plain text environment variables for sensitive information, such as credential data.",
		Action:       compose.WithProject(factory, compose.ProjectCreate, true),
		Flags:        append(append(deploymentConfigFlags(true), append(loadBalancerFlags(), flags.OptionalConfigFlags()...)...), flags.OptionalLaunchTypeFlag()),
		OnUsageError: flags.UsageErrorFactory("create"),
	}
}

func startServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "start",
		Usage:        "Starts one copy of each of the containers on an existing ECS service by setting the desired count to 1 (only if the current desired count is 0).",
		Action:       compose.WithProject(factory, compose.ProjectStart, true),
		Flags:        append(flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag()),
		OnUsageError: flags.UsageErrorFactory("start"),
	}
}

func upServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "up",
		Usage:        "Creates a new ECS service or updates an existing one according to your compose file. For new services or existing services with a current desired count of 0, the desired count for the service is set to 1. For existing services with non-zero desired counts, a new task definition is created to reflect any changes to the compose file and the service is updated to use that task definition. In this case, the desired count does not change.",
		Action:       compose.WithProject(factory, compose.ProjectUp, true),
		Flags:        append(append(append(deploymentConfigFlags(true), append(loadBalancerFlags(), flags.OptionalConfigFlags()...)...), ComposeServiceTimeoutFlag()), flags.OptionalLaunchTypeFlag()),
		OnUsageError: flags.UsageErrorFactory("up"),
	}
}

func psServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "ps",
		Aliases:      []string{"list"},
		Usage:        "Lists all the containers in your cluster that belong to the service created with the compose project.",
		Action:       compose.WithProject(factory, compose.ProjectPs, true),
		Flags:        flags.OptionalConfigFlags(),
		OnUsageError: flags.UsageErrorFactory("ps"),
	}
}

func scaleServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "scale",
		Usage:        "ecs-cli compose service scale [count] - scales the desired count of the service to the specified count",
		Action:       compose.WithProject(factory, compose.ProjectScale, true),
		Flags:        append(append(deploymentConfigFlags(false), flags.OptionalConfigFlags()...), ComposeServiceTimeoutFlag()),
		OnUsageError: flags.UsageErrorFactory("scale"),
	}
}

func stopServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "stop",
		Usage:        "Stops the running tasks that belong to the service created with the compose project. This command updates the desired count of the service to 0.",
		Action:       compose.WithProject(factory, compose.ProjectStop, true),
		Flags:        append(flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag()),
		OnUsageError: flags.UsageErrorFactory("stop"),
	}
}

func rmServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "rm",
		Aliases:      []string{"delete", "down"},
		Usage:        "Updates the desired count of the service to 0 and then deletes the service.",
		Action:       compose.WithProject(factory, compose.ProjectDown, true),
		Flags:        append(flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag()),
		OnUsageError: flags.UsageErrorFactory("rm"),
	}
}

func deploymentConfigFlags(specifyDefaults bool) []cli.Flag {
	maxPercentUsageString := "[Optional] Specifies the upper limit (as a percentage of the service's desiredCount) of the number of running tasks that can be running in a service during a deployment."
	minHealthyPercentUsageString := "[Optional] Specifies the lower limit (as a percentage of the service's desiredCount) of the number of running tasks that must remain running and healthy in a service during a deployment."
	if specifyDefaults {
		maxPercentUsageString += fmt.Sprintf(" Defaults to %d.", flags.DeploymentMaxPercentDefaultValue)
		minHealthyPercentUsageString += fmt.Sprintf(" Defaults to %d.", flags.DeploymentMinHealthyPercentDefaultValue)
	}
	return []cli.Flag{
		cli.StringFlag{
			Name:  flags.DeploymentMaxPercentFlag,
			Usage: maxPercentUsageString,
		},
		cli.StringFlag{
			Name:  flags.DeploymentMinHealthyPercentFlag,
			Usage: minHealthyPercentUsageString,
		},
	}
}

func loadBalancerFlags() []cli.Flag {
	targetGroupArnUsageString := "[Optional] Specifies the full Amazon Resource Name (ARN) of a previously configured Elastic Load Balancing target group to associate with your service."
	containerNameUsageString := "[Optional] Specifies the container name (as it appears in a container definition). This parameter is required if a load balancer or target group is specified."
	containerPortUsageString := "[Optional] Specifies the port on the container to associate with the load balancer. This port must correspond to a containerPort in the service's task definition. This parameter is required if a load balancer or target group is specified."
	loadBalancerNameUsageString := "[Optional] Specifies the name of a previously configured Elastic Load Balancing load balancer to associate with your service."
	roleUsageString := "[Optional] Specifies the name or full Amazon Resource Name (ARN) of the IAM role that allows Amazon ECS to make calls to your load balancer or target group on your behalf. This parameter is required if you are using a load balancer or target group with your service. If you specify the role parameter, you must also specify a load balancer name or target group ARN, along with a container name and container port."

	return []cli.Flag{
		cli.StringFlag{
			Name:  flags.TargetGroupArnFlag,
			Usage: targetGroupArnUsageString,
		},
		cli.StringFlag{
			Name:  flags.ContainerNameFlag,
			Usage: containerNameUsageString,
		},
		cli.StringFlag{
			Name:  flags.ContainerPortFlag,
			Usage: containerPortUsageString,
		},
		cli.StringFlag{
			Name:  flags.LoadBalancerNameFlag,
			Usage: loadBalancerNameUsageString,
		},
		cli.StringFlag{
			Name:  flags.RoleFlag,
			Usage: roleUsageString,
		},
	}
}

// ComposeServiceTimeoutFlag allows user to specify a custom timeout
func ComposeServiceTimeoutFlag() cli.Flag {
	return cli.Float64Flag{
		Name:  flags.ComposeServiceTimeOutFlag,
		Value: service.DefaultUpdateServiceTimeout,
		Usage: fmt.Sprintf(
			"Specifies the timeout value in minutes (decimals supported) to wait for the running task count to change. If the running task count has not changed for the specified period of time, then the CLI times out and returns an error. Setting the timeout to 0 will cause the command to return without checking for success.",
		),
	}
}
