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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/stretchr/testify/assert"
)

const (
	testClusterName   = "test-cluster"
	testAWSProfile    = "aws-profile"
	testRegion        = "us-west-2"
	testClusterConfig = "prod-config"
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

func saveClusterConfig(t *testing.T, parser *YAMLReadWriter, dest *Destination) {
	cluster := Cluster{Cluster: testClusterName, Region: testRegion}
	err := parser.SaveCluster(testClusterConfig, &cluster)
	assert.NoError(t, err, "Error saving mock config")
}

func TestConfigPermissions(t *testing.T) {
	dest, err := newMockDestination()

	assert.NoError(t, err, "Error creating mock config destination")

	parser := setupParser(t, dest, false)

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Create config file and confirm it has expected initial permissions
	saveClusterConfig(t, parser, dest)

	path := ConfigFilePath(dest)
	confirmConfigMode(t, path, configFileMode)

	// Now set the config mode to something bad
	badMode := os.FileMode(0777)
	err = os.Chmod(path, badMode)
	assert.NoError(t, err, "Unable to change mode of new config %v", path)

	confirmConfigMode(t, path, badMode)

	// Save the config and confirm it's fixed again
	cluster := &Cluster{Cluster: testClusterName}
	err = parser.SaveCluster("clusterConfig", cluster)
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

	err = ioutil.WriteFile(ConfigFilePath(dest), []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Reinitialize from the written file.
	parser := setupParser(t, dest, true)

	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, testClusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")
	assert.Equal(t, iniConfigVersion, readConfig.Version, "Expected ini config version to be set.")
	assert.Empty(t, readConfig.DefaultLaunchType, "Expected launch type to be empty.")
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
	config, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, flags.ComposeServiceNamePrefixDefaultValue, config.ComposeServiceNamePrefix, "ComposeServiceNamePrefix should be set to the default value.")
	assert.Equal(t, flags.CFNStackNamePrefixDefaultValue, config.CFNStackNamePrefix, "CFNStackNamePrefix should be set to the default value.")
	assert.Equal(t, iniConfigVersion, config.Version, "Expected ini config version to be set.")
	assert.Empty(t, config.DefaultLaunchType, "Expected launch type to be empty.")
}

func TestReadCredentialsFile(t *testing.T) {
	configContents := `default: Default
ecs_profiles:
  Default:
    aws_access_key_id: default_key_id
    aws_secret_access_key: default_key
  Alt:
    aws_access_key_id: alt_key_id
    aws_secret_access_key: alt_key
    aws_session_token: alt_token
`

	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Save the profile
	err = ioutil.WriteFile(dest.Path+"/"+profileConfigFileName, []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Read
	parser := setupParser(t, dest, false)

	// Test read the default profile
	config, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, "default_key_id", config.AWSAccessKey, "Access Key should be present.")
	assert.Equal(t, "default_key", config.AWSSecretKey, "Secret key should be present.")
	assert.Equal(t, yamlConfigVersion, config.Version, "Expected yaml config version to be set.")

	// Test read a specific profile
	config, err = parser.Get("", "Alt")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, "alt_key_id", config.AWSAccessKey, "Access Key should be present.")
	assert.Equal(t, "alt_key", config.AWSSecretKey, "Secret key should be present.")
	assert.Equal(t, "alt_token", config.AWSSessionToken, "Session token should be present.")
	assert.Equal(t, yamlConfigVersion, config.Version, "Expected yaml config version to be set.")
}

func TestReadClusterConfigFileNoLaunchType(t *testing.T) {
	configContents := `default: prod_config
clusters:
  gamma_config:
    cluster: cli-demo-gamma
    region: us-west-1
    compose-service-name-prefix: custom-service-
    cfn-stack-name: cfn-custom-cli-demo-gamma
  beta_config:
    cluster: cli-demo-beta
    region: us-west-2
  prod_config:
    cluster: cli-demo-prod
    region: us-east-2
`

	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Save the profile
	err = ioutil.WriteFile(dest.Path+"/"+clusterConfigFileName, []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Read
	parser := setupParser(t, dest, false)

	// Test read the default config
	config, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, "cli-demo-prod", config.Cluster, "Cluster should be present.")
	assert.Equal(t, "us-east-2", config.Region, "Region should be present.")
	assert.Equal(t, yamlConfigVersion, config.Version, "Expected yaml config version to be set.")
	assert.Empty(t, config.DefaultLaunchType, "Expected launch type to be empty.")

	// Test read a specific config
	config, err = parser.Get("gamma_config", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, "cli-demo-gamma", config.Cluster, "Cluster should be present.")
	assert.Equal(t, "us-west-1", config.Region, "Region should be present.")
	assert.Equal(t, "custom-service-", config.ComposeServiceNamePrefix, "ComposeServiceNamePrefix should be present.")
	assert.Equal(t, "cfn-custom-cli-demo-gamma", config.CFNStackName, "CFNStackName Name should be present.")
	assert.Empty(t, config.CFNStackNamePrefix, "Expected CFNStackNamePrefix to be empty.")
	assert.Empty(t, config.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty.")
	assert.Equal(t, yamlConfigVersion, config.Version, "Expected yaml config version to be set.")
	assert.Empty(t, config.DefaultLaunchType, "Expected launch type to be empty.")
}

func TestReadClusterConfigFileWithLaunchType(t *testing.T) {
	configContents := `default: prod_config
clusters:
  gamma_config:
    cluster: cli-demo-gamma
    region: us-west-1
    compose-service-name-prefix: custom-service-
    cfn-stack-name: cfn-custom-cli-demo-gamma
    default_launch_type: EC2
  beta_config:
    cluster: cli-demo-beta
    region: us-west-2
  prod_config:
    cluster: cli-demo-prod
    region: us-east-2
    default_launch_type: FARGATE
`

	dest, err := newMockDestination()
	assert.NoError(t, err, "Error creating mock config destination")

	err = os.MkdirAll(dest.Path, *dest.Mode)
	assert.NoError(t, err, "Could not create config directory")

	defer os.RemoveAll(dest.Path)

	// Save the profile
	err = ioutil.WriteFile(dest.Path+"/"+clusterConfigFileName, []byte(configContents), *dest.Mode)
	assert.NoError(t, err)

	// Read
	parser := setupParser(t, dest, false)

	// Test read the default config
	config, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, "cli-demo-prod", config.Cluster, "Cluster should be present.")
	assert.Equal(t, "us-east-2", config.Region, "Region should be present.")
	assert.Equal(t, yamlConfigVersion, config.Version, "Expected yaml config version to be set.")
	assert.Equal(t, LaunchTypeFargate, config.DefaultLaunchType)

	// Test read a specific config
	config, err = parser.Get("gamma_config", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, "cli-demo-gamma", config.Cluster, "Cluster should be present.")
	assert.Equal(t, "us-west-1", config.Region, "Region should be present.")
	assert.Equal(t, "custom-service-", config.ComposeServiceNamePrefix, "ComposeServiceNamePrefix should be present.")
	assert.Equal(t, "cfn-custom-cli-demo-gamma", config.CFNStackName, "CFNStackName Name should be present.")
	assert.Empty(t, config.CFNStackNamePrefix, "Expected CFNStackNamePrefix to be empty.")
	assert.Empty(t, config.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty.")
	assert.Equal(t, yamlConfigVersion, config.Version, "Expected yaml config version to be set.")
	assert.Equal(t, LaunchTypeEC2, config.DefaultLaunchType)
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
	saveClusterConfig(t, parser, dest)

	// Ensure that what has been read is correct
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, testClusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")
	assert.Equal(t, yamlConfigVersion, readConfig.Version, "Expected yaml config version to be set.")
	assert.Empty(t, readConfig.DefaultLaunchType, "Expected launch type to be empty.")
}
