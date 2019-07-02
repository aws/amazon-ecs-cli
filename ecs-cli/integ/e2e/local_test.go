// +build integ

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package e2e

import (
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	project "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/localproject"
	"github.com/stretchr/testify/require"
)

func TestECSLocal(t *testing.T) {
	t.Parallel()

	type commandTest struct {
		args    []string
		execute func(t *testing.T, args []string)
	}

	tests := map[string]struct {
		sequence []commandTest
	}{
		"clean state": {
			sequence: []commandTest{
				{
					args: []string{"local", "ps"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.Error(t, err, "expected args=%v to fail", args)
						stdout.TestHasAllSubstrings(t, []string{
							fmt.Sprintf("%s does not exist", project.LocalOutDefaultFileName),
						})
					},
				},
				{
					args: []string{"local", "ps", "--all"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"CONTAINER ID",
							"IMAGE",
							"STATUS",
							"PORTS",
							"NAMES",
							"TASKDEFINITION",
						})
					},
				},
				{
					args: []string{"local", "ps", "--all", "--json"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.NoError(t, err)
						stdout.TestHasAllSubstrings(t, []string{
							"[]",
						})
					},
				},
				{
					args: []string{"local", "down"},
					execute: func(t *testing.T, args []string) {
						stdout, err := integ.RunCmd(t, args)
						require.Error(t, err, "expected args=%v to fail", args)
						stdout.TestHasAllSubstrings(t, []string{
							fmt.Sprintf("%s does not exist", project.LocalOutDefaultFileName),
						})
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for _, cmd := range tc.sequence {
				cmd.execute(t, cmd.args)
			}
		})
	}
}
