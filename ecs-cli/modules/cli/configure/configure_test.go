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
	secondCluster = "alternateCluster"
	stackName     = "defaultCluster"
	profileName   = "defaultProfile"
	profileName2  = "alternate"
	region        = "us-west-1"
	awsAccessKey  = "AKID"
	awsAccessKey2 = "AKID2"
	awsSecretKey  = "SKID"
	awsSecretKey2 = "SKID2"
)

func createClusterConfig(name string, cluster string) *cli.Context {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.RegionFlag, region, "")
	flags.String(command.ClusterFlag, cluster, "")
	flags.String(command.ConfigNameFlag, name, "")
	return cli.NewContext(nil, flags, nil)
}

func createProfileConfig(name string, accessKey string, secretKey string) *cli.Context {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.AccessKeyFlag, accessKey, "")
	flags.String(command.SecretKeyFlag, secretKey, "")
	flags.String(command.ProfileNameFlag, name, "")
	return cli.NewContext(nil, flags, nil)
}

func TestDefaultCluster(t *testing.T) {
	config1 := createClusterConfig(profileName, clusterName)
	config2 := createClusterConfig(profileName2, secondCluster)
	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	err = Cluster(config1)
	assert.NoError(t, err, "Unexpected error configuring cluster")
	err = Cluster(config2)
	assert.NoError(t, err, "Unexpected error configuring cluster")
	err = DefaultCluster(config2)
	assert.NoError(t, err, "Unexpected error configuring cluster")

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, region, readConfig.Region, "Region mismatch in config.")
	assert.Equal(t, secondCluster, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")

}

func TestDefaultProfile(t *testing.T) {
	config1 := createProfileConfig(profileName, awsAccessKey, awsSecretKey)
	config2 := createProfileConfig(profileName2, awsAccessKey2, awsSecretKey2)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	err = Profile(config1)
	assert.NoError(t, err, "Unexpected error configuring profile")
	err = Profile(config2)
	assert.NoError(t, err, "Unexpected error configuring profile")
	err = DefaultProfile(config2)
	assert.NoError(t, err, "Unexpected error configuring profile")

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, awsAccessKey2, readConfig.AWSAccessKey, "Access Key mismatch in config.")
	assert.Equal(t, awsSecretKey2, readConfig.AWSSecretKey, "Secret Key name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")

}

func TestConfigureProfile(t *testing.T) {
	config1 := createProfileConfig(profileName, awsAccessKey, awsSecretKey)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	err = Profile(config1)
	assert.NoError(t, err, "Unexpected error configuring profile")

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, awsAccessKey, readConfig.AWSAccessKey, "Access Key mismatch in config.")
	assert.Equal(t, awsSecretKey, readConfig.AWSSecretKey, "Secret Key name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")

}

func TestConfigureCluster(t *testing.T) {
	config1 := createClusterConfig(profileName, clusterName)
	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	err = Cluster(config1)
	assert.NoError(t, err, "Unexpected error configuring cluster")

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, region, readConfig.Region, "Region mismatch in config.")
	assert.Equal(t, clusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeProjectNamePrefix, "Compose project prefix name should be empty.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackNamePrefix, "CFNStackNamePrefix should be empty.")

}

func TestConfigureClusterNoCluster(t *testing.T) {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.RegionFlag, region, "")
	flags.String(command.ConfigNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flags, nil)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	err = Cluster(config1)
	assert.Error(t, err, "Expected error configuring cluster.")

}

func TestConfigureClusterNoRegion(t *testing.T) {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.ClusterFlag, clusterName, "")
	flags.String(command.ConfigNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flags, nil)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	err = Cluster(config1)
	assert.Error(t, err, "Expected error configuring cluster.")

}

func TestConfigureClusterNoConfigName(t *testing.T) {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.ClusterFlag, clusterName, "")
	flags.String(command.RegionFlag, region, "")
	config1 := cli.NewContext(nil, flags, nil)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	err = Cluster(config1)
	assert.Error(t, err, "Expected error configuring cluster.")
}

func TestConfigureProfileNoAccessKey(t *testing.T) {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.SecretKeyFlag, awsSecretKey, "")
	flags.String(command.ProfileNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flags, nil)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	err = Profile(config1)
	assert.Error(t, err, "Expected error configuring profile")

}

func TestConfigureProfileNoSecretKey(t *testing.T) {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.AccessKeyFlag, awsAccessKey, "")
	flags.String(command.ProfileNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flags, nil)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	err = Profile(config1)
	assert.Error(t, err, "Expected error configuring profile")

}

func TestConfigureProfileNoProfileName(t *testing.T) {
	flags := flag.NewFlagSet("ecs-cli", 0)
	flags.String(command.AccessKeyFlag, awsAccessKey, "")
	flags.String(command.SecretKeyFlag, awsSecretKey, "")
	config1 := cli.NewContext(nil, flags, nil)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	err = Profile(config1)
	assert.Error(t, err, "Expected error configuring profile")

}

func TestDefaultClusterDoesNotExist(t *testing.T) {
	config1 := createClusterConfig(profileName, clusterName)
	config2 := createClusterConfig(profileName2, secondCluster)
	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	err = Cluster(config1)
	assert.NoError(t, err, "Unexpected error configuring cluster")
	err = DefaultCluster(config2)
	assert.Error(t, err, "Expected error configuring cluster")
}

func TestDefaultProfileDoesNotExist(t *testing.T) {
	config1 := createProfileConfig(profileName, awsAccessKey, awsSecretKey)
	config2 := createProfileConfig(profileName2, awsAccessKey2, awsSecretKey2)

	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// configure 2 profiles and set one as default
	err = Profile(config1)
	assert.NoError(t, err, "Unexpected error configuring profile")
	err = DefaultProfile(config2)
	assert.Error(t, err, "Expected error configuring profile")

}
