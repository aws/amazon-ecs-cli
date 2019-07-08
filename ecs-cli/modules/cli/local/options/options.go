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

// Package options implements utility functions around ECS local flags.
package options

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/localproject"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type flagPair struct {
	first  string
	second string
}

// ValidateFlagPairs returns an error if two flags can not be used together.
func ValidateFlagPairs(c *cli.Context) error {
	invalid := []flagPair{
		{
			flags.TaskDefinitionFile,
			flags.TaskDefinitionRemote,
		},
		{
			flags.TaskDefinitionFile,
			flags.TaskDefinitionCompose,
		},
		{
			flags.TaskDefinitionRemote,
			flags.TaskDefinitionCompose,
		},
		{
			flags.Output,
			flags.TaskDefinitionCompose,
		},
	}

	for _, pair := range invalid {
		if c.String(pair.first) != "" && c.String(pair.second) != "" {
			return fmt.Errorf("%s and %s can not be used together", pair.first, pair.second)
		}
	}
	return nil
}

// ContainerSearchInfo show the task definition filepath or arn/family message on local down.
func ContainerSearchInfo(c *cli.Context) {
	if c.IsSet(flags.TaskDefinitionFile) {
		logrus.Infof("Searching for containers matching --%s=%s", flags.TaskDefinitionFile, c.String(flags.TaskDefinitionFile))
		return
	}
	if c.IsSet(flags.TaskDefinitionRemote) {
		logrus.Infof("Searching for containers matching --%s=%s", flags.TaskDefinitionRemote, c.String(flags.TaskDefinitionRemote))
		return
	}
	if c.Bool(flags.All) {
		logrus.Info("Searching for all running containers")
		return
	}
	logrus.Infof("Searching for containers matching --%s=%s", flags.TaskDefinitionFile, localproject.LocalInFileName)
	return
}
