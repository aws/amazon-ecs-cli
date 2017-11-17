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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	clusterName              = "defaultCluster"
	secondCluster            = "alternateCluster"
	stackName                = "defaultCluster"
	profileName              = "defaultProfile"
	profileName2             = "alternate"
	region                   = "us-west-1"
	awsAccessKey             = "AKID"
	awsAccessKey2            = "AKID2"
	awsSecretKey             = "SKID"
	awsSecretKey2            = "SKID2"
	awsProfile               = "awsprofile"
	composeServiceNamePrefix = "ecs-"
	cfnStackNamePrefix       = "cfn-"
	composeProjectNamePrefix = "ecs-compose-"
)

func createClusterConfig(name string, cluster string, launchType string) *cli.Context {
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.RegionFlag, region, "")
	flagSet.String(flags.ClusterFlag, cluster, "")
	flagSet.String(flags.ConfigNameFlag, name, "")
	flagSet.String(flags.DefaultLaunchTypeFlag, launchType, "")
	return cli.NewContext(nil, flagSet, nil)
}

func createProfileConfig(name string, accessKey string, secretKey string) *cli.Context {
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.AccessKeyFlag, accessKey, "")
	flagSet.String(flags.SecretKeyFlag, secretKey, "")
	flagSet.String(flags.ProfileNameFlag, name, "")
	return cli.NewContext(nil, flagSet, nil)
}

func TestDefaultCluster(t *testing.T) {
	config1 := createClusterConfig(profileName, clusterName, config.LaunchTypeEC2)
	config2 := createClusterConfig(profileName2, secondCluster, config.LaunchTypeFargate)
	// Create a temporary directory for the dummy ecs config
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
	assert.Equal(t, config.LaunchTypeFargate, readConfig.DefaultLaunchType, "Launch Type mismatch in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")

}

func TestDefaultProfile(t *testing.T) {
	config1 := createProfileConfig(profileName, awsAccessKey, awsSecretKey)
	config2 := createProfileConfig(profileName2, awsAccessKey2, awsSecretKey2)

	// Create a temporary directory for the dummy ecs config
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
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")

}

func TestConfigureProfile(t *testing.T) {
	config1 := createProfileConfig(profileName, awsAccessKey, awsSecretKey)

	// Create a temporary directory for the dummy ecs config
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
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")

}

func TestConfigureClusterNoLaunchType(t *testing.T) {
	config1 := createClusterConfig(profileName, clusterName, "")
	// Create a temporary directory for the dummy ecs config
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
	assert.Empty(t, readConfig.DefaultLaunchType, "Launch Type expected to be empty")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")

}

func TestConfigureClusterFargate(t *testing.T) {
	config1 := createClusterConfig(profileName, clusterName, config.LaunchTypeFargate)
	// Create a temporary directory for the dummy ecs config
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
	assert.Equal(t, config.LaunchTypeFargate, readConfig.DefaultLaunchType, "Launch Type mismatch in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")

}

func TestConfigureClusterEC2(t *testing.T) {
	config1 := createClusterConfig(profileName, clusterName, config.LaunchTypeEC2)
	// Create a temporary directory for the dummy ecs config
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
	assert.Equal(t, config.LaunchTypeEC2, readConfig.DefaultLaunchType, "Launch Type mismatch in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")

}

func TestConfigureClusterNoCluster(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.RegionFlag, region, "")
	flagSet.String(flags.ConfigNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flagSet, nil)

	// Create a temporary directory for the dummy ecs config
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
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.ClusterFlag, clusterName, "")
	flagSet.String(flags.ConfigNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flagSet, nil)

	// Create a temporary directory for the dummy ecs config
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
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.SecretKeyFlag, awsSecretKey, "")
	flagSet.String(flags.ProfileNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flagSet, nil)

	// Create a temporary directory for the dummy ecs config
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
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.AccessKeyFlag, awsAccessKey, "")
	flagSet.String(flags.ProfileNameFlag, profileName, "")
	config1 := cli.NewContext(nil, flagSet, nil)

	// Create a temporary directory for the dummy ecs config
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
	config1 := createClusterConfig(profileName, clusterName, "")
	config2 := createClusterConfig(profileName2, secondCluster, "")
	// Create a temporary directory for the dummy ecs config
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

	// Create a temporary directory for the dummy ecs config
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

func TestMigratePrefixesPresent(t *testing.T) {
	configContents := `[ecs]
cluster = defaultCluster
aws_profile =
region = us-west-1
aws_access_key_id = AKID
aws_secret_access_key = SKID
compose-project-name-prefix =
compose-service-name-prefix = ecs-
cfn-stack-name-prefix = cfn-
`
	// Create a temporary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// save the old config
	fileInfo, err := os.Stat(tempDirName)
	assert.NoError(t, err)
	mode := fileInfo.Mode()
	err = os.MkdirAll(tempDirName+"/.ecs", mode)
	assert.NoError(t, err, "Could not create config directory")
	defer os.RemoveAll(tempDirName)
	err = ioutil.WriteFile(tempDirName+"/.ecs/config", []byte(configContents), mode)
	assert.NoError(t, err)

	// migrate
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.Bool(flags.ForceFlag, true, "")
	context := cli.NewContext(nil, flagSet, nil)

	err = Migrate(context)
	assert.NoError(t, err, "Unexpected error configuring cluster")

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, region, readConfig.Region, "Region mismatch in config.")
	assert.Equal(t, clusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Equal(t, composeServiceNamePrefix, readConfig.ComposeServiceNamePrefix, "Compose service prefix name was the incorrect value.")
	assert.Equal(t, cfnStackNamePrefix+clusterName, readConfig.CFNStackName, "CFNStackName should be empty.")
	assert.Equal(t, awsAccessKey, readConfig.AWSAccessKey, "Access Key mismatch in config.")
	assert.Equal(t, awsSecretKey, readConfig.AWSSecretKey, "Secret Key name mismatch in config.")
	assert.Empty(t, readConfig.DefaultLaunchType, "Launch Type expected to be empty")

}

func TestMigratePrefixEmpty(t *testing.T) {
	configContents := `[ecs]
cluster = defaultCluster
aws_profile =
region = us-west-1
aws_access_key_id = AKID
aws_secret_access_key = SKID
compose-project-name-prefix =
compose-service-name-prefix =
cfn-stack-name-prefix =
`
	// Create a temporary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// save the old config
	fileInfo, err := os.Stat(tempDirName)
	assert.NoError(t, err)
	mode := fileInfo.Mode()
	err = os.MkdirAll(tempDirName+"/.ecs", mode)
	assert.NoError(t, err, "Could not create config directory")
	defer os.RemoveAll(tempDirName)
	err = ioutil.WriteFile(tempDirName+"/.ecs/config", []byte(configContents), mode)
	assert.NoError(t, err)

	// migrate
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.Bool(flags.ForceFlag, true, "")
	context := cli.NewContext(nil, flagSet, nil)

	err = Migrate(context)
	assert.NoError(t, err, "Unexpected error configuring cluster")

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, region, readConfig.Region, "Region mismatch in config.")
	assert.Equal(t, clusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Empty(t, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be empty.")
	assert.Equal(t, clusterName, readConfig.CFNStackName, "CFNStackName should be empty.")
	assert.Equal(t, awsAccessKey, readConfig.AWSAccessKey, "Access Key mismatch in config.")
	assert.Equal(t, awsSecretKey, readConfig.AWSSecretKey, "Secret Key name mismatch in config.")
	assert.Empty(t, readConfig.DefaultLaunchType, "Launch Type expected to be empty")

}

func TestMigratePrefixDefault(t *testing.T) {
	configContents := `[ecs]
cluster = defaultCluster
aws_profile =
region = us-west-1
aws_access_key_id = AKID
aws_secret_access_key = SKID
`
	// Create a temporary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatal("Error while creating the dummy ecs config directory")
	}
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")
	defer os.RemoveAll(tempDirName)

	// save the old config
	fileInfo, err := os.Stat(tempDirName)
	assert.NoError(t, err)
	mode := fileInfo.Mode()
	err = os.MkdirAll(tempDirName+"/.ecs", mode)
	assert.NoError(t, err, "Could not create config directory")
	defer os.RemoveAll(tempDirName)
	err = ioutil.WriteFile(tempDirName+"/.ecs/config", []byte(configContents), mode)
	assert.NoError(t, err)

	// migrate
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.Bool(flags.ForceFlag, true, "")
	context := cli.NewContext(nil, flagSet, nil)

	err = Migrate(context)
	assert.NoError(t, err, "Unexpected error configuring cluster")

	parser, err := config.NewReadWriter()
	assert.NoError(t, err, "Error reading config")
	readConfig, err := parser.Get("", "")
	assert.NoError(t, err, "Error reading config")
	assert.Equal(t, region, readConfig.Region, "Region mismatch in config.")
	assert.Equal(t, clusterName, readConfig.Cluster, "Cluster name mismatch in config.")
	assert.Equal(t, flags.ComposeServiceNamePrefixDefaultValue, readConfig.ComposeServiceNamePrefix, "Compose service prefix name should be default.")
	assert.Empty(t, readConfig.CFNStackName, "CFNStackName should be empty.")
	assert.Equal(t, awsAccessKey, readConfig.AWSAccessKey, "Access Key mismatch in config.")
	assert.Equal(t, awsSecretKey, readConfig.AWSSecretKey, "Secret Key name mismatch in config.")
	assert.Empty(t, readConfig.DefaultLaunchType, "Launch Type expected to be empty")

}

func TestMigrateWarningConfigNotModified(t *testing.T) {
	// Test case left for posterity. Currently migrateWarning
	// uses pass by value so it can't modify the config.
	cliConfig := config.CLIConfig{Cluster: clusterName,
		Region:                   region,
		AWSProfile:               awsProfile,
		AWSAccessKey:             awsAccessKey,
		AWSSecretKey:             awsSecretKey,
		ComposeServiceNamePrefix: composeServiceNamePrefix,
		ComposeProjectNamePrefix: composeProjectNamePrefix,
		CFNStackNamePrefix:       cfnStackNamePrefix,
		CFNStackName:             cfnStackNamePrefix,
	}
	migrateWarning(cliConfig)

	assert.Equal(t, region, cliConfig.Region)
	assert.Equal(t, awsProfile, cliConfig.AWSProfile)
	assert.Equal(t, awsAccessKey, cliConfig.AWSAccessKey)
	assert.Equal(t, awsSecretKey, cliConfig.AWSSecretKey)
	assert.Equal(t, composeServiceNamePrefix, cliConfig.ComposeServiceNamePrefix)
	assert.Equal(t, composeProjectNamePrefix, cliConfig.ComposeProjectNamePrefix)
	assert.Equal(t, cfnStackNamePrefix, cliConfig.CFNStackNamePrefix)
	assert.Equal(t, cfnStackNamePrefix, cliConfig.CFNStackName)
}
