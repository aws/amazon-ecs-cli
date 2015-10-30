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

package command

import (
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/codegangsta/cli"
)

const (
	asgMaxSizeFlag    = "size"
	vpcAzFlag         = "azs"
	securityGroupFlag = "security-group"
	sourceCidrFlag    = "cidr"
	ecsPortFlag       = "port"
	subnetIdsFlag     = "subnets"
	vpcIdFlag         = "vpc"
	instanceTypeFlag  = "instance-type"
	imageIdFlag       = "image-id"
	keypairNameFlag   = "keypair"
	capabilityIAMFlag = "capability-iam"
	forceFlag         = "force"
)

func UpCommand() cli.Command {
	return cli.Command{
		Name:   "up",
		Usage:  "Create the ECS Cluster (if it does not already exist) and the AWS resources required to set up the cluster.",
		Before: ecscli.BeforeApp,
		Action: ClusterUp,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name: ecscli.VerboseFlag + ",debug",
			},
			cli.StringFlag{
				Name:  keypairNameFlag,
				Usage: "Specify the name of an existing Amazon EC2 key pair to enable SSH access to the EC2 instances in your cluster.",
			},
			cli.BoolFlag{
				Name:  capabilityIAMFlag,
				Usage: "Acknowledge that this command may create IAM resources.",
			},
			cli.StringFlag{
				Name:  asgMaxSizeFlag,
				Usage: "[Optional] Specify the number of instances to register to the cluster. The default is 1.",
			},
			cli.StringFlag{
				Name:  vpcAzFlag,
				Usage: "[Optional] Specify a comma-separated list of 2 VPC availability zones in which to create subnets (these AZs must be in the 'available' status). This option is recommended if you do not specify a VPC ID with the --vpc option. WARNING: Leaving this option blank can result in failure to launch container instances if an unavailable AZ is chosen at random.",
			},
			cli.StringFlag{
				Name:  securityGroupFlag,
				Usage: "[Optional] Specify an existing security group to associate it with container instances. Defaults to creating a new one.",
			},
			cli.StringFlag{
				Name:  sourceCidrFlag,
				Usage: "[Optional] Specify a CIDR/IP range for the security group to use for container instances in your cluster. Defaults to 0.0.0.0/0 if --security-group is not specified",
			},
			cli.StringFlag{
				Name:  ecsPortFlag,
				Usage: "[Optional] Specify a port to open on a new security group that is created for your container instances if an existing security group is not specified with the --security-group option. Defaults to port 80.",
			},
			cli.StringFlag{
				Name:  subnetIdsFlag,
				Usage: "[Optional] Specify a comma-separated list of existing VPC Subnet IDs in which to launch your container instances. This option is required if you specify a VPC with the --vpc option.",
			},
			cli.StringFlag{
				Name:  vpcIdFlag,
				Usage: "[Optional] Specify the ID of an existing VPC in which to launch your container instances. If you specify a VPC ID, you must specify a list of existing subnets in that VPC with the --subnets option. If you do not specify a VPC ID, a new VPC is created with two subnets.",
			},
			cli.StringFlag{
				Name:  instanceTypeFlag,
				Usage: "[Optional] Specify the EC2 instance type for your container instances.",
			},
			cli.StringFlag{
				Name:  imageIdFlag,
				Usage: "[Optional] Specify the ID of the AMI for your container Instances.",
			},
		},
	}
}

func DownCommand() cli.Command {
	return cli.Command{
		Name:   "down",
		Usage:  "Delete the ECS Cluster and associated resources in the CloudFormation stack.",
		Action: ClusterDown,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  forceFlag + ", f",
				Usage: "Overrides cofirmation prompt before deleting resources",
			},
		},
	}
}

func ScaleCommand() cli.Command {
	return cli.Command{
		Name:   "scale",
		Usage:  "Modify the number of container instances in your cluster.",
		Action: ClusterScale,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  capabilityIAMFlag,
				Usage: "Acknowledge that this command may create IAM resources.",
			},
			cli.StringFlag{
				Name:  asgMaxSizeFlag,
				Usage: "Specify the number of instances to maintain in your cluster.",
			},
		},
	}
}

func PsCommand() cli.Command {
	return cli.Command{
		Name:   "ps",
		Usage:  "List all of the running containers in your ECS Cluster.",
		Action: ClusterPS,
	}
}
