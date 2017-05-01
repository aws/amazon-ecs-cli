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

package entity

import (
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/project/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ec2/mock"
	ecsClient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
	"github.com/golang/mock/gomock"
)

type validateListTasksInput func(*ecs.ListTasksInput, string, *testing.T)
type setupEntityForTestInfo func(*context.Context) ProjectEntity

func TestInfo(setupEntity setupEntityForTestInfo, validateFunc validateListTasksInput, t *testing.T, filterLocal bool) {
	projectName := "project"
	containerInstance := "containerInstance"
	ec2InstanceId := "ec2instanceId"
	ec2Instance := &ec2.Instance{
		PublicIpAddress: aws.String("publicIpAddress"),
	}

	instanceIdsMap := make(map[string]string)
	instanceIdsMap[containerInstance] = ec2InstanceId

	ec2InstancesMap := make(map[string]*ec2.Instance)
	ec2InstancesMap[ec2InstanceId] = ec2Instance

	container := &ecs.Container{
		Name:         aws.String("contName"),
		ContainerArn: aws.String("contArn/contId"),
		LastStatus:   aws.String("lastStatus"),
	}

	ecsTask := &ecs.Task{
		TaskArn:              aws.String("taskArn/taskId"),
		Containers:           []*ecs.Container{container},
		ContainerInstanceArn: aws.String(containerInstance),
	}

	runningTasks := []*ecs.Task{ecsTask}
	stoppedTasks := []*ecs.Task{ecsTask, ecsTask}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)

	gomock.InOrder(

		mockEcs.EXPECT().GetTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			// verify input fields
			req := x.(*ecs.ListTasksInput)
			validateFunc(req, projectName, t)
			if ecs.DesiredStatusRunning != aws.StringValue(req.DesiredStatus) {
				t.Errorf("Expected DesiredStatus to be [%s] but got [%s]",
					ecs.DesiredStatusRunning, aws.StringValue(req.DesiredStatus))
			}
			// execute the function passed as input
			funct := y.(ecsClient.ProcessTasksAction)
			funct(runningTasks)
		}).Return(nil),
		mockEcs.EXPECT().GetTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			// verify input fields
			req := x.(*ecs.ListTasksInput)
			validateFunc(req, projectName, t)
			if ecs.DesiredStatusStopped != aws.StringValue(req.DesiredStatus) {
				t.Errorf("Expected DesiredStatus to be [%s] but got [%s]",
					ecs.DesiredStatusStopped, aws.StringValue(req.DesiredStatus))
			}
			// execute the function passed as input
			funct := y.(ecsClient.ProcessTasksAction)
			funct(stoppedTasks)
		}).Return(nil),
		mockEcs.EXPECT().GetEC2InstanceIDs([]*string{&containerInstance}).Return(instanceIdsMap, nil),
		mockEc2.EXPECT().DescribeInstances([]*string{&ec2InstanceId}).Return(ec2InstancesMap, nil),
	)

	context := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
		ECSParams: &config.CliParams{},
		Context: project.Context{
			ProjectName: projectName,
		},
	}
	entity := setupEntity(context)
	infoSet, err := entity.Info(filterLocal)
	if err != nil {
		t.Fatal(err)
	}
	expectedCountOfContainers := len(runningTasks) + len(stoppedTasks)
	if expectedCountOfContainers != len(infoSet) {
		t.Errorf("Expected count to be [%s] but got [%s]",
			expectedCountOfContainers, len(infoSet))
	}
}
