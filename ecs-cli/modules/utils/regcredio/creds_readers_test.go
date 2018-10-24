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

package regcredio

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCredsInput(t *testing.T) {
	credsInputString := `version: 1
registry_credentials:
  registry.io:
    username: some_user_name
    password: myl337p4$$w0rd!<bz*
    kms_key_id: aws:arn:kms:key/iuytre-jhgfd
    container_names:
      - nginx-custom
      - logging
  other-registry.net:
    secrets_manager_arn: aws:arn:secretsmanager:secret/repocreds-776ytg
    container_names:
      - metrics`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(credsInputString))
	assert.NoError(t, err, "Unexpected error writing file")
	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	credsResult, err := ReadCredsInput(tmpfile.Name())
	assert.NoError(t, err, "Unexpected error reading file")

	// assert expected values match
	assert.Equal(t, "1", credsResult.Version)
	assert.Equal(t, 2, len(credsResult.RegistryCredentials))

	firstRegResult := credsResult.RegistryCredentials["registry.io"]
	assert.NotEmpty(t, firstRegResult)
	assert.Equal(t, "some_user_name", firstRegResult.Username)
	assert.Equal(t, "myl337p4$$w0rd!<bz*", firstRegResult.Password)
	assert.Equal(t, "aws:arn:kms:key/iuytre-jhgfd", firstRegResult.KmsKeyID)
	assert.Equal(t, 2, len(firstRegResult.ContainerNames))

	otherRegResult := credsResult.RegistryCredentials["other-registry.net"]
	assert.NotEmpty(t, otherRegResult)
	assert.Equal(t, "aws:arn:secretsmanager:secret/repocreds-776ytg", otherRegResult.SecretManagerARN)
	assert.Equal(t, 1, len(otherRegResult.ContainerNames))
}

func TestReadCredsInputWithEnvVarsFromShell(t *testing.T) {
	// setup test env vars
	secretEnvKey := "MY_SECRET_ARN"
	secretEnvVal := "aws:arn:secretmanager:secret/regsecret-1"

	usrnameEnvKey := "MY_REG_USRNAME"
	usrnameEnvVal := "myname@example.net"

	passwrdEnvKey := "MY_REG_PASSWORD"
	passwrdEnvVal := "ne4t04905e867uyrdtoilfgkj"

	kmsEnvKey := "MY_KEY_ARN"
	kmsEnvVal := "aws:arn:kms:key/iuytre-yhe4"

	os.Setenv(usrnameEnvKey, usrnameEnvVal)
	os.Setenv(passwrdEnvKey, passwrdEnvVal)
	os.Setenv(kmsEnvKey, kmsEnvVal)
	os.Setenv(secretEnvKey, secretEnvVal)
	defer func() {
		os.Unsetenv(usrnameEnvKey)
		os.Unsetenv(passwrdEnvKey)
		os.Unsetenv(kmsEnvKey)
		os.Unsetenv(secretEnvKey)
	}()

	inputFileString := `version: 1
registry_credentials:
  myrepo.someregistry.io:
    secrets_manager_arn: ${MY_SECRET_ARN}
    username: ${MY_REG_USRNAME}
    password: ${MY_REG_PASSWORD}
    kms_key_id: ${MY_KEY_ARN}
    container_names:
      - test`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(inputFileString))
	assert.NoError(t, err, "Unexpected error writing file")
	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	credsResult, err := ReadCredsInput(tmpfile.Name())
	assert.NoError(t, err, "Unexpected error reading file")

	// assert expected values match
	assert.Equal(t, "1", credsResult.Version)
	assert.Equal(t, 1, len(credsResult.RegistryCredentials))

	credEntry := credsResult.RegistryCredentials["myrepo.someregistry.io"]
	assert.NotEmpty(t, credEntry)
	assert.Equal(t, usrnameEnvVal, credEntry.Username)
	assert.Equal(t, passwrdEnvVal, credEntry.Password)
	assert.Equal(t, kmsEnvVal, credEntry.KmsKeyID)
	assert.Equal(t, secretEnvVal, credEntry.SecretManagerARN)
	assert.Equal(t, 1, len(credEntry.ContainerNames))
}

func TestReadCredsInput_ErrorFileNotFound(t *testing.T) {
	var fakeFileName = "/missingFile"
	_, err := ReadCredsInput(fakeFileName)
	assert.Error(t, err, "Expected error on missing file")
}

func TestReadCredsInput_ErrorBadYaml(t *testing.T) {
	badCredEntryFileString := `version: 1
registry_credentials:
  myrepo.someregistry.io:
  secrets_manager_arn: arn:aws:secretmanager:some-secret
  container_names:
	  - test`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(badCredEntryFileString))
	assert.NoError(t, err, "Unexpected error writing file")
	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	_, err = ReadCredsInput(tmpfile.Name())
	assert.Error(t, err, "Expected error on bad file YAML")
}

func TestReadCredsOutput(t *testing.T) {
	credsOutputString := `version: "1"
registry_credential_outputs:
  task_execution_role: someTestRole
  container_credentials:
    my.example.registry.net:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net
      container_names:
      - web
    another.example.io:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-another.example.io
      kms_key_id: arn:aws:kms::key/some-key-57yrt
      container_names:
      - test`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(credsOutputString))
	assert.NoError(t, err, "Unexpected error writing file")
	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	credsOutput, err := ReadCredsOutput(tmpfile.Name())
	assert.NoError(t, err, "Unexpected error reading creds output file")

	// assert expected values match
	assert.Equal(t, "1", credsOutput.Version)
	assert.Equal(t, "someTestRole", credsOutput.CredentialResources.TaskExecutionRole)
	assert.Equal(t, 2, len(credsOutput.CredentialResources.ContainerCredentials))

	firstOutputEntry := credsOutput.CredentialResources.ContainerCredentials["my.example.registry.net"]
	assert.NotEmpty(t, firstOutputEntry)
	assert.Equal(t, "arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net", firstOutputEntry.CredentialARN)
	assert.Equal(t, "", firstOutputEntry.KMSKeyID)
	assert.ElementsMatch(t, []string{"web"}, firstOutputEntry.ContainerNames)

	secondOutputEntry := credsOutput.CredentialResources.ContainerCredentials["another.example.io"]
	assert.NotEmpty(t, secondOutputEntry)
	assert.Equal(t, "arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-another.example.io", secondOutputEntry.CredentialARN)
	assert.Equal(t, "arn:aws:kms::key/some-key-57yrt", secondOutputEntry.KMSKeyID)
	assert.ElementsMatch(t, []string{"test"}, secondOutputEntry.ContainerNames)
}

func TestReadCredsOutput_ErrorOnBadYaml(t *testing.T) {
	badCredsOutputFileString := `version: 1
registry_credential_outputs:
task_execution_role: someTestRole
container_credentials:
	myrepo.someregistry.io:
      credentials_parameter: arn:aws:secretmanager:some-secret
	  container_names:
		  - test`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(badCredsOutputFileString))
	assert.NoError(t, err, "Unexpected error writing file")
	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	_, err = ReadCredsOutput(tmpfile.Name())
	assert.Error(t, err, "Expected error on bad file YAML")
}

func TestReadCredsOutput_ErrorFileNotFound(t *testing.T) {
	var fakeFileName = "/missingFile"
	_, err := ReadCredsOutput(fakeFileName)
	assert.Error(t, err, "Expected error on missing file")
}

func TestFindLatestRegCredsOutputFile(t *testing.T) {
	testCases := []struct {
		description    string
		inputFileNames []string
		expectedLatest string
	}{
		{"Find latest of 3 valid output files", []string{"ecs-registry-creds_20171117T125102Z.yml", "ecs-registry-creds_20181012T215145Z.yml", "ecs-registry-creds_20181017T125102Z.yml"}, "ecs-registry-creds_20181017T125102Z.yml"},
		{"Find latest valid file out of mixed valid/invalid output files", []string{"ecs-registry-creds_3.yml", "ecs-registry-creds_20181013T125105Z.yml"}, "ecs-registry-creds_20181013T125105Z.yml"},
		{"Return no file if no valid file found", []string{"ecs-registry-creds_3.yml", "ecs-registry-creds_TEST.yml"}, ""},
		{"Return no file if no files found", []string{}, ""},
	}
	for _, test := range testCases {
		t.Run(test.description, func(t *testing.T) {
			// setup test dir & file
			testOutputDir, err := ioutil.TempDir("", "test")
			assert.NoError(t, err, "Unexpected error creating temp directory")

			defer os.RemoveAll(testOutputDir)

			for _, fileName := range test.inputFileNames {
				err := ioutil.WriteFile(testOutputDir+string(os.PathSeparator)+fileName, []byte("creds go here"), os.ModeTemporary)
				assert.NoError(t, err, "Unexpected error creating test file")
			}

			expectedLatestFile := ""
			if test.expectedLatest != "" {
				expectedLatestFile = testOutputDir + string(os.PathSeparator) + test.expectedLatest
			}

			actualLatestFile, err := FindLatestRegCredsOutputFile(testOutputDir)
			assert.NoError(t, err, "Unexpected error finding latest creds file")
			assert.Equal(t, expectedLatestFile, actualLatestFile)
		})
	}
}
