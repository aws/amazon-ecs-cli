// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ec2/mock/sdk"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/golang/mock/gomock"
)

func TestDescribeInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEC2 := mock_ec2iface.NewMockEC2API(ctrl)
	client := NewEC2Client(&config.CliParams{})
	client.(*ec2Client).client = mockEC2
	defer ctrl.Finish()

	// empty list of input ids
	output, err := client.DescribeInstances([]*string{})
	if err != nil {
		t.Errorf("Expected no error for empty input, but got [%v]", err)
	}
	if len(output) != 0 {
		t.Errorf("Expected empty output map for empty input list, but got [%v]", output)
	}

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
               if len(expectedIds) != len(observedIds.InstanceIds) {
                       t.Fatalf("Expected request to have ids set to [%v] but got [%v]", expectedIds, observedIds.InstanceIds)
               }
               for idx, _ := range expectedIds {
                       if aws.StringValue(expectedIds[idx]) != aws.StringValue(observedIds.InstanceIds[idx]) {
                               t.Fatalf("Expected request to have ids set to [%s] but got [%s]",
                                       aws.StringValue(expectedIds[idx]), aws.StringValue(observedIds.InstanceIds[idx]))
                       }
               }
	}).Return(result, nil)

	output, err = client.DescribeInstances(expectedIds)
	if err != nil {
		t.Fatalf("Expected no error while Describing EC2 Instances, but got [%v]", err)
	}
	if len(output) == 0 {
		t.Fatalf("Expected output to be of length [%s] but got 0", len(reservation.Instances))
	}
	for _, id := range expectedIds {
		if output[aws.StringValue(id)] == nil {
			t.Errorf("Expected output to have an instance for [%s] but got 0", aws.StringValue(id))
		}
	}
}

func TestDescribeInstancesErrorCases(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockEC2 := mock_ec2iface.NewMockEC2API(ctrl)
	client := NewEC2Client(&config.CliParams{})
	client.(*ec2Client).client = mockEC2
	defer ctrl.Finish()

	expectedIds := []*string{aws.String("id1"), aws.String("id2")}

	// Describe returned error
	mockEC2.EXPECT().DescribeInstances(gomock.Any()).Return(nil, errors.New("something failed"))
	_, err := client.DescribeInstances(expectedIds)
	if err == nil {
		t.Error("Expected error while Describing EC2 Instances, but got none")
	}

	// Describe returned nil reservations in the response
	mockEC2.EXPECT().DescribeInstances(gomock.Any()).Return(&ec2.DescribeInstancesOutput{}, nil)
	_, err = client.DescribeInstances(expectedIds)
	if err == nil {
		t.Error("Expected error for nil reservations, but got none")
	}

	// Describe returned empty reservations in the response
	mockEC2.EXPECT().DescribeInstances(gomock.Any()).Return(
		&ec2.DescribeInstancesOutput{Reservations: []*ec2.Reservation{}}, nil)
	_, err = client.DescribeInstances(expectedIds)
	if err == nil {
		t.Error("Expected error for empty reservations, but got none")
	}
}
