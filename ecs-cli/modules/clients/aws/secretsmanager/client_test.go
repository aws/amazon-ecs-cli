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

package secretsmanager

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/secretsmanager/mock/sdk"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCreateSecret(t *testing.T) {
	mockSM, client := setupTestController(t)

	expectedInput := secretsmanager.CreateSecretInput{
		Name:         aws.String("some-secret"),
		SecretString: aws.String("{\"username\":\"bob\"},{\"password\":\"abc123xyz456\"}"),
	}
	expectedOutput := secretsmanager.CreateSecretOutput{
		ARN:  aws.String("arn:aws:secretsmanager:some-secret-123"),
		Name: aws.String("some-secret"),
	}
	mockSM.EXPECT().CreateSecret(&expectedInput).Return(&expectedOutput, nil)

	output, err := client.CreateSecret(expectedInput)
	assert.NoError(t, err, "Expected no error when Creating Secret")
	assert.Equal(t, &expectedOutput, output, "Expected CreateSecret output to match")
}

func TestCreateSecretErrorCase(t *testing.T) {
	mockSM, client := setupTestController(t)

	mockSM.EXPECT().CreateSecret(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := client.CreateSecret(secretsmanager.CreateSecretInput{})
	assert.Error(t, err, "Expected error when Creating Secret")
}

func TestDescribeSecret(t *testing.T) {
	mockSM, client := setupTestController(t)

	testSecretName := "some-secret-1"
	expectedOutput := secretsmanager.DescribeSecretOutput{
		ARN:  aws.String("arn:aws:secretsmanager:secret:" + testSecretName),
		Name: aws.String(testSecretName),
	}

	mockSM.EXPECT().DescribeSecret(gomock.Any()).Return(&expectedOutput, nil)

	output, err := client.DescribeSecret(testSecretName)
	assert.NoError(t, err, "Expected no error when Describing Secret")
	assert.Equal(t, &expectedOutput, output, "Expected DescribeSecret output to match")
}

func TestDescribeSecretErrorCase(t *testing.T) {
	mockSM, client := setupTestController(t)

	mockSM.EXPECT().DescribeSecret(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := client.DescribeSecret("fake-secret-name")
	assert.Error(t, err, "Expected error when Describing Secret")
}

func TestListSecrets(t *testing.T) {
	mockSM, client := setupTestController(t)

	secretList := []*secretsmanager.SecretListEntry{&secretsmanager.SecretListEntry{Name: aws.String("my-secret")}}
	mockListResponse := &secretsmanager.ListSecretsOutput{
		SecretList: secretList,
	}
	mockSM.EXPECT().ListSecrets(gomock.Any()).Return(mockListResponse, nil)

	output, err := client.ListSecrets(nil)
	assert.NoError(t, err, "Expected no error when listing secrets")
	assert.NotEmpty(t, output, "Expected ListSecrets output to be non-empty")
}

func TestListSecretsWithNextToken(t *testing.T) {
	mockSM, client := setupTestController(t)
	tokenString := "someNextToken"
	nextToken := &tokenString

	secretList := []*secretsmanager.SecretListEntry{&secretsmanager.SecretListEntry{Name: aws.String("my-secret")}}
	mockListResponse := &secretsmanager.ListSecretsOutput{
		SecretList: secretList,
	}
	expectedInput := secretsmanager.ListSecretsInput{NextToken: nextToken}

	mockSM.EXPECT().ListSecrets(&expectedInput).Return(mockListResponse, nil)

	output, err := client.ListSecrets(nextToken)
	assert.NoError(t, err, "Expected no error when listing secrets")
	assert.NotEmpty(t, output, "Expected ListSecrets output to be non-empty")
}

func TestListSecretsErrorCase(t *testing.T) {
	mockSM, client := setupTestController(t)

	mockSM.EXPECT().ListSecrets(gomock.Any()).Return(nil, errors.New("something went wrong"))
	_, err := client.ListSecrets(nil)

	assert.Error(t, err, "Expected error when Listing Secrets")
}

func TestPutSecretValue(t *testing.T) {
	mockSM, client := setupTestController(t)

	expectedInput := secretsmanager.PutSecretValueInput{
		SecretId:     aws.String("my-test-secret"),
		SecretString: aws.String("{\"username\":\"testUser1\"},{\"password\":\"p4$$w0rd\"}"),
	}
	expectedOutput := secretsmanager.PutSecretValueOutput{
		ARN: aws.String("arn:aws:secretsmanager:my-test-secret"),
	}

	mockSM.EXPECT().PutSecretValue(&expectedInput).Return(&expectedOutput, nil)

	output, err := client.PutSecretValue(expectedInput)
	assert.NoError(t, err, "Expected no error when Putting Secret value")
	assert.Equal(t, &expectedOutput, output, "Expected PutSecretValue to match")
}

func TestPutSecretValueErrorCase(t *testing.T) {
	mockSM, client := setupTestController(t)

	mockSM.EXPECT().PutSecretValue(gomock.Any()).Return(nil, errors.New("something went wrong"))
	_, err := client.PutSecretValue(secretsmanager.PutSecretValueInput{})

	assert.Error(t, err, "Expected error when Putting Secret value")
}

func setupTestController(t *testing.T) (*mock_secretsmanageriface.MockSecretsManagerAPI, SMClient) {
	ctrl := gomock.NewController(t)
	mockSM := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)
	client := newClient(mockSM)

	return mockSM, client
}
