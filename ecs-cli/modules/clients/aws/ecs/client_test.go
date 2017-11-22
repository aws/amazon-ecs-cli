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

package ecs

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock/sdk"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/cache/mocks"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/version"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const clusterName = "clusterName"

// mockReadWriter implements ReadWriter interface to return just the cluster
// field whenperforming read.
type mockReadWriter struct{}

func (rdwr *mockReadWriter) Get(cluster string, profile string) (*config.CLIConfig, error) {
	return config.NewCLIConfig(clusterName), nil
}

func (rdwr *mockReadWriter) SaveProfile(configName string, profile *config.Profile) error {
	return nil
}

func (rdwr *mockReadWriter) SaveCluster(configName string, cluster *config.Cluster) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultProfile(configName string) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultCluster(configName string) error {
	return nil
}

func TestNewECSClientWithRegion(t *testing.T) {
	// TODO: Re-enable by making an integ test target in Makefile.
	t.Skip("Integ test, Re-enable Me!")
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}
	_, err := config.NewCLIParams(context, rdwr)
	assert.Error(t, err, "Expected error when region not specified")

	globalSet.String("region", "us-east-1", "")
	globalContext = cli.NewContext(nil, globalSet, nil)
	context = cli.NewContext(nil, nil, globalContext)
	params, err := config.NewCLIParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating opts")

	client := NewECSClient()
	client.Initialize(params)

	// test for UserAgent
	realClient, ok := client.(*ecsClient).client.(*ecs.ECS)
	assert.True(t, ok, "Could not cast client to ecs.ECS")

	buildHandlerList := realClient.Handlers.Build
	request := &request.Request{
		HTTPRequest: &http.Request{
			Header: http.Header{},
		},
	}
	buildHandlerList.Run(request)
	expectedUserAgentString := fmt.Sprintf("%s %s %s/%s",
		version.AppName, version.Version, aws.SDKName, aws.SDKVersion)
	userAgent := request.HTTPRequest.Header.Get(clients.UserAgentHeader)
	assert.Equal(t, expectedUserAgentString, userAgent, "Wrong User-Agent string")
}

func TestRegisterTDWithCache(t *testing.T) {
	defer os.Clearenv()

	mockEcs, mockCache, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	registerTaskDefinitionInput1 := ecs.RegisterTaskDefinitionInput{
		Family: aws.String("family1"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name: aws.String("foo"),
			},
		},
	}
	registerTaskDefinitionInput2 := ecs.RegisterTaskDefinitionInput{
		Family: aws.String("family2"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name: aws.String("foo"),
			},
		},
	}

	taskDefinition1 := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(1),
		Status:            aws.String(ecs.TaskDefinitionStatusActive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:1"),
	}
	taskDefinition2 := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput2.Family,
		Revision:          aws.Int64(1),
		Status:            aws.String(ecs.TaskDefinitionStatusActive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family2:1"),
	}

	describeTaskDefinitionInput1 := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: registerTaskDefinitionInput1.Family,
	}
	describeTaskDefinitionInput2 := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: registerTaskDefinitionInput2.Family,
	}
	describeTaskDefinitionInput1WithRevision := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskDefinition1.TaskDefinitionArn,
	}

	cache := make(map[string]interface{})

	gomock.InOrder(
		//First, we will mock the call to DescribeTaskDefinition
		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput1).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition1}, nil),

		// Next, expect a cache miss when it tries to register, so it actually
		// registers
		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Return(errors.New("MISS")),

		mockEcs.EXPECT().RegisterTaskDefinition(&registerTaskDefinitionInput1).
			Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &taskDefinition1}, nil),

		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			cache[x.(string)] = y.(*ecs.TaskDefinition)
		}).Return(nil),

		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput1).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition1}, nil),

		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			td := y.(*ecs.TaskDefinition)
			cached := cache[x.(string)].(*ecs.TaskDefinition)
			*td = *cached
		}).Return(nil),

		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput1WithRevision).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition1}, nil),

		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput2).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition2}, nil),

		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Return(errors.New("MISS")),

		mockEcs.EXPECT().RegisterTaskDefinition(&registerTaskDefinitionInput2).
			Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &taskDefinition2}, nil),

		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			_, ok := cache[x.(string)]
			assert.False(t, ok, "there shouldn't be a cached family2 entry")
		}).Return(nil),
	)

	resp1, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	assert.NoError(t, err, "Unexpected error when calling RegisterTaskDefinition")

	resp2, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	assert.NoError(t, err, "Unexpected error when calling RegisterTaskDefinition")
	assert.Equal(t, aws.StringValue(resp1.Family), aws.StringValue(resp2.Family), "Expected family to match")
	assert.Equal(t, aws.Int64Value(resp1.Revision), aws.Int64Value(resp2.Revision), "Expected revision to match")

	_, err = client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput2, mockCache)
	assert.NoError(t, err, "Unexpected error when calling RegisterTaskDefinition")
}

func TestRegisterTaskDefinitionIfNeededTDBecomesInactive(t *testing.T) {
	defer os.Clearenv()

	mockEcs, mockCache, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	registerTaskDefinitionInput1 := ecs.RegisterTaskDefinitionInput{
		Family: aws.String("family1"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name: aws.String("foo"),
			},
		},
	}

	describeTaskDefinitionInput1 := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: registerTaskDefinitionInput1.Family,
	}

	taskDefinition1 := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(1),
		Status:            aws.String(ecs.TaskDefinitionStatusActive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:1"),
	}

	taskDefinition1Inactive := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(1),
		Status:            aws.String(ecs.TaskDefinitionStatusInactive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:1"),
	}
	taskDefinition1Revision2 := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(2),
		Status:            aws.String(ecs.TaskDefinitionStatusActive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:2"),
	}

	cache := make(map[string]interface{})

	gomock.InOrder(
		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput1).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition1}, nil),

		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Return(errors.New("MISS")),

		mockEcs.EXPECT().RegisterTaskDefinition(&registerTaskDefinitionInput1).
			Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &taskDefinition1}, nil),

		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			cache[x.(string)] = y.(*ecs.TaskDefinition)
		}).Return(nil),

		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput1).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition1Inactive}, nil),

		mockEcs.EXPECT().RegisterTaskDefinition(&registerTaskDefinitionInput1).
			Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &taskDefinition1Revision2}, nil),

		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			cache[x.(string)] = y.(*ecs.TaskDefinition)
			if len(cache) != 1 {
				t.Fatal("There should only be one entry in the cache, since the previous INACTIVE task should have the same hash")
			}
		}).Return(nil),
	)

	resp1, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	assert.NoError(t, err, "Unexpected error when calling RegisterTaskDefinition")

	resp2, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	assert.NoError(t, err, "Unexpected error when calling RegisterTaskDefinition")

	assert.NotEqual(t, aws.Int64Value(resp1.Revision), aws.Int64Value(resp2.Revision), "Expected revision to be incremented")
}

func TestRegisterTaskDefinitionIfNeededFamilyNameNotProvided(t *testing.T) {
	_, _, client, ctrl := setupTestController(t, nil)
	defer ctrl.Finish()

	_, err := client.RegisterTaskDefinitionIfNeeded(&ecs.RegisterTaskDefinitionInput{}, nil)

	assert.Error(t, err, "Expected an error if the Family name was not provided.")
}

func TestRegisterTaskDefinitionIfNeededTDLatestTDRevisionIsInactive(t *testing.T) {
	defer os.Clearenv()

	mockEcs, mockCache, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	registerTaskDefinitionInput1 := ecs.RegisterTaskDefinitionInput{
		Family: aws.String("family1"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name: aws.String("foo"),
			},
		},
	}
	describeTaskDefinitionInput1 := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: registerTaskDefinitionInput1.Family,
	}
	taskDefinition1 := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(2),
		Status:            aws.String(ecs.TaskDefinitionStatusActive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:2"),
	}

	taskDefinition1Inactive := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(1),
		Status:            aws.String(ecs.TaskDefinitionStatusInactive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:1"),
	}

	gomock.InOrder(
		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput1).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition1Inactive}, nil),

		mockEcs.EXPECT().RegisterTaskDefinition(&registerTaskDefinitionInput1).
			Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &taskDefinition1}, nil),

		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil),
	)

	resp1, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	assert.NoError(t, err, "Unexpected error when calling RegisterTaskDefinition")
	assert.Condition(t, func() (success bool) {
		return aws.Int64Value(resp1.Revision) > aws.Int64Value(taskDefinition1Inactive.Revision)
	}, "Expected revison of response to be incremented because the latest task definition was INACTIVE")
}

func TestRegisterTaskDefinitionIfNeededCachedTDIsInactive(t *testing.T) {
	defer os.Clearenv()

	mockEcs, mockCache, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	registerTaskDefinitionInput1 := ecs.RegisterTaskDefinitionInput{
		Family: aws.String("family1"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name: aws.String("foo"),
			},
		},
	}
	taskDefinition2 := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(2),
		Status:            aws.String(ecs.TaskDefinitionStatusActive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:2"),
	}
	taskDefinition1CachedInactive := ecs.TaskDefinition{
		Family:            registerTaskDefinitionInput1.Family,
		Revision:          aws.Int64(1),
		Status:            aws.String(ecs.TaskDefinitionStatusInactive),
		TaskDefinitionArn: aws.String("arn:aws:ecs:region1:123456:task-definition/family1:1"),
	}
	describeTaskDefinitionInput2 := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: registerTaskDefinitionInput1.Family,
	}
	describeTaskDefinitionInput1Inactive := ecs.DescribeTaskDefinitionInput{
		TaskDefinition: taskDefinition1CachedInactive.TaskDefinitionArn,
	}

	gomock.InOrder(
		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput2).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition2}, nil),

		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			*y.(*ecs.TaskDefinition) = taskDefinition1CachedInactive
		}).Return(nil),

		mockEcs.EXPECT().DescribeTaskDefinition(&describeTaskDefinitionInput1Inactive).
			Return(&ecs.DescribeTaskDefinitionOutput{TaskDefinition: &taskDefinition1CachedInactive}, nil),

		mockEcs.EXPECT().RegisterTaskDefinition(&registerTaskDefinitionInput1).
			Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &taskDefinition2}, nil),

		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Return(nil),
	)

	resp1, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	assert.NoError(t, err, "Unexpected error when calling RegisterTaskDefinition")
	assert.Condition(t, func() (success bool) {
		return aws.Int64Value(resp1.Revision) > aws.Int64Value(taskDefinition1CachedInactive.Revision)
	}, "Expected revison of response to be incremented because the cached task definition is INACTIVE")
}

func TestGetTasksPages(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	family := "taskDefinitionFamily"
	taskIds := []*string{aws.String("taskId")}
	taskDetail := &ecs.Task{
		TaskArn: taskIds[0],
	}
	listTasksInput := &ecs.ListTasksInput{
		Family: aws.String(family),
	}

	mockEcs.EXPECT().ListTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		// verify input fields
		req := x.(*ecs.ListTasksInput)
		assert.Equal(t, clusterName, aws.StringValue(req.Cluster), "Expected clusterName to match")
		assert.Equal(t, aws.StringValue(listTasksInput.Family), aws.StringValue(req.Family), "Expected Family to match")

		// execute the function passed as input
		funct := y.(func(page *ecs.ListTasksOutput, end bool) bool)
		funct(&ecs.ListTasksOutput{TaskArns: taskIds}, false)
	}).Return(nil)

	mockEcs.EXPECT().DescribeTasks(gomock.Any()).Do(func(input interface{}) {
		// verify input fields
		req := input.(*ecs.DescribeTasksInput)
		assert.Equal(t, clusterName, aws.StringValue(req.Cluster), "Expected clusterName to match")
		assert.Equal(t, len(taskIds), len(req.Tasks), "Expected tasks length to match")
		assert.Equal(t, aws.StringValue(taskIds[0]), aws.StringValue(req.Tasks[0]), "Expected taskId to match")
	}).Return(&ecs.DescribeTasksOutput{Tasks: []*ecs.Task{taskDetail}}, nil)

	// make actual call
	client.GetTasksPages(listTasksInput, func(tasks []*ecs.Task) error {
		assert.Len(t, tasks, 1, "Expected exactly 1 task")
		assert.Equal(t, aws.StringValue(taskDetail.TaskArn), aws.StringValue(tasks[0].TaskArn), "Expected TaskArn to match")
		return nil
	})

}

func TestRunTask(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	td := "taskDef"
	group := "taskGroup"
	count := 5

	mockEcs.EXPECT().RunTask(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecs.RunTaskInput)
		assert.Equal(t, clusterName, aws.StringValue(req.Cluster), "Expected clusterName to match")
		assert.Equal(t, td, aws.StringValue(req.TaskDefinition), "Expected taskDefinition to match")
		assert.Equal(t, group, aws.StringValue(req.Group), "Expected group to match")
		assert.Equal(t, int64(count), aws.Int64Value(req.Count), "Expected count to match")
		assert.Nil(t, req.NetworkConfiguration, "Expected Network Config to be nil.")
		assert.Nil(t, req.LaunchType, "Expected Launch Type to be nil.")
	}).Return(&ecs.RunTaskOutput{}, nil)

	_, err := client.RunTask(td, group, count, nil, "")
	assert.NoError(t, err, "Unexpected error when calling RunTask")
}

func TestRunTaskWithLaunchTypeEC2(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getCLIConfigParamsWithLaunchType(t, "EC2"))
	defer ctrl.Finish()

	td := "taskDef"
	group := "taskGroup"
	count := 5

	mockEcs.EXPECT().RunTask(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecs.RunTaskInput)
		assert.Equal(t, clusterName, aws.StringValue(req.Cluster), "Expected clusterName to match")
		assert.Equal(t, td, aws.StringValue(req.TaskDefinition), "Expected taskDefinition to match")
		assert.Equal(t, group, aws.StringValue(req.Group), "Expected group to match")
		assert.Equal(t, int64(count), aws.Int64Value(req.Count), "Expected count to match")
		assert.Equal(t, "EC2", aws.StringValue(req.LaunchType))
		assert.Nil(t, req.NetworkConfiguration, "Expected Network Config to be nil.")
	}).Return(&ecs.RunTaskOutput{}, nil)

	_, err := client.RunTask(td, group, count, nil, "EC2")
	assert.NoError(t, err, "Unexpected error when calling RunTask")
}

func TestRunTaskWithLaunchTypeFargate(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getCLIConfigParamsWithLaunchType(t, "FARGATE"))
	defer ctrl.Finish()

	td := "taskDef"
	group := "taskGroup"
	count := 5

	subnets := []*string{aws.String("subnet-feedface")}
	securityGroups := []*string{aws.String("sg-c0ffeefe")}
	awsVpcConfig := &ecs.AwsVpcConfiguration{
		Subnets:        subnets,
		SecurityGroups: securityGroups,
		AssignPublicIp: aws.String("ENABLED"),
	}
	networkConfig := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: awsVpcConfig,
	}

	mockEcs.EXPECT().RunTask(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecs.RunTaskInput)
		assert.Equal(t, clusterName, aws.StringValue(req.Cluster), "Expected clusterName to match")
		assert.Equal(t, td, aws.StringValue(req.TaskDefinition), "Expected taskDefinition to match")
		assert.Equal(t, group, aws.StringValue(req.Group), "Expected group to match")
		assert.Equal(t, int64(count), aws.Int64Value(req.Count), "Expected count to match")
		assert.Equal(t, "FARGATE", aws.StringValue(req.LaunchType))
		assert.NotNil(t, req.NetworkConfiguration, "Expected Network Config to not be nil.")
	}).Return(&ecs.RunTaskOutput{}, nil)

	_, err := client.RunTask(td, group, count, networkConfig, "FARGATE")
	assert.NoError(t, err, "Unexpected error when calling RunTask")
}

func TestRunTask_WithTaskNetworking(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	td := "taskDef"
	group := "taskGroup"
	count := 5

	subnets := []*string{aws.String("subnet-feedface")}
	securityGroups := []*string{aws.String("sg-c0ffeefe")}
	awsVpcConfig := &ecs.AwsVpcConfiguration{
		Subnets:        subnets,
		SecurityGroups: securityGroups,
	}
	networkConfig := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: awsVpcConfig,
	}

	mockEcs.EXPECT().RunTask(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecs.RunTaskInput)
		assert.Equal(t, clusterName, aws.StringValue(req.Cluster), "Expected clusterName to match")
		assert.Equal(t, td, aws.StringValue(req.TaskDefinition), "Expected taskDefinition to match")
		assert.Equal(t, group, aws.StringValue(req.Group), "Expected group to match")
		assert.Equal(t, int64(count), aws.Int64Value(req.Count), "Expected count to match")
		assert.Equal(t, networkConfig, req.NetworkConfiguration, "Expected networkConfiguration to match")
	}).Return(&ecs.RunTaskOutput{}, nil)

	_, err := client.RunTask(td, group, count, networkConfig, "")
	assert.NoError(t, err, "Unexpected error when calling RunTask")
}

func TestIsActiveCluster(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, nil)
	defer ctrl.Finish()

	// API error
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(nil, errors.New("describe-clusters error"))
	_, err := client.IsActiveCluster("")
	assert.Error(t, err, "Expected error when calling IsActiveCluster")

	// Non 0 failures
	output := &ecs.DescribeClustersOutput{
		Failures: []*ecs.Failure{&ecs.Failure{}},
	}
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(output, nil)
	active, err := client.IsActiveCluster("")
	assert.NoError(t, err, "Unexpected error when calling IsActiveCluster")
	assert.False(t, active, "Expected IsActiveCluster to return false when API returned failures")

	// Inactive cluster
	output = &ecs.DescribeClustersOutput{
		Clusters: []*ecs.Cluster{&ecs.Cluster{Status: aws.String("INACTIVE")}},
	}
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(output, nil)
	active, err = client.IsActiveCluster("")
	assert.NoError(t, err, "Unexpected error when calling IsActiveCluster")
	assert.False(t, active, "Expected IsActiveCluster to return false when API returned inactive cluster")

	// Active cluster
	output = &ecs.DescribeClustersOutput{
		Clusters: []*ecs.Cluster{&ecs.Cluster{Status: aws.String("ACTIVE")}},
	}
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(output, nil)
	active, err = client.IsActiveCluster("")
	assert.NoError(t, err, "Unexpected error when calling IsActiveCluster")
	assert.True(t, active, "Expected IsActiveCluster to return true when API returned active cluster")
}

func TestGetEC2InstanceIDs(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	containerInstanceArn := "containerInstanceArn"
	containerInstanceArns := []*string{aws.String(containerInstanceArn)}
	ec2InstanceID := "ec2InstanceId"
	containerInstances := []*ecs.ContainerInstance{
		&ecs.ContainerInstance{
			ContainerInstanceArn: aws.String(containerInstanceArn),
			Ec2InstanceId:        aws.String(ec2InstanceID),
		},
	}

	mockEcs.EXPECT().DescribeContainerInstances(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecs.DescribeContainerInstancesInput)
		assert.Equal(t, clusterName, aws.StringValue(req.Cluster), "Expected clusterName to match")
		assert.Equal(t, len(containerInstanceArns), len(req.ContainerInstances), "Expected ContainerInstances to be the same length")
		assert.Equal(t, containerInstanceArn, aws.StringValue(req.ContainerInstances[0]), "Expected containerInstanceArn to match")
	}).Return(&ecs.DescribeContainerInstancesOutput{
		ContainerInstances: containerInstances,
	}, nil)

	containerToEC2InstanceMap, err := client.GetEC2InstanceIDs(containerInstanceArns)
	assert.NoError(t, err, "Unexpected error when calling GetEC2InstanceIDs")
	assert.Equal(t, ec2InstanceID, containerToEC2InstanceMap[containerInstanceArn], "Ec2InstanceId should match")
}

func TestGetEC2InstanceIDsWithEmptyArns(t *testing.T) {
	_, _, client, ctrl := setupTestController(t, nil)
	defer ctrl.Finish()

	containerToEC2InstanceMap, err := client.GetEC2InstanceIDs([]*string{})
	assert.NoError(t, err, "Unexpected error when calling GetEC2InstanceIDs")
	assert.Empty(t, containerToEC2InstanceMap, "containerToEC2InstanceMap should be empty")
}

func TestGetEC2InstanceIDsWithNoEc2InstanceID(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	containerInstanceArn := "containerInstanceArn"
	containerInstanceArns := []*string{aws.String(containerInstanceArn)}
	containerInstances := []*ecs.ContainerInstance{
		&ecs.ContainerInstance{
			ContainerInstanceArn: aws.String(containerInstanceArn),
		},
	}

	mockEcs.EXPECT().DescribeContainerInstances(gomock.Any()).Return(&ecs.DescribeContainerInstancesOutput{
		ContainerInstances: containerInstances,
	}, nil)

	containerToEC2InstanceMap, err := client.GetEC2InstanceIDs(containerInstanceArns)
	assert.NoError(t, err, "Unexpected error when calling GetEC2InstanceIDs")
	assert.Empty(t, containerToEC2InstanceMap, "containerToEC2InstanceMap should be empty")
}

func TestGetEC2InstanceIDsErrorCase(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, getDefaultCLIConfigParams(t))
	defer ctrl.Finish()

	containerInstanceArn := "containerInstanceArn"
	containerInstanceArns := []*string{aws.String(containerInstanceArn)}

	mockEcs.EXPECT().DescribeContainerInstances(gomock.Any()).Return(nil, errors.New("something wrong"))

	_, err := client.GetEC2InstanceIDs(containerInstanceArns)
	assert.Error(t, err, "Expected error when calling GetEC2InstanceIDs")
}

/*
	Helpers
*/
func setupTestController(t *testing.T, configParams *config.CLIParams) (*mock_ecsiface.MockECSAPI, *mock_cache.MockCache,
	ECSClient, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockEcs := mock_ecsiface.NewMockECSAPI(ctrl)
	mockCache := mock_cache.NewMockCache(ctrl)
	client := NewECSClient()

	if configParams != nil {
		client.Initialize(configParams)
	}

	client.(*ecsClient).client = mockEcs

	return mockEcs, mockCache, client, ctrl
}

func getDefaultCLIConfigParams(t *testing.T) *config.CLIParams {
	setDefaultAWSEnvVariables()

	testSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	return &config.CLIParams{
		Cluster: clusterName,
		Session: testSession,
	}
}

func getCLIConfigParamsWithLaunchType(t *testing.T, launchType string) *config.CLIParams {
	setDefaultAWSEnvVariables()

	testSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	return &config.CLIParams{
		Cluster:    clusterName,
		Session:    testSession,
		LaunchType: launchType,
	}
}

func setDefaultAWSEnvVariables() {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	os.Setenv("AWS_REGION", "region1")
}
