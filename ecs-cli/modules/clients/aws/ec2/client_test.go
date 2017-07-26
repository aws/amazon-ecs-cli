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

package ec2

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ec2/mock/sdk"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDescribeInstances(t *testing.T) {
	mockEC2, client := setupTest(t)

	// 2 ids in the input list
	expectedIds := []*string{aws.String("id1"), aws.String("id2")}

	instance1 := &ec2.Instance{InstanceId: expectedIds[0]}
	instance2 := &ec2.Instance{InstanceId: expectedIds[1]}
	instance3 := &ec2.Instance{InstanceId: aws.String("id3")}
	reservation := &ec2.Reservation{
		Instances: []*ec2.Instance{instance1, instance2, instance3},
	}
	result := &ec2.DescribeInstancesOutput{
		Reservations: []*ec2.Reservation{reservation},
	}

	mockEC2.EXPECT().DescribeInstances(gomock.Any()).Do(func(input interface{}) {
		observedIds := input.(*ec2.DescribeInstancesInput)
		assert.Equal(t, len(expectedIds), len(observedIds.InstanceIds), "Expected request to have ids set")

		for idx := range expectedIds {
			assert.Equal(t, aws.StringValue(expectedIds[idx]), aws.StringValue(observedIds.InstanceIds[idx]), "Expected request instance ids to match")
		}
	}).Return(result, nil)

	output, err := client.DescribeInstances(expectedIds)
	assert.NoError(t, err, "Expected no error while Describing EC2 Instances")
	assert.NotEmpty(t, output, "Expected output to be of length")

	for _, id := range expectedIds {
		assert.NotNil(t, output[aws.StringValue(id)], "Expected output to have an instance")
	}
}

func TestDescribeInstancesWithEmptyList(t *testing.T) {
	_, client := setupTest(t)

	// empty list of input ids
	output, err := client.DescribeInstances([]*string{})
	assert.NoError(t, err, "Expected no error for empty input")
	assert.Empty(t, output, "Expected empty output map for empty input list")

}

func TestDescribeInstancesErrorCase(t *testing.T) {
	mockEC2, client := setupTest(t)

	expectedIds := []*string{aws.String("id1"), aws.String("id2")}

	mockEC2.EXPECT().DescribeInstances(gomock.Any()).Return(nil, errors.New("something failed"))

	_, err := client.DescribeInstances(expectedIds)

	assert.Error(t, err, "Expected error while Describing EC2 Instances")
}

func TestDescribeInstancesErrorCaseWithEmptyOutput(t *testing.T) {
	mockEC2, client := setupTest(t)

	expectedIds := []*string{aws.String("id1"), aws.String("id2")}

	// Describe returned nil reservations in the response
	mockEC2.EXPECT().DescribeInstances(gomock.Any()).Return(&ec2.DescribeInstancesOutput{}, nil)

	_, err := client.DescribeInstances(expectedIds)

	assert.Error(t, err, "Expected error for nil reservations")
}

func TestDescribeInstancesErrorCaseWithEmptyReservation(t *testing.T) {
	mockEC2, client := setupTest(t)

	expectedIds := []*string{aws.String("id1"), aws.String("id2")}

	// Describe returned empty reservations in the response
	mockEC2.EXPECT().DescribeInstances(gomock.Any()).Return(
		&ec2.DescribeInstancesOutput{Reservations: []*ec2.Reservation{}}, nil)

	_, err := client.DescribeInstances(expectedIds)

	assert.Error(t, err, "Expected error for empty reservations")
}

func setupTest(t *testing.T) (*mock_ec2iface.MockEC2API, EC2Client) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEC2 := mock_ec2iface.NewMockEC2API(ctrl)
	mockSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	client := newClient(&config.CLIParams{Session: mockSession}, mockEC2)

	return mockEC2, client
}
