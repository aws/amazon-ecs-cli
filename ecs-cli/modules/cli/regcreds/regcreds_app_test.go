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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/iam/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/kms/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/secretsmanager/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcredio"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
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
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		testRegistryName: getTestCredsEntry("", testUsername, testPassword, "", testContainers),
	}

	expectedCreateInput := secretsmanager.CreateSecretInput{
		Name:         generateECSResourceName(testRegistryName),
		SecretString: generateSecretString(testUsername, testPassword),
		Description:  generateSecretDescription(testRegistryName),
	}
	responseARN := "arn:aws:secretsmanager:examplereg.net-123"

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockSM.EXPECT().DescribeSecret(gomock.Any()).Return(nil, nil),
		mocks.MockSM.EXPECT().CreateSecret(expectedCreateInput).Return(&secretsmanager.CreateSecretOutput{ARN: aws.String(responseARN)}, nil),
	)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, false)
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
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		testRegistryName: getTestCredsEntry("", testUsername, testPassword, testKmsKeyID, testContainers),
	}

	expectedCreateInput := secretsmanager.CreateSecretInput{
		Name:         generateECSResourceName(testRegistryName),
		SecretString: generateSecretString(testUsername, testPassword),
		KmsKeyId:     aws.String(testKmsKeyID),
		Description:  generateSecretDescription(testRegistryName),
	}
	responseARN := "arn:aws:secretsmanager:examplereg.net-123"

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockSM.EXPECT().DescribeSecret(gomock.Any()).Return(nil, nil),
		mocks.MockSM.EXPECT().CreateSecret(expectedCreateInput).Return(&secretsmanager.CreateSecretOutput{ARN: aws.String(responseARN)}, nil),
	)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, false)
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
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		testRegistryName: getTestCredsEntry("", testUsername, testPassword, "", testContainers),
	}

	responseARN := "arn:aws:secretsmanager:examplereg.net-123"

	mocks := setupTestController(t)
	mocks.MockSM.EXPECT().DescribeSecret(gomock.Any()).Return(&secretsmanager.DescribeSecretOutput{ARN: aws.String(responseARN)}, nil)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, false)
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
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		testRegistryName: getTestCredsEntry(testSecretARN, "", "", "", testContainers),
	}

	mocks := setupTestController(t)
	mocks.MockSM.EXPECT().DescribeSecret(gomock.Any()).Return(&secretsmanager.DescribeSecretOutput{}, nil)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, false)
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
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		testRegistryName: getTestCredsEntry(testSecretARN, testUsername, testPassword, "", testContainers),
	}

	expectedPutSecretValueInput := secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(testSecretARN),
		SecretString: generateSecretString(testUsername, testPassword),
	}

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockSM.EXPECT().PutSecretValue(expectedPutSecretValueInput).Return(&secretsmanager.PutSecretValueOutput{}, nil),
		mocks.MockSM.EXPECT().DescribeSecret(gomock.Any()).Return(&secretsmanager.DescribeSecretOutput{}, nil),
	)

	// call with updateAllowed = true
	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, true)
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
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		testRegistryName: getTestCredsEntry(testSecretARN, testUsername, testPassword, "", testContainers),
	}

	// call with updateAllowed = false
	mocks := setupTestController(t)
	mocks.MockSM.EXPECT().DescribeSecret(gomock.Any()).Return(&secretsmanager.DescribeSecretOutput{}, nil)

	credsOutput, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, false)
	assert.NoError(t, err, "Expected no error when using existing secren ARN")

	actualCredEntry := credsOutput[testRegistryName]
	assert.NotEmpty(t, actualCredEntry)
	assert.ElementsMatch(t, testContainers, actualCredEntry.ContainerNames)
}

func TestGetOrCreateRegistryCredentials_ErrorOnCreate(t *testing.T) {
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		"testRegistry": getTestCredsEntry("", "testUsername", "testPassword", "", []string{"test"}),
	}

	mocks := setupTestController(t)
	gomock.InOrder(
		mocks.MockSM.EXPECT().DescribeSecret(gomock.Any()).Return(nil, nil),
		mocks.MockSM.EXPECT().CreateSecret(gomock.Any()).Return(nil, errors.New("something went wrong")),
	)

	_, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, false)
	assert.Error(t, err)
}

func TestGetOrCreateRegistryCredentials_ErrorOnUpdate(t *testing.T) {
	testRegistryCreds := map[string]regcredio.RegistryCredEntry{
		"testRegistry": getTestCredsEntry("arn:aws:secretsmanager:secret:test", "testUsername", "testPassword", "", []string{"test"}),
	}

	mocks := setupTestController(t)
	mocks.MockSM.EXPECT().PutSecretValue(gomock.Any()).Return(nil, errors.New("something went wrong"))

	_, err := getOrCreateRegistryCredentials(testRegistryCreds, mocks.MockSM, true)
	assert.Error(t, err)
}

func TestValidateCredsInput_ErrorEmptyCreds(t *testing.T) {
	emptyCredMap := make(map[string]regcredio.RegistryCredEntry)
	emptyCredsInput := regcredio.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: emptyCredMap,
	}

	_, err := validateCredsInput(emptyCredsInput, nil)
	assert.Error(t, err, "Expected empty creds to return error")
}

func TestValidateCredsInput_ErrorOnMissingReqFields(t *testing.T) {
	mapWithEmptyCredEntry := map[string]regcredio.RegistryCredEntry{
		"example.com": regcredio.RegistryCredEntry{},
	}

	testCredsInput := regcredio.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: mapWithEmptyCredEntry,
	}

	_, err := validateCredsInput(testCredsInput, nil)
	assert.Error(t, err, "Expected creds with empty entry to return error")
}

func TestValidateCredsInput_ErrorOnDuplicateContainers(t *testing.T) {
	duplicateContainer := "http"
	regCreds := map[string]regcredio.RegistryCredEntry{
		"registry-1.net": regcredio.RegistryCredEntry{
			SecretManagerARN: "arn:aws:secretsmanager:some-secret",
			ContainerNames:   []string{duplicateContainer, "logging"},
		},
		"registry-2.net": regcredio.RegistryCredEntry{
			SecretManagerARN: "arn:aws:secretsmanager:some-other-secret",
			ContainerNames:   []string{"metrics", duplicateContainer},
		},
	}

	testCredsInput := regcredio.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: regCreds,
	}

	_, err := validateCredsInput(testCredsInput, nil)
	assert.Error(t, err, "Expected creds with duplicate containers to return error")
}

func TestValidateCredsInput_KeyAliasDescribed(t *testing.T) {
	mocks := setupTestController(t)
	testRegName := "testRegistry"
	regCreds := map[string]regcredio.RegistryCredEntry{
		testRegName: getTestCredsEntry("", "testuser", "testPassword", "alias/someKey", []string{"test"}),
	}

	testCredsInput := regcredio.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: regCreds,
	}

	expectedKeyARN := "arn:aws:kms:key/56yrtgf-4etrfgd-34erfd"
	expectKeyMetadata := kms.KeyMetadata{
		Arn: aws.String(expectedKeyARN),
	}

	gomock.InOrder(
		mocks.MockKMS.EXPECT().GetValidKeyARN("alias/someKey").Return(expectedKeyARN, nil),
		mocks.MockKMS.EXPECT().DescribeKey("alias/someKey").Return(&kms.DescribeKeyOutput{KeyMetadata: &expectKeyMetadata}, nil),
	)

	validatedOutput, err := validateCredsInput(testCredsInput, mocks.MockKMS)
	assert.NoError(t, err, "Unexpected error on Describe Key")
	assert.Equal(t, expectedKeyARN, validatedOutput[testRegName].KmsKeyID)
}

func TestValidateCredsInput_NoDescribeOnKeyARN(t *testing.T) {
	mocks := setupTestController(t)
	testKeyARN := "arn:aws:kms:key/7457r6ythfg-5rythgf"
	regCreds := map[string]regcredio.RegistryCredEntry{
		"testRegistry": getTestCredsEntry("", "testuser", "testPassword", testKeyARN, []string{"test"}),
	}

	testCredsInput := regcredio.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: regCreds,
	}

	mocks.MockKMS.EXPECT().GetValidKeyARN(testKeyARN).Return(testKeyARN, nil)

	validatedOutput, err := validateCredsInput(testCredsInput, mocks.MockKMS)
	assert.NoError(t, err, "Unexpected error when validating reg creds")
	assert.Equal(t, testKeyARN, validatedOutput["testRegistry"].KmsKeyID)
}

func TestValidateCredsInput_ErrorOnDescribeFail(t *testing.T) {
	mocks := setupTestController(t)
	regCreds := map[string]regcredio.RegistryCredEntry{
		"testRegistry": getTestCredsEntry("", "testuser", "testPassword", "alias/someKey", []string{"test"}),
	}
	testCredsInput := regcredio.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: regCreds,
	}

	gomock.InOrder(
		mocks.MockKMS.EXPECT().GetValidKeyARN(gomock.Any()).Return("", errors.New("something went wrong")),
		mocks.MockKMS.EXPECT().DescribeKey(gomock.Any()).Return(nil, errors.New("something went wrong")),
	)

	_, err := validateCredsInput(testCredsInput, mocks.MockKMS)
	assert.Error(t, err, "Expected error when Describe Key fails")
}

func TestValidateCredsInput_ErrorOnRegionMismatch(t *testing.T) {
	testKeyARN := "arn:aws:kms:us-east-1:1234567:key/765ythgf-45erfd"
	regCreds := map[string]regcredio.RegistryCredEntry{
		"testRegistry": getTestCredsEntry("arn:aws:secretsmanager:us-west-2:1234567:secret/some-secret", "testuser", "testPassword", testKeyARN, []string{"test"}),
	}
	testCredsInput := regcredio.ECSRegCredsInput{
		Version:             "1",
		RegistryCredentials: regCreds,
	}

	mocks := setupTestController(t)
	mocks.MockKMS.EXPECT().GetValidKeyARN(testKeyARN).Return(testKeyARN, nil)

	_, err := validateCredsInput(testCredsInput, mocks.MockKMS)
	assert.Error(t, err, "Expected error when secret and key regions don't match")
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

func getTestCredsEntry(secretARN, username, password, kmsKey string, containers []string) regcredio.RegistryCredEntry {
	return regcredio.RegistryCredEntry{
		SecretManagerARN: secretARN,
		Username:         username,
		Password:         password,
		KmsKeyID:         kmsKey,
		ContainerNames:   containers,
	}
}

type testClients struct {
	MockIAM *mock_iam.MockClient
	MockKMS *mock_kms.MockClient
	MockSM  *mock_secretsmanager.MockSMClient
}

func setupTestController(t *testing.T) testClients {
	ctrl := gomock.NewController(t)

	clients := testClients{
		MockIAM: mock_iam.NewMockClient(ctrl),
		MockKMS: mock_kms.NewMockClient(ctrl),
		MockSM:  mock_secretsmanager.NewMockSMClient(ctrl),
	}

	return clients
}
