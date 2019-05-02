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

// Package ecs contains validation functions against resources in ECS.
package ecs

import (
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"

	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/require"
)

// TestClusterSize validates that a cluster has the expected number of
// instances by periodically querying against ECS.
func TestClusterSize(t *testing.T, clusterName string, wantedSize int) {
	client := newClient(t)
	f := func(t *testing.T) bool {
		resp, err := client.ListContainerInstances(&ecs.ListContainerInstancesInput{
			Cluster: aws.String(clusterName),
		})
		if err != nil {
			t.Logf("Unexpected error: %v while listing container instances for %s", err, clusterName)
			return false
		}
		if len(resp.ContainerInstanceArns) != wantedSize {
			t.Logf("Number of container instances mismatch, wanted = %d, got = %d", wantedSize, len(resp.ContainerInstanceArns))
			return false
		}
		return true
	}
	timeoutInS := 600 * time.Second // 10 mins
	sleepInS := 15 * time.Second
	require.True(t, integ.RetryUntilTimeout(t, f, timeoutInS, sleepInS), "Failed to list container instances")
	t.Logf("Cluster %s has %d instances", clusterName, wantedSize)
}

// TestListTasks validates that a cluster has the expected number of
// tasks by periodically querying against ECS.
func TestListTasks(t *testing.T, clusterName string, wantedSize int) {
	client := newClient(t)
	f := func(t *testing.T) bool {
		resp, err := client.ListTasks(&ecs.ListTasksInput{
			Cluster: aws.String(clusterName),
		})
		if err != nil {
			t.Logf("Unexpected error: %v while listing tasks for %s", err, clusterName)
			return false
		}
		if len(resp.TaskArns) != wantedSize {
			t.Logf("Number of tasks mismatch, wanted = %d, got = %d", wantedSize, len(resp.TaskArns))
			return false
		}
		return true
	}
	timeoutInS := 180 * time.Second // 3 mins
	sleepInS := 15 * time.Second
	require.True(t, integ.RetryUntilTimeout(t, f, timeoutInS, sleepInS), "Failed to list tasks")
	t.Logf("Cluster %s has %d tasks", clusterName, wantedSize)
}

func newClient(t *testing.T) *ecs.ECS {
	sess, err := session.NewSession()
	require.NoError(t, err, "failed to create new session for ecs")
	conf := aws.NewConfig()
	return ecs.New(sess, conf)
}
