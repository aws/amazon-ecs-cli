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
	"testing"

	command "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	clusterName  = "defaultCluster"
	stackName    = "defaultCluster"
	profileName  = "defaultProfile"
	region       = "us-west-1"
	awsAccessKey = "AKID"
	awsSecretKey = "SKID"
)

func TestConfigureWithoutKeysOrProfile(t *testing.T) {
	// Config init when just cluster and region are specified
	setNoKeysNoProfile := flag.NewFlagSet("ecs-cli", 0)
	setNoKeysNoProfile.String(command.RegionFlag, region, "")
	setNoKeysNoProfile.String(command.ClusterFlag, clusterName, "")
	context := cli.NewContext(nil, setNoKeysNoProfile, nil)
	cfg, err := createECSConfigFromCli(context)
	assert.NoError(t, err, "Unexpected error initializing region and cluster")
	assert.Equal(t, clusterName, cfg.Cluster, "Expected cluster name to match")
	assert.Equal(t, region, cfg.Region, "Expected region to match")
	assert.Empty(t, cfg.AWSProfile, "Expected AWS profile to be empty")
	assert.Empty(t, cfg.AWSAccessKey, "Expected access key to be empty")
	assert.Empty(t, cfg.AWSSecretKey, "Expected secret key to be empty")
}

func TestConfigtWithSecretAndAccessKeys(t *testing.T) {
	// Config init when all non profile params are specified.
	setSecretAndAccessKeys := flag.NewFlagSet("ecs-cli", 0)
	setSecretAndAccessKeys.String(command.ClusterFlag, clusterName, "")
	setSecretAndAccessKeys.String(command.RegionFlag, region, "")
	setSecretAndAccessKeys.String(command.SecretKeyFlag, awsSecretKey, "")
	setSecretAndAccessKeys.String(command.AccessKeyFlag, awsAccessKey, "")
	context := cli.NewContext(nil, setSecretAndAccessKeys, nil)
	cfg, err := createECSConfigFromCli(context)
	assert.NoError(t, err, "Unexpected error reading config from rdwr")
	assert.Equal(t, clusterName, cfg.Cluster, "Expected cluster name to match")
	assert.Equal(t, region, cfg.Region, "Expected region to match")
	assert.Empty(t, cfg.AWSProfile, "Expected AWS profile to be empty")
	assert.Equal(t, awsAccessKey, cfg.AWSAccessKey, "Expected access key to match")
	assert.Equal(t, awsSecretKey, cfg.AWSSecretKey, "Expected secret key to match")
}

func TestConfigInitWithProfile(t *testing.T) {
	// Config init with profile.
	setProfile := flag.NewFlagSet("ecs-cli", 0)
	setProfile.String(command.ProfileFlag, profileName, "")
	setProfile.String(command.ClusterFlag, clusterName, "")
	setProfile.String(command.RegionFlag, region, "")
	context := cli.NewContext(nil, setProfile, nil)
	cfg, err := createECSConfigFromCli(context)
	assert.NoError(t, err, "Unexpected error reading config from rdwr")
	assert.Equal(t, clusterName, cfg.Cluster, "Expected cluster name to match")
	assert.Equal(t, region, cfg.Region, "Expected region to match")
	assert.Equal(t, profileName, cfg.AWSProfile, "Expected AWS profile to match")
	assert.Empty(t, cfg.AWSAccessKey, "Expected access key to be empty")
	assert.Empty(t, cfg.AWSSecretKey, "Expected secret key to be empty")
}

func TestConfigInitWithoutCluster(t *testing.T) {
	// Config init with no cluster should fail.
	setProfileNoCluster := flag.NewFlagSet("ecs-cli", 0)
	setProfileNoCluster.String(command.ProfileFlag, profileName, "")
	setProfileNoCluster.String(command.RegionFlag, region, "")
	context := cli.NewContext(nil, setProfileNoCluster, nil)
	_, err := createECSConfigFromCli(context)
	assert.Error(t, err, "Expected error when cluster is not specified")
}

func TestConfigInitWithProfileAndKeys(t *testing.T) {
	// Config init with all params will attempt to use the credentials keys specified in the ecs profile
	setEverything := flag.NewFlagSet("ecs-cli", 0)
	setEverything.String(command.ProfileFlag, profileName, "")
	setEverything.String(command.ClusterFlag, clusterName, "")
	setEverything.String(command.RegionFlag, region, "")
	setEverything.String(command.SecretKeyFlag, awsSecretKey, "")
	setEverything.String(command.AccessKeyFlag, awsAccessKey, "")
	context := cli.NewContext(nil, setEverything, nil)
	_, err := createECSConfigFromCli(context)
	assert.Error(t, err, "Expected error when both AWS Profile and access keys are specified")
}

func TestConfigInitWithPrefixes(t *testing.T) {
	setPrefixes := flag.NewFlagSet("ecs-cli", 0)
	setPrefixes.String(command.ProfileFlag, profileName, "")
	setPrefixes.String(command.ClusterFlag, clusterName, "")

	composeProjectName := "projectName"
	composeServiceName := "serviceName"
	cfnStackName := "stackName"

	setPrefixes.String(command.ComposeProjectNamePrefixFlag, composeProjectName, "")
	setPrefixes.String(command.ComposeServiceNamePrefixFlag, composeServiceName, "")
	setPrefixes.String(command.CFNStackNamePrefixFlag, cfnStackName, "")

	context := cli.NewContext(nil, setPrefixes, nil)

	cfg, err := createECSConfigFromCli(context)
	assert.NoError(t, err, "Unexpected error reading config from rdwr")
	assert.Equal(t, composeProjectName, cfg.ComposeProjectNamePrefix, "Expected ComposeProjectName to match in config")
	assert.Equal(t, composeServiceName, cfg.ComposeServiceNamePrefix, "Expected ComposeServiceName to match in config")
	assert.Equal(t, cfnStackName, cfg.CFNStackNamePrefix, "Expected CfnStackName to match in config")
}

func TestConfigInitWithoutPrefixes(t *testing.T) {
	setNoPrefixes := flag.NewFlagSet("ecs-cli", 0)
	setNoPrefixes.String(command.ProfileFlag, profileName, "")
	setNoPrefixes.String(command.ClusterFlag, clusterName, "")

	context := cli.NewContext(nil, setNoPrefixes, nil)

	cfg, err := createECSConfigFromCli(context)
	assert.NoError(t, err, "Unexpected error reading config from rdwr")
	assert.Empty(t, cfg.ComposeProjectNamePrefix, "Expected ComposeProjectNamePrefix to be empty")
	assert.Empty(t, cfg.ComposeServiceNamePrefix, "Expected ComposeServiceNamePrefix to be empty")
	assert.Empty(t, cfg.CFNStackNamePrefix, "Expected CFNStackNamePrefix to be empty")
}
