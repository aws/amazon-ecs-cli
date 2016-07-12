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

package cloudformation

import (
	"errors"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/cloudformation/mock/sdk"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/mock/gomock"
)

type noopsleeper struct{}

func (s *noopsleeper) Sleep(d time.Duration) {
}

func createStackEvent(status string) *cloudformation.DescribeStackEventsOutput {
	output := &cloudformation.DescribeStackEventsOutput{}
	output.StackEvents = []*cloudformation.StackEvent{
		&cloudformation.StackEvent{ResourceStatus: aws.String(status)},
	}

	return output
}

func createDescribeStacksOutput(status string) *cloudformation.DescribeStacksOutput {
	return &cloudformation.DescribeStacksOutput{
		Stacks: []*cloudformation.Stack{&cloudformation.Stack{StackStatus: aws.String(status)}},
	}
}

func TestWaitUntilCreateCompletes(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	eventCreateComplete := createStackEvent(cloudformation.ResourceStatusCreateComplete)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventCreateComplete, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusCreateComplete), nil)
	err := cfnClient.WaitUntilCreateComplete("")
	if err != nil {
		t.Error("Error waiting for create completion:", err)
	}
}

func TestWaitUntilCreateCompleteFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	eventCreateInProgress := createStackEvent(cloudformation.ResourceStatusCreateInProgress)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventCreateInProgress, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusCreateInProgress), nil)
	eventCreateFailed := createStackEvent(cloudformation.ResourceStatusCreateFailed)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventCreateFailed, nil)

	err := cfnClient.WaitUntilCreateComplete("")
	if err == nil {
		t.Error("Expected error waiting for create completion")
	}
}

func TestWaitUntilDeleteCompletes(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	eventDeleteComplete := createStackEvent(cloudformation.ResourceStatusDeleteComplete)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventDeleteComplete, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusDeleteComplete), nil)
	err := cfnClient.WaitUntilDeleteComplete("")
	if err != nil {
		t.Error("Error waiting for create completion:", err)
	}
}

func TestWaitUntilDeleteCompleteFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	eventDeleteInProgress := createStackEvent(cloudformation.ResourceStatusDeleteInProgress)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventDeleteInProgress, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusDeleteInProgress), nil)
	eventDeleteFailed := createStackEvent(cloudformation.ResourceStatusDeleteFailed)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventDeleteFailed, nil)

	err := cfnClient.WaitUntilDeleteComplete("")
	if err == nil {
		t.Error("Expected error waiting for create completion")
	}
}

func TestWaitUntilUpdateCompletes(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	eventInProgress := createStackEvent(cloudformation.ResourceStatusUpdateInProgress)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventInProgress, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusUpdateInProgress), nil)
	eventUpdateComplete := createStackEvent(cloudformation.ResourceStatusUpdateComplete)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventUpdateComplete, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusUpdateComplete), nil)
	err := cfnClient.WaitUntilUpdateComplete("")
	if err != nil {
		t.Error("Error waiting for update completion:", err)
	}
}

func TestWaitUntilUpdateCompleteFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	eventInProgress := createStackEvent(cloudformation.ResourceStatusUpdateInProgress)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventInProgress, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusUpdateInProgress), nil)
	eventUpdateFailed := createStackEvent(cloudformation.ResourceStatusUpdateFailed)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).Return(eventUpdateFailed, nil)

	err := cfnClient.WaitUntilUpdateComplete("")
	if err == nil {
		t.Error("Expected error waiting for update completion")
	}
}

func TestWaitDescribeEventsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).AnyTimes().Return(nil, errors.New(""))

	err := cfnClient.(*cloudformationClient).waitUntilComplete("", failureInCreateEvent, "", createStackFailures, 10)
	if err == nil {
		t.Error("Expected error waiting for create completion")
	}

	err = cfnClient.(*cloudformationClient).waitUntilComplete("", failureInDeleteEvent, "", deleteStackFailures, 10)
	if err == nil {
		t.Error("Expected error waiting for delete completion")
	}

	err = cfnClient.(*cloudformationClient).waitUntilComplete("", failureInUpdateEvent, "", updateStackFailures, 10)
	if err == nil {
		t.Error("Expected error waiting for update completion")
	}
}

func TestWaitExhaustRetries(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	eventCreateInProgress := createStackEvent(cloudformation.ResourceStatusCreateInProgress)
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).AnyTimes().Return(eventCreateInProgress, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).AnyTimes().Return(createDescribeStacksOutput(cloudformation.StackStatusCreateInProgress), nil)

	err := cfnClient.(*cloudformationClient).waitUntilComplete("", failureInCreateEvent, "", createStackFailures, 10)
	if err == nil {
		t.Error("Expected error waiting for create completion")
	}

	err = cfnClient.(*cloudformationClient).waitUntilComplete("", failureInDeleteEvent, "", deleteStackFailures, 10)
	if err == nil {
		t.Error("Expected error waiting for delete completion")
	}

	err = cfnClient.(*cloudformationClient).waitUntilComplete("", failureInUpdateEvent, "", updateStackFailures, 10)
	if err == nil {
		t.Error("Expected error waiting for update completion")
	}
}

func TestWaitDescribeStackFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	cfnClient.(*cloudformationClient).sleeper = &noopsleeper{}
	defer ctrl.Finish()

	// Create some stack events for firstStackEventWithFailure() to process.
	// latest event, no error.
	eventsWithFailure := createStackEvent(cloudformation.ResourceStatusCreateInProgress)
	eventsWithFailure.StackEvents = append(eventsWithFailure.StackEvents, &cloudformation.StackEvent{
		ResourceStatus: aws.String(cloudformation.ResourceStatusCreateInProgress),
	})
	// second event. failure.
	eventsWithFailure.StackEvents = append(eventsWithFailure.StackEvents, &cloudformation.StackEvent{
		ResourceStatus:       aws.String(cloudformation.ResourceStatusCreateFailed),
		ResourceStatusReason: aws.String("do you really wanna know?"),
	})
	// oldest event, no error.
	eventsWithFailure.StackEvents = append(eventsWithFailure.StackEvents, &cloudformation.StackEvent{
		ResourceStatus: aws.String(cloudformation.ResourceStatusCreateInProgress),
	})
	mockCfn.EXPECT().DescribeStackEvents(gomock.Any()).AnyTimes().Return(eventsWithFailure, nil)
	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(cloudformation.StackStatusCreateFailed), nil)

	err := cfnClient.(*cloudformationClient).waitUntilComplete("", failureInCreateEvent, "", createStackFailures, 10)
	if err == nil {
		t.Error("Expected error waiting for create completion")
	}
}

func TestFailureInCreateEvent(t *testing.T) {
	eventInProgress := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusCreateInProgress)}
	failed := failureInCreateEvent(eventInProgress)
	if failed {
		t.Fatal("Unexpected failure determining if create failed for in-progress event")
	}

	eventCreateFailed := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusCreateFailed)}
	failed = failureInCreateEvent(eventCreateFailed)
	if !failed {
		t.Fatal("Expected failure determining if create failed for rollback create failed event")
	}

	eventCreateComplete := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusCreateComplete)}
	failed = failureInCreateEvent(eventCreateComplete)
	if failed {
		t.Fatal("Unexpected failure determining if create failed for create complete event")
	}
}

func TestFailureInDeleteEvent(t *testing.T) {
	eventInProgress := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusCreateInProgress)}
	failed := failureInDeleteEvent(eventInProgress)
	if failed {
		t.Fatal("Unexpected failure determining if delete failed for in-progress event")
	}

	eventDeleteFailed := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusDeleteFailed)}
	failed = failureInDeleteEvent(eventDeleteFailed)
	if !failed {
		t.Fatal("Expected failure determining if delete failed for delete-failed event")
	}

	eventDeleteComplete := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusDeleteComplete)}
	failed = failureInDeleteEvent(eventDeleteComplete)
	if failed {
		t.Fatal("Unexpected failure determining if delete failed for delete complete event")
	}
}

func TestFailureInUpdateEvent(t *testing.T) {
	eventInProgress := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusUpdateInProgress)}
	failed := failureInUpdateEvent(eventInProgress)
	if failed {
		t.Fatal("Unexpected failure determining if update failed for in-progress event")
	}

	eventUpdateFailed := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusUpdateFailed)}
	failed = failureInUpdateEvent(eventUpdateFailed)
	if !failed {
		t.Fatal("Expected failure determining if update failed for update-failed event")
	}

	eventUpdateComplete := &cloudformation.StackEvent{ResourceStatus: aws.String(cloudformation.ResourceStatusUpdateComplete)}
	failed = failureInUpdateEvent(eventUpdateComplete)
	if failed {
		t.Fatal("Unexpected failure determining if update failed for update complete event")
	}
}

func TestValidateStackExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockCfn := mock_cloudformationiface.NewMockCloudFormationAPI(ctrl)
	cfnClient := NewCloudformationClient()
	cfnClient.(*cloudformationClient).client = mockCfn
	defer ctrl.Finish()

	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(nil, errors.New("describe-stacks error"))
	err := cfnClient.ValidateStackExists("")
	if err == nil {
		t.Error("Expected error validating if stack exists")
	}

	mockCfn.EXPECT().DescribeStacks(gomock.Any()).Return(createDescribeStacksOutput(""), nil)
	err = cfnClient.ValidateStackExists("")
	if err != nil {
		t.Error("Unexpected error validating if stack exists", err)
	}
}
