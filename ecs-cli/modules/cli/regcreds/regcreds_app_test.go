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
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcreds"
	"github.com/stretchr/testify/assert"
)

func TestRegCregsUp(t *testing.T) {
	testFileString := `version: 1
registry_credentials:
  myrepo.someregistry.io:
    username: some_user_name
    password: myl337p4$$w0rd!<bz*
    container_names:
      - test`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(testFileString))
	assert.NoError(t, err, "Unexpected error writing file")
	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	//todo: add client and call mocks when added
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
