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

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/cfn"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"

	"github.com/stretchr/testify/assert"
)

const ecsCLIStackNamePrefix = "amazon-ecs-cli-setup-"

// A VPC is a virtual private cloud.
type VPC struct {
	ID      string
	Subnets []string
}

// TestUp runs `ecs-cli up` given a CLI configuration and returns the created VPC.
func TestUp(t *testing.T, conf *CLIConfig) *VPC {
	// Given
	args := []string{
		"up",
		"--cluster-config",
		conf.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if err != nil {
		assert.FailNowf(t, "Failed to create cluster", "Error %v running %v", err, args)
	}

	// Then
	stdout.Stdout(out).TestHasAllSnippets(t, []string{
		"VPC created",
		"Subnet created",
		"Cluster creation succeeded",
	})
	cfn.TestStackNameExists(t, stackName(conf.ClusterName))
	return parseVPC(out)
}

func stackName(clusterName string) string {
	return ecsCLIStackNamePrefix + clusterName
}

func parseVPC(stdout []byte) *VPC {
	vpc := VPC{}

	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "VPC created:") {
			vpc.ID = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "Subnet created:") {
			vpc.Subnets = append(vpc.Subnets, strings.TrimSpace(strings.Split(line, ":")[1]))
		}
	}
	return &vpc
}
