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

package task

import (
	"strings"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
)

func TestTaskCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskDefinition := ecs.TaskDefinition{
		Family:               aws.String("family"),
		ContainerDefinitions: []*ecs.ContainerDefinition{},
		Volumes:              []*ecs.Volume{},
	}
	respTaskDef := taskDefinition
	respTaskDef.TaskDefinitionArn = aws.String("taskDefinitionArn")

	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		// verify input fields
		req := x.(*ecs.RegisterTaskDefinitionInput)
		if aws.StringValue(taskDefinition.Family) != aws.StringValue(req.Family) {
			t.Errorf("Expected taskDefintion family to be [%s] but got [%s]",
				aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family))
		}
	}).Return(&respTaskDef, nil)

	context := &context.Context{
		ECSClient: mockEcs,
		ECSParams: &config.CliParams{},
	}
	task := NewTask(context)
	task.SetTaskDefinition(&taskDefinition)

	err := task.Create()
	if err != nil {
		t.Fatal("Unexpected error while create")
	}
	if aws.StringValue(respTaskDef.TaskDefinitionArn) != aws.StringValue(task.TaskDefinition().TaskDefinitionArn) {
		t.Errorf("Expected task's TaskDefArn to be [%s] but got [%s]",
			aws.StringValue(respTaskDef.TaskDefinitionArn), aws.StringValue(task.TaskDefinition().TaskDefinitionArn))
	}
}

func TestTaskInfoFilterLocal(t *testing.T) {
	entity.TestInfo(func(context *context.Context) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		if !strings.Contains(aws.StringValue(req.StartedBy), projectName) {
			t.Errorf("Expected startedby to contain projectName [%s] but got [%s]",
				projectName, aws.StringValue(req.StartedBy))
		}
	}, t, true)
}

func TestTaskInfoAll(t *testing.T) {
	entity.TestInfo(func(context *context.Context) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		if req.StartedBy != nil {
			t.Error("Expected startedby to be not set")
		}
	}, t, false)
}

// TODO: Test UP
