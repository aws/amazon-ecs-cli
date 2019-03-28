// +build integ

// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package cmd_up

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestCusterCreationWithGPUInstances runs the 'ecs-cli up -c <clusterName> --instance-type <gpuInstance> --capability-iam --force' command.
func TestCusterCreationWithGPUInstances(t *testing.T) {
	// Given
	cfnClient, ecsClient, clusterName := setup(t)
	cmd := integ.GetCommand([]string{"up", "-c", clusterName, "--instance-type", "p2.xlarge", "--capability-iam", "--force"})

	// When
	stdout, err := cmd.Output()
	assert.NoError(t, err, fmt.Sprintf("Error running %v\nStdout: %s", cmd.Args, string(stdout)))

	// Then
	assertHasCFNStack(t, cfnClient, clusterName)
	assertHasActiveContainerInstances(t, ecsClient, clusterName, 1) // by default we only activate 1 instance
	assertHasGPUResources(t, ecsClient, clusterName)

	// Cleanup the created resources
	after(cfnClient, ecsClient, clusterName)
}

func assertHasGPUResources(t *testing.T, client *ecs.ECS, clusterName string) {
	cluster, err := client.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	})

	if err != nil {
		assert.FailNow(t, "failed to list container instances in the cluster", clusterName, err.Error())
	}
	instances, err := client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(clusterName),
		ContainerInstances: cluster.ContainerInstanceArns,
	})
	if err != nil {
		assert.FailNow(t, "failed to describe container instances in the cluster", clusterName, err.Error())
	}

	for _, instance := range instances.ContainerInstances {
		hasGPU := false
		for _, resource := range instance.RegisteredResources {
			hasGPU = hasGPU || *resource.Name == "GPU"
		}
		assert.True(t, hasGPU, "Expected instance %s to have GPU resource", *instance.Ec2InstanceId)
	}
}
