// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package sts

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/sts/mock/sdk"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetAWSAccountID(t *testing.T) {
	mockSts, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	expectedAccountID := "123456789"

	mockSts.EXPECT().GetCallerIdentity(gomock.Any()).Return(&sts.GetCallerIdentityOutput{
		Account: aws.String(expectedAccountID),
	}, nil)

	accountID, err := client.GetAWSAccountID()
	assert.NoError(t, err, "GetAWSAccountId")
	assert.Equal(t, expectedAccountID, accountID, "Expected account id to match")
}

func TestGetAWSAccountIDErrorCase(t *testing.T) {
	mockSts, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	mockSts.EXPECT().GetCallerIdentity(gomock.Any()).Return(nil, errors.New("something failed"))

	_, err := client.GetAWSAccountID()
	assert.Error(t, err, "Expected error while GetAWSAccountId is called")

}

func setupTestController(t *testing.T) (*mock_stsiface.MockSTSAPI, Client, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockSts := mock_stsiface.NewMockSTSAPI(ctrl)
	mockSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	client := newClient(&config.CliParams{Session: mockSession}, mockSts)

	return mockSts, client, ctrl
}
