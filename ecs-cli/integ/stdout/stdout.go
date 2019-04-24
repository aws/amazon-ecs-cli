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
	"testing"

	"github.com/stretchr/testify/require"
)

// Stdout is the standard output content from running a test.
type Stdout []byte

// TestHasAllSubstrings returns true if stdout contains each snippet in wantedSnippets, false otherwise.
func (b Stdout) TestHasAllSubstrings(t *testing.T, wantedSubstrings []string) {
	s := string(b)
	for _, substring := range wantedSubstrings {
		require.Contains(t, s, substring)
	}
}
