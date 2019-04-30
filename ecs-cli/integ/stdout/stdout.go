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

package stdout

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Stdout is the standard output content from running a test.
type Stdout []byte

// Lines returns a slice of lines in stdout without the coverage statistics from the ecs-cli.test binary.
func (b Stdout) Lines() []string {
	// Each command against the ecs-cli.test binary produces an output like:
	// > PASS
	// > coverage: 2.3% of statements in ./ecs-cli/modules/...
	// >
	// We need to remove these lines from stdout.
	coverageLineCount := 3

	lines := strings.Split(string(b), "\n")
	return lines[:len(lines)-coverageLineCount]
}

// TestHasAllSubstrings returns true if stdout contains each snippet in wantedSnippets, false otherwise.
func (b Stdout) TestHasAllSubstrings(t *testing.T, wantedSubstrings []string) {
	s := strings.Join(b.Lines(), "\n")
	for _, substring := range wantedSubstrings {
		require.Contains(t, s, substring)
	}
}
