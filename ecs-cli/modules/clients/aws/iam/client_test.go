// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package iam

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/iam/mock/sdk"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	testPolicyArn = "arn:aws:iam:policy/SomePolicy"
	testRoleName  = "myFancyTestRole"
)

func TestAttachRolePolicy(t *testing.T) {
	mockIAM, client := setupTestController(t)

	expectedInput := iam.AttachRolePolicyInput{
		PolicyArn: aws.String(testPolicyArn),
		RoleName:  aws.String(testRoleName),
	}
	mockIAM.EXPECT().AttachRolePolicy(&expectedInput).Return(&iam.AttachRolePolicyOutput{}, nil)

	_, err := client.AttachRolePolicy(testPolicyArn, testRoleName)
	assert.NoError(t, err, "Expected no error when Attaching Role Policy")
}

func TestAttachRolePolicy_ErrorCase(t *testing.T) {
	mockIAM, client := setupTestController(t)
	mockIAM.EXPECT().AttachRolePolicy(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := client.AttachRolePolicy(testPolicyArn, testRoleName)
	assert.Error(t, err, "Unexpected error when Attaching Role Policy")
}

func TestCreateRole(t *testing.T) {
	mockIAM, client := setupTestController(t)

	testDescription := "This is a role I'll use with ECS"
	testAssumeRolePolicyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow"},"Principal":{"Service":["TEST"]},"Action":"TEST"]}`

	expectedInput := iam.CreateRoleInput{
		RoleName:                 aws.String(testRoleName),
		Description:              aws.String(testDescription),
		AssumeRolePolicyDocument: aws.String(testAssumeRolePolicyDoc),
	}
	expectedOutputRole := iam.Role{
		Arn:                      aws.String("arn:" + testRoleName),
		RoleName:                 aws.String(testRoleName),
		Description:              aws.String(testDescription),
		AssumeRolePolicyDocument: aws.String(testAssumeRolePolicyDoc),
	}
	mockIAM.EXPECT().CreateRole(&expectedInput).Return(&iam.CreateRoleOutput{Role: &expectedOutputRole}, nil)

	output, err := client.CreateRole(expectedInput)
	assert.NoError(t, err, "Unexpected error when Creating Role")
	actualRole := *output.Role
	assert.NotEmpty(t, actualRole)
	assert.Equal(t, expectedOutputRole, actualRole)
}

func TestCreateRole_ErrorCase(t *testing.T) {
	mockIAM, client := setupTestController(t)
	mockIAM.EXPECT().CreateRole(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := client.CreateRole(iam.CreateRoleInput{})
	assert.Error(t, err, "Expected error when Creating Role")
}

func TestCreatePolicy(t *testing.T) {
	mockIAM, client := setupTestController(t)

	testPolicyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"test:TestAction","Resource":"arn:MyStuff"}]}`
	testPolicyName := "myFancyExecutionRolePolicy"
	testPolicyDescription := "This policy will let me do things"

	expectedInput := iam.CreatePolicyInput{
		PolicyDocument: aws.String(testPolicyDoc),
		PolicyName:     aws.String(testPolicyName),
		Description:    aws.String(testPolicyDescription),
	}
	expectedPolicy := iam.Policy{
		Arn:         aws.String("arn:" + testPolicyName),
		PolicyName:  aws.String(testPolicyName),
		Description: aws.String(testPolicyDescription),
	}
	mockIAM.EXPECT().CreatePolicy(&expectedInput).Return(&iam.CreatePolicyOutput{Policy: &expectedPolicy}, nil)

	output, err := client.CreatePolicy(expectedInput)
	assert.NoError(t, err, "Unexpected error when Creating Policy")
	actualPolicy := *output.Policy
	assert.NotEmpty(t, actualPolicy)
}

func TestCreatePolicy_ErrorCase(t *testing.T) {
	mockIAM, client := setupTestController(t)
	mockIAM.EXPECT().CreatePolicy(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := client.CreatePolicy(iam.CreatePolicyInput{})
	assert.Error(t, err, "Expected error when Creating Policy")
}

func setupTestController(t *testing.T) (*mock_iamiface.MockIAMAPI, Client) {
	ctrl := gomock.NewController(t)
	mockIAM := mock_iamiface.NewMockIAMAPI(ctrl)
	client := newClient(mockIAM)

	return mockIAM, client
}
