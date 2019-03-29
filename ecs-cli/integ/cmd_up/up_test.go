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

package integ

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const (
	// timeoutForWaitingOnActiveInstancesInS is how long we are willing wait
	// for the container instances to become ACTIVE in the cluster.
	timeoutForWaitingOnActiveInstancesInS = 300

	// sleepDurationInBetweenRetriesInS is how long we sleep in between retrying requests that fail.
	sleepDurationInBetweenRetriesInS = 30
)

type upTest struct {
	clusterName                 string
	cmdArgs                     []string
	wantedStdoutSnippets        []string
	wantedClusterSize           int
	wantedInstanceAttributes    []*ecs.Attribute
	wantedInstanceResourceNames []string
}

var upTests = []upTest{
	{
		// Test cluster creation with default configurations
		integ.SuggestedResourceName("TestCmdUpWithDefaultConfig"),
		[]string{"up", "--capability-iam"},
		[]string{
			"Using recommended Amazon Linux 2 AMI",
			"VPC created",
			"Security Group created",
			"Subnet created",
			"Cluster creation succeeded",
		},
		1,
		[]*ecs.Attribute{
			{
				Name:  aws.String("ecs.instance-type"),
				Value: aws.String("t2.micro"),
			},
		},
		[]string{},
	},
	{
		// Test cluster creation with a GPU instance
		integ.SuggestedResourceName("TestCmdUpWithGPUInstance"),
		[]string{"up", "--instance-type", "p2.xlarge", "--capability-iam"},
		[]string{
			"Using GPU ecs-optimized AMI because instance type was p2.xlarge",
			"VPC created",
			"Security Group created",
			"Subnet created",
			"Cluster creation succeeded",
		},
		1,
		[]*ecs.Attribute{
			{
				Name:  aws.String("ecs.instance-type"),
				Value: aws.String("p2.xlarge"),
			},
		},
		[]string{"GPU"},
	},
}

// TestCmd_UP runs the 'ecs-cli up [command options] [arguments...]' command
// given a list of configurations and expected results.
func TestCmd_UP(t *testing.T) {
	cfnClient, ecsClient := setup(t)
	for _, test := range upTests {
		t.Run(fmt.Sprintf("create cluster %s", test.clusterName), func( *testing.T) {
			// Given
			cmdArgs := append(test.cmdArgs, "-c", test.clusterName)
			cmd := integ.GetCommand(cmdArgs)

			// When
			out, err := cmd.Output()
			if err != nil {
				assert.NoError(t, err, fmt.Sprintf("Error running %v\nStdout: %s", cmd.Args, string(out)))
				return
			}

			// Then
			if ok := integ.Stdout(out).HasAllSnippets(t, test.wantedStdoutSnippets); !ok {
				// assert failures don't halt the test, to move on to the next test case on failures we return
				return
			}
			if ok := hasCFNStackWithClusterName(t, cfnClient, test.clusterName); !ok {
				return
			}
			if ok := hasClusterWithWantedConfig(t, ecsClient, &test); !ok {
				return
			}

			// Cleanup the created resources
			after(cfnClient, ecsClient, test.clusterName)
		})
	}
}

// setup initializes all the clients needed by the upTest.
func setup(t *testing.T) (cfnClient *cloudformation.CloudFormation, ecsClient *ecs.ECS) {
	sess, err := session.NewSession()
	if err != nil {
		// Fail the upTest immediately if we won't be able to evaluate it
		assert.FailNowf(t, "failed to create new session for upTest clients", "%v", err)
	}

	conf := aws.NewConfig()
	cfnClient = cloudformation.New(sess, conf)
	ecsClient = ecs.New(sess, conf)
	return
}

// hasCFNStackWithClusterName returns true if the CFN stack was created successfully, false otherwise.
func hasCFNStackWithClusterName(t *testing.T, client *cloudformation.CloudFormation, clusterName string) bool {
	resp, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName(clusterName)),
	})
	if err != nil {
		return assert.Failf(t, "unexpected CloudFormation error during DescribeStacks", "wanted no errors, got %v", err)
	}
	if resp.Stacks == nil {
		return assert.Fail(t, "stacks should not be nil")
	}
	if len(resp.Stacks) != 1 {
		return assert.Failf(t, "did not receive only 1 stack", "wanted only one stack, got %d", len(resp.Stacks))
	}
	if *resp.Stacks[0].StackName != stackName(clusterName) {
		return assert.Failf(t, "unexpected stack name", "wanted %s, got %s", stackName(clusterName), *resp.Stacks[0].StackName)
	}
	return true
}

// hasClusterWithWantedConfig returns true if the instances in the cluster are all active, have the required attributes
// and resource, false otherwise.
func hasClusterWithWantedConfig(t *testing.T, client *ecs.ECS, test *upTest) bool {
	instances, err := getContainerInstances(t, client, test.clusterName, test.wantedClusterSize)
	if err != nil {
		assert.NoError(t, err)
		return false
	}
	if !areInstancesActive(t, instances) {
		return false
	}
	if !areInstancesWithAttributes(t, instances, test.wantedInstanceAttributes) {
		return false
	}
	if !areInstancesWithResourceNames(t, instances, test.wantedInstanceResourceNames) {
		return false
	}
	return true
}

// getContainerInstances returns the list of instances in the cluster.
func getContainerInstances(t *testing.T, client *ecs.ECS, clusterName string, clusterSize int) ([]*ecs.ContainerInstance, error) {
	maxNumRetries := timeoutForWaitingOnActiveInstancesInS / sleepDurationInBetweenRetriesInS
	for retryCount := 0; retryCount < maxNumRetries; retryCount++ {
		cluster, err := client.ListContainerInstances(&ecs.ListContainerInstancesInput{
			Cluster: aws.String(clusterName),
		})
		if err != nil {
			t.Logf("Unexpected error %v while listing container instances, retry...", err)
			time.Sleep(sleepDurationInBetweenRetriesInS * time.Second)
			continue
		}
		if len(cluster.ContainerInstanceArns) != clusterSize {
			t.Logf("Unexpected number of container instances wanted %d, got %d, retry...", clusterSize, len(cluster.ContainerInstanceArns))
			time.Sleep(sleepDurationInBetweenRetriesInS * time.Second)
			continue
		}
		instances, err := client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(clusterName),
			ContainerInstances: cluster.ContainerInstanceArns,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unexpected error while describing container instances")
		}
		return instances.ContainerInstances, nil
	}
	return nil, errors.New(fmt.Sprintf("timeout after %d seconds while retrieving container instances from cluster %s", timeoutForWaitingOnActiveInstancesInS, clusterName))
}

// areInstancesActive returns true if all the instances have the status ACTIVE, false otherwise.
func areInstancesActive(t *testing.T, instances []*ecs.ContainerInstance) bool {
	for _, instance := range instances {
		if !assert.Equal(t, *instance.Status, ecs.ContainerInstanceStatusActive) {
			return false
		}
	}
	return true
}

// areInstancesWithAttributes returns true if all the instances have the required attributes, false otherwise.
func areInstancesWithAttributes(t *testing.T, instances []*ecs.ContainerInstance, attributes []*ecs.Attribute) bool {
	for _, instance := range instances {
		instanceAttrs := make(map[string]*ecs.Attribute)
		for _, attr := range instance.Attributes {
			instanceAttrs[*attr.Name] = attr
		}

		for _, wantedAttr := range attributes {
			actualAttr, hasKey := instanceAttrs[*wantedAttr.Name]
			if !hasKey {
				t.Logf("instance %s is missing attribute name %s", *instance.Ec2InstanceId, *wantedAttr.Name)
				return false
			}
			if !assert.Equal(t, wantedAttr, actualAttr) {
				return false
			}
		}
	}
	return true
}

// areInstancesWithResourceNames returns true if all the instances have the required resource names, false otherwise.
func areInstancesWithResourceNames(t *testing.T, instances []*ecs.ContainerInstance, resourceNames []string) bool {
	for _, instance := range instances {
		instanceResources := make(map[string]bool)
		for _, resource := range instance.RegisteredResources {
			instanceResources[*resource.Name] = true
		}

		for _, wantedName := range resourceNames {
			if _, hasKey := instanceResources[wantedName]; !hasKey {
				t.Logf("instance %s is missing resource name %s", *instance.Ec2InstanceId, wantedName)
				return false
			}
		}
	}
	return true
}

// after best-effort deletes any resources created by the upTest.
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
