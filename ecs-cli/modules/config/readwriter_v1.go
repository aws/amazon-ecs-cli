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
	"path/filepath"

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
	Save(*CLIConfig) error
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
	dest, err := newDefaultDestination()
	if err != nil {
		return nil, err
	}

	return &YAMLReadWriter{destination: dest}, nil
}

func readINI(dest *Destination, cliConfig *CLIConfig) error {
	iniReadWriter, err := NewINIReadWriter(dest)
	if err != nil {
		return err
	}
	return iniReadWriter.GetConfig(cliConfig)
}

// readClusterConfig does all the work to read and parse the yaml cluster config
func readClusterConfig(path string, clusterConfigKey string, cliConfig *CLIConfig) error {
	// read cluster file
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "Failed to read config file: "+path)
	}
	config := ClusterConfig{Clusters: make(map[string]Cluster)}
	if err = yaml.Unmarshal(dat, &config); err != nil {
		return errors.Wrap(err, "Failed to parse yaml file: "+path)
	}

	// get the correct cluster
	chosenCluster := clusterConfigKey
	if clusterConfigKey == "" {
		chosenCluster = config.Default
	}

	cluster := config.Clusters[chosenCluster]
	// Get the info out of the cluster
	cliConfig.Region = cluster.Region
	cliConfig.Cluster = cluster.Cluster
	cliConfig.ComposeServiceNamePrefix = cluster.ComposeServiceNamePrefix
	return nil

}

// readProfileConfig does all the work to read and parse the yaml cluster config
func readProfileConfig(path string, profileConfigKey string, cliConfig *CLIConfig) error {
	// read cluster file
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, "Failed to read config file: "+path)
	}

	config := ProfileConfig{Profiles: make(map[string]Profile)}
	if err = yaml.Unmarshal(dat, &config); err != nil {
		return errors.Wrap(err, "Failed to parse yaml file: "+path)
	}

	// get the correct profile
	chosenProfile := profileConfigKey
	if profileConfigKey == "" {
		chosenProfile = config.Default
	}
	profile := config.Profiles[chosenProfile]

	// Get the info out of the cluster
	cliConfig.AWSSecretKey = profile.AWSSecretKey
	cliConfig.AWSAccessKey = profile.AWSAccessKey

	return nil

}

// Get gets the ecs-cli config object from the config file(s).
// This function either reads the old single configuration file
// Or if the new files are present, it reads from them instead
func (rdwr *YAMLReadWriter) Get(clusterConfig string, profileConfig string) (*CLIConfig, error) {
	cliConfig := &CLIConfig{} // read the raw bytes of the config file
	profilePath := credentialsFilePath(rdwr.destination)
	configPath := configFilePath(rdwr.destination)

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

// Save saves the CLIConfig to a yaml formatted file
func (rdwr *YAMLReadWriter) Save(cliConfig *CLIConfig) error {
	destMode := rdwr.destination.Mode
	if err := os.MkdirAll(rdwr.destination.Path, *destMode); err != nil {
		return errors.Wrapf(err, "Could not make directory %s", rdwr.destination.Path)
	}

	path := configFilePath(rdwr.destination)

	// If config file exists, set permissions first, because we may be writing creds.
	// This is necessary because ioutil.WriteFile only sets the permissions
	// if the file is being created for the first time; this handles the case
	// where the file already exists
	if _, err := os.Stat(path); err == nil {
		if err = os.Chmod(path, configFileMode); err != nil {
			return errors.Wrapf(err, "Could not set file permissions, %s, for path %s", configFileMode, path)
		}
	}

	data, err := yaml.Marshal(cliConfig)
	if err != nil {
		return errors.Wrap(err, "Unable to parse the config")
	}
	// WriteFile will not change the permissions of an existing file
	if err = ioutil.WriteFile(path, data, configFileMode.Perm()); err != nil {
		return errors.Wrapf(err, "Could not write config file %s", path)
	}

	return nil
}

func credentialsFilePath(dest *Destination) string {
	return filepath.Join(dest.Path, profileConfigFileName)
}

func configFilePath(dest *Destination) string {
	return filepath.Join(dest.Path, clusterConfigFileName)
}
