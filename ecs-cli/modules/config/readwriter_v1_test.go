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

func setupParser(t *testing.T, dest *Destination, shouldBeInitialized bool) *YamlReadWriter {

	parser := &YamlReadWriter{destination: dest}

	return parser
}

func saveConfigWithCluster(t *testing.T, parser *YamlReadWriter, dest *Destination) {
	saveConfig(t, parser, dest, &CliConfig{Cluster: testClusterName, ComposeProjectNamePrefix: "", ComposeServiceNamePrefix: "", CFNStackNamePrefix: ""})
}

func saveConfig(t *testing.T, parser *YamlReadWriter, dest *Destination, newConfig *CliConfig) {

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

	path := yamlConfigPath(dest)
	confirmConfigMode(t, path, configFileMode)

	// Now set the config mode to something bad
	badMode := os.FileMode(0777)
	err = os.Chmod(path, badMode)
	assert.NoError(t, err, "Unable to change mode of new config %v", path)

	confirmConfigMode(t, path, badMode)

	// Save the config and confirm it's fixed again
	cliConfig := &CliConfig{Cluster: testClusterName}
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

func TestPrefixesEmptyNewYamlFormat(t *testing.T) {
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	parser := setupParser(t, dest, false)

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	saveConfigWithCluster(t, parser, dest)

	// Reinitialize from the written file.
	parser = setupParser(t, dest, true)

	readConfig, configMap, err := parser.GetConfig()
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, testClusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	_, ok := configMap[composeProjectNamePrefixKey]
	assert.True(t, ok, "Compose project prefix name should exist in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	_, ok = configMap[composeServiceNamePrefixKey]
	assert.True(t, ok, "Compose service name prefix should exist in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	_, ok = configMap[cfnStackNamePrefixKey]
	assert.True(t, ok, "CFNStackNamePrefix should exist in config.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")
}

func TestPrefixesEmptyOldIniFormat(t *testing.T) {
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

	err = ioutil.WriteFile(iniConfigPath(dest), []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Reinitialize from the written file.
	parser := setupParser(t, dest, true)

	readConfig, configMap, err := parser.GetConfig()
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, testClusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	_, ok := configMap[composeProjectNamePrefixKey]
	assert.True(t, ok, "Compose project prefix name should exist in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	_, ok = configMap[composeServiceNamePrefixKey]
	assert.True(t, ok, "Compose service name prefix should exist in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	_, ok = configMap[cfnStackNamePrefixKey]
	assert.True(t, ok, "CFNStackNamePrefix should exist in config.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")
}

func TestPrefixesDefaultOldIniFormat(t *testing.T) {
	configContents := `[ecs]
cluster = test
aws_profile =
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

	err = ioutil.WriteFile(iniConfigPath(dest), []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	parser := setupParser(t, dest, true)
	readConfig, configMap, err := parser.GetConfig()
	assert.NoError(t, err, "Error reading config")
	_, ok := configMap[cfnStackNamePrefixKey]
	assert.True(t, ok, "CFNStackNamePrefix should exist in config")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	_, ok = configMap[composeServiceNamePrefixKey]
	assert.True(t, ok, "Compose service name prefix should exist in config")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	_, ok = configMap[composeProjectNamePrefixKey]
	assert.True(t, ok, "Compose project name prefix should exist in config")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")
}

func TestMissingPrefixesOldIniFormat(t *testing.T) {
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
	_, configMap, err := parser.GetConfig()
	assert.NoError(t, err, "Error reading config")
	_, ok := configMap[cfnStackNamePrefixKey]
	assert.False(t, ok, "CFNStackNamePrefix should not exist in config")
	_, ok = configMap[composeServiceNamePrefixKey]
	assert.False(t, ok, "Compose service name prefix should not exist in config")
	_, ok = configMap[composeProjectNamePrefixKey]
	assert.False(t, ok, "Compose project name prefix should not exist in config")
}

func TestMissingPrefixesNewYamlFormat(t *testing.T) {
	configContentsNoPrefixes := `
version: v0
cluster: test
aws_profile:
region:us-west-2:
aws_access_key_id:
aws_secret_access_key:
`
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	err = ioutil.WriteFile(dest.Path+"/"+yamlConfigFileName, []byte(configContentsNoPrefixes), *dest.Mode)
	assert.NoError(t, err)

	parser := setupParser(t, dest, true)
	_, configMap, err := parser.GetConfig()
	assert.NoError(t, err, "Error reading config")
	_, ok := configMap[cfnStackNamePrefixKey]
	assert.False(t, ok, "CFNStackNamePrefix should not exist in config")
	_, ok = configMap[composeServiceNamePrefixKey]
	assert.False(t, ok, "Compose service name prefix should not exist in config")
	_, ok = configMap[composeProjectNamePrefixKey]
	assert.False(t, ok, "Compose project name prefix should not exist in config")
}

func TestPrefixesDefaultNewYamlFormat(t *testing.T) {
	configContents := `
version: v0
cluster: test
aws_profile:
region: us-west-2
aws_access_key_id:
aws_secret_access_key:
compose-project-name-prefix:
compose-service-name-prefix:
cfn-stack-name-prefix:
`
	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	err = ioutil.WriteFile(yamlConfigPath(dest), []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	parser := setupParser(t, dest, true)
	readConfig, configMap, err := parser.GetConfig()
	assert.NoError(t, err, "Error reading config")
	_, ok := configMap[cfnStackNamePrefixKey]
	assert.True(t, ok, "CFNStackNamePrefix should exist in config")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	_, ok = configMap[composeServiceNamePrefixKey]
	assert.True(t, ok, "Compose service name prefix should exist in config")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	_, ok = configMap[composeProjectNamePrefixKey]
	assert.True(t, ok, "Compose project name prefix should exist in config")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")
}

func TestConfigFileTruncation(t *testing.T) {
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
	configSmallerContents := `[ecs]
cluster = short-name
aws_profile = profile
region = us-west-2
aws_access_key_id =
aws_secret_access_key =
`

	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Save config for the first time
	err = ioutil.WriteFile(dest.Path+"/"+iniConfigFileName, []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Save config with shorter cluster name
	err = ioutil.WriteFile(dest.Path+"/"+iniConfigFileName, []byte(configSmallerContents), *dest.Mode)
	assert.NoError(t, err)

	_, err = newIniConfig(dest)
	assert.NoError(t, err)
}
