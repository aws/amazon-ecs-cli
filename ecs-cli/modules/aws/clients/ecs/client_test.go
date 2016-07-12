// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs/mock/sdk"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/version"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils/cache/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/codegangsta/cli"
	"github.com/golang/mock/gomock"
)

var clusterName = "test"
var defaultCliConfigParams = config.CliParams{Cluster: "cluster", Config: &aws.Config{Region: aws.String("region1")}}

// mockReadWriter implements ReadWriter interface to return just the cluster
// field whenperforming read.
type mockReadWriter struct{}

func (rdwr *mockReadWriter) GetConfig() (*config.CliConfig, error) {
	return config.NewCliConfig(clusterName), nil
}

func (rdwr *mockReadWriter) ReadFrom(ecsConfig *config.CliConfig) error {
	return nil
}

func (rdwr *mockReadWriter) IsInitialized() (bool, error) {
	return true, nil
}

func (rdwr *mockReadWriter) Save(dest *config.Destination) error {
	return nil
}

func (rdwr *mockReadWriter) IsKeyPresent(section, key string) bool {
	return true
}

func TestNewECSClientWithRegion(t *testing.T) {
	// TODO: Re-enable by making an integ test target in Makefile.
	t.Skip("Integ test, Re-enable Me!")
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)
	context := cli.NewContext(nil, nil, globalContext)
	rdwr := &mockReadWriter{}
	_, err := config.NewCliParams(context, rdwr)
	if err == nil {
		t.Errorf("Expected error when region not specified")
	}

	globalSet.String("region", "us-east-1", "")
	globalContext = cli.NewContext(nil, globalSet, nil)
	context = cli.NewContext(nil, nil, globalContext)
	params, err := config.NewCliParams(context, rdwr)
	if err != nil {
		t.Errorf("Unexpected error creating opts: ", err)
	}
	client := NewECSClient()
	client.Initialize(params)

	// test for UserAgent
	realClient, ok := client.(*ecsClient).client.(*ecs.ECS)
	if !ok {
		t.Fatal("Could not cast client to ecs.ECS")
	}
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
	if userAgent != expectedUserAgentString {
		t.Errorf("Wrong User-Agent string, expected \"%s\" but was \"%s\"",
			expectedUserAgentString, userAgent)
	}
}

func setupTestController(t *testing.T, configParams *config.CliParams) (*mock_ecsiface.MockECSAPI, *mock_cache.MockCache,
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

func TestRegisterTDWithCache(t *testing.T) {
	mockEcs, mockCache, client, ctrl := setupTestController(t, &defaultCliConfigParams)
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
			if _, ok := cache[x.(string)]; ok {
				t.Fatal("there shouldn't be a cached family2 entry")
			}
		}).Return(nil),
	)

	resp1, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	if err != nil {
		t.Fatal(err)
	}

	if *resp1.Family != *resp2.Family || *resp1.Revision != *resp2.Revision {
		t.Errorf("Expected family/revision to match: %v:%v, %v:%v", *resp1.Family, *resp1.Revision, *resp2.Family, *resp2.Revision)
	}

	_, err = client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput2, mockCache)
	if err != nil {
		t.Error(err)
	}
}

func TestRegisterTaskDefinitionIfNeededTDBecomesInactive(t *testing.T) {
	mockEcs, mockCache, client, ctrl := setupTestController(t, &defaultCliConfigParams)
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
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := client.RegisterTaskDefinitionIfNeeded(&registerTaskDefinitionInput1, mockCache)
	if err != nil {
		t.Fatal(err)
	}

	if *resp1.Revision == *resp2.Revision {
		t.Errorf("Expected revison of second response to be incremented because the task definition is INACTIVE: %v:%v, %v:%v",
			*resp1.Family, *resp1.Revision, *resp2.Family, *resp2.Revision)
	}

}

func TestRegisterTaskDefinitionIfNeededFamilyNameNotProvided(t *testing.T) {
	_, _, client, ctrl := setupTestController(t, nil)
	defer ctrl.Finish()

	_, err := client.RegisterTaskDefinitionIfNeeded(&ecs.RegisterTaskDefinitionInput{}, nil)
	if err == nil {
		t.Fatal("Expected an error if the Family name was not provided.", err)
	}

}

func TestRegisterTaskDefinitionIfNeededTDLatestTDRevisionIsInactive(t *testing.T) {
	mockEcs, mockCache, client, ctrl := setupTestController(t, &defaultCliConfigParams)
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
	if err != nil {
		t.Fatal(err)
	}

	if *resp1.Revision <= *taskDefinition1Inactive.Revision {
		t.Errorf("Expected revison of response to be incremented because the latest task definition was INACTIVE: %v:%v",
			*taskDefinition1Inactive.Revision, *resp1.Revision)
	}

}

func TestRegisterTaskDefinitionIfNeededCachedTDIsInactive(t *testing.T) {
	mockEcs, mockCache, client, ctrl := setupTestController(t, &defaultCliConfigParams)
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
	if err != nil {
		t.Fatal(err)
	}

	if *resp1.Revision <= *taskDefinition1CachedInactive.Revision {
		t.Errorf("Expected revison of response to be incremented because the cached task definition is INACTIVE: %v:%v",
			*taskDefinition1CachedInactive.Revision, *resp1.Revision)
	}

}

func TestGetTasksPages(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, nil)
	defer ctrl.Finish()

	clusterName := "clusterName"
	client.(*ecsClient).params = &config.CliParams{
		Cluster: clusterName,
	}

	startedBy := "startedBy"
	taskIds := []*string{aws.String("taskId")}
	taskDetail := &ecs.Task{
		TaskArn: taskIds[0],
	}
	listTasksInput := &ecs.ListTasksInput{
		StartedBy: aws.String(startedBy),
	}

	mockEcs.EXPECT().ListTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		// verify input fields
		input := x.(*ecs.ListTasksInput)
		if clusterName != *input.Cluster {
			t.Errorf("Expected request.cluster to be [%s] but got [%s]",
				clusterName, *input.Cluster)
		}
		if *listTasksInput.StartedBy != *input.StartedBy {
			t.Errorf("Expected request.StartedBy to be [%v] but got [%v]",
				*listTasksInput.StartedBy, *input.StartedBy)
		}

		// execute the function passed as input
		funct := y.(func(page *ecs.ListTasksOutput, end bool) bool)
		funct(&ecs.ListTasksOutput{TaskArns: taskIds}, false)
	}).Return(nil)

	mockEcs.EXPECT().DescribeTasks(gomock.Any()).Do(func(input interface{}) {
		// verify input fields
		req := input.(*ecs.DescribeTasksInput)
		if clusterName != *req.Cluster {
			t.Errorf("Expected request.cluster to be [%s] but got [%s]",
				clusterName, *req.Cluster)
		}
		if len(taskIds) != len(req.Tasks) || *taskIds[0] != *req.Tasks[0] {
			t.Errorf("Expected request.tasks to be [%v] but got [%v]", taskIds, req.Tasks)
		}
	}).Return(&ecs.DescribeTasksOutput{Tasks: []*ecs.Task{taskDetail}}, nil)

	// make actual call
	client.GetTasksPages(listTasksInput, func(tasks []*ecs.Task) error {
		if len(tasks) != 1 {
			t.Fatalf("Expected tasks [%v] but got [%v]", taskDetail, tasks)
		}
		if *taskDetail.TaskArn != *tasks[0].TaskArn {
			t.Errorf("Expected TaskArn [%s] but got [%s]", *taskDetail.TaskArn, *tasks[0].TaskArn)
		}
		return nil
	})

}

func TestRunTask(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, nil)
	defer ctrl.Finish()

	clusterName := "clusterName"
	td := "taskDef"
	startedBy := "startedBy"
	count := 5
	client.(*ecsClient).params = &config.CliParams{
		Cluster: clusterName,
	}

	mockEcs.EXPECT().RunTask(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecs.RunTaskInput)
		if clusterName != aws.StringValue(req.Cluster) {
			t.Errorf("clusterName should be [%s]. Got [%s]", clusterName, aws.StringValue(req.Cluster))
		}
		if td != aws.StringValue(req.TaskDefinition) {
			t.Errorf("taskDefinition should be [%s]. Got [%s]", td, aws.StringValue(req.TaskDefinition))
		}
		if startedBy != aws.StringValue(req.StartedBy) {
			t.Errorf("startedBy should be [%s]. Got [%s]", startedBy, aws.StringValue(req.StartedBy))
		}
		if int64(count) != aws.Int64Value(req.Count) {
			t.Errorf("count should be [%s]. Got [%s]", count, aws.Int64Value(req.Count))
		}
	}).Return(&ecs.RunTaskOutput{}, nil)

	_, err := client.RunTask(td, startedBy, count)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIsActiveCluster(t *testing.T) {
	mockEcs, _, client, ctrl := setupTestController(t, nil)
	defer ctrl.Finish()

	// API error
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(nil, errors.New("describe-clusters error"))
	_, err := client.IsActiveCluster("")
	if err == nil {
		t.Error("Expected IsActiveCluster to return error on api error")
	}

	// Non 0 failures
	output := &ecs.DescribeClustersOutput{
		Failures: []*ecs.Failure{&ecs.Failure{}},
	}
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(output, nil)
	active, err := client.IsActiveCluster("")
	if err != nil {
		t.Fatal("Error in IsActiveCluster: ", err)
	}

	if active {
		t.Error("Expected IsActiveCluster to return false on api returning failures")
	}

	// Inactive cluster
	output = &ecs.DescribeClustersOutput{
		Clusters: []*ecs.Cluster{&ecs.Cluster{Status: aws.String("INACTIVE")}},
	}
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(output, nil)
	active, err = client.IsActiveCluster("")
	if err != nil {
		t.Fatal("Error in IsActiveCluster: ", err)
	}

	if active {
		t.Error("Expected IsActiveCluster to return false on api returning an inactive cluster")
	}

	// Active cluster
	output = &ecs.DescribeClustersOutput{
		Clusters: []*ecs.Cluster{&ecs.Cluster{Status: aws.String("ACTIVE")}},
	}
	mockEcs.EXPECT().DescribeClusters(gomock.Any()).Return(output, nil)
	active, err = client.IsActiveCluster("")
	if err != nil {
		t.Fatal("Error in IsActiveCluster: ", err)
	}

	if !active {
		t.Error("Expected IsActiveCluster to return true on api returning an active cluster")
	}

}
