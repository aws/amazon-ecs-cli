// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// package serviceCommand defines the subcommands for compose service workflows
package serviceCommand

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/service"
	composeFactory "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/factory"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/usage"
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
		Usage: usage.Service,
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
		Usage:        usage.ServiceCreate,
		Action:       compose.WithProject(factory, compose.ProjectCreate, true),
		Flags:        flags.AppendFlags(deploymentConfigFlags(true), loadBalancerFlags(), flags.OptionalConfigFlags(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag(), serviceDiscoveryFlags(), flags.OptionalSchedulingStrategyFlag(), taggingFlags()),
		OnUsageError: flags.UsageErrorFactory("create"),
	}
}

func startServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "start",
		Usage:        usage.ServiceStart,
		Action:       compose.WithProject(factory, compose.ProjectStart, true),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag(), flags.OptionalCreateLogsFlag(), ForceNewDeploymentFlag(), EnableExecuteCommandFlag()),
		OnUsageError: flags.UsageErrorFactory("start"),
	}
}

func upServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "up",
		Usage:        usage.ServiceUp,
		Action:       compose.WithProject(factory, compose.ProjectUp, true),
		Flags:        flags.AppendFlags(deploymentConfigFlags(true), loadBalancerFlags(), flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag(), flags.OptionalLaunchTypeFlag(), flags.OptionalCreateLogsFlag(), ForceNewDeploymentFlag(), EnableExecuteCommandFlag(), serviceDiscoveryFlags(), updateServiceDiscoveryFlags(), flags.OptionalSchedulingStrategyFlag(), taggingFlags()),
		OnUsageError: flags.UsageErrorFactory("up"),
	}
}

func psServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "ps",
		Aliases:      []string{"list"},
		Usage:        usage.ServicePs,
		Action:       compose.WithProject(factory, compose.ProjectPs, true),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), flags.OptionalDesiredStatusFlag()),
		OnUsageError: flags.UsageErrorFactory("ps"),
	}
}

func scaleServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "scale",
		Usage:        usage.ServiceScale,
		Action:       compose.WithProject(factory, compose.ProjectScale, true),
		Flags:        flags.AppendFlags(deploymentConfigFlags(false), flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag()),
		OnUsageError: flags.UsageErrorFactory("scale"),
	}
}

func stopServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "stop",
		Usage:        usage.ServiceStop,
		Action:       compose.WithProject(factory, compose.ProjectStop, true),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag()),
		OnUsageError: flags.UsageErrorFactory("stop"),
	}
}

func rmServiceCommand(factory composeFactory.ProjectFactory) cli.Command {
	return cli.Command{
		Name:         "rm",
		Aliases:      []string{"delete", "down"},
		Usage:        usage.ServiceRm,
		Action:       compose.WithProject(factory, compose.ProjectDown, true),
		Flags:        flags.AppendFlags(flags.OptionalConfigFlags(), ComposeServiceTimeoutFlag(), deleteServiceDiscoveryFlags()),
		OnUsageError: flags.UsageErrorFactory("rm"),
	}
}

func serviceDiscoveryFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.EnableServiceDiscoveryFlag,
			Usage: "[Service Discovery] Enable Service Discovery for your ECS Service.",
		},
		cli.StringFlag{
			Name:  flags.VpcIdFlag,
			Usage: "[Service Discovery] The VPC that will be attached to the private DNS namespace.",
		},
		cli.StringFlag{
			Name:  flags.PrivateDNSNamespaceNameFlag,
			Usage: "[Service Discovery] The name of the private DNS namespace to use with Service Discovery. The CLI creates the namespace if it doesn't already exist. For example, if the name is 'corp' than a service 'foo' will be reachable via DNS at 'foo.corp'.",
		},
		cli.StringFlag{
			Name:  flags.PrivateDNSNamespaceIDFlag,
			Usage: "[Service Discovery] The ID of an existing private DNS namespace to use with Service Discovery.",
		},
		cli.StringFlag{
			Name:  flags.PublicDNSNamespaceIDFlag,
			Usage: "[Service Discovery] The ID of an existing public DNS namespace to use with Service Discovery.",
		},
		cli.StringFlag{
			Name:  flags.PublicDNSNamespaceNameFlag,
			Usage: "[Service Discovery] The name of an existing public DNS namespace to use with Service Discovery. For example, if the name is 'corp' than a service 'foo' will be reachable via DNS at 'foo.corp'.",
		},
		cli.StringFlag{
			Name:  flags.ServiceDiscoveryContainerNameFlag,
			Usage: "[Service Discovery] The name of the container (service name in compose) that will use Service Discovery.",
		},
		cli.StringFlag{
			Name:  flags.ServiceDiscoveryContainerPortFlag,
			Usage: "[Service Discovery] The port on the container used for Service Discovery.",
		},
		cli.StringFlag{
			Name:  flags.DNSTTLFlag,
			Usage: "[Service Discovery] The TTL of the DNS Records used with the Route53 Service Discovery Resource. Default value is 60 seconds.",
		},
		cli.StringFlag{
			Name:  flags.DNSTypeFlag,
			Usage: "[Service Discovery] The type of the DNS Records used with the Route53 Service Discovery Resource (A or SRV). Note that SRV records require container name and container port.",
		},
		cli.StringFlag{
			Name:  flags.HealthcheckCustomConfigFailureThresholdFlag,
			Usage: "[Service Discovery] The number of 30-second intervals that you want service discovery service to wait after receiving an UpdateInstanceCustomHealthStatus request before it changes the health status of a service instance. Default value is 1.",
		},
	}
}

func updateServiceDiscoveryFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.UpdateServiceDiscoveryFlag,
			Usage: "[Optional] [Service Discovery] Allows update of Service Discovery Service settings DNS TTL and Failure Threshold.",
		},
	}
}

func deleteServiceDiscoveryFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.DeletePrivateNamespaceFlag,
			Usage: "[Optional] [Service Discovery] Deletes the private namespace created by the ECS CLI",
		},
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
	targetGroupArnUsageString := fmt.Sprintf("[Deprecated] Specifies the full Amazon Resource Name (ARN) of a previously configured target group for an Application Load Balancer or Network Load Balancer to associate with your service. NOTE: For Classic Load Balancers, use the --%s flag.", flags.LoadBalancerNameFlag)
	containerNameUsageString := fmt.Sprintf("[Deprecated] Specifies the container name (as it appears in a container definition). This parameter is required if --%s or --%s is specified.", flags.LoadBalancerNameFlag, flags.TargetGroupArnFlag)
	containerPortUsageString := fmt.Sprintf("[Deprecated] Specifies the port on the container to associate with the load balancer. This port must correspond to a containerPort in the service's task definition. This parameter is required if --%s or --%s is specified.", flags.LoadBalancerNameFlag, flags.TargetGroupArnFlag)
	loadBalancerNameUsageString := fmt.Sprintf("[Deprecated] Specifies the name of a previously configured Classic Elastic Load Balancing load balancer to associate with your service. NOTE: For Application Load Balancers or Network Load Balancers, use the --%s flag.", flags.TargetGroupArnFlag)
	targetGroupsUsageString := fmt.Sprintf("[Optional] Specifies multiple target groups to register with a service. Can't be used with --%s flag or --%s at the same time. To specify multiple target groups, add multiple seperate --%s flags Example: ecs-cli compose service create --target-groups targetGroupArn=arn,containerName=nginx,containerPort=80 --target-groups targetGroupArn=arn,containerName=database,containerPort=3306", flags.LoadBalancerNameFlag, flags.TargetGroupArnFlag, flags.TargetGroupsFlag)
	roleUsageString := fmt.Sprintf("[Optional] Specifies the name or full Amazon Resource Name (ARN) of the IAM role that allows Amazon ECS to make calls to your load balancer or target group on your behalf. This parameter requires either --%s or --%s to be specified.", flags.LoadBalancerNameFlag, flags.TargetGroupArnFlag)
	healthCheckGracePeriodString := "[Optional] Specifies the period of time, in seconds, that the Amazon ECS service scheduler should ignore unhealthy Elastic Load Balancing target health checks after a task has first started."

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
		cli.StringFlag{
			Name:  flags.HealthCheckGracePeriodFlag,
			Usage: healthCheckGracePeriodString,
		},
		cli.StringSliceFlag{
			Name:  flags.TargetGroupsFlag,
			Usage: targetGroupsUsageString,
			Value: &cli.StringSlice{},
		},
	}
}

// ComposeServiceTimeoutFlag allows user to specify a custom timeout
func ComposeServiceTimeoutFlag() []cli.Flag {
	return []cli.Flag{
		cli.Float64Flag{
			Name:  flags.ComposeServiceTimeOutFlag,
			Value: service.DefaultUpdateServiceTimeout,
			Usage: fmt.Sprintf(
				"Specifies the timeout value in minutes (decimals supported) to wait for the running task count to change. If the running task count has not changed for the specified period of time, then the CLI times out and returns an error. Setting the timeout to 0 will cause the command to return without checking for success.",
			),
		},
	}
}

func EnableExecuteCommandFlag() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.EnableExecuteCommandFlag,
			Usage: "[Optional] Whether or not to enable the execute command functionality on the deployed service/tasks.",
		},
	}
}

func ForceNewDeploymentFlag() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.ForceDeploymentFlag,
			Usage: "[Optional] Whether or not to force a new deployment of the service.",
		},
	}
}

func taggingFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  flags.DisableECSManagedTagsFlag,
			Usage: "[Optional] Disable ECS Managed Tags (Cluster name and Service name tags will not be automatically added to tasks). Only affects new Services.",
		},
		cli.StringFlag{
			Name:  flags.ResourceTagsFlag,
			Usage: "[Optional] Specify resource tags for your ECS Service and Task Definition; tags are only added when resources are created. Tags will be propogated from your task definition to tasks created by the service. Specify tags in the format 'key1=value1,key2=value2,key3=value3'",
		},
	}
}
