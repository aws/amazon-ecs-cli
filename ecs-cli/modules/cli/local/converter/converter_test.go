// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package converter implements the logic to translate an ecs.TaskDefinition
// structure to a docker compose schema, which will be written to a
// docker-compose.local.yml file.

package converter

import (
	// "errors"
	// "os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	composeV3 "github.com/docker/cli/cli/compose/types"

	"github.com/stretchr/testify/assert"
)

func TestConvertToComposeService(t *testing.T) {
	// GIVEN
	expectedImage := "nginx"
	expectedName := "web"
	expectedCommand := []string{"CMD-SHELL", "curl -f http://localhost"}
	expectedEntrypoint := []string{"sh", "-c"}
	expectedWorkingDir := "./app"
	expectedHostname := "myHost"
	expectedLinks := []string{"container1"}
	expectedDNS := []string{"1.2.3.4"}
	expectedDNSSearch := []string{"search.example.com"}
	expectedUser := "admin"
	expectedSecurityOpt := []string{"label:type:test_virt"}
	expectedTty := true
	expectedPrivileged := true
	expectedReadOnly := true

	taskDefinition := &ecs.TaskDefinition{
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Image: aws.String(expectedImage),
				Name: aws.String(expectedName),
				Command: aws.StringSlice(expectedCommand),
				EntryPoint: aws.StringSlice(expectedEntrypoint),
				WorkingDirectory: aws.String(expectedWorkingDir),
				Hostname: aws.String(expectedHostname),
				Links: aws.StringSlice(expectedLinks),
				DnsServers: aws.StringSlice(expectedDNS),
				DnsSearchDomains: aws.StringSlice(expectedDNSSearch),
				User: aws.String(expectedUser),
				DockerSecurityOptions: aws.StringSlice(expectedSecurityOpt),
				PseudoTerminal: aws.Bool(expectedTty),
				Privileged: aws.Bool(expectedPrivileged),
				ReadonlyRootFilesystem: aws.Bool(expectedReadOnly),
			},
		},
	}

	containerDef := taskDefinition.ContainerDefinitions[0]

	// WHEN
	service, err := convertToComposeService(containerDef)

	// THEN
	assert.NoError(t, err, "Unexpected error converting Container Definition")
	assert.Equal(t, expectedName, service.Name, "Expected Name to match")
	assert.Equal(t, expectedImage, service.Image, "Expected Image to match")
	assert.Equal(t, composeV3.ShellCommand(expectedCommand), service.Command, "Expected Command to match")
	assert.Equal(t, composeV3.ShellCommand(expectedEntrypoint), service.Entrypoint, "Expected Entry point to match")
	assert.Equal(t, expectedWorkingDir, service.WorkingDir, "Expected WorkingDir to match")
	assert.Equal(t, expectedHostname, service.Hostname, "Expected Hostname to match")
	assert.Equal(t, expectedLinks, service.Links, "Expected Links to match")
	assert.Equal(t, composeV3.StringList(expectedDNS), service.DNS, "Expected DNS to match")
	assert.Equal(t, composeV3.StringList(expectedDNSSearch), service.DNSSearch, "Expected DNSSearch to match")
	assert.Equal(t, expectedUser, service.User, "Expected User to match")
	assert.Equal(t, expectedSecurityOpt, service.SecurityOpt, "Expected SecurityOpt to match")
	assert.Equal(t, expectedTty, service.Tty, "Expected Tty to match")
	assert.Equal(t, expectedPrivileged, service.Privileged, "Expected Privileged to match")
	assert.Equal(t, expectedReadOnly, service.ReadOnly, "Expected ReadOnly to match")
}
