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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/secretsmanager/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcreds"
	"github.com/aws/aws-sdk-go/aws"
	secretsmanager "github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestGetOrCreateRegistryCredentials_WithCredPair(t *testing.T) {
	testUsername := "someUser1"
	testPassword := "someP4$$w0rd"
	testRegistryName := "examplereg.net"
	testContainers := []string{"logging", "web"}

	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds[testRegistryName] = getTestCredsEntry("", testUsername, testPassword, "", testContainers)

	expectedCreateInput := secretsmanager.CreateSecretInput{
		Name:         generateSecretName(testRegistryName),
		SecretString: generateSecretString(testUsername, testPassword),
		Description:  generateSecretDescription(testRegistryName),
	}
	responseARN := "arn:aws:secretsmanager:examplereg.net-123"

	mockSM := setupTestController(t)
	gomock.InOrder(
		mockSM.EXPECT().DescribeSecret(gomock.Any()).Return(nil, nil),
		mockSM.EXPECT().CreateSecret(expectedCreateInput).Return(&secretsmanager.CreateSecretOutput{ARN: aws.String(responseARN)}, nil),
	)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mockSM, false)
	assert.NoError(t, err, "Expected no error when creating secret with cred pair")

	actualCredEntry := credsOutput[testRegistryName]
	assert.NotEmpty(t, actualCredEntry)
	assert.Equal(t, responseARN, actualCredEntry.CredentialARN)
	assert.ElementsMatch(t, testContainers, actualCredEntry.ContainerNames)
}

func TestGetOrCreateRegistryCredentials_WithCredPairAndKmsKey(t *testing.T) {
	testUsername := "someUser1"
	testPassword := "someP4$$w0rd"
	testKmsKeyID := "my-fav-key"
	testRegistryName := "examplereg.net"
	testContainers := []string{"logging", "web"}

	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds[testRegistryName] = getTestCredsEntry("", testUsername, testPassword, testKmsKeyID, testContainers)

	expectedCreateInput := secretsmanager.CreateSecretInput{
		Name:         generateSecretName(testRegistryName),
		SecretString: generateSecretString(testUsername, testPassword),
		KmsKeyId:     aws.String(testKmsKeyID),
		Description:  generateSecretDescription(testRegistryName),
	}
	responseARN := "arn:aws:secretsmanager:examplereg.net-123"

	mockSM := setupTestController(t)
	gomock.InOrder(
		mockSM.EXPECT().DescribeSecret(gomock.Any()).Return(nil, nil),
		mockSM.EXPECT().CreateSecret(expectedCreateInput).Return(&secretsmanager.CreateSecretOutput{ARN: aws.String(responseARN)}, nil),
	)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mockSM, false)
	assert.NoError(t, err, "Expected no error when creating secret with cred pair")

	actualCredEntry := credsOutput[testRegistryName]
	assert.NotEmpty(t, actualCredEntry)
	assert.Equal(t, responseARN, actualCredEntry.CredentialARN)
	assert.ElementsMatch(t, testContainers, actualCredEntry.ContainerNames)
}

func TestGetOrCreateRegistryCredentials_WithCredPairAndExistingFound(t *testing.T) {
	testUsername := "someUser1"
	testPassword := "someP4$$w0rd"
	testRegistryName := "examplereg.net"
	testContainers := []string{"logging", "web"}

	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds[testRegistryName] = getTestCredsEntry("", testUsername, testPassword, "", testContainers)

	responseARN := "arn:aws:secretsmanager:examplereg.net-123"

	mockSM := setupTestController(t)
	mockSM.EXPECT().DescribeSecret(gomock.Any()).Return(&secretsmanager.DescribeSecretOutput{ARN: aws.String(responseARN)}, nil)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mockSM, false)
	assert.NoError(t, err, "Expected no error when creating secret with cred pair")

	actualCredEntry := credsOutput[testRegistryName]
	assert.NotEmpty(t, actualCredEntry)
	assert.Equal(t, responseARN, actualCredEntry.CredentialARN)
	assert.ElementsMatch(t, testContainers, actualCredEntry.ContainerNames)
}

func TestGetOrCreateRegistryCredentials_WithSecretArnOnly(t *testing.T) {
	testSecretARN := "arn:aws:secretsmanager:examplereg.net-123"
	testRegistryName := "examplereg.net"
	testContainers := []string{"logging", "web"}

	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds[testRegistryName] = getTestCredsEntry(testSecretARN, "", "", "", testContainers)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, setupTestController(t), false)
	assert.NoError(t, err, "Expected no error when using existing secren ARN")

	actualCredEntry := credsOutput[testRegistryName]
	assert.NotEmpty(t, actualCredEntry)
	assert.Equal(t, testSecretARN, actualCredEntry.CredentialARN)
	assert.ElementsMatch(t, testContainers, actualCredEntry.ContainerNames)
}

func TestGetOrCreateRegistryCredentials_WithExistingAndCredsUpdateOk(t *testing.T) {
	testSecretARN := "arn:aws:secretsmanager:examplereg.net-123"
	testUsername := "someUser1"
	testPassword := "someP4$$w0rd"
	testRegistryName := "examplereg.net"
	testContainers := []string{"logging", "web"}

	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds[testRegistryName] = getTestCredsEntry(testSecretARN, testUsername, testPassword, "", testContainers)

	expectedPutSecretValueInput := secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(testSecretARN),
		SecretString: generateSecretString(testUsername, testPassword),
	}

	mockSM := setupTestController(t)
	mockSM.EXPECT().PutSecretValue(expectedPutSecretValueInput).Return(&secretsmanager.PutSecretValueOutput{}, nil)

	// call with updateAllowed = true
	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mockSM, true)
	assert.NoError(t, err, "Expected no error when updating existing secren ARN")

	actualCredEntry := credsOutput[testRegistryName]
	assert.NotEmpty(t, actualCredEntry)
	assert.ElementsMatch(t, testContainers, actualCredEntry.ContainerNames)
}

func TestGetOrCreateRegistryCredentials_WithExistingAndCredsNoUpdate(t *testing.T) {
	testSecretARN := "arn:aws:secretsmanager:examplereg.net-123"
	testUsername := "someUser1"
	testPassword := "someP4$$w0rd"
	testRegistryName := "examplereg.net"
	testContainers := []string{"logging", "web"}

	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds[testRegistryName] = getTestCredsEntry(testSecretARN, testUsername, testPassword, "", testContainers)

	// call with updateAllowed = false
	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, setupTestController(t), false)
	assert.NoError(t, err, "Expected no error when using existing secren ARN")

	actualCredEntry := credsOutput[testRegistryName]
	assert.NotEmpty(t, actualCredEntry)
	assert.ElementsMatch(t, testContainers, actualCredEntry.ContainerNames)
}

func TestGetOrCreateRegistryCredentials_ErrorOnCreate(t *testing.T) {
	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds["testRegistry"] = getTestCredsEntry("", "testUsername", "testPassword", "", []string{"test"})

	mockSM := setupTestController(t)
	gomock.InOrder(
		mockSM.EXPECT().DescribeSecret(gomock.Any()).Return(nil, nil),
		mockSM.EXPECT().CreateSecret(gomock.Any()).Return(nil, errors.New("something went wrong")),
	)

	_, err := getOrCreateRegistryCredentials(testRegistryCreds, mockSM, false)
	assert.Error(t, err)
}

func TestGetOrCreateRegistryCredentials_ErrorOnUpdate(t *testing.T) {
	testRegistryCreds := make(map[string]readers.RegistryCredEntry)
	testRegistryCreds["testRegistry"] = getTestCredsEntry("arn:aws:secretsmanager:secret:test", "testUsername", "testPassword", "", []string{"test"})

	mockSM := setupTestController(t)
	mockSM.EXPECT().PutSecretValue(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := getOrCreateRegistryCredentials(testRegistryCreds, mockSM, true)
	assert.Error(t, err)
}

func TestValidateCredsInput_ErrorEmptyCreds(t *testing.T) {
	emptyCredMap := make(map[string]readers.RegistryCredEntry)
	emptyCredsInput := readers.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: emptyCredMap,
	}

	err := validateCredsInput(emptyCredsInput)
	assert.Error(t, err, "Expected empty creds to return error")
}

func TestValidateCredsInput_ErrorOnMissingReqFields(t *testing.T) {
	mapWithEmptyCredEntry := make(map[string]readers.RegistryCredEntry)
	mapWithEmptyCredEntry["example.com"] = readers.RegistryCredEntry{}

	testCredsInput := readers.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: mapWithEmptyCredEntry,
	}

	err := validateCredsInput(testCredsInput)
	assert.Error(t, err, "Expected creds with empty entry to return error")
}

func TestValidateCredsInput_ErrorOnDuplicateContainers(t *testing.T) {
	duplicateContainer := "http"
	regCreds := make(map[string]readers.RegistryCredEntry)
	regCreds["registry-1.net"] = readers.RegistryCredEntry{
		SecretManagerARN: "arn:aws:secretsmanager:some-secret",
		ContainerNames:   []string{duplicateContainer, "logging"},
	}
	regCreds["registry-2.net"] = readers.RegistryCredEntry{
		SecretManagerARN: "arn:aws:secretsmanager:some-other-secret",
		ContainerNames:   []string{"metrics", duplicateContainer},
	}

	testCredsInput := readers.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: regCreds,
	}

	err := validateCredsInput(testCredsInput)
	assert.Error(t, err, "Expected creds with duplicate containers to return error")
}

func TestGenerateSecretString(t *testing.T) {
	type ECSRegistrySecret struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	testCases := []struct {
		inputUsername  string
		inputPassword  string
		expectedSecret ECSRegistrySecret
	}{
		{"user1", "l33tp4$$w0rd", ECSRegistrySecret{"user1", "l33tp4$$w0rd"}},
		{"someUserNameThatIsVeryLong0987654321", "*3G7nMl6W*Pi#*erjm", ECSRegistrySecret{"someUserNameThatIsVeryLong0987654321", "*3G7nMl6W*Pi#*erjm"}},
		{"myemail@example.com", "some-dashed-psswrd-64", ECSRegistrySecret{"myemail@example.com", "some-dashed-psswrd-64"}},
	}
	for _, test := range testCases {
		t.Run(fmt.Sprintf("Parse registry secret %s", test.inputUsername), func(t *testing.T) {

			actualSecretString := generateSecretString(test.inputUsername, test.inputPassword)
			assert.NotEmpty(t, *actualSecretString)

			regSecret := &ECSRegistrySecret{}
			err := json.Unmarshal([]byte(*actualSecretString), regSecret)
			assert.NoError(t, err, "Unexpected error when unmarshalling registry secret")
			assert.Equal(t, test.expectedSecret.Username, regSecret.Username, "Expected username to match")
			assert.Equal(t, test.expectedSecret.Password, regSecret.Password, "Expected password to match")
		})
	}
}

func getTestCredsEntry(secretARN, username, password, kmsKey string, containers []string) readers.RegistryCredEntry {
	return readers.RegistryCredEntry{
		SecretManagerARN: secretARN,
		Username:         username,
		Password:         password,
		KmsKeyID:         kmsKey,
		ContainerNames:   containers,
	}
}

func setupTestController(t *testing.T) *mock_secretsmanager.MockSMClient {
	ctrl := gomock.NewController(t)
	client := mock_secretsmanager.NewMockSMClient(ctrl)

	return client
}
