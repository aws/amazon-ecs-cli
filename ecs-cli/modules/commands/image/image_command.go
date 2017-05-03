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

package imageCommand

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/image"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/urfave/cli"
)

// PushCommand push ECR image
func PushCommand() cli.Command {
	return cli.Command{
		Name:      "push",
		Usage:     "Push an image to an Amazon ECR repository.",
		ArgsUsage: image.PushImageFormat,
		Before:    app.BeforeApp,
		Action:    image.ImagePush,
		Flags:     imagePushFlags(),
	}
}

// PullCommand pull ECR image
func PullCommand() cli.Command {
	return cli.Command{
		Name:      "pull",
		Usage:     "Pull an image from an Amazon ECR repository.",
		ArgsUsage: image.PullImageFormat,
		Before:    app.BeforeApp,
		Action:    image.ImagePull,
		Flags:     imagePullFlags(),
	}
}

// ImagesCommand list images in ECR
func ImagesCommand() cli.Command {
	return cli.Command{
		Name:      "images",
		Usage:     "List images an Amazon ECR repository.",
		ArgsUsage: image.ListImageFormat,
		Before:    app.BeforeApp,
		Action:    image.ImageList,
		Flags:     imagesFlags(),
	}
}

func imagePushFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  command.RegistryIdFlag,
			Usage: "[Optional] Specifies the Amazon ECR registry ID to push the image to. By default, images are pushed to the current AWS account.",
		},
	}
}

func imagePullFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  command.RegistryIdFlag,
			Usage: "[Optional] Specifies the the Amazon ECR registry ID to pull the image from. By default, images are pulled from the current AWS account.",
		},
	}
}

func imagesFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  command.RegistryIdFlag,
			Usage: "[Optional] Specifies the the Amazon ECR registry ID to pull the image from. By default, images are pulled from the current AWS account.",
		},
		cli.BoolFlag{
			Name:  command.TaggedFlag,
			Usage: "[Optional] Filters the result to show only tagged images",
		},
		cli.BoolFlag{
			Name:  command.UntaggedFlag,
			Usage: "[Optional] Filters the result to show only untagged images",
		},
	}
}
