// +build testrunmain

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

package main

import (
	"testing"
)

// TestRunMain is a mandatory test to generate the binary ecs-cli.test that's used by our integration tests.
// It creates a wrapper test binary that's the same as the ecs-cli that also outputs the line coverage
// after each execution.
// See https://www.elastic.co/blog/code-coverage-for-your-golang-system-tests for further details.
func TestRunMain(t *testing.T) {
	main()
}
