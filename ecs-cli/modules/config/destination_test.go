// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDefaultDestinationLinux(t *testing.T) {
	// Mock GetOSName in the test, then reset it after the test
	oldGetOSName := getOSName
	getOSName = func() string {
		return "linux"
	}
	defer func() { getOSName = oldGetOSName }()

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	defer os.Remove(tempDirName)

	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	dest, err := NewDefaultDestination()
	assert.NoError(t, err, "Unexpected error creating new config path")
	assert.Equal(t, filepath.Join(tempDirName, ".ecs"), dest.Path)
	assert.True(t, dest.Mode.IsDir(), "Expected user home directory to be in directory mode")
}

func TestNewDefaultDestinationDarwin(t *testing.T) {
	// Mock GetOSName in the test, then reset it after the test
	oldGetOSName := getOSName
	getOSName = func() string {
		return "darwin"
	}
	defer func() { getOSName = oldGetOSName }()

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	defer os.Remove(tempDirName)

	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	dest, err := NewDefaultDestination()
	assert.NoError(t, err, "Unexpected error creating new config path")
	assert.Equal(t, filepath.Join(tempDirName, ".ecs"), dest.Path)
	assert.True(t, dest.Mode.IsDir(), "Expected user home directory to be in directory mode")
}

func TestNewDefaultDestinationWindows(t *testing.T) {
	// Mock GetOSName in the test, then reset it after the test
	oldGetOSName := getOSName
	getOSName = func() string {
		return "windows"
	}
	defer func() { getOSName = oldGetOSName }()

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	defer os.Remove(tempDirName)

	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	dest, err := NewDefaultDestination()
	assert.NoError(t, err, "Unexpected error creating new config path")
	assert.Equal(t, filepath.Join(tempDirName, "AppData", "local", "ecs"), dest.Path)
	assert.True(t, dest.Mode.IsDir(), "Expected user home directory to be in directory mode")
}
