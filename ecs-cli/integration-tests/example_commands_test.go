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

package integration

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComposeServiceUp(t *testing.T) {
	LongRunningTest(t)

	// run the command
	cmdArgs := []string{"compose", "service", "-f", "compose-files/docker-compose.yml", "up"}
	if _, err := exec.Command("ecs-cli", cmdArgs...).Output(); err != nil {
		assert.NoError(t, err, "Unexpected error when running compose service up command")
	}

	ecsClient, err := CreateEcsClient()
	assert.NoError(t, err, "Unexpected Error creating ecs client")
	ecsClient.DescribeService("compose-service-name-prefix-integration-tests") //TODO: Check that this is the right name

}
