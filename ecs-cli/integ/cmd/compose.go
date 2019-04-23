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
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"

	"github.com/stretchr/testify/assert"
)

// A Project is the configuration needed to create an ECS Service.
type Project struct {
	Name       string
	ConfigName string
}

// NewProject creates a new Project for an ECS Service.
func NewProject(name string, configName string) *Project {
	return &Project{
		Name:       name,
		ConfigName: configName,
	}
}

// TestServiceUp runs `ecs-cli compose service up` for a project.
func TestServiceUp(t *testing.T, p *Project) {
	// Given
	args := []string{
		"compose",
		"--project-name",
		p.Name,
		"service",
		"up",
		"--create-log-groups",
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if err != nil {
		assert.FailNowf(t, "Failed to create service", "Error %v running %v", err, args)
	}

	// Then
	stdout.Stdout(out).TestHasAllSnippets(t, []string{
		"ECS Service has reached a stable state",
		"desiredCount=1",
		"serviceName=" + p.Name,
	})

	t.Logf("Created service with name %s", p.Name)
}

// TestServicePs runs `ecs-cli compose service ps` for a project.
func TestServicePs(t *testing.T, p *Project, wantedNumOfContainers int) {
	f := func(t *testing.T) bool {
		return testServiceHasAllRunningContainers(t, p, wantedNumOfContainers)
	}
	timeoutInS := 120 * time.Second // 2 mins
	sleepInS := 15 * time.Second
	if ok := integ.RetryUntilTimeout(t, f, timeoutInS, sleepInS); !ok {
		assert.Fail(t, "failed to get RUNNING containers")
	}
	t.Logf("Project %s has %d running containers", p.Name, wantedNumOfContainers)
}

// TestServiceScale runs `ecs-cli compose service scale` for a project.
func TestServiceScale(t *testing.T, p *Project, scale int) {
	// Given
	args := []string{
		"compose",
		"--project-name",
		p.Name,
		"service",
		"scale",
		strconv.Itoa(scale),
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if err != nil {
		assert.FailNowf(t, "Failed to scale service", "Error %v running %v", err, args)
	}

	// Then
	stdout.Stdout(out).TestHasAllSnippets(t, []string{
		"ECS Service has reached a stable state",
		fmt.Sprintf("desiredCount=%d", scale),
		fmt.Sprintf("runningCount=%d", scale),
		"serviceName=" + p.Name,
	})
	t.Logf("Scaled the service %s to %d tasks", p.Name, scale)
}

// TestServiceDown runs `ecs-cli compose service down` for a project.
func TestServiceDown(t *testing.T, p *Project) {
	// Given
	args := []string{
		"compose",
		"--project-name",
		p.Name,
		"service",
		"down",
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if err != nil {
		assert.FailNowf(t, "Failed to create service", "Error %v running %v", err, args)
	}

	// Then
	stdout.Stdout(out).TestHasAllSnippets(t, []string{
		"Deleted ECS service",
		"ECS Service has reached a stable state",
		"desiredCount=0",
		"runningCount=0",
		"serviceName=" + p.Name,
	})
	t.Logf("Deleted service %s", p.Name)
}

func testServiceHasAllRunningContainers(t *testing.T, p *Project, wantedNumOfContainers int) bool {
	// Given
	args := []string{
		"compose",
		"--project-name",
		p.Name,
		"service",
		"ps",
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if !assert.NoErrorf(t, err, "Failed to list containers for service %v", args) {
		return false
	}

	// Then
	lines := strings.Split(string(out), "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1] // Drop the last new line
	}
	containers := lines[1:] // Drop the headers
	if !assert.Equal(t, wantedNumOfContainers, len(containers), "Number of running containers mismatch") {
		return false
	}
	for _, container := range containers {
		status := integ.GetRowValues(container)[1]
		if !assert.Equal(t, "RUNNING", status, "Unexpected container status") {
			t.Logf("Container is not RUNNING: %s", container)
			return false
		}
	}
	return true
}
