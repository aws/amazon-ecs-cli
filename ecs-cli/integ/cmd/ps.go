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

package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/stretchr/testify/require"
)

// TestPsRunning runs `ecs-cli ps` for a project and validates that the
// number of running containers is equal to the wanted number of containers.
func TestPsRunning(t *testing.T, p *Project, wantedNumOfContainers int) {
	f := func(t *testing.T) bool {
		return testClusterHasAllRunningContainers(t, p, wantedNumOfContainers)
	}
	timeoutInS := 120 * time.Second // 2 mins
	sleepInS := 15 * time.Second
	require.True(t, integ.RetryUntilTimeout(t, f, timeoutInS, sleepInS), "Failed to get RUNNING containers")
	t.Logf("Project %s has %d running containers", p.Name, wantedNumOfContainers)
}

func testClusterHasAllRunningContainers(t *testing.T, p *Project, wantedNumOfContainers int) bool {
	// Given
	args := []string{
		"ps",
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if err != nil {
		t.Logf("Failed to list containers for cluster %v", args)
		return false
	}

	// Then
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		t.Logf("No running containers yet, out = %v", lines)
		return false
	}
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1] // Drop the last new line
	}
	containers := lines[1:] // Drop the headers
	numActiveContainers := 0
	for _, container := range containers {
		status := integ.GetRowValues(container)[1]
		if status == "RUNNING" {
			numActiveContainers++
		}
	}
	if wantedNumOfContainers != numActiveContainers {
		t.Logf("Wanted = %d, got = %d running containers", wantedNumOfContainers, numActiveContainers)
		t.Logf("Containers = %s", strings.Join(containers, "\n"))
		return false
	}
	return true
}
