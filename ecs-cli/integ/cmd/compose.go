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
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"
)

// A Project is the configuration needed to create an ECS Service.
type Project struct {
	Name              string
	ComposeFileName   string
	ECSParamsFileName string
	ConfigName        string
}

// NewProject creates a new Project for an ECS Service.
func NewProject(name string, configName string) *Project {
	return &Project{
		Name:       name,
		ConfigName: configName,
	}
}

// TestTaskUp runs `ecs-cli compose up` for a project.
func TestTaskUp(t *testing.T, p *Project) {
	// Given
	args := []string{
		"compose",
		"--file",
		p.ComposeFileName,
		"--ecs-params",
		p.ECSParamsFileName,
		"--project-name",
		p.Name,
		"up",
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	require.NoErrorf(t, err, "Failed to create task definition", "error %v, running %v, out: %s", err, args, string(out))

	// Then
	stdout.Stdout(out).TestHasAllSubstrings(t, []string{
		"Started container",
	})
	t.Logf("Created containers for %s", p.Name)
}

// TestServiceUp runs `ecs-cli compose service up` for a project.
func TestServiceUp(t *testing.T, p *Project) {
	// Given
	args := []string{
		"compose",
		"--file",
		p.ComposeFileName,
		"--ecs-params",
		p.ECSParamsFileName,
		"--project-name",
		p.Name,
		"service",
		"up",
		"--cluster-config",
		p.ConfigName,
		"--create-log-groups",
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	require.NoErrorf(t, err, "Failed to create service", "error %v, running %v, out: %s", err, args, string(out))

	// Then
	stdout.Stdout(out).TestHasAllSubstrings(t, []string{
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
	require.True(t, integ.RetryUntilTimeout(t, f, timeoutInS, sleepInS), "Failed to get RUNNING containers")
	t.Logf("Project %s has %d running containers", p.Name, wantedNumOfContainers)
}

// TestTaskScale runs `ecs-cli compose scale` for a project.
func TestTaskScale(t *testing.T, p *Project, scale int) {
	args := []string{
		"compose",
		"--file",
		p.ComposeFileName,
		"--ecs-params",
		p.ECSParamsFileName,
		"--project-name",
		p.Name,
		"scale",
		strconv.Itoa(scale),
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	require.NoErrorf(t, err, "Failed to scale task", "error %v, running %v, out: %s", err, args, string(out))

	// Then
	stdout.Stdout(out).TestHasAllSubstrings(t, []string{
		"Started container",
	})
	t.Logf("Scaled the task %s to %d", p.Name, scale)
}

// TestServiceScale runs `ecs-cli compose service scale` for a project.
func TestServiceScale(t *testing.T, p *Project, scale int) {
	// Given
	args := []string{
		"compose",
		"--file",
		p.ComposeFileName,
		"--ecs-params",
		p.ECSParamsFileName,
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
	require.NoErrorf(t, err, "Failed to scale service", "error %v, running %v, out: %s", err, args, string(out))

	// Then
	stdout.Stdout(out).TestHasAllSubstrings(t, []string{
		"ECS Service has reached a stable state",
		fmt.Sprintf("desiredCount=%d", scale),
		"serviceName=" + p.Name,
	})
	t.Logf("Scaled the service %s to %d tasks", p.Name, scale)
}

// TestTaskDown runs `ecs-cli compose down` for a project.
func TestTaskDown(t *testing.T, p *Project) {
	args := []string{
		"compose",
		"--file",
		p.ComposeFileName,
		"--ecs-params",
		p.ECSParamsFileName,
		"--project-name",
		p.Name,
		"down",
		"--cluster-config",
		p.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	require.NoErrorf(t, err, "Failed to delete task", "error %v, running %v, out: %s", err, args, string(out))

	// Then
	stdout.Stdout(out).TestHasAllSubstrings(t, []string{
		"Stopped container",
	})
	t.Logf("Deleted task %s", p.Name)
}

// TestServiceDown runs `ecs-cli compose service down` for a project.
func TestServiceDown(t *testing.T, p *Project) {
	// Given
	args := []string{
		"compose",
		"--file",
		p.ComposeFileName,
		"--ecs-params",
		p.ECSParamsFileName,
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
	require.NoErrorf(t, err, "Failed to delete service", "error %v, running %v, out: %s", err, args, string(out))

	// Then
	stdout.Stdout(out).TestHasAllSubstrings(t, []string{
		"Deleted ECS service",
		"ECS Service has reached a stable state",
		"desiredCount=0",
		"serviceName=" + p.Name,
	})
	t.Logf("Deleted service %s", p.Name)
}

func testServiceHasAllRunningContainers(t *testing.T, p *Project, wantedNumOfContainers int) bool {
	// Given
	args := []string{
		"compose",
		"--file",
		p.ComposeFileName,
		"--ecs-params",
		p.ECSParamsFileName,
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
	if err != nil {
		t.Logf("Failed to list containers for service %v", args)
		return false
	}

	// Then
	lines := stdout.Stdout(out).Lines()
	if len(lines) < 2 {
		t.Logf("No running containers yet, out = %v", lines)
		return false
	}
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1] // Drop the last new line
	}
	containers := lines[1:] // Drop the headers
	if wantedNumOfContainers != len(containers) {
		t.Logf("Wanted = %d, got = %d running containers", wantedNumOfContainers, len(containers))
		return false
	}
	for _, container := range containers {
		status := integ.GetRowValues(container)[1]
		if status != "RUNNING" {
			t.Logf("Container is not RUNNING: %s", container)
			return false
		}
	}
	return true
}
