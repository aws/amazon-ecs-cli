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
	GetConfigs(string, string) (*CLIConfig, map[interface{}]interface{}, error)
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

func readINI(dest *Destination) (*CLIConfig, map[interface{}]interface{}, error) {
	iniReadWriter, err := NewINIReadWriter(dest)
	if err != nil {
		return nil, nil, err
	}
	return iniReadWriter.GetConfig()
}

func readYAML(path string, configMap map[interface{}]interface{}) error {
	// read cluster file
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// convert cluster yaml to a map (replaces IsKeyPresent functionality)
	if err = yaml.Unmarshal(dat, &configMap); err != nil {
		return err
	}
	return nil
}

// GetConfigs gets the ecs-cli config object from the config file(s).
// map contains the keys that are present in the config file (maps string field name to string field value)
// map is type map[interface{}]interface{} to ensure fowards compatibility with changes that will
// cause certain keys to be mapped to maps of keys
// This function either reads the old single configuration file
// Or if the new files are present, it reads from them instead
func (rdwr *YAMLReadWriter) GetConfigs(clusterConfig string, profileConfig string) (*CLIConfig, map[interface{}]interface{}, error) {
	cliConfig := &CLIConfig{}
	configMap := make(map[interface{}]interface{})
	// read the raw bytes of the config file
	profilePath := credentialsFilePath(rdwr.destination)
	configPath := configFilePath(rdwr.destination)

	// Try to read the config as YAML
	clusterMap := make(map[interface{}]interface{})
	errYAMLConfig := readYAML(configPath, clusterMap)
	if errYAMLConfig == nil {
		// cluster config exists and is YAML
		processClusterMap(clusterConfig, clusterMap, configMap, cliConfig)
	}

	if _, err := os.Stat(profilePath); err != nil {
		// credentials file exists- so that means we are using the new style configs
		profileMap := make(map[interface{}]interface{})
		if err = readYAML(profilePath, profileMap); err != nil {
			return nil, nil, errors.Wrapf(err, "Error reading or parsing credentials file, %s", profilePath)
		}

		if err = processProfileMap(profileConfig, profileMap, configMap, cliConfig); err != nil {
			return nil, nil, errors.Wrapf(err, "Error processing credentials file, %s", profilePath)
		}
		return cliConfig, configMap, nil
	}

	if to, configMap, errINI := readINI(rdwr.destination); errINI == nil {
		// File is INI
		return to, configMap, nil
	}

	if _, err := os.Stat(configPath); err != nil {
		// Cluster config file exists, but it wasn't INI, so that means it was YAML but there was a parse error
		return nil, nil, errYAMLConfig
	}

	// if no configs exist, we return empty objects
	return cliConfig, configMap, nil

}

func processProfileMap(profileKey string, profileMap map[interface{}]interface{}, configMap map[interface{}]interface{}, cliConfig *CLIConfig) error {
	if profileKey == "" {
		var ok bool
		if profileKey, ok = profileMap["default"].(string); !ok {
			return errors.New("Format issue with profile config file; expected key not found.")
		}
	}
	profile, ok := profileMap["ecs_profiles"].(map[interface{}]interface{})[profileKey].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}

	configMap[awsAccessKey] = profile[awsAccessKey]
	configMap[awsSecretKey] = profile[awsSecretKey]
	if cliConfig.AWSAccessKey, ok = profile[awsAccessKey].(string); !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}
	if cliConfig.AWSSecretKey, ok = profile[awsSecretKey].(string); !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}

	return nil

}

func processClusterMap(clusterConfigKey string, clusterMap map[interface{}]interface{}, configMap map[interface{}]interface{}, cliConfig *CLIConfig) error {
	if clusterConfigKey == "" {
		var ok bool
		if clusterConfigKey, ok = clusterMap["default"].(string); !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	cluster, ok := clusterMap["clusters"].(map[interface{}]interface{})[clusterConfigKey].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}

	configMap[clusterKey] = cluster[clusterKey]
	configMap[regionKey] = cluster[regionKey]
	if cliConfig.Cluster, ok = cluster[clusterKey].(string); !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}
	if cliConfig.Region, ok = cluster[regionKey].(string); !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}

	// Prefixes
	// ComposeProjectNamePrefix
	if _, ok := cluster[composeProjectNamePrefixKey]; ok {
		configMap[composeProjectNamePrefixKey] = cluster[composeProjectNamePrefixKey]
		if cliConfig.ComposeProjectNamePrefix, ok = cluster[composeProjectNamePrefixKey].(string); !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	// ComposeServiceNamePrefix
	if _, ok := cluster[composeServiceNamePrefixKey]; ok {
		configMap[composeServiceNamePrefixKey] = cluster[composeServiceNamePrefixKey]
		if cliConfig.ComposeServiceNamePrefix, ok = cluster[composeServiceNamePrefixKey].(string); !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	// CFNStackNamePrefix
	if _, ok := cluster[cfnStackNamePrefixKey]; ok {
		configMap[cfnStackNamePrefixKey] = cluster[cfnStackNamePrefixKey]
		if cliConfig.CFNStackNamePrefix, ok = cluster[cfnStackNamePrefixKey].(string); !ok {
			return errors.New("Format issue with profile cluster file; expected key not found.")
		}
	}

	return nil

}

// Save saves the CLIConfig to a yaml formatted file
func (rdwr *YAMLReadWriter) Save(cliConfig *CLIConfig) error {
	destMode := rdwr.destination.Mode
	if err := os.MkdirAll(rdwr.destination.Path, *destMode); err != nil {
		return errors.Wrapf(err, "Could not make directory %s", rdwr.destination.Path)
	}
	// set version
	cliConfig.Version = configVersion

	path := configPath(rdwr.destination)

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
