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
	"flag"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
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
		assert.Equal(t, aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family), "Expected Task Defintion family to match.")
	}).Return(&respTaskDef, nil)

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)

	context := &context.Context{
		ECSClient:  mockEcs,
		CLIParams:  &config.CLIParams{},
		CLIContext: cliContext,
	}
	task := NewTask(context)
	task.SetTaskDefinition(&taskDefinition)

	err := task.Create()
	assert.NoError(t, err, "Unexpected error while create")
	assert.Equal(t, aws.StringValue(respTaskDef.TaskDefinitionArn), aws.StringValue(task.TaskDefinition().TaskDefinitionArn), "Expected TaskDefArn to match.")
}

func TestTaskInfoFilterLocal(t *testing.T) {
	entity.TestInfo(func(context *context.Context) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Equal(t, projectName, aws.StringValue(req.Family), "Expected Task Definition Family to be project name")
	}, t, true)
}

func TestTaskInfoAll(t *testing.T) {
	entity.TestInfo(func(context *context.Context) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Nil(t, req.StartedBy, "Unexpected filter on StartedBy")
		assert.Nil(t, req.Family, "Unexpected filter on Task Definition family")
		assert.Nil(t, req.ServiceName, "Unexpected filter on Service Name")
	}, t, false)
}

// TODO: Test UP
