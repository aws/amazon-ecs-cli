// +build integ

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

package integ

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/container"
	"github.com/stretchr/testify/assert"
)

const (
	integClusterName = "ecs-cli-integ"
)

// Test assumes binary has been built, named cluster contains 3 containers
func TestCmd_PS(t *testing.T) {
	// set up command
	ps_args := []string{"ps", "--cluster", integClusterName}
	cmd := integ.GetCommand(ps_args)

	var stdoutWriter bytes.Buffer
	writer := io.MultiWriter(&stdoutWriter)

	cmd.Stdout = writer

	// execute command
	err := cmd.Run()
	assert.NoError(t, err, "Unexpected error starting 'ps'")

	// assert on result
	actualStdout := stdoutWriter.String()
	assert.NotEmpty(t, actualStdout)

	stdoutLines := strings.Split(actualStdout, "\n")
	length := len(stdoutLines)

	// trim off empty last row is needed
	if stdoutLines[length-1] == "" {
		stdoutLines = stdoutLines[:length-1]
	}

	headers := integ.GetRowValues(stdoutLines[0])
	assert.Equal(t, container.ContainerInfoColumns, headers)

	runningContainers := stdoutLines[1:]
	assert.Equal(t, 3, len(runningContainers))
}
