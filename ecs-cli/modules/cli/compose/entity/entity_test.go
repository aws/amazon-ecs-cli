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
	"errors"
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ec2/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	eniIdentifier    = "eni-123456"
	taskArn          = "arn:123346:task/mytask"
	taskDefArn       = "arn:123456:taskdefinition/mytaskdef"
	privateIPAddress = "10.0.0.1"
	publicIPAddress  = "55.241.196.185"
)

func TestGetContainersForTasks(t *testing.T) {
	containerInstanceArn := "containerInstanceArn"
	ec2InstanceID := "ec2InstanceId"
	ec2Instance := &ec2.Instance{PublicIpAddress: aws.String(publicIPAddress)}

	ecsTasks := []*ecs.Task{
		&ecs.Task{
			Containers: []*ecs.Container{
				&ecs.Container{
					Name: aws.String("containerName"),
				},
			},
			ContainerInstanceArn: aws.String(containerInstanceArn),
		},
	}
	containerInstances := make(map[string]string)
	containerInstances[containerInstanceArn] = ec2InstanceID

	ec2Instances := make(map[string]*ec2.Instance)
	ec2Instances[ec2InstanceID] = ec2Instance

	mockProjectEntity := setupMocks(t, []*string{aws.String(containerInstanceArn)}, containerInstances,
		[]*string{aws.String(ec2InstanceID)}, ec2Instances)

	containers, err := getContainersForTasks(mockProjectEntity, ecsTasks, nil)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasks")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Equal(t, aws.StringValue(ec2Instance.PublicIpAddress), containers[0].EC2IPAddress, "Expects PublicIpAddress to match")
}

func ecsTask(launchType string) *ecs.Task {
	return ecsTaskWithOptions(launchType, ENIStatusAttached, ecs.DesiredStatusRunning)
}

func ecsTaskWithOptions(launchType string, attachmentStatus string, taskStatus string) *ecs.Task {
	return &ecs.Task{
		TaskDefinitionArn: aws.String(taskDefArn),
		TaskArn:           aws.String(taskArn),
		LaunchType:        aws.String(launchType),
		LastStatus:        aws.String(taskStatus),
		Attachments: []*ecs.Attachment{
			&ecs.Attachment{
				Status: aws.String(attachmentStatus),
				Type:   aws.String(ENIAttachmentType),
				Details: []*ecs.KeyValuePair{
					&ecs.KeyValuePair{
						Name:  aws.String(eniIDKey),
						Value: aws.String(eniIdentifier),
					},
				},
			},
		},
		Containers: []*ecs.Container{
			&ecs.Container{
				NetworkInterfaces: []*ecs.NetworkInterface{
					&ecs.NetworkInterface{
						PrivateIpv4Address: aws.String(privateIPAddress),
					},
				},
				Name: aws.String("containerName"),
			},
		},
	}
}

func taskDefinition() *ecs.TaskDefinition {
	return &ecs.TaskDefinition{
		TaskDefinitionArn: aws.String(taskDefArn),
		ContainerDefinitions: []*ecs.ContainerDefinition{
			&ecs.ContainerDefinition{
				Name: aws.String("containerName"),
				PortMappings: []*ecs.PortMapping{
					&ecs.PortMapping{
						ContainerPort: aws.Int64(80),
						HostPort:      aws.Int64(80),
						Protocol:      aws.String("tcp"),
					},
				},
			},
		},
	}
}

// Test case ensures that GetContainersForTasksWithTaskNetworking() can be given
// a list of tasks without task networking and return without error.
func TestGetContainersForTasksWithTaskNetworkingNoNetworkInterfaces(t *testing.T) {
	ecsTasks := []*ecs.Task{
		&ecs.Task{
			TaskDefinitionArn: aws.String(taskDefArn),
			TaskArn:           aws.String(taskArn),
			Containers: []*ecs.Container{
				&ecs.Container{
					Name: aws.String("containerName"),
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 0, "Expected to have 0 containers")
	assert.Len(t, tasks, 1, "Expected to have 1 tasks without task networking")
}

func TestGetContainersForTasksWithTaskNetworkingFargateENIDeleted(t *testing.T) {
	ecsTasks := []*ecs.Task{
		ecsTaskWithOptions("FARGATE", "DELETED", ecs.DesiredStatusRunning),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}
	taskDef := taskDefinition()

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Len(t, tasks, 0, "Expected to have 0 tasks without task networking")
	assert.Equal(t, privateIPAddress, containers[0].EC2IPAddress)
	assert.Equal(t, privateIPAddress+":80->80/tcp", containers[0].PortString())
}

func TestGetContainersForTasksWithTaskNetworkingFargateTaskStopped(t *testing.T) {
	ecsTasks := []*ecs.Task{
		ecsTaskWithOptions("FARGATE", "DELETED", ecs.DesiredStatusStopped),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}
	taskDef := taskDefinition()

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Len(t, tasks, 0, "Expected to have 0 tasks without task networking")
	assert.Equal(t, privateIPAddress, containers[0].EC2IPAddress)
	assert.Equal(t, privateIPAddress+":80->80/tcp", containers[0].PortString())
}

func TestGetContainersForTasksWithTaskNetworkingEC2TaskStopped(t *testing.T) {
	ecsTasks := []*ecs.Task{
		ecsTaskWithOptions("EC2", "DELETED", ecs.DesiredStatusStopped),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}
	taskDef := taskDefinition()

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Len(t, tasks, 0, "Expected to have 0 tasks without task networking")
	assert.Equal(t, privateIPAddress, containers[0].EC2IPAddress)
	assert.Equal(t, privateIPAddress+":80->80/tcp", containers[0].PortString())
}

func TestGetContainersForTasksWithTaskNetworkingEC2(t *testing.T) {
	ecsTasks := []*ecs.Task{
		ecsTask("EC2"),
	}

	taskDef := taskDefinition()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Len(t, tasks, 0, "Expected to have 0 tasks without task networking")
	assert.Equal(t, privateIPAddress, containers[0].EC2IPAddress)
	assert.Equal(t, privateIPAddress+":80->80/tcp", containers[0].PortString())
}

func TestGetContainersForTasksWithTaskNetworkingFargate(t *testing.T) {
	ecsTasks := []*ecs.Task{
		ecsTask("FARGATE"),
	}

	taskDef := taskDefinition()

	eni := &ec2.NetworkInterface{
		NetworkInterfaceId: aws.String(eniIdentifier),
		Association: &ec2.NetworkInterfaceAssociation{
			PublicIp: aws.String(publicIPAddress),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEc2.EXPECT().DescribeNetworkInterfaces(gomock.Any()).Do(func(x interface{}) {
			eniIDs := x.([]*string)
			assert.Equal(t, eniIdentifier, aws.StringValue(eniIDs[0]))
		}).Return([]*ec2.NetworkInterface{eni}, nil),
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Len(t, tasks, 0, "Expected to have 0 tasks without task networking")
	assert.Equal(t, publicIPAddress, containers[0].EC2IPAddress)
	assert.Equal(t, publicIPAddress+":80->80/tcp", containers[0].PortString())
}

func TestGetContainersForTasksWithTaskNetworkingFargateENIDescribeFails(t *testing.T) {
	ecsTasks := []*ecs.Task{
		ecsTask("FARGATE"),
	}

	taskDef := taskDefinition()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEc2.EXPECT().DescribeNetworkInterfaces(gomock.Any()).Do(func(x interface{}) {
			eniIDs := x.([]*string)
			assert.Equal(t, eniIdentifier, aws.StringValue(eniIDs[0]))
		}).Return(nil, fmt.Errorf("Some API Error")),
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Len(t, tasks, 0, "Expected to have 0 tasks without task networking")
	assert.Equal(t, privateIPAddress, containers[0].EC2IPAddress)
	assert.Equal(t, privateIPAddress+":80->80/tcp", containers[0].PortString())
}

func TestGetContainersForTasksWithTaskNetworkingFargateENIWithoutPublicIP(t *testing.T) {
	ecsTasks := []*ecs.Task{
		ecsTask("FARGATE"),
	}

	taskDef := taskDefinition()

	eni := &ec2.NetworkInterface{
		NetworkInterfaceId: aws.String(eniIdentifier),
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)
	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEc2.EXPECT().DescribeNetworkInterfaces(gomock.Any()).Do(func(x interface{}) {
			eniIDs := x.([]*string)
			assert.Equal(t, eniIdentifier, aws.StringValue(eniIDs[0]))
		}).Return([]*ec2.NetworkInterface{eni}, nil),
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().DescribeTaskDefinition(taskDefArn).Return(taskDef, nil),
	)

	containers, tasks, err := getContainersForTasksWithTaskNetworking(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasksWithTaskNetworking")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Len(t, tasks, 0, "Expected to have 0 tasks without task networking")
	assert.Equal(t, privateIPAddress, containers[0].EC2IPAddress)
	assert.Equal(t, privateIPAddress+":80->80/tcp", containers[0].PortString())
}

func TestGetContainersForTasksPrivateIP(t *testing.T) {
	containerInstanceArn := "containerInstanceArn"
	ec2InstanceID := "ec2InstanceId"
	ec2Instance := &ec2.Instance{PrivateIpAddress: aws.String(privateIPAddress)}

	ecsTasks := []*ecs.Task{
		&ecs.Task{
			Containers: []*ecs.Container{
				&ecs.Container{
					Name: aws.String("containerName"),
				},
			},
			ContainerInstanceArn: aws.String(containerInstanceArn),
		},
	}
	containerInstances := make(map[string]string)
	containerInstances[containerInstanceArn] = ec2InstanceID

	ec2Instances := make(map[string]*ec2.Instance)
	ec2Instances[ec2InstanceID] = ec2Instance

	mockProjectEntity := setupMocks(t, []*string{aws.String(containerInstanceArn)}, containerInstances,
		[]*string{aws.String(ec2InstanceID)}, ec2Instances)

	containers, err := getContainersForTasks(mockProjectEntity, ecsTasks, nil)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasks")
	assert.Len(t, containers, 1, "Expected to have 1 container")
	assert.Equal(t, aws.StringValue(ec2Instance.PrivateIpAddress), containers[0].EC2IPAddress, "Expects PublicIpAddress to match")
}

func TestGetContainersForTasksWithoutEc2InstanceID(t *testing.T) {
	containerInstanceArn := "containerInstanceArn"
	ec2InstanceID := "ec2InstanceId"

	ecsTasks := []*ecs.Task{
		&ecs.Task{
			Containers: []*ecs.Container{
				&ecs.Container{
					Name: aws.String("containerName"),
				},
			},
			ContainerInstanceArn: aws.String(containerInstanceArn),
		},
	}
	containerInstances := make(map[string]string)
	containerInstances[containerInstanceArn] = ec2InstanceID

	// No ec2InstanceID is found
	ec2Instances := make(map[string]*ec2.Instance)

	projectEntity := setupMocks(t, []*string{aws.String(containerInstanceArn)}, containerInstances,
		[]*string{aws.String(ec2InstanceID)}, ec2Instances)

	containers, err := getContainersForTasks(projectEntity, ecsTasks, nil)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasks")
	assert.Len(t, containers, 1, "Expects to have 1 containers")
	assert.Empty(t, containers[0].EC2IPAddress, "Expects ec2IpAddress to be empty")
}

func TestGetContainersForTasksErrorCases(t *testing.T) {
	containerInstanceArn := "containerInstanceArn"
	ec2InstanceID := "ec2InstanceId"

	ecsTasks := []*ecs.Task{
		&ecs.Task{
			Containers: []*ecs.Container{
				&ecs.Container{
					Name: aws.String("containerName"),
				},
			},
			ContainerInstanceArn: aws.String(containerInstanceArn),
		},
	}
	containerInstances := make(map[string]string)
	containerInstances[containerInstanceArn] = ec2InstanceID

	mockEc2, mockEcs, mockProjectEntity := setupTest(t)
	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}
	// GetEC2InstanceIDs failed
	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().GetEC2InstanceIDs(gomock.Any()).Return(nil, errors.New("something wrong")),
	)

	_, err := getContainersForTasks(mockProjectEntity, ecsTasks, nil)
	assert.Error(t, err, "Expected error when calling getContainersForTasks")

	// DescribeInstances failed
	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().GetEC2InstanceIDs(gomock.Any()).Return(containerInstances, nil),
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEc2.EXPECT().DescribeInstances(gomock.Any()).Return(nil, errors.New("something wrong")),
	)
	_, err = getContainersForTasks(mockProjectEntity, ecsTasks, nil)
	assert.Error(t, err, "Expected error when calling getContainersForTasks")
}

func setupTest(t *testing.T) (*mock_ec2.MockEC2Client, *mock_ecs.MockECSClient, *mock_entity.MockProjectEntity) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockEc2 := mock_ec2.NewMockEC2Client(ctrl)

	mockProjectEntity := mock_entity.NewMockProjectEntity(ctrl)

	return mockEc2, mockEcs, mockProjectEntity
}

func setupMocks(t *testing.T, getEc2InstanceIDsRequest []*string, getEc2InstanceIDsResult map[string]string,
	describeInstancesRequest []*string, describeInstancesResult map[string]*ec2.Instance) *mock_entity.MockProjectEntity {

	mockEc2, mockEcs, mockProjectEntity := setupTest(t)

	mockContext := &context.Context{
		ECSClient: mockEcs,
		EC2Client: mockEc2,
	}

	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().GetEC2InstanceIDs(getEc2InstanceIDsRequest).Return(getEc2InstanceIDsResult, nil),
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEc2.EXPECT().DescribeInstances(describeInstancesRequest).Return(describeInstancesResult, nil),
	)
	return mockProjectEntity
}
