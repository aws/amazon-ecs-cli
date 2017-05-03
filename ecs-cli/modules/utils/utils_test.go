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

package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHomeDir(t *testing.T) {
	tempDirName := tempDir(t)
	defer os.Remove(tempDirName)
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	_, err := GetHomeDir()
	assert.NoError(t, err, "Unexpected error getting home dir")
}

func TestGetHomeDirInWindows(t *testing.T) {
	tempDirName := tempDir(t)
	defer os.Remove(tempDirName)
	os.Setenv("USERPROFILE", tempDirName)
	defer os.Unsetenv("USERPROFILE")

	_, err := GetHomeDir()
	assert.NoError(t, err, "Unexpected error getting home dir")
}

func tempDir(t *testing.T) string {
	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	assert.NoError(t, err, "Unexpected error while creating the dummy ecs config directory")
	return tempDirName
}
