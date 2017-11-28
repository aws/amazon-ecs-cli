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

package flags

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"
)

// Flag names used by the cli.
const (
	// Configure
	AccessKeyFlag           = "access-key"
	SecretKeyFlag           = "secret-key"
	RegionFlag              = "region"
	AwsRegionEnvVar         = "AWS_REGION"
	AwsDefaultRegionEnvVar  = "AWS_DEFAULT_REGION"
	AwsDefaultProfileEnvVar = "AWS_DEFAULT_PROFILE"
	ProfileFlag             = "profile"
	ClusterFlag             = "cluster"
	ClusterEnvVar           = "ECS_CLUSTER"
	VerboseFlag             = "verbose"
	ClusterConfigFlag       = "cluster-config"
	ECSProfileFlag          = "ecs-profile"
	ProfileNameFlag         = "profile-name"
	ConfigNameFlag          = "config-name"
	AWSProfileFlag          = "aws-profile"
	ECSProfileEnvVar        = "ECS_PROFILE"
	AWSProfileEnvVar        = "AWS_PROFILE"
	AWSAccessKeyEnvVar      = "AWS_ACCESS_KEY_ID"
	AWSSecretKeyEnvVar      = "AWS_SECRET_ACCESS_KEY"

	// logs
	TaskIDFlag         = "task-id"
	TaskDefinitionFlag = "task-def"
	FollowLogsFlag     = "follow"
	FilterPatternFlag  = "filter-pattern"
	SinceFlag          = "since"
	StartTimeFlag      = "start-time"
	EndTimeFlag        = "end-time"
	TimeStampsFlag     = "timestamps"

	ComposeProjectNamePrefixFlag         = "compose-project-name-prefix"
	ComposeProjectNamePrefixDefaultValue = "ecscompose-"
	ComposeServiceNamePrefixFlag         = "compose-service-name-prefix"
	ComposeServiceNamePrefixDefaultValue = ComposeProjectNamePrefixDefaultValue + "service-"
	CFNStackNameFlag                     = "cfn-stack-name"
	CFNStackNamePrefixDefaultValue       = "amazon-ecs-cli-setup-"

	LaunchTypeFlag        = "launch-type"
	DefaultLaunchTypeFlag = "default-launch-type"

	// Cluster
	AsgMaxSizeFlag                  = "size"
	VpcAzFlag                       = "azs"
	SecurityGroupFlag               = "security-group"
	SourceCidrFlag                  = "cidr"
	EcsPortFlag                     = "port"
	SubnetIdsFlag                   = "subnets"
	VpcIdFlag                       = "vpc"
	InstanceTypeFlag                = "instance-type"
	InstanceRoleFlag                = "instance-role"
	ImageIdFlag                     = "image-id"
	KeypairNameFlag                 = "keypair"
	CapabilityIAMFlag               = "capability-iam"
	NoAutoAssignPublicIPAddressFlag = "no-associate-public-ip-address"
	ForceFlag                       = "force"

	// Image
	RegistryIdFlag = "registry-id"
	TaggedFlag     = "tagged"
	UntaggedFlag   = "untagged"

	// Compose
	ProjectNameFlag       = "project-name"
	ComposeFileNameFlag   = "file"
	TaskRoleArnFlag       = "task-role-arn"
	ECSParamsFileNameFlag = "ecs-params"

	// Compose Service
	CreateServiceCommandName                = "create"
	DeploymentMaxPercentDefaultValue        = 200
	DeploymentMaxPercentFlag                = "deployment-max-percent"
	DeploymentMinHealthyPercentDefaultValue = 100
	DeploymentMinHealthyPercentFlag         = "deployment-min-healthy-percent"
	TargetGroupArnFlag                      = "target-group-arn"
	ContainerNameFlag                       = "container-name"
	ContainerPortFlag                       = "container-port"
	LoadBalancerNameFlag                    = "load-balancer-name"
	RoleFlag                                = "role"
	ComposeServiceTimeOutFlag               = "timeout"
)

// OptionalRegionAndProfileFlags provides these flags:
// OptionalRegionFlag inline overrides region
// OptionalClusterConfigFlag specifies the cluster profile to read from config
// OptionalProfileConfigFlag specifies the credentials profile to read from the config
// OptionalAWSProfileFlag specifies the AWS Profile to use for credential information
func OptionalRegionAndProfileFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name: RegionFlag + ", r",
			Usage: fmt.Sprintf(
				"[Optional] Specifies the AWS region to use. Defaults to the region configured using the configure command",
			),
		},
		cli.StringFlag{
			Name: ClusterConfigFlag,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the name of the ECS cluster configuration to use. Defaults to the default cluster configuration.",
			),
		},
		cli.StringFlag{
			Name:   ECSProfileFlag,
			EnvVar: ECSProfileEnvVar,
			Usage: fmt.Sprintf(
				"[Optional] Specifies the name of the ECS profle configuration to use. Defaults to the default profile configuration.",
			),
		},
		cli.StringFlag{
			Name:   AWSProfileFlag,
			EnvVar: AWSProfileEnvVar,
			Usage: fmt.Sprintf(
				"[Optional]  Use the AWS credentials from an existing named profile in ~/.aws/credentials.",
			),
		},
	}
}

// OptionalClusterFlag inline overrides cluster
func OptionalClusterFlag() cli.Flag {
	return cli.StringFlag{
		Name: ClusterFlag + ", c",
		Usage: fmt.Sprintf(
			"[Optional] Specifies the ECS cluster name to use. Defaults to the cluster configured using the configure command",
		),
	}
}

// OptionalConfigFlags returns the concatenation of OptionalRegionAndProfileFlags and OptionalClusterFlag
func OptionalConfigFlags() []cli.Flag {
	return append(OptionalRegionAndProfileFlags(), OptionalClusterFlag())
}

// OptionalLaunchTypeFlag allows users to specify the launch type for their task/service/cluster
func OptionalLaunchTypeFlag() cli.Flag {
	return cli.StringFlag{
		Name: LaunchTypeFlag,
		Usage: fmt.Sprintf(
			"[Optional] Specifies the launch type. Options: EC2 or FARGATE. Overrides the default launch type stored in your cluster configuration. Defaults to EC2 if a cluster configuration is not used.",
		),
	}
}

// UsageErrorFactory Returns a usage error function for the specified command
func UsageErrorFactory(command string) func(*cli.Context, error, bool) error {
	return func(c *cli.Context, err error, isSubcommand bool) error {
		if err != nil {
			logrus.Error(err)
		}
		err = cli.ShowCommandHelp(c, command)
		if err != nil {
			logrus.Debug(err)
		}
		os.Exit(1)
		return err
	}
}
