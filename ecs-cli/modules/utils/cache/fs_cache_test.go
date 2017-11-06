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

package cache

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheCreatesDir(t *testing.T) {
	tempDirName := tempDir(t)
	defer os.Remove(tempDirName)
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	created := make(chan string)
	osMkdirAll = func(path string, perms os.FileMode) error {
		if perms != 0700 {
			t.Errorf("directory created with more open perms than expected; %v", perms)
		}
		go func() { created <- path }()
		return nil
	}
	NewFSCache("foo")
	dir := <-created

	expected, _ := cacheDir("foo")
	if dir != expected {
		t.Errorf("expected %v, got %v", expected, dir)
	}
}

func TestCacheDirLinux(t *testing.T) {
	// Mock GetOSName in the test, then reset it after the test
	oldGetOSName := getOSName
	getOSName = func() string {
		return "linux"
	}
	defer func() { getOSName = oldGetOSName }()

	// Create a temporary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	defer os.Remove(tempDirName)

	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	path, err := cacheDir("name")
	assert.NoError(t, err, "Unexpected error creating new config path")
	assert.Equal(t, filepath.Join(tempDirName, ".cache", cachePrefix, "name"), path)
}

func TestCacheDirDarwin(t *testing.T) {
	// Mock GetOSName in the test, then reset it after the test
	oldGetOSName := getOSName
	getOSName = func() string {
		return "darwin"
	}
	defer func() { getOSName = oldGetOSName }()

	// Create a temporary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	defer os.Remove(tempDirName)

	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	path, err := cacheDir("name")
	assert.NoError(t, err, "Unexpected error creating new config path")
	assert.Equal(t, filepath.Join(tempDirName, ".cache", cachePrefix, "name"), path)
}

func TestCacheDirWindows(t *testing.T) {
	// Mock GetOSName in the test, then reset it after the test
	oldGetOSName := getOSName
	getOSName = func() string {
		return "windows"
	}
	defer func() { getOSName = oldGetOSName }()

	// Create a temporary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	defer os.Remove(tempDirName)

	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	path, err := cacheDir("name")
	assert.NoError(t, err, "Unexpected error creating new config path")
	assert.Equal(t, filepath.Join(tempDirName, "AppData", "local", "ecs", "cache", "name"), path)
}

func tempDir(t *testing.T) string {
	// Create a temporary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	assert.NoError(t, err, "Unexpected error while creating the dummy ecs config directory")
	return tempDirName
}
