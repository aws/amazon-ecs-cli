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

// Package integ contains utility functions to run ECS CLI commands and check their outputs.
package integ

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

const (
	binPath = "../../../bin/local/ecs-cli" // TODO: use abs path or env var
)

// GetCommand returns a Cmd struct with the right binary path & arguments
func GetCommand(args []string) *exec.Cmd {
	cmdPath := binPath

	if runtime.GOOS == "windows" {
		cmdPath = cmdPath + ".exe"
	}

	cmd := exec.Command(cmdPath, args...)
	return cmd
}

// GetRowValues takes a row of stdout and returns a slice of strings split by arbirary whitespace
func GetRowValues(row string) []string {
	spaces := regexp.MustCompile(`\s+`)
	return strings.Split(spaces.ReplaceAllString(row, " "), " ")
}

// SuggestedResourceName returns a resource name matching the template "{CODEBUILD_BUILD_ID}-{identifier}".
// The CODEBUILD_BUILD_ID in the name can be used to safely delete a resource if it belongs to an old test build.
// The identifier can be used to give a human-friendly resource name.
func SuggestedResourceName(identifiers ...string) string {
	return fmt.Sprintf("%s-%s", getBuildId(), strings.Join(identifiers, "-"))
}

// getBuildId returns the CodeBuild ID compatible with CloudFormation.
func getBuildId() string {
	return strings.Replace(os.Getenv("CODEBUILD_BUILD_ID"), ":", "-", -1) // replace all occurrences
}

// RetryUntilTimeout retries function f every sleepInS seconds until the timeoutInS expires.
func RetryUntilTimeout(t *testing.T, f func(t *testing.T) bool, timeoutInS time.Duration, sleepInS time.Duration) bool {
	numRetries := int64(timeoutInS) / int64(sleepInS)
	var i int64
	for ; i < numRetries; i++ {
		if ok := f(t); ok {
			return true
		}
		t.Logf("Sleeping for %v", sleepInS)
		time.Sleep(sleepInS * time.Second)
	}
	return false
}
