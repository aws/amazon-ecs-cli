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
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"
	"github.com/stretchr/testify/require"
)

const (
	binPath = "../../../bin/local/ecs-cli.test" // TODO: use abs path or env var
)

// GetCommand returns a Cmd struct with the right binary path & arguments
func GetCommand(args []string) *exec.Cmd {
	cmdPath := binPath

	if runtime.GOOS == "windows" {
		cmdPath = cmdPath + ".exe"
	}

	fname, err := createTempCoverageFile()
	if err != nil {
		return exec.Command(cmdPath, args...)
	}
	args = append([]string{fmt.Sprintf("-test.coverprofile=%s", fname)}, args...)
	return exec.Command(cmdPath, args...)
}

// RunCmd runs a command with args and returns its Stdout.
func RunCmd(t *testing.T, args []string) stdout.Stdout {
	cmd := GetCommand(args)

	out, err := cmd.Output()
	require.NoErrorf(t, err, "Failed running command", fmt.Sprintf("args=%v, stdout=%s, err=%v", args, string(out), err))

	return stdout.Stdout(out)
}

// createTempCoverageFile creates a coverage file for a CLI command under $TMPDIR.
func createTempCoverageFile() (string, error) {
	tmpfile, err := ioutil.TempFile("", "coverage-*.out")
	if err != nil {
		return "", err
	}

	err = tmpfile.Close()
	if err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
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
		t.Logf("Current timestamp=%v, sleeping for %v", time.Now(), sleepInS)
		time.Sleep(sleepInS)
	}
	return false
}
