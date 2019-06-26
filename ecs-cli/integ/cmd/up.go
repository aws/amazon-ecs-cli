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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/cfn"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"
)

const ecsCLIStackNamePrefix = "amazon-ecs-cli-setup-"

// A VPC is a virtual private cloud.
type VPC struct {
	ID      string
	Subnets []string
}

// TestUp runs `ecs-cli up` given a CLI configuration and returns the created VPC.
func TestUp(t *testing.T, conf *CLIConfig, options ...func([]string) []string) *VPC {
	// Given
	args := []string{
		"up",
		"--cluster-config",
		conf.ConfigName,
	}

	// forces the recreation of any existing resources that match current configuration
	// in case of a previously failed integ test
	args = append(args, "--force")

	for _, option := range options {
		args = option(args)
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	require.NoErrorf(t, err, "Failed to create cluster", "error %v, running %v, out: %s", err, args, string(out))

	// Then
	stdout.Stdout(out).TestHasAllSubstrings(t, []string{
		"VPC created",
		"Subnet created",
		"Cluster creation succeeded",
	})
	cfn.TestStackNameExists(t, stackName(conf.ClusterName))

	t.Logf("Created cluster %s in stack %s", conf.ClusterName, stackName(conf.ClusterName))
	return parseVPC(out)
}

// WithCapabilityIAM acknowledges that this command may create IAM resources.
func WithCapabilityIAM() func(args []string) []string {
	return func(args []string) []string {
		args = append(args, "--capability-iam")
		return args
	}
}

// WithSize sets the number of instances for the cluster.
func WithSize(size int) func(args []string) []string {
	return func(args []string) []string {
		args = append(args, "--size")
		args = append(args, fmt.Sprintf("%d", size))
		return args
	}
}

// WithInstanceType sets the EC2 instance type for the cluster
func WithInstanceType(instanceType string) func(args []string) []string {
	return func(args []string) []string {
		args = append(args, "--instance-type")
		args = append(args, "t2.medium")
		return args
	}
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
