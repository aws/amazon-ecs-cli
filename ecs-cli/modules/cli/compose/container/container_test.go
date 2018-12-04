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

package container

import (
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

const (
	taskId       = "taskId"
	taskArn      = "taskArn/" + taskId
	contId       = "contId"
	contArn      = "contArn/" + contId
	contName     = "contName"
	ec2IPAddress = "127.0.0.1"
)

func TestId(t *testing.T) {
	container := setupContainer()
	if contId != container.Id() {
		t.Errorf("Expected container id to be [%s] but got [%s]", contId, container.Id())
	}

	// incorrect arn
	container.ecsContainer.ContainerArn = aws.String("")

	if container.Id() != "" {
		t.Errorf("Expected container id to be empty but got [%s]", contId)
	}
}

func TestName(t *testing.T) {
	container := setupContainer()
	expectedContName := taskId + "/" + contName
	if expectedContName != container.Name() {
		t.Errorf("Expected container name to be [%s] but got [%s]", expectedContName, container.Name())
	}
}

func TestStatus(t *testing.T) {
	lastStatus := ecs.DesiredStatusStopped
	exitCode := 1
	reason := "reason"

	container := setupContainer()
	ecsCont := container.ecsContainer

	// just last status
	ecsCont.LastStatus = aws.String(lastStatus)
	state := container.State()
	if lastStatus != state {
		t.Errorf("Expected state to be [%s] but got [%s]", lastStatus, state)
	}

	// status with reason
	ecsCont.Reason = aws.String(reason)
	state = container.State()
	if !strings.Contains(state, reason) {
		t.Errorf("Expected state to contain [%s] but got [%s]", reason, state)
	}

	// status with exit code
	ecsCont.ExitCode = aws.Int64(int64(exitCode))
	state = container.State()
	if !strings.Contains(state, strconv.Itoa(exitCode)) {
		t.Errorf("Expected state to contain [%d] but got [%s]", exitCode, state)
	}
}

func TestPortString(t *testing.T) {
	contPort := 8000
	hostPort := 80
	ipAddr := "0.0.0.0"
	udp := ecs.TransportProtocolUdp

	binding1 := &ecs.NetworkBinding{
		BindIP:        aws.String(ipAddr),
		Protocol:      aws.String(udp),
		ContainerPort: aws.Int64(int64(contPort)),
		HostPort:      aws.Int64(int64(hostPort)),
	}
	expectedBinding1 := ec2IPAddress + ":80->8000/udp"

	binding2 := &ecs.NetworkBinding{
		BindIP:        aws.String(""),
		ContainerPort: aws.Int64(int64(contPort)),
		HostPort:      aws.Int64(int64(hostPort)),
	}
	expectedBinding2 := ":80->8000/tcp"

	container := setupContainer()
	container.networkBindings = []*ecs.NetworkBinding{binding1, binding2}

	portString := container.PortString()
	if !strings.Contains(portString, expectedBinding1) {
		t.Errorf("Expected portString to contain [%s] but got [%s]", expectedBinding1, portString)
	}
	if !strings.Contains(portString, expectedBinding2) {
		t.Errorf("Expected portString to contain [%s] but got [%s]", expectedBinding2, portString)
	}

	// container without ec2IPAddress
	container = setupContainer()
	container.EC2IPAddress = ""
	container.networkBindings = []*ecs.NetworkBinding{binding1}
	expectedBinding1WithoutEC2IpAddr := ipAddr + ":80->8000/udp"
	portString = container.PortString()
	if !strings.Contains(portString, expectedBinding1WithoutEC2IpAddr) {
		t.Errorf("Expected portString to contain [%s] but got [%s]", expectedBinding1WithoutEC2IpAddr, portString)
	}
}

func TestHealthStatus(t *testing.T) {
	containerHealth := ecs.HealthStatusHealthy
	container := setupContainer()
	container.ecsContainer.SetHealthStatus(containerHealth)

	// container.HealthStatus() should simply report the value returned by the ECS API
	assert.Equal(t, containerHealth, container.HealthStatus())
}

func setupContainer() Container {
	ecsContainer := &ecs.Container{
		ContainerArn: aws.String(contArn),
		Name:         aws.String(contName),
	}
	ecsTask := &ecs.Task{
		TaskArn: aws.String(taskArn),
	}
	return NewContainer(ecsTask, ec2IPAddress, ecsContainer, nil)
}
