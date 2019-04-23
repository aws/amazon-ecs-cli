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

// Package cmd contains ECS CLI command tests.
package cmd

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"

	"github.com/stretchr/testify/assert"
)

// A CLIConfig holds the basic lookup information used by the ECS CLI for a cluster.
type CLIConfig struct {
	ClusterName string
	ConfigName  string
}

// TestFargateConfig runs `ecs-cli configure` with a FARGATE launch type.
func TestFargateConfig(t *testing.T) *CLIConfig {
	conf := CLIConfig{
		ClusterName: integ.SuggestedResourceName("fargate", "cluster"),
		ConfigName:  integ.SuggestedResourceName("fargate", "config"),
	}
	testConfig(t, conf.ClusterName, "FARGATE", conf.ConfigName)
	t.Logf("Created config %s", conf.ConfigName)
	return &conf
}

func testConfig(t *testing.T, clusterName, launchType, configName string) {
	args := []string{
		"configure",
		"--cluster",
		clusterName,
		"--region",
		os.Getenv("AWS_REGION"),
		"--default-launch-type",
		launchType,
		"--config-name",
		configName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if err != nil {
		assert.FailNowf(t, "Failed to configure CLI", "Error %v running %v", err, args)
	}

	// Then
	stdout.Stdout(out).TestHasAllSnippets(t, []string{
		fmt.Sprintf("Saved ECS CLI cluster configuration %s", configName),
	})
}
