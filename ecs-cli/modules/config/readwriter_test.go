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
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testClusterName = "test-cluster"
	testAWSProfile  = "aws-profile"
)

func newMockDestination() (*Destination, error) {
	tmpPath, err := ioutil.TempDir(os.TempDir(), "ecs-cli-test-")
	if err != nil {
		return nil, err
	}

	mode, err := GetFilePermissions(tmpPath)
	if err != nil {
		return nil, err
	}

	return &Destination{Path: tmpPath, Mode: mode}, nil
}

func setupParser(t *testing.T, dest *Destination, shouldBeInitialized bool) *IniReadWriter {
	iniCfg, err := newIniConfig(dest)
	assert.NoError(t, err, "Error creating config ini")

	parser := &IniReadWriter{Destination: dest, cfg: iniCfg}

	// Test when unitialized.
	initialized, err := parser.IsInitialized()
	assert.NoError(t, err, "Error getting if initialized from ini")
	assert.Equal(t, shouldBeInitialized, initialized, "Unexpected state during parser initialization.")

	return parser
}

func saveConfigWithCluster(t *testing.T, parser *IniReadWriter, dest *Destination) {
	saveConfig(t, parser, dest, &SectionKeys{Cluster: testClusterName})
}

func saveConfig(t *testing.T, parser *IniReadWriter, dest *Destination, sectionKeys *SectionKeys) {
	// Create a new config file
	newConfig := &CliConfig{sectionKeys}
	err := parser.ReadFrom(newConfig)
	assert.NoError(t, err, "Could not create config from struct")

	err = parser.Save(dest)
	assert.NoError(t, err, "Could not save config file")
}

func TestConfigPermissions(t *testing.T) {
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	parser := setupParser(t, dest, false)

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Create config file and confirm it has expected initial permissions
	saveConfigWithCluster(t, parser, dest)

	path := configPath(dest)
	confirmConfigMode(t, path, configFileMode)

	// Now set the config mode to something bad
	badMode := os.FileMode(0777)
	err = os.Chmod(path, badMode)
	assert.NoError(t, err, "Unable to change mode of new config %v", path)

	confirmConfigMode(t, path, badMode)

	// Save the config and confirm it's fixed again
	err = parser.Save(dest)
	assert.NoError(t, err, "Unable to save to new config %v", path)

	confirmConfigMode(t, path, configFileMode)
}

func confirmConfigMode(t *testing.T, path string, expected os.FileMode) {
	info, err := os.Stat(path)
	assert.NoError(t, err, "Unable to stat config file %s", path)

	mode := info.Mode()
	assert.Equal(t, expected, mode, "Made of config does not match")

}

func TestNewConfigReadWriter(t *testing.T) {
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	parser := setupParser(t, dest, false)

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	saveConfigWithCluster(t, parser, dest)

	// Reinitialize from the written file.
	parser = setupParser(t, dest, true)

	readConfig, err := parser.GetConfig()
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, testClusterName, readConfig.Cluster, "Cluster name mismatch in config.")
}

func TestConfigFileTruncation(t *testing.T) {
	configContents := `[ecs]
cluster = very-long-cluster-name
aws_profile = some-long-profile
region = us-west-2
aws_access_key_id =
aws_secret_access_key =
`
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	parser := setupParser(t, dest, false)

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Save config for the first time
	err = ioutil.WriteFile(dest.Path+"/"+configFileName, []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Save config with shorter cluster name
	saveConfigWithCluster(t, parser, dest)

	_, err = newIniConfig(dest)
	assert.NoError(t, err)
}
