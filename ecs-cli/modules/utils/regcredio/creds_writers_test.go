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
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateCredsOutput(t *testing.T) {
	// setup test resources
	testOutputDir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err, "Unexpected error creating temp directory")

	defer os.RemoveAll(testOutputDir)

	testRoleName := "myTestCredsRole"
	testCreds := make(map[string]CredsOutputEntry)
	testReg1 := "my.example.net"
	testReg2 := "example.io"

	testCreds[testReg1] = BuildOutputEntry("arn:aws:secretsmanager:secret/test", "", []string{"web", "test"})
	testCreds[testReg2] = BuildOutputEntry("arn:aws:secretsmanager:secret/testTwo", "arn:aws:kms:key/test-546yrtgf", []string{"metrics"})

	// generate file
	currTime := time.Now().UTC()
	err = GenerateCredsOutput(testCreds, testRoleName, testOutputDir, &currTime)
	assert.NoError(t, err, "Unexpected error when generating creds output")

	// assert output file was produced, is legible
	outputFile, err := filepath.Glob(testOutputDir + string(os.PathSeparator) + ECSCredFileBaseName + "*.yml")
	assert.NoError(t, err, "Unexpected error finding output file")

	actualCredsOutput, err := ReadCredsOutput(outputFile[0])
	assert.NoError(t, err, "Unexpected error parsing output file")
	assert.Equal(t, actualCredsOutput.Version, "1")

	actualRegCreds := actualCredsOutput.CredentialResources
	assert.NotEmpty(t, actualRegCreds)
	assert.Equal(t, testCreds[testReg1], actualRegCreds.ContainerCredentials[testReg1])
	assert.Equal(t, testCreds[testReg2], actualRegCreds.ContainerCredentials[testReg2])
	assert.Equal(t, actualRegCreds.TaskExecutionRole, testRoleName)
}
