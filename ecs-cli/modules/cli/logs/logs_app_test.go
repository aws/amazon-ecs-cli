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

package logs

import (
	"flag"
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudwatchlogs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	taskID              = "task1234"
	taskDefArn          = "arn:aws:ecs:us-west-2:123412341234:task-definition/myTaskDef:1"
	taskDefName         = "testTaskDef:7"
	containerName       = "wordpress"
	containerImage      = "wordpress"
	containerName2      = "mysql"
	containerImage2     = "mysql"
	logRegion1          = "us-east-2"
	logRegion2          = "us-east-1"
	logGroup1           = "testlogs"
	logGroup2           = "testlogs2"
	logPrefix1          = "testpre1"
	logPrefix2          = "testpre2"
	clientErrorMesssage = "Some Error with CloudWatch Logs Client"
)

func dummyTaskDef(containers []*ecs.ContainerDefinition) *ecs.TaskDefinition {
	taskDef := &ecs.TaskDefinition{}
	taskDef.SetContainerDefinitions(containers)
	taskDef.SetTaskDefinitionArn(taskDefArn)

	return taskDef
}

func dummyContainerDefFromLogOptions(logRegion string, logGroup string, logPrefix string) *ecs.ContainerDefinition {
	return dummyContainerDef(logRegion, logGroup, logPrefix, "awslogs", containerName, containerImage)
}

func dummyContainerDef(logRegion string, logGroup string, logPrefix string, logDriver string, name string, image string) *ecs.ContainerDefinition {
	container := &ecs.ContainerDefinition{}
	container.SetName(name)
	container.SetImage(image)
	logConfig := &ecs.LogConfiguration{}
	logConfig.SetLogDriver(logDriver)
	options := map[string]*string{
		"awslogs-stream-prefix": aws.String(logPrefix),
		"awslogs-group":         aws.String(logGroup),
		"awslogs-region":        aws.String(logRegion),
	}

	logConfig.SetOptions(options)
	container.SetLogConfiguration(logConfig)

	return container
}

func TestLogsRequestOneContainer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	var containers []*ecs.ContainerDefinition
	containers = append(containers, dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix1))
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	request, logRegion, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.NoError(t, err, "Unexpected error getting logs")
	assert.Equal(t, logRegion1, logRegion)
	assert.Equal(t, logGroup1, aws.StringValue(request.LogGroupName))
	assert.Equal(t, logPrefix1+"/"+containerName+"/"+taskID, aws.StringValue(request.LogStreamNames[0]))
	assert.Equal(t, 1, len(request.LogStreamNames))
}

func TestLogsRequestTwoContainers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	container1 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix1)
	container2 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix2)
	containers := []*ecs.ContainerDefinition{container1, container2}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	request, logRegion, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.NoError(t, err, "Unexpected error getting logs")
	assert.Equal(t, logRegion1, logRegion)
	assert.Equal(t, logGroup1, aws.StringValue(request.LogGroupName))
	assert.Equal(t, 2, len(request.LogStreamNames))
	assert.Contains(t, aws.StringValueSlice(request.LogStreamNames), "testpre1/wordpress/task1234")
	assert.Contains(t, aws.StringValueSlice(request.LogStreamNames), "testpre2/wordpress/task1234")
}

func TestLogsRequestNoLogConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	containers := []*ecs.ContainerDefinition{
		&ecs.ContainerDefinition{
			Name:  aws.String(containerName),
			Image: aws.String(containerImage),
		},
	}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	_, _, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.Error(t, err, "Unexpected error getting logs")
}

func TestLogsRequestTwoContainersDifferentPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	container1 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix1)
	container2 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix2)
	containers := []*ecs.ContainerDefinition{container1, container2}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	request, logRegion, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.NoError(t, err, "Unexpected error getting logs")
	assert.Equal(t, logRegion1, logRegion)
	assert.Equal(t, logGroup1, aws.StringValue(request.LogGroupName))
	assert.Equal(t, 2, len(request.LogStreamNames))
	assert.Contains(t, aws.StringValueSlice(request.LogStreamNames), "testpre1/wordpress/task1234")
	assert.Contains(t, aws.StringValueSlice(request.LogStreamNames), "testpre2/wordpress/task1234")
}

func TestLogsRequestWithTaskDefFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	container1 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix1)
	containers := []*ecs.ContainerDefinition{container1}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTaskDefinition(taskDefName).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	flagSet.String(flags.TaskDefinitionFlag, taskDefName, "")
	context := cli.NewContext(nil, flagSet, nil)

	request, logRegion, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.NoError(t, err, "Unexpected error getting logs")
	assert.Equal(t, logRegion1, logRegion)
	assert.Equal(t, logGroup1, aws.StringValue(request.LogGroupName))
	assert.Equal(t, logPrefix1+"/"+containerName+"/"+taskID, aws.StringValue(request.LogStreamNames[0]))
	assert.Equal(t, 1, len(request.LogStreamNames))
}

func TestLogsRequestContainerFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	container1 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix1)
	container2 := dummyContainerDef(logRegion1, logGroup1, logPrefix1, "awslogs", containerName2, containerImage2)
	containers := []*ecs.ContainerDefinition{container1, container2}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	flagSet.String(flags.ContainerNameFlag, containerName, "")
	context := cli.NewContext(nil, flagSet, nil)

	request, logRegion, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.NoError(t, err, "Unexpected error getting logs")
	assert.Equal(t, logRegion1, logRegion)
	assert.Equal(t, logGroup1, aws.StringValue(request.LogGroupName))
	assert.Equal(t, logPrefix1+"/"+containerName+"/"+taskID, aws.StringValue(request.LogStreamNames[0]))
	assert.Equal(t, 1, len(request.LogStreamNames))
}

// Error Cases

func TestLogsRequestMismatchRegionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	container1 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix1)
	container2 := dummyContainerDefFromLogOptions(logRegion2, logGroup1, logPrefix2)
	containers := []*ecs.ContainerDefinition{container1, container2}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	_, _, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.Error(t, err, "Expected error getting logs")
}

func TestLogsRequestMismatchLogGroupError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	container1 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, logPrefix1)
	container2 := dummyContainerDefFromLogOptions(logRegion1, logGroup2, logPrefix2)
	containers := []*ecs.ContainerDefinition{container1, container2}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	_, _, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.Error(t, err, "Expected error getting logs")
}

func TestLogsRequestWrongLogDriver(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	container1 := dummyContainerDef(logRegion1, logGroup1, logPrefix1, "mylogs", containerName, containerImage)
	containers := []*ecs.ContainerDefinition{container1}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	_, _, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.Error(t, err, "Expected error getting logs")
}

func TestLogsRequestNoPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTask := &ecs.Task{}
	ecsTask.SetTaskDefinitionArn(taskDefArn)
	ecsTasks := []*ecs.Task{ecsTask}

	container1 := dummyContainerDefFromLogOptions(logRegion1, logGroup1, "")
	containers := []*ecs.ContainerDefinition{container1}
	taskDef := dummyTaskDef(containers)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
		mockECS.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	_, _, err := logsRequest(context, mockECS, &config.CLIParams{})
	assert.Error(t, err, "Expected error getting logs")
}

func TestLogsRequestTaskNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)

	ecsTasks := make([]*ecs.Task, 0)

	gomock.InOrder(
		mockECS.EXPECT().DescribeTasks(gomock.Any()).Return(ecsTasks, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TaskIDFlag, taskID, "")
	context := cli.NewContext(nil, flagSet, nil)

	// Error message for this case includes info obtained from the params
	params := &config.CLIParams{}
	params.Cluster = "Cluster"
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-west-2")})
	assert.NoError(t, err)
	params.Session = sess

	_, _, err = logsRequest(context, mockECS, params)
	assert.Error(t, err, "Expected error getting logs")
}

/* Create Logs */

func TestCreateLogGroups(t *testing.T) {
	taskDef := dummyTaskDef([]*ecs.ContainerDefinition{
		dummyContainerDef(logRegion1, logGroup1, logPrefix1, "awslogs", containerName, containerImage),
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogFactory := mock_cloudwatchlogs.NewMockLogClientFactory(ctrl)
	mockLogClient := mock_cloudwatchlogs.NewMockClient(ctrl)

	gomock.InOrder(
		mockLogFactory.EXPECT().Get(logRegion1).Return(mockLogClient),
		mockLogClient.EXPECT().CreateLogGroup(gomock.Any()),
	)

	err := CreateLogGroups(taskDef, mockLogFactory)
	assert.NoError(t, err, "Unexpected error in call to CreateLogGroups()")
}

func TestCreateLogGroupsTwoContainers(t *testing.T) {
	taskDef := dummyTaskDef([]*ecs.ContainerDefinition{
		dummyContainerDef(logRegion1, logGroup1, logPrefix1, "awslogs", containerName, containerImage),
		dummyContainerDef(logRegion2, logGroup2, logPrefix1, "awslogs", containerName, containerImage),
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogFactory := mock_cloudwatchlogs.NewMockLogClientFactory(ctrl)
	mockLogClient := mock_cloudwatchlogs.NewMockClient(ctrl)

	gomock.InOrder(
		mockLogFactory.EXPECT().Get(logRegion1).Return(mockLogClient),
		mockLogClient.EXPECT().CreateLogGroup(gomock.Any()),
		mockLogFactory.EXPECT().Get(logRegion2).Return(mockLogClient),
		mockLogClient.EXPECT().CreateLogGroup(gomock.Any()),
	)

	err := CreateLogGroups(taskDef, mockLogFactory)
	assert.NoError(t, err, "Unexpected error in call to CreateLogGroups()")
}

func TestCreateLogGroupsWrongDriver(t *testing.T) {
	taskDef := dummyTaskDef([]*ecs.ContainerDefinition{
		dummyContainerDef(logRegion1, logGroup1, logPrefix1, "catsanddogslogger", containerName, containerImage),
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogFactory := mock_cloudwatchlogs.NewMockLogClientFactory(ctrl)

	err := CreateLogGroups(taskDef, mockLogFactory)
	assert.Error(t, err, "Expected error in call to CreateLogGroups()")
}

func TestCreateLogGroupsLogGroupAlreadyExists(t *testing.T) {
	taskDef := dummyTaskDef([]*ecs.ContainerDefinition{
		dummyContainerDef(logRegion1, logGroup1, logPrefix1, "awslogs", containerName, containerImage),
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogFactory := mock_cloudwatchlogs.NewMockLogClientFactory(ctrl)
	mockLogClient := mock_cloudwatchlogs.NewMockClient(ctrl)

	alreadyExistsErr := awserr.New(cloudwatchlogs.ErrCodeResourceAlreadyExistsException, "Resource Already Exists Exception", nil)

	gomock.InOrder(
		mockLogFactory.EXPECT().Get(logRegion1).Return(mockLogClient),
		mockLogClient.EXPECT().CreateLogGroup(gomock.Any()).Return(alreadyExistsErr),
	)

	err := CreateLogGroups(taskDef, mockLogFactory)
	assert.NoError(t, err, "Unexpected error in call to CreateLogGroups()")
}

func TestCreateLogGroupsErrorCase(t *testing.T) {
	taskDef := dummyTaskDef([]*ecs.ContainerDefinition{
		dummyContainerDef(logRegion1, logGroup1, logPrefix1, "awslogs", containerName, containerImage),
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLogFactory := mock_cloudwatchlogs.NewMockLogClientFactory(ctrl)
	mockLogClient := mock_cloudwatchlogs.NewMockClient(ctrl)

	someErr := fmt.Errorf(clientErrorMesssage)

	gomock.InOrder(
		mockLogFactory.EXPECT().Get(logRegion1).Return(mockLogClient),
		mockLogClient.EXPECT().CreateLogGroup(gomock.Any()).Return(someErr),
	)

	err := CreateLogGroups(taskDef, mockLogFactory)
	assert.Error(t, err, "Expected error in call to CreateLogGroups()")
	assert.Equal(t, clientErrorMesssage, err.Error())
}
