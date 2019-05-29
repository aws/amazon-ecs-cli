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
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"
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
						// Given
						cmd := integ.GetCommand(args)

						// When
						out, err := cmd.Output()
						require.NoErrorf(t, err, "Failed local ps", fmt.Sprintf("args=%v, stdout=%s, err=%v", args, string(out), err))

						// Then
						stdout := stdout.Stdout(out)
						require.Equal(t, 1, len(stdout.Lines()), "Expected only the table header")
						stdout.TestHasAllSubstrings(t, []string{
							"CONTAINER ID",
							"IMAGE",
							"STATUS",
							"PORTS",
							"NAMES",
							"TASKDEFINITIONARN",
							"TASKFILEPATH",
						})
					},
				},
				{
					args: []string{"local", "ps", "--json"},
					execute: func(t *testing.T, args []string) {
						// Given
						cmd := integ.GetCommand(args)

						// When
						out, err := cmd.Output()
						require.NoErrorf(t, err, "Failed local ps", fmt.Sprintf("args=%v, stdout=%s, err=%v", args, string(out), err))

						// Then
						stdout := stdout.Stdout(out)
						stdout.TestHasAllSubstrings(t, []string{"[]"})
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
