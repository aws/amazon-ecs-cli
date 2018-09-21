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

package kms

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/kms/mock/sdk"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestDescribeKey(t *testing.T) {
	mockKMS, client := setupTestController(t)

	testKeyID := "r6utfygh-677u-8765ytg00000"
	expectedResponseMetadata := kms.KeyMetadata{
		Arn:   aws.String("arn:aws:key/65rtfg-54erfd"),
		KeyId: aws.String(testKeyID),
	}

	expectedDescribeKeyResponse := kms.DescribeKeyOutput{
		KeyMetadata: &expectedResponseMetadata,
	}
	mockKMS.EXPECT().DescribeKey(gomock.Any()).Return(&expectedDescribeKeyResponse, nil)

	output, err := client.DescribeKey(testKeyID)
	assert.NoError(t, err, "Unexpected error when describing key")
	assert.Equal(t, &expectedDescribeKeyResponse, output, "Expected DescribeKey output to match")

}

func TestDescribeKey_ErrorCase(t *testing.T) {
	mockKMS, client := setupTestController(t)

	mockKMS.EXPECT().DescribeKey(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := client.DescribeKey("r6utfygh-677u-8765ytg00000")
	assert.Error(t, err, "Expected error when Describing Key")
}

func setupTestController(t *testing.T) (*mock_kmsiface.MockKMSAPI, Client) {
	ctrl := gomock.NewController(t)
	mockKMS := mock_kmsiface.NewMockKMSAPI(ctrl)
	client := newClient(mockKMS)

	return mockKMS, client
}
