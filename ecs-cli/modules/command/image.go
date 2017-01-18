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

package command

import (
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/urfave/cli"
)

const (
	PUSH_IMAGE_FORMAT = "REPOSITORY[:TAG]"
	PULL_IMAGE_FORMAT = "REPOSITORY_NAME[:TAG|@DIGEST]"
)

// PushCommand push ECR image
func PushCommand() cli.Command {
	return cli.Command{
		Name:      "push",
		Usage:     "Push an image to an Amazon ECR repository.",
		ArgsUsage: "[" + PUSH_IMAGE_FORMAT + "]",
		Before:    ecscli.BeforeApp,
		Action:    ImagePush,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  ecscli.RegistryIdFlag,
				Usage: "[Optional] Specifies the Amazon ECR registry ID to push the image to. By default, images are pushed to the current AWS account.",
			},
			cli.StringFlag{
				Name:  ecscli.FromFlag,
				Usage: "[Optional] Specifies the image to push.",
			},
			cli.StringFlag{
				Name:  ecscli.ToFlag,
				Usage: "[Optional] Specifies the ECR repository and tag to push your image to.",
			},
		},
	}
}

// PullCommand pull ECR image
func PullCommand() cli.Command {
	return cli.Command{
		Name:      "pull",
		Usage:     "Pull an image from an Amazon ECR repository.",
		ArgsUsage: PULL_IMAGE_FORMAT,
		Before:    ecscli.BeforeApp,
		Action:    ImagePull,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  ecscli.RegistryIdFlag,
				Usage: "[Optional] Specifies the the Amazon ECR registry ID to pull the image from. By default, images are pulled from the current AWS account.",
			},
		},
	}
}
