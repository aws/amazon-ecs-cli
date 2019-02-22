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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
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
		assert.Equal(t, aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family), "Expected Task Definition family to match.")
	}).Return(&respTaskDef, nil)

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)

	context := &context.ECSContext{
		ECSClient:     mockEcs,
		CommandConfig: &config.CommandConfig{},
		CLIContext:    cliContext,
	}
	task := NewTask(context)
	task.SetTaskDefinition(&taskDefinition)

	err := task.Create()
	assert.NoError(t, err, "Unexpected error while create")
	assert.Equal(t, aws.StringValue(respTaskDef.TaskDefinitionArn), aws.StringValue(task.TaskDefinition().TaskDefinitionArn), "Expected TaskDefArn to match.")
}

func TestTaskCreateWithTags(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskDefinition := ecs.TaskDefinition{
		Family:               aws.String("family"),
		ContainerDefinitions: []*ecs.ContainerDefinition{},
		Volumes:              []*ecs.Volume{},
	}
	respTaskDef := taskDefinition
	respTaskDef.TaskDefinitionArn = aws.String("taskDefinitionArn")

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.ResourceTagsFlag, "holmes=watson", "")
	cliContext := cli.NewContext(nil, flagSet, nil)

	mockEcs := mock_ecs.NewMockECSClient(ctrl)

	context := &context.ECSContext{
		ECSClient:     mockEcs,
		CommandConfig: &config.CommandConfig{},
		CLIContext:    cliContext,
	}

	expectedTags := []*ecs.Tag{
		&ecs.Tag{
			Key:   aws.String("holmes"),
			Value: aws.String("watson"),
		},
	}

	mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		// verify input fields
		req := x.(*ecs.RegisterTaskDefinitionInput)
		assert.Equal(t, aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family), "Expected Task Definition family to match.")
		assert.ElementsMatch(t, expectedTags, req.Tags, "Expected resource tags to match")
	}).Return(&respTaskDef, nil)

	task := NewTask(context)
	task.SetTaskDefinition(&taskDefinition)

	err := task.Create()
	assert.NoError(t, err, "Unexpected error while create")
	assert.Equal(t, aws.StringValue(respTaskDef.TaskDefinitionArn), aws.StringValue(task.TaskDefinition().TaskDefinitionArn), "Expected TaskDefArn to match.")
}

func TestTaskInfoFilterLocal(t *testing.T) {
	entity.TestInfo(func(context *context.ECSContext) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Equal(t, projectName, aws.StringValue(req.Family), "Expected Task Definition Family to be project name")
	}, t, true, "")
}

func TestTaskInfoAll(t *testing.T) {
	entity.TestInfo(func(context *context.ECSContext) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Nil(t, req.StartedBy, "Unexpected filter on StartedBy")
		assert.Nil(t, req.Family, "Unexpected filter on Task Definition family")
		assert.Nil(t, req.ServiceName, "Unexpected filter on Service Name")
	}, t, false, "")
}

func TestTaskInfoRunning(t *testing.T) {
	entity.TestInfo(func(context *context.ECSContext) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Nil(t, req.StartedBy, "Unexpected filter on StartedBy")
		assert.Nil(t, req.Family, "Unexpected filter on Task Definition family")
		assert.Nil(t, req.ServiceName, "Unexpected filter on Service Name")
		assert.Equal(t, ecs.DesiredStatusRunning, aws.StringValue(req.DesiredStatus), "Expected Desired status to match")
	}, t, false, ecs.DesiredStatusRunning)
}

func TestTaskInfoStopped(t *testing.T) {
	entity.TestInfo(func(context *context.ECSContext) entity.ProjectEntity {
		return NewTask(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Nil(t, req.StartedBy, "Unexpected filter on StartedBy")
		assert.Nil(t, req.Family, "Unexpected filter on Task Definition family")
		assert.Nil(t, req.ServiceName, "Unexpected filter on Service Name")
		assert.Equal(t, ecs.DesiredStatusStopped, aws.StringValue(req.DesiredStatus), "Expected Desired status to match")
	}, t, false, ecs.DesiredStatusStopped)
}

// TODO: Test UP

// tests for helpers
func TestConvertToECSTaskOverride(t *testing.T) {
	container := "railsapp"
	command := []string{"bundle exec puma -C config/puma.rb"}

	input := map[string][]string{
		container: command,
	}

	expected := &ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{
			{
				Name:    aws.String(container),
				Command: aws.StringSlice(command),
			},
		},
	}

	actual, err := convertToECSTaskOverride(input)

	if assert.NoError(t, err) {
		assert.Equal(t, expected, actual)
	}
}

func TestConvertToECSTaskOverride_WithNil(t *testing.T) {
	var input map[string][]string

	actual, err := convertToECSTaskOverride(input)

	if assert.NoError(t, err) {
		assert.Nil(t, actual)
	}
}

func TestBuildRuntaskInput(t *testing.T) {
	taskDef := "clydeApp"
	count := 1
	cluster := "myCluster"
	launchType := "EC2"

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)
	ctrl := gomock.NewController(t)
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	context := &context.ECSContext{
		ECSClient:  mockEcs,
		CLIContext: cliContext,
		CommandConfig: &config.CommandConfig{
			Cluster:    cluster,
			LaunchType: launchType,
		},
	}

	task := &Task{
		ecsContext: context,
	}

	req, err := task.buildRunTaskInput(taskDef, count, nil)

	if assert.NoError(t, err) {
		assert.Equal(t, aws.String(cluster), req.Cluster)
		assert.Equal(t, aws.String(taskDef), req.TaskDefinition)
		assert.Equal(t, aws.String(launchType), req.LaunchType)
		assert.Equal(t, int64(count), aws.Int64Value(req.Count))
		assert.Nil(t, req.Overrides)
	}
}

func TestBuildRuntaskInput_WithOverride(t *testing.T) {
	taskDef := "clydeApp"
	count := 1
	cluster := "myCluster"
	container := "railsapp"
	launchType := "EC2"
	command := []string{"bundle exec puma -C config/puma.rb"}
	override := map[string][]string{
		container: command,
	}

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)
	ctrl := gomock.NewController(t)
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	context := &context.ECSContext{
		ECSClient:  mockEcs,
		CLIContext: cliContext,
		CommandConfig: &config.CommandConfig{
			Cluster:    cluster,
			LaunchType: launchType,
		},
	}

	task := &Task{
		ecsContext: context,
	}

	expectedOverride := &ecs.TaskOverride{
		ContainerOverrides: []*ecs.ContainerOverride{
			{
				Name:    aws.String("railsapp"),
				Command: aws.StringSlice(command),
			},
		},
	}

	req, err := task.buildRunTaskInput(taskDef, count, override)

	if assert.NoError(t, err) {
		assert.Equal(t, aws.String(cluster), req.Cluster)
		assert.Equal(t, aws.String(taskDef), req.TaskDefinition)
		assert.Equal(t, aws.String(launchType), req.LaunchType)
		assert.Equal(t, int64(count), aws.Int64Value(req.Count))
		assert.Equal(t, expectedOverride, req.Overrides)
	}
}
