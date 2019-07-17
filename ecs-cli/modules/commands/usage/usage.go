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

// Package usage contains the usage documentation for all ECS-CLI commands and subcommands
package usage

import (
	// "fmt"
)

// String displayed as usage for command. Constant should match the command
// prefix for corresponding command or combined command + subcommand, e.g.
// const Local contains the docstring for the `local` command; `LocalUp`
// contains the docstring for `local up`.
const (
	Attributechecker = "Checks if a given list of container instances can run a given task definition by checking their attributes. Outputs attributes that are required by the task definition but not present on the container instances."
)
