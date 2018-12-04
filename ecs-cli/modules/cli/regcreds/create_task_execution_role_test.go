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

package regcreds

import (
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcredio"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCreateTaskExecutionRole(t *testing.T) {
	testRegistry := "myreg.test.io"
	testRegCredARN := "arn:aws:secret/some-test-arn"
	testRegKMSKey := "arn:aws:kms:key/67yt-756yth"

	testCreds := map[string]regcredio.CredsOutputEntry{
		testRegistry: regcredio.BuildOutputEntry(testRegCredARN, testRegKMSKey, []string{"test"}),
	}

	testRoleName := "myNginxProjectRole"

	testPolicyArn := aws.String("arn:aws:iam::policy/" + testRoleName + "-policy")
	testRoleArn := aws.String("arn:aws:iam::role/" + testRoleName)

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockIAM.EXPECT().CreateOrFindRole(testRoleName, roleDescriptionString, assumeRolePolicyDocString).Return(*testRoleArn, nil),
		mocks.MockIAM.EXPECT().CreateRole(gomock.Any()).Return(&iam.CreateRoleOutput{Role: &iam.Role{Arn: testRoleArn}}, nil),
	)
	gomock.InOrder(
		// If KMSKeyID present, first thing to happen should be verifying its ARN
		mocks.MockKMS.EXPECT().GetValidKeyARN(testRegKMSKey).Return(testRegKMSKey, nil),
		mocks.MockIAM.EXPECT().CreatePolicy(gomock.Any()).Return(&iam.CreatePolicyOutput{Policy: &iam.Policy{Arn: testPolicyArn}}, nil),
		mocks.MockIAM.EXPECT().AttachRolePolicy(getExecutionRolePolicyARN("us-west-2"), testRoleName).Return(nil, nil),
		mocks.MockIAM.EXPECT().AttachRolePolicy(*testPolicyArn, testRoleName).Return(nil, nil),
	)

	testParams := executionRoleParams{
		CredEntries: testCreds,
		RoleName:    testRoleName,
		Region:      "us-west-2",
	}

	policyCreateTime, err := createTaskExecutionRole(testParams, mocks.MockIAM, mocks.MockKMS)
	assert.NoError(t, err, "Unexpected error when creating task execution role")
	assert.NotNil(t, policyCreateTime, "Expected policy create time to be non-nil")
}

func TestCreateTaskExecutionRole_NoKMSKey(t *testing.T) {
	testRegistry := "myreg.test.io"
	testRegCredARN := "arn:aws:secret/some-test-arn"
	testCreds := map[string]regcredio.CredsOutputEntry{
		testRegistry: regcredio.BuildOutputEntry(testRegCredARN, "", []string{""}),
	}
	testRoleName := "myNginxProjectRole"

	testPolicyArn := aws.String("arn:aws:iam::policy/" + testRoleName + "-policy")
	testRoleArn := aws.String("arn:aws:iam::role/" + testRoleName)

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockIAM.EXPECT().CreateOrFindRole(testRoleName, roleDescriptionString, assumeRolePolicyDocString).Return(*testRoleArn, nil),
		mocks.MockIAM.EXPECT().CreateRole(gomock.Any()).Return(&iam.CreateRoleOutput{Role: &iam.Role{Arn: testRoleArn}}, nil),
	)
	gomock.InOrder(
		mocks.MockIAM.EXPECT().CreatePolicy(gomock.Any()).Return(&iam.CreatePolicyOutput{Policy: &iam.Policy{Arn: testPolicyArn}}, nil),
		mocks.MockIAM.EXPECT().AttachRolePolicy(getExecutionRolePolicyARN("us-west-2"), testRoleName).Return(nil, nil),
		mocks.MockIAM.EXPECT().AttachRolePolicy(*testPolicyArn, testRoleName).Return(nil, nil),
	)

	testParams := executionRoleParams{
		CredEntries: testCreds,
		RoleName:    testRoleName,
		Region:      "us-west-2",
	}

	policyCreateTime, err := createTaskExecutionRole(testParams, mocks.MockIAM, mocks.MockKMS)
	assert.NoError(t, err, "Unexpected error when creating task execution role")
	assert.NotNil(t, policyCreateTime, "Expected policy create time to be non-nil")
}

func TestCreateTaskExecutionRole_RoleExists(t *testing.T) {
	testRegistry := "myreg.test.io"
	testRegCredARN := "arn:aws:secret/some-test-arn"
	testCreds := map[string]regcredio.CredsOutputEntry{
		testRegistry: regcredio.BuildOutputEntry(testRegCredARN, "", []string{""}),
	}
	testRoleName := "myNginxProjectRole"

	testPolicyArn := aws.String("arn:aws:iam::policy/" + testRoleName + "-policy")
	roleExistsError := awserr.New("EntityAlreadyExists", "Didn't you see the error code? This role already exists.", errors.New("something went wrong"))

	mocks := setupTestController(t)
	gomock.InOrder(
		// CreateOrFindRole should return nil if given role already exists
		mocks.MockIAM.EXPECT().CreateOrFindRole(testRoleName, roleDescriptionString, assumeRolePolicyDocString).Return("", nil),
		mocks.MockIAM.EXPECT().CreateRole(gomock.Any()).Return(nil, roleExistsError),
	)
	gomock.InOrder(
		mocks.MockIAM.EXPECT().CreatePolicy(gomock.Any()).Return(&iam.CreatePolicyOutput{Policy: &iam.Policy{Arn: testPolicyArn}}, nil),
		mocks.MockIAM.EXPECT().AttachRolePolicy(getExecutionRolePolicyARN("us-west-2"), testRoleName).Return(nil, nil),
		mocks.MockIAM.EXPECT().AttachRolePolicy(*testPolicyArn, testRoleName).Return(nil, nil),
	)

	testParams := executionRoleParams{
		CredEntries: testCreds,
		RoleName:    testRoleName,
		Region:      "us-west-2",
	}

	policyCreateTime, err := createTaskExecutionRole(testParams, mocks.MockIAM, mocks.MockKMS)
	assert.NoError(t, err, "Unexpected error when creating task execution role")
	assert.NotNil(t, policyCreateTime, "Expected policy create time to be non-nil")
}

func TestCreateTaskExecutionRole_ErrorOnCreateRoleFails(t *testing.T) {
	testRegistry := "myreg.test.io"
	testRegCredARN := "arn:aws:secret/some-test-arn"
	testCreds := map[string]regcredio.CredsOutputEntry{
		testRegistry: regcredio.BuildOutputEntry(testRegCredARN, "", []string{""}),
	}
	testRoleName := "myNginxProjectRole"

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockIAM.EXPECT().CreateOrFindRole(testRoleName, roleDescriptionString, assumeRolePolicyDocString).Return("", errors.New("something went wrong")),
		mocks.MockIAM.EXPECT().CreateRole(gomock.Any()).Return(nil, errors.New("something went wrong")),
	)

	testParams := executionRoleParams{
		CredEntries: testCreds,
		RoleName:    testRoleName,
		Region:      "us-west-2",
	}

	_, err := createTaskExecutionRole(testParams, mocks.MockIAM, mocks.MockKMS)
	assert.Error(t, err, "Expected error when CreateRole fails")
}

func TestCreateTaskExecutionRole_ErrorOnCreatePolicyFails(t *testing.T) {
	testRegistry := "myreg.test.io"
	testRegCredARN := "arn:aws:secret/some-test-arn"
	testCreds := map[string]regcredio.CredsOutputEntry{
		testRegistry: regcredio.BuildOutputEntry(testRegCredARN, "", []string{""}),
	}
	testRoleName := "myNginxProjectRole"

	testRoleArn := aws.String("arn:aws:iam::role/" + testRoleName)

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockIAM.EXPECT().CreateOrFindRole(testRoleName, roleDescriptionString, assumeRolePolicyDocString).Return(*testRoleArn, nil),
		mocks.MockIAM.EXPECT().CreateRole(gomock.Any()).Return(&iam.CreateRoleOutput{Role: &iam.Role{Arn: testRoleArn}}, nil),
	)
	gomock.InOrder(
		mocks.MockIAM.EXPECT().CreatePolicy(gomock.Any()).Return(nil, errors.New("something went wrong")),
	)

	testParams := executionRoleParams{
		CredEntries: testCreds,
		RoleName:    testRoleName,
		Region:      "us-west-2",
	}

	_, err := createTaskExecutionRole(testParams, mocks.MockIAM, mocks.MockKMS)
	assert.Error(t, err, "Expected error when CreatePolicy fails")
}
