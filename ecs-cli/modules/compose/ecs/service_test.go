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
	"strings"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs/mock"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskDefinition := ecs.TaskDefinition{
		Family:               aws.String("family"),
		ContainerDefinitions: []*ecs.ContainerDefinition{},
		Volumes:              []*ecs.Volume{},
	}
	taskDefId := "taskDefinitionId"
	respTaskDef := taskDefinition
	respTaskDef.TaskDefinitionArn = aws.String("arn/" + taskDefId)

	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	gomock.InOrder(
		mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			// verify input fields
			req := x.(*ecs.RegisterTaskDefinitionInput)
			if aws.StringValue(taskDefinition.Family) != aws.StringValue(req.Family) {
				t.Errorf("Expected taskDefintion family to be [%s] but got [%s]",
					aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family))
			}
		}).Return(&respTaskDef, nil),
		mockEcs.EXPECT().CreateService(gomock.Any(), taskDefId).Return(nil),
	)
	context := &Context{
		ECSClient: mockEcs,
	}
	service := NewService(context)
	service.SetTaskDefinition(&taskDefinition)

	err := service.Create()
	if err != nil {
		t.Fatal("Unexpected error while create")
	}
	// task definition should be set
	if aws.StringValue(respTaskDef.TaskDefinitionArn) != aws.StringValue(service.TaskDefinition().TaskDefinitionArn) {
		t.Errorf("Expected service TaskDefArn to be [%s] but got [%s]",
			aws.StringValue(respTaskDef.TaskDefinitionArn), aws.StringValue(service.TaskDefinition().TaskDefinitionArn))
	}
}

func TestServiceInfo(t *testing.T) {
	testInfo(func(context *Context) ProjectEntity {
		return NewService(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		if !strings.Contains(aws.StringValue(req.ServiceName), projectName) {
			t.Errorf("Expected serviceName to contain projectName [%s] but got [%s]",
				projectName, aws.StringValue(req.ServiceName))
		}
		if req.StartedBy != nil {
			t.Error("Expected startedby to be not set")
		}
	}, t, true)
}

func TestServiceRun(t *testing.T) {
	service := NewService(&Context{})
	if err := service.Run(map[string]string{}); err == nil {
		t.Error("Expected unsupported error")
	}
}
