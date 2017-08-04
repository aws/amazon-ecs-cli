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

	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/stretchr/testify/assert"
)

const (
	testClusterName = "test-cluster"
	testAWSProfile  = "aws-profile"
	testRegion      = "us-west-2"
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

func setupParser(t *testing.T, dest *Destination, shouldBeInitialized bool) *YAMLReadWriter {

	parser := &YAMLReadWriter{destination: dest}

	return parser
}

func saveConfigWithCluster(t *testing.T, parser *YAMLReadWriter, dest *Destination) {
	saveConfig(t, parser, dest, &CLIConfig{Cluster: testClusterName, ComposeProjectNamePrefix: "", ComposeServiceNamePrefix: "", CFNStackNamePrefix: ""})
}

func saveConfig(t *testing.T, parser *YAMLReadWriter, dest *Destination, newConfig *CLIConfig) {

	err := parser.Save(newConfig)
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

	path := configFilePath(dest)
	confirmConfigMode(t, path, configFileMode)

	// Now set the config mode to something bad
	badMode := os.FileMode(0777)
	err = os.Chmod(path, badMode)
	assert.NoError(t, err, "Unable to change mode of new config %v", path)

	confirmConfigMode(t, path, badMode)

	// Save the config and confirm it's fixed again
	cliConfig := &CLIConfig{Cluster: testClusterName}
	err = parser.Save(cliConfig)
	assert.NoError(t, err, "Unable to save to new config %v", path)

	confirmConfigMode(t, path, configFileMode)
}

func confirmConfigMode(t *testing.T, path string, expected os.FileMode) {
	info, err := os.Stat(path)
	assert.NoError(t, err, "Unable to stat config file %s", path)

	mode := info.Mode()
	assert.Equal(t, expected, mode, "Made of config does not match")

}

func TestPrefixesEmptyOldINIFormat(t *testing.T) {
	configContents := `[ecs]
cluster = test-cluster
aws_profile = testProfile
region = us-west-2
aws_access_key_id =
aws_secret_access_key =
compose-project-name-prefix =
compose-service-name-prefix =
cfn-stack-name-prefix =
`
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")
	defer os.RemoveAll(dest.Path)

	err = ioutil.WriteFile(configFilePath(dest), []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Reinitialize from the written file.
	parser := setupParser(t, dest, true)

	readConfig, err := parser.GetConfigs("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, testClusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")
}

func TestPrefixesDefaultOldINIFormat(t *testing.T) {
	configContentsNoPrefixes := `[ecs]
cluster = test
aws_profile =
region = us-west-2
aws_access_key_id =
aws_secret_access_key =
`
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	err = ioutil.WriteFile(dest.Path+"/"+iniConfigFileName, []byte(configContentsNoPrefixes), *dest.Mode)
	assert.NoError(t, err)

	parser := setupParser(t, dest, true)
	config, err := parser.GetConfigs("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, ecscli.ComposeProjectNamePrefixDefaultValue, config.ComposeProjectNamePrefix, "ComposeProjectNamePrefix should be set to the default value.")
	assert.Equal(t, ecscli.ComposeServiceNamePrefixDefaultValue, config.ComposeServiceNamePrefix, "ComposeServiceNamePrefix should be set to the default value.")
	assert.Equal(t, ecscli.CFNStackNamePrefixDefaultValue, config.CFNStackNamePrefix, "CFNStackNamePrefix should be set to the default value.")
}

func TestOverwriteINIConfigFile(t *testing.T) {
	configContents := `[ecs]
cluster = very-long-cluster-name
aws_profile = some-long-profile
region = us-west-2
aws_access_key_id =
aws_secret_access_key =
compose-project-name-prefix = ecscompose-
compose-service-name-prefix = ecscompose-service-
cfn-stack-name-prefix = amazon-ecs-cli-setup-
`

	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Save old ini config file
	err = ioutil.WriteFile(dest.Path+"/"+iniConfigFileName, []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Overwrite
	parser := setupParser(t, dest, false)
	saveConfigWithCluster(t, parser, dest)

	// Ensure that what has been read is correct
	readConfig, err := parser.GetConfigs("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, testClusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")

}
