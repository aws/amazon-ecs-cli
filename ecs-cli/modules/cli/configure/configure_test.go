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

package configure

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	command "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	clusterName   = "defaultCluster"
	stackName     = "defaultCluster"
	profileName   = "defaultProfile"
	secondProfile = "alternate"
	region        = "us-west-1"
	awsAccessKey  = "AKID"
	awsSecretKey  = "SKID"
)

func createClusterConfig(name string) *cli.Context {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.RegionFlag, region, "")
	flags.String(command.ClusterFlag, name, "")
	flags.String(command.ConfigNameFlag, name, "")
	return cli.NewContext(nil, flags, nil)
}

func createProfileConfig(name string) *cli.Context {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.AccessKeyFlag, name, "")
	flags.String(command.SecretKeyFlag, awsSecretKey, "")
	flags.String(command.ProfileNameFlag, name, "")
	return cli.NewContext(nil, flags, nil)
}

func TestClusterConfigureAndSetDefault(t *testing.T) {
	// Config init when just cluster and region are specified
	config1 := createClusterConfig(profileName)
	config2 := createClusterConfig(secondProfile)
	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	ConfigureCluster(config1)
	ConfigureCluster(config2)
	ConfigureDefaultCluster(config2)

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, region, readConfig.Region, "Region mismatch in config.")
	assert.Equal(t, secondProfile, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")

}

func TestProfileConfigureAndSetDefault(t *testing.T) {
	// Config init when just cluster and region are specified
	config1 := createProfileConfig(profileName)
	config2 := createProfileConfig(secondProfile)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	ConfigureProfile(config1)
	ConfigureProfile(config2)
	ConfigureDefaultProfile(config2)

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, secondProfile, readConfig.AWSAccessKey, "Access Key mismatch in config.")
	assert.Equal(t, awsSecretKey, readConfig.AWSSecretKey, "Secret Key name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")

}
