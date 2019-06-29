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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

type flagPair struct {
	first  string
	second string
}

// ValidateCombinations returns an error if two flags can not be used together.
func ValidateCombinations(c *cli.Context) error {
	notTogether := []flagPair{
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

	for _, pair := range notTogether {
		if c.String(pair.first) != "" && c.String(pair.second) != "" {
			return fmt.Errorf("%s and %s can not be used together", pair.first, pair.second)
		}
	}
	return nil
}

// HasTaskDefInputFlag returns true if any --task-def-* flag was set, false otherwise.
func HasTaskDefInputFlag(c *cli.Context) bool {
	if c.String(flags.TaskDefinitionFile) != "" {
		return true
	}
	if c.String(flags.TaskDefinitionRemote) != "" {
		return true
	}
	if c.String(flags.TaskDefinitionCompose) != "" {
		return true
	}
	return false
}
