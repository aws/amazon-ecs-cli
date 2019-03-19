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

// Package cmd_up tests the "ecs-cli up" command with various configurations.
package cmd_up

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	// timeoutForWaitingOnActiveContainersInS is how long we are willing wait
	// for the container instances to become ACTIVE in the cluster.
	timeoutForWaitingOnActiveContainersInS = 300

	// sleepDurationInBetweenRetriesInS is how long we sleep in between retrying requests that fail.
	sleepDurationInBetweenRetriesInS       = 30
)

// TestClusterCreation runs the 'ecs-cli up -c <clusterName> --capability-iam --force' command.
// If there is no CloudFormation stack created, then the test fails.
func TestClusterCreation(t *testing.T) {
	// Given
	cfnClient, ecsClient, clusterName := setup(t)
	cmd := integ.GetCommand([]string{"up", "-c", clusterName, "--capability-iam", "--force"})

	// When
	stdout, err := cmd.Output()
	assert.NoError(t, err, fmt.Sprintf("Error running %v\nStdout: %s", cmd.Args, string(stdout)))

	// Then
	assertHasCFNStack(t, cfnClient, clusterName)
	assertHasActiveContainerInstances(t, ecsClient, clusterName, 1) // by default we only bring 1 t2.micro instance

	// Cleanup the created resources
	after(cfnClient, ecsClient, clusterName)
}

// setup initializes all the clients needed by the test.
func setup(t *testing.T) (cfnClient *cloudformation.CloudFormation, ecsClient *ecs.ECS, clusterName string) {
	sess, err := session.NewSession()
	// Fail the test immediately if we won't be able to evaluate it
	assert.NoError(t, err, "failed to create new session")

	conf := aws.NewConfig()
	cfnClient = cloudformation.New(sess, conf)
	ecsClient = ecs.New(sess, conf)
	clusterName = integ.SuggestedResourceName()
	return
}

// assertHasCFNStack validates that the CFN stack was created successfully
func assertHasCFNStack(t *testing.T, client *cloudformation.CloudFormation, clusterName string) {
	resp, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName(clusterName)),
	})
	assert.NoError(t, err, "unexpected CloudFormation error during DescribeStacks")
	assert.NotNil(t, resp.Stacks)
	assert.Len(t, resp.Stacks, 1)
	assert.Equal(t, *resp.Stacks[0].StackName, stackName(clusterName))
}

// assertHasActiveContainerInstances validates that the containers in the cluster are all eventually ACTIVE
func assertHasActiveContainerInstances(t *testing.T, client *ecs.ECS, clusterName string, clusterSize int) {
	maxNumRetries := timeoutForWaitingOnActiveContainersInS / sleepDurationInBetweenRetriesInS
	for retryCount := 0; retryCount < maxNumRetries; retryCount++ {
		cluster, err := client.ListContainerInstances(&ecs.ListContainerInstancesInput{
			Cluster: aws.String(clusterName),
		})
		if err != nil || len(cluster.ContainerInstanceArns) != clusterSize {
			t.Log("No available container instances in the cluster, retry...")
			time.Sleep(sleepDurationInBetweenRetriesInS * time.Second)
			continue
		}

		instances, err := client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(clusterName),
			ContainerInstances: cluster.ContainerInstanceArns,
		})
		if err != nil {
			t.Logf("Unexpected error while describing container instances, retry... %v", err)
			time.Sleep(sleepDurationInBetweenRetriesInS * time.Second)
			continue
		}

		hasAllInstancesActive := true
		for _, instance := range instances.ContainerInstances {
			hasAllInstancesActive = hasAllInstancesActive && *instance.Status == ecs.ContainerInstanceStatusActive
		}

		// All instances are up, we can exit successfully
		if hasAllInstancesActive {
			return
		}
		t.Log("Not all instances are active yet, retrying...")
		time.Sleep(sleepDurationInBetweenRetriesInS * time.Second)
	}
	assert.FailNow(t, "no active instances in the cluster",
		"The cluster %s failed to get active instances after %d seconds",
		clusterName,
		timeoutForWaitingOnActiveContainersInS)
}

// after best-effort deletes any resources created by the test.
func after(cfnClient *cloudformation.CloudFormation, ecsClient *ecs.ECS, clusterName string) {
	deleteStack(cfnClient, clusterName)
	deleteCluster(ecsClient, clusterName)
}

func deleteStack(client *cloudformation.CloudFormation, clusterName string) {
	client.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName(clusterName)),
	})
}

func deleteCluster(client *ecs.ECS, clusterName string) {
	client.DeleteCluster(&ecs.DeleteClusterInput{
		Cluster: aws.String(clusterName),
	})
}

func stackName(clusterName string) string {
	const ecsCLIStackNamePrefix = "amazon-ecs-cli-setup-"
	return ecsCLIStackNamePrefix + clusterName
}
