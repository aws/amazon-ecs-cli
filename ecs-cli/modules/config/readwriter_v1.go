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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

const (
	clusterConfigFileName = "config"
	profileConfigFileName = "credentials"
	configFileMode        = os.FileMode(0600) // Owner=read/write, Other=None
)

// ReadWriter interface has methods to read and write ecs-cli config to and from the config file.
type ReadWriter interface {
	SaveProfile(string, *Profile) error
	SaveCluster(string, *Cluster) error
	SetDefaultProfile(string) error
	SetDefaultCluster(string) error
	Get(string, string) (*CLIConfig, error)
}

// YAMLReadWriter implments the ReadWriter interfaces. It can be used to save and load
// ecs-cli config. Sample ecs-cli config:
// cluster: test
// aws_profile:
// region: us-west-2
// aws_access_key_id:
// aws_secret_access_key:
// compose-project-name-prefix: ecscompose-
// compose-service-name-prefix:
// cfn-stack-name-prefix: ecs-cli-
type YAMLReadWriter struct {
	destination *Destination
}

// NewReadWriter creates a new Parser object.
func NewReadWriter() (*YAMLReadWriter, error) {
	dest, err := NewDefaultDestination()
	if err != nil {
		return nil, err
	}

	return &YAMLReadWriter{destination: dest}, nil
}

func readINI(dest *Destination, cliConfig *CLIConfig) error {
	// Only read if the file exists; ini library is not good about throwing file not exist errors
	if _, err := os.Stat(ConfigFilePath(dest)); err == nil {
		iniReadWriter, err := NewINIReadWriter(dest)
		if err != nil {
			return err
		}
		return iniReadWriter.GetConfig(cliConfig)
	}

	return nil
}

// ReadClusterFile reads the cluster config file and returns a cluster config object
func ReadClusterFile(path string) (*ClusterConfig, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read config file: "+path)
	}
	config := ClusterConfig{Clusters: make(map[string]Cluster)}
	if err = yaml.Unmarshal(dat, &config); err != nil {
		return nil, errors.Wrap(err, "Failed to parse yaml file: "+path)
	}

	return &config, nil
}

// ReadCredFile reads the cluster config file and returns a profile config object
func ReadCredFile(path string) (*ProfileConfig, error) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to read config file: "+path)
	}
	config := ProfileConfig{Profiles: make(map[string]Profile)}
	if err = yaml.Unmarshal(dat, &config); err != nil {
		return nil, errors.Wrap(err, "Failed to parse yaml file: "+path)
	}

	return &config, nil
}

// readClusterConfig does all the work to read and parse the yaml cluster config
func readClusterConfig(path string, clusterConfigKey string, cliConfig *CLIConfig) error {
	// read cluster file
	config, err := ReadClusterFile(path)
	if err != nil {
		return err
	}
	// get the correct cluster
	chosenCluster := clusterConfigKey
	if clusterConfigKey == "" {
		chosenCluster = config.Default
	}

	cluster, ok := config.Clusters[chosenCluster]
	if !ok {
		return fmt.Errorf("Cluster Configuration %s could not be found. Configure clusters using 'ecs-cli configure'.", chosenCluster)
	}

	// Get the info out of the cluster
	cliConfig.Region = cluster.Region
	cliConfig.Cluster = cluster.Cluster
	cliConfig.ComposeServiceNamePrefix = cluster.ComposeServiceNamePrefix
	cliConfig.CFNStackName = cluster.CFNStackName
	cliConfig.DefaultLaunchType = cluster.DefaultLaunchType
	// Fields must be explicitly set as empty because the iniReadWriter will set them to default
	cliConfig.ComposeProjectNamePrefix = ""
	cliConfig.CFNStackNamePrefix = ""
	cliConfig.Version = yamlConfigVersion
	return nil

}

// readProfileConfig does all the work to read and parse the yaml cluster config
func readProfileConfig(path string, profileConfigKey string, cliConfig *CLIConfig) error {
	// read profile file
	config, err := ReadCredFile(path)
	if err != nil {
		return err
	}

	// get the correct profile
	chosenProfile := profileConfigKey
	if profileConfigKey == "" {
		chosenProfile = config.Default
	}
	profile, ok := config.Profiles[chosenProfile]
	if !ok {
		return fmt.Errorf("ECS Profile %s could not be found. Configure profiles using 'ecs-cli configure profile'.", chosenProfile)
	}

	// Get the info out of the cluster
	cliConfig.AWSSecretKey = profile.AWSSecretKey
	cliConfig.AWSAccessKey = profile.AWSAccessKey
	cliConfig.Version = yamlConfigVersion

	return nil

}

// Get gets the ecs-cli config object from the config file(s).
// This function either reads the old single configuration file
// Or if the new files are present, it reads from them instead
func (rdwr *YAMLReadWriter) Get(clusterConfig string, profileConfig string) (*CLIConfig, error) {
	cliConfig := &CLIConfig{}
	profilePath := credentialsFilePath(rdwr.destination)
	configPath := ConfigFilePath(rdwr.destination)

	// try to readINI first; it is either sucessful or it
	// set cliConfig to be its default value (all fields empty strings)
	readINI(rdwr.destination, cliConfig)

	// Try to read the config as YAML
	// nothing will happen if it fails
	errYAML := readClusterConfig(configPath, clusterConfig, cliConfig)

	if _, err := os.Stat(profilePath); err == nil {
		// credentials file exists- so that means we are using the new style configs
		err = readProfileConfig(profilePath, profileConfig, cliConfig)
		if err != nil {
			return nil, err
		}
	}

	// Check if there was a format error on the config file
	// this happens when both ini and yaml readers fail to read anything
	// but the file does exist (the files are allowed to not exist).
	if _, err := os.Stat(configPath); err == nil && cliConfig.Cluster == "" && cliConfig.Region == "" {
		return nil, errors.Wrapf(errYAML, "Error parsing %s", configPath)
	}

	// if no configs exist, we return an empty object
	return cliConfig, nil

}

func (rdwr *YAMLReadWriter) saveConfig(path string, config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return errors.Wrap(err, "Error saving config file")
	}

	destMode := rdwr.destination.Mode
	if err = os.MkdirAll(rdwr.destination.Path, *destMode); err != nil {
		return err
	}

	// If config file exists, set permissions first, because we may be writing creds.
	if _, err = os.Stat(path); err == nil {
		if err = os.Chmod(path, configFileMode); err != nil {
			logrus.Errorf("Unable to chmod %s to mode %s", path, configFileMode)
			return err
		}
	}

	if err = ioutil.WriteFile(path, data, configFileMode.Perm()); err != nil {
		logrus.Errorf("Unable to write configuration to %s", path)
		return err
	}

	return nil
}

// SaveProfile saves a single credential configuration
func (rdwr *YAMLReadWriter) SaveProfile(configName string, profile *Profile) error {
	path := credentialsFilePath(rdwr.destination)

	config := &ProfileConfig{Profiles: make(map[string]Profile), Version: configVersion}
	if _, err := os.Stat(path); err == nil {
		// an existing config file is there
		config, err = ReadCredFile(path)
		if err != nil {
			return err
		}
	}

	config.Profiles[configName] = *profile
	if len(config.Profiles) == 1 {
		config.Default = configName
	}

	// save the modified config
	return rdwr.saveConfig(path, config)
}

// SaveCluster save a single cluster configuration
func (rdwr *YAMLReadWriter) SaveCluster(configName string, cluster *Cluster) error {
	path := ConfigFilePath(rdwr.destination)

	// if no err on read- then existing yaml config
	config, err := ReadClusterFile(path)
	if err != nil {
		// err on read: this means that no yaml file currently exists
		config = &ClusterConfig{Clusters: make(map[string]Cluster), Version: configVersion}
	}

	config.Clusters[configName] = *cluster
	if len(config.Clusters) == 1 {
		config.Default = configName
	}

	// save the modified config
	return rdwr.saveConfig(path, config)
}

// SetDefaultProfile updates which set of credentials is defined as default
func (rdwr *YAMLReadWriter) SetDefaultProfile(configName string) error {
	path := credentialsFilePath(rdwr.destination)
	config, err := ReadCredFile(path)
	if err != nil {
		return err
	}

	if _, ok := config.Profiles[configName]; !ok {
		return fmt.Errorf("%s must be defined as a profile before it can be set as default. ", configName)
	}

	config.Default = configName

	// save the modified config
	return rdwr.saveConfig(path, config)
}

// SetDefaultCluster updates which cluster configuration is default
func (rdwr *YAMLReadWriter) SetDefaultCluster(configName string) error {
	path := ConfigFilePath(rdwr.destination)
	config, err := ReadClusterFile(path)
	if err != nil {
		return err
	}

	if _, ok := config.Clusters[configName]; !ok {
		return fmt.Errorf("%s must be defined as a profile before it can be set as default. ", configName)
	}

	config.Default = configName

	// save the modified config
	return rdwr.saveConfig(path, config)
}

func credentialsFilePath(dest *Destination) string {
	return filepath.Join(dest.Path, profileConfigFileName)
}

func ConfigFilePath(dest *Destination) string {
	return filepath.Join(dest.Path, clusterConfigFileName)
}
