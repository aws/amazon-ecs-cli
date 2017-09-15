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

func TestGetContainersForTasks(t *testing.T) {
	containerInstanceArn := "containerInstanceArn"
	ec2InstanceID := "ec2InstanceId"
	ec2Instance := &ec2.Instance{PublicIpAddress: aws.String("publicIpAddress")}

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
	containerInstancesMap := make(map[string]string)
	containerInstancesMap[containerInstanceArn] = ec2InstanceID

	ec2InstancesMap := make(map[string]*ec2.Instance)
	ec2InstancesMap[ec2InstanceID] = ec2Instance

	mockProjectEntity := setupMocks(t, []*string{aws.String(containerInstanceArn)}, containerInstancesMap,
		[]*string{aws.String(ec2InstanceID)}, ec2InstancesMap)

	containers, err := getContainersForTasks(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasks")
	assert.Len(t, containers, 1, "Expects to have 1 containers")
	assert.Equal(t, containers[0].Ec2IPAddress, aws.StringValue(ec2Instance.PublicIpAddress), "Expects PublicIpAddress to match")
}

func TestGetContainersForTasksPrivateIP(t *testing.T) {
	containerInstanceArn := "containerInstanceArn"
	ec2InstanceID := "ec2InstanceId"
	ec2Instance := &ec2.Instance{PrivateIpAddress: aws.String("publicIpAddress")}

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
	containerInstancesMap := make(map[string]string)
	containerInstancesMap[containerInstanceArn] = ec2InstanceID

	ec2InstancesMap := make(map[string]*ec2.Instance)
	ec2InstancesMap[ec2InstanceID] = ec2Instance

	mockProjectEntity := setupMocks(t, []*string{aws.String(containerInstanceArn)}, containerInstancesMap,
		[]*string{aws.String(ec2InstanceID)}, ec2InstancesMap)

	containers, err := getContainersForTasks(mockProjectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasks")
	assert.Len(t, containers, 1, "Expects to have 1 containers")
	assert.Equal(t, containers[0].Ec2IPAddress, aws.StringValue(ec2Instance.PrivateIpAddress), "Expects PublicIpAddress to match")
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
	containerInstancesMap := make(map[string]string)
	containerInstancesMap[containerInstanceArn] = ec2InstanceID

	// No ec2InstanceID is found
	ec2InstancesMap := make(map[string]*ec2.Instance)

	projectEntity := setupMocks(t, []*string{aws.String(containerInstanceArn)}, containerInstancesMap,
		[]*string{aws.String(ec2InstanceID)}, ec2InstancesMap)

	containers, err := getContainersForTasks(projectEntity, ecsTasks)
	assert.NoError(t, err, "Unexpected error when calling getContainersForTasks")
	assert.Len(t, containers, 1, "Expects to have 1 containers")
	assert.Empty(t, containers[0].Ec2IPAddress, "Expects ec2IpAddress to be empty")
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
	containerInstancesMap := make(map[string]string)
	containerInstancesMap[containerInstanceArn] = ec2InstanceID

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

	_, err := getContainersForTasks(mockProjectEntity, ecsTasks)
	assert.Error(t, err, "Expected error when calling getContainersForTasks")

	// DescribeInstances failed
	gomock.InOrder(
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEcs.EXPECT().GetEC2InstanceIDs(gomock.Any()).Return(containerInstancesMap, nil),
		mockProjectEntity.EXPECT().Context().Return(mockContext),
		mockEc2.EXPECT().DescribeInstances(gomock.Any()).Return(nil, errors.New("something wrong")),
	)
	_, err = getContainersForTasks(mockProjectEntity, ecsTasks)
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
