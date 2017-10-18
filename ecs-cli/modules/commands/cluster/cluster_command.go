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

package clusterCommand

import (
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/cluster"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/urfave/cli"
)

func UpCommand() cli.Command {
	return cli.Command{
		Name:         "up",
		Usage:        "Creates the ECS cluster (if it does not already exist) and the AWS resources required to set up the cluster.",
		Before:       ecscli.BeforeApp,
		Action:       cluster.ClusterUp,
		Flags:        append(clusterUpFlags(), command.OptionalConfigFlags()...),
		OnUsageError: command.UsageErrorFactory("up"),
	}
}

func DownCommand() cli.Command {
	return cli.Command{
		Name:         "down",
		Usage:        "Deletes the CloudFormation stack that was created by ecs-cli up and the associated resources. The --force option is required.",
		Action:       cluster.ClusterDown,
		Flags:        append(clusterDownFlags(), command.OptionalConfigFlags()...),
		OnUsageError: command.UsageErrorFactory("down"),
	}
}

func ScaleCommand() cli.Command {
	return cli.Command{
		Name:         "scale",
		Usage:        "Modifies the number of container instances in your cluster. This command changes the desired and maximum instance count in the Auto Scaling group created by the ecs-cli up command. You can use this command to scale up (increase the number of instances) or scale down (decrease the number of instances) your cluster.",
		Action:       cluster.ClusterScale,
		Flags:        append(clusterScaleFlags(), command.OptionalConfigFlags()...),
		OnUsageError: command.UsageErrorFactory("scale"),
	}
}

func PsCommand() cli.Command {
	return cli.Command{
		Name:         "ps",
		Usage:        "Lists all of the running containers in your ECS cluster",
		Action:       cluster.ClusterPS,
		Flags:        command.OptionalConfigFlags(),
		OnUsageError: command.UsageErrorFactory("ps"),
	}
}

func clusterUpFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name: command.VerboseFlag + ",debug",
		},
		cli.StringFlag{
			Name:  command.KeypairNameFlag,
			Usage: "Specifies the name of an existing Amazon EC2 key pair to enable SSH access to the EC2 instances in your cluster.",
		},
		cli.BoolFlag{
			Name:  command.CapabilityIAMFlag,
			Usage: "Acknowledges that this command may create IAM resources. Required if --instance-role is not specified.",
		},
		cli.StringFlag{
			Name:  command.AsgMaxSizeFlag,
			Usage: "[Optional] Specifies the number of instances to launch and register to the cluster. Defaults to 1.",
		},
		cli.StringFlag{
			Name:  command.VpcAzFlag,
			Usage: "[Optional] Specifies a comma-separated list of 2 VPC Availability Zones in which to create subnets (these zones must have the available status). This option is recommended if you do not specify a VPC ID with the --vpc option. WARNING: Leaving this option blank can result in failure to launch container instances if an unavailable zone is chosen at random.",
		},
		cli.StringFlag{
			Name:  command.SecurityGroupFlag,
			Usage: "[Optional] Specifies a comma-separated list of existing security groups to associate with your container instances. If you do not specify a security group here, then a new one is created.",
		},
		cli.StringFlag{
			Name:  command.SourceCidrFlag,
			Usage: "[Optional] Specifies a CIDR/IP range for the security group to use for container instances in your cluster. This parameter is ignored if an existing security group is specified with the --security-group option. Defaults to 0.0.0.0/0.",
		},
		cli.StringFlag{
			Name:  command.EcsPortFlag,
			Usage: "[Optional] Specifies a port to open on the security group to use for container instances in your cluster. This parameter is ignored if an existing security group is specified with the --security-group option. Defaults to port 80.",
		},
		cli.StringFlag{
			Name:  command.SubnetIdsFlag,
			Usage: "[Optional] Specifies a comma-separated list of existing VPC Subnet IDs in which to launch your container instances. This option is required if you specify a VPC with the --vpc option.",
		},
		cli.StringFlag{
			Name:  command.VpcIdFlag,
			Usage: "[Optional] Specifies the ID of an existing VPC in which to launch your container instances. If you specify a VPC ID, you must specify a list of existing subnets in that VPC with the --subnets option. If you do not specify a VPC ID, a new VPC is created with two subnets.",
		},
		cli.StringFlag{
			Name:  command.InstanceTypeFlag,
			Usage: "[Optional] Specifies the EC2 instance type for your container instances. Defaults to t2.micro.",
		},
		cli.StringFlag{
			Name:  command.ImageIdFlag,
			Usage: "[Optional] Specify the AMI ID for your container instances. Defaults to amazon-ecs-optimized AMI.",
		},
		cli.BoolFlag{
			Name:  command.NoAutoAssignPublicIPAddressFlag,
			Usage: "[Optional] Do not assign public IP addresses to new instances in this VPC. Unless this option is specified, new instances in this VPC receive an automatically assigned public IP address.",
		},
		cli.BoolFlag{
			Name:  command.ForceFlag + ", f",
			Usage: "[Optional] Forces the recreation of any existing resources that match your current configuration. This option is useful for cleaning up stale resources from previous failed attempts.",
		},
		cli.StringFlag{
			Name:  command.InstanceRoleFlag,
			Usage: "[Optional] Specifies a custom IAM Role for instances in your cluster. Required if --capability-iam is not specified.",
		},
	}
}

func clusterDownFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  command.ForceFlag + ", f",
			Usage: "Acknowledges that this command permanently deletes resources.",
		},
	}
}

func clusterScaleFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  command.CapabilityIAMFlag,
			Usage: "Acknowledges that this command may create IAM resources.",
		},
		cli.StringFlag{
			Name:  command.AsgMaxSizeFlag,
			Usage: "Specifies the number of instances to maintain in your cluster.",
		},
	}
}
