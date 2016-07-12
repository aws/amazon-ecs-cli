// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package command

import (
	"flag"
	"testing"

	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/codegangsta/cli"
)

const (
	clusterName  = "defaultCluster"
	profileName  = "defaultProfile"
	region       = "us-west-1"
	awsAccessKey = "AKID"
	awsSecretKey = "SKID"
)

func TestConfigureWithoutKeysOrProfile(t *testing.T) {
	// Config init when just cluster and region are specified
	setNoKeysNoProfile := flag.NewFlagSet("ecs-cli", 0)
	setNoKeysNoProfile.String(ecscli.RegionFlag, region, "")
	setNoKeysNoProfile.String(ecscli.ClusterFlag, clusterName, "")
	context := cli.NewContext(nil, setNoKeysNoProfile, nil)
	cfg, err := createECSConfigFromCli(context)
	if err != nil {
		t.Error("Error initializing region and cluster: ", err)
	}
	if clusterName != cfg.Cluster {
		t.Errorf("Cluster name mismtach in config. Expected [%s] Got [%s]", clusterName, cfg.Cluster)
	}
	if region != cfg.Region {
		t.Errorf("Region mismatch in config. Expected [%s] Got [%s]", region, cfg.Region)
	}
	if "" != cfg.AwsProfile {
		t.Errorf("Expected empty string for profile. Got [%s]", cfg.AwsProfile)
	}
	if "" != cfg.AwsAccessKey {
		t.Errorf("Expected empty string for acess key. Got [%s]", cfg.AwsAccessKey)
	}
	if "" != cfg.AwsSecretKey {
		t.Errorf("Expected empty string for profile. Got [%s]", cfg.AwsSecretKey)
	}
}

func TestConfigtWithSecretAndAccessKeys(t *testing.T) {
	// Config init when all non profile params are specified.
	setSecretAndAccessKeys := flag.NewFlagSet("ecs-cli", 0)
	setSecretAndAccessKeys.String(ecscli.ClusterFlag, clusterName, "")
	setSecretAndAccessKeys.String(ecscli.RegionFlag, region, "")
	setSecretAndAccessKeys.String(ecscli.SecretKeyFlag, awsSecretKey, "")
	setSecretAndAccessKeys.String(ecscli.AccessKeyFlag, awsAccessKey, "")
	context := cli.NewContext(nil, setSecretAndAccessKeys, nil)
	cfg, err := createECSConfigFromCli(context)
	if err != nil {
		t.Errorf("Error reading config from rdwr: ", err)
	}
	if clusterName != cfg.Cluster {
		t.Errorf("Cluster name mismtach in config. Expected [%s] Got [%s]", clusterName, cfg.Cluster)
	}
	if region != cfg.Region {
		t.Errorf("Region mismatch in config. Expected [%s] Got [%s]", region, cfg.Region)
	}
	if awsAccessKey != cfg.AwsAccessKey {
		t.Errorf("Access key mismatch in config. Expected [%s] Got [%s]", awsSecretKey, cfg.AwsAccessKey)
	}
	if awsSecretKey != cfg.AwsSecretKey {
		t.Errorf("Secret key mismatch in config. Expected [%s] Got [%s]", awsSecretKey, cfg.AwsSecretKey)
	}
	if "" != cfg.AwsProfile {
		t.Errorf("Expected empty string for profile. Got [%s]", cfg.AwsProfile)
	}
}

func TestConfigInitWithProfile(t *testing.T) {
	// Config init with profile.
	setProfile := flag.NewFlagSet("ecs-cli", 0)
	setProfile.String(ecscli.ProfileFlag, profileName, "")
	setProfile.String(ecscli.ClusterFlag, clusterName, "")
	setProfile.String(ecscli.RegionFlag, region, "")
	context := cli.NewContext(nil, setProfile, nil)
	cfg, err := createECSConfigFromCli(context)
	if err != nil {
		t.Errorf("Error reading config from rdwr: ", err)
	}
	if clusterName != cfg.Cluster {
		t.Errorf("Cluster name mismtach in config. Expected [%s] Got [%s]", clusterName, cfg.Cluster)
	}
	if profileName != cfg.AwsProfile {
		t.Errorf("Profile name mismatch in config. Expected [%s] Got [%s]", profileName, cfg.AwsProfile)
	}
	if region != cfg.Region {
		t.Errorf("Region mismatch in config. Expected [%s] Got [%s]", region, cfg.Region)
	}
	if "" != cfg.AwsAccessKey {
		t.Errorf("Expected empty string for acess key. Got [%s]", cfg.AwsAccessKey)
	}
	if "" != cfg.AwsSecretKey {
		t.Errorf("Expected empty string for profile. Got [%s]", cfg.AwsSecretKey)
	}
}

func TestConfigInitWithoutCluster(t *testing.T) {
	// Config init with no cluster should fail.
	setProfileNoCluster := flag.NewFlagSet("ecs-cli", 0)
	setProfileNoCluster.String(ecscli.ProfileFlag, profileName, "")
	setProfileNoCluster.String(ecscli.RegionFlag, region, "")
	context := cli.NewContext(nil, setProfileNoCluster, nil)
	_, err := createECSConfigFromCli(context)
	if err == nil {
		t.Errorf("Expected error when cluster is not specified")
	}
}

func TestConfigInitWithProfileAndKeys(t *testing.T) {
	// Config init with all params will attempt to use the credentials keys specified in the ecs profile
	setEverything := flag.NewFlagSet("ecs-cli", 0)
	setEverything.String(ecscli.ProfileFlag, profileName, "")
	setEverything.String(ecscli.ClusterFlag, clusterName, "")
	setEverything.String(ecscli.RegionFlag, region, "")
	setEverything.String(ecscli.SecretKeyFlag, awsSecretKey, "")
	setEverything.String(ecscli.AccessKeyFlag, awsAccessKey, "")
	context := cli.NewContext(nil, setEverything, nil)
	_, err := createECSConfigFromCli(context)
	if err == nil {
		t.Errorf("Expected error when both AWS Profile and access keys are specified")
	}
}

func TestConfigInitWithPrefixes(t *testing.T) {
	setPrefixes := flag.NewFlagSet("ecs-cli", 0)
	setPrefixes.String(ecscli.ProfileFlag, profileName, "")
	setPrefixes.String(ecscli.ClusterFlag, clusterName, "")

	composeProjectName := "projectName"
	composeServiceName := "serviceName"
	cfnStackName := "stackName"

	setPrefixes.String(ecscli.ComposeProjectNamePrefixFlag, composeProjectName, "")
	setPrefixes.String(ecscli.ComposeServiceNamePrefixFlag, composeServiceName, "")
	setPrefixes.String(ecscli.CFNStackNamePrefixFlag, cfnStackName, "")

	context := cli.NewContext(nil, setPrefixes, nil)

	cfg, err := createECSConfigFromCli(context)
	if err != nil {
		t.Errorf("Error reading config from rdwr: ", err)
	}
	if composeProjectName != cfg.ComposeProjectNamePrefix {
		t.Errorf("ComposeProjectName mismtach in config. Expected [%s] Got [%s]", clusterName, cfg.ComposeProjectNamePrefix)
	}
	if composeServiceName != cfg.ComposeServiceNamePrefix {
		t.Errorf("ComposeServiceName mismatch in config. Expected [%s] Got [%s]", composeServiceName, cfg.ComposeServiceNamePrefix)
	}
	if cfnStackName != cfg.CFNStackNamePrefix {
		t.Errorf("CFNStackNamePrefix mismatch in config. Expected [%s] Got [%s]", cfnStackName, cfg.CFNStackNamePrefix)
	}
}

func TestConfigInitWithoutPrefixes(t *testing.T) {
	setNoPrefixes := flag.NewFlagSet("ecs-cli", 0)
	setNoPrefixes.String(ecscli.ProfileFlag, profileName, "")
	setNoPrefixes.String(ecscli.ClusterFlag, clusterName, "")

	context := cli.NewContext(nil, setNoPrefixes, nil)

	cfg, err := createECSConfigFromCli(context)
	if err != nil {
		t.Errorf("Error reading config from rdwr: ", err)
	}
	if "" != cfg.ComposeProjectNamePrefix {
		t.Errorf("ComposeProjectName mismtach in config. Expected empty string Got [%s]", cfg.ComposeProjectNamePrefix)
	}
	if "" != cfg.ComposeServiceNamePrefix {
		t.Errorf("ComposeServiceName mismatch in config. Expected empty string Got [%s]", cfg.ComposeServiceNamePrefix)
	}
	if "" != cfg.CFNStackNamePrefix {
		t.Errorf("CFNStackNamePrefix mismatch in config. Expected empty string Got [%s]", cfg.CFNStackNamePrefix)
	}
}
