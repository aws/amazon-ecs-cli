// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

func TestRegisterTDWithCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEcs := mock_ecsiface.NewMockECSAPI(ctrl)
	mockCache := mock_cache.NewMockCache(ctrl)
	client := NewECSClient()
	client.(*ecsClient).client = mockEcs
	defer ctrl.Finish()

	td1 := ecs.RegisterTaskDefinitionInput{
		Family: aws.String("family1"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name: aws.String("foo"),
			},
		},
	}

	td2 := ecs.RegisterTaskDefinitionInput{
		Family: aws.String("family2"),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			{
				Name: aws.String("foo"),
			},
		},
	}

	cache := make(map[string]interface{})

	gomock.InOrder(
		// First, expect a cache miss when it tries to register, so it actually
		// registers
		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Return(errors.New("MISS")),
		mockEcs.EXPECT().RegisterTaskDefinition(gomock.Any()).Do(func(input interface{}) {
			td := input.(*ecs.RegisterTaskDefinitionInput)
			if *td.Family != "family1" {
				t.Fatal("First td should have been family1")
			}
		}).Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &ecs.TaskDefinition{Family: aws.String("family1"), Revision: aws.Int64(1)}}, nil),
		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			cache[x.(string)] = y.(*ecs.TaskDefinition)
		}).Return(nil),
		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			td := y.(*ecs.TaskDefinition)
			cached := cache[x.(string)].(*ecs.TaskDefinition)
			*td = *cached
		}).Return(nil),
		// Doesn't get called a second time for family1 because of the cache
		mockCache.EXPECT().Get(gomock.Any(), gomock.Any()).Return(errors.New("MISS")),
		mockEcs.EXPECT().RegisterTaskDefinition(gomock.Any()).Do(func(input interface{}) {
			td := input.(*ecs.RegisterTaskDefinitionInput)
			if *td.Family != "family2" {
				t.Fatal("second td should have been family2")
			}
		}).Return(&ecs.RegisterTaskDefinitionOutput{TaskDefinition: &ecs.TaskDefinition{Family: aws.String("family2"), Revision: aws.Int64(1)}}, nil),
		mockCache.EXPECT().Put(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			if _, ok := cache[x.(string)]; ok {
				t.Fatal("there shouldn't be a cached family2 entry")
			}
		}).Return(nil),
	)

	resp1, err := client.RegisterTaskDefinitionIfNeeded(&td1, mockCache)
	if err != nil {
		t.Fatal(err)
	}
	resp2, err := client.RegisterTaskDefinitionIfNeeded(&td1, mockCache)
	if err != nil {
		t.Fatal(err)
	}

	if *resp1.Family != *resp2.Family || *resp1.Revision != *resp2.Revision {
		t.Errorf("Expected family/revision to match: %v:%v, %v:%v", *resp1.Family, *resp1.Revision, *resp2.Family, *resp2.Revision)
	}

	_, err = client.RegisterTaskDefinitionIfNeeded(&td2, mockCache)
	if err != nil {
		t.Error(err)
	}
}

func TestGetTasksPages(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEcs := mock_ecsiface.NewMockECSAPI(ctrl)
	client := NewECSClient()
	client.(*ecsClient).client = mockEcs
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
	ctrl := gomock.NewController(t)
	mockEcs := mock_ecsiface.NewMockECSAPI(ctrl)
	client := NewECSClient()
	client.(*ecsClient).client = mockEcs
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
	ctrl := gomock.NewController(t)
	mockEcs := mock_ecsiface.NewMockECSAPI(ctrl)
	client := NewECSClient()
	client.(*ecsClient).client = mockEcs
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
