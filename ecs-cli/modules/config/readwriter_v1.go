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

// func readYAML(yamlPath string, configMap map[interface{}]interface{}, cliConfig *CLIConfig) error {
// 	// convert yaml to CliConfig
// 	dat, err := ioutil.ReadFile(yamlPath)
// 	if err != nil {
// 		return errors.Wrapf(err, "Error reading yaml file %s", yamlPath)
// 	}
// 	if err = yaml.Unmarshal(dat, cliConfig); err != nil {
// 		return errors.Wrapf(err, "Error parsing yaml file %s", yamlPath)
// 	}
//
// 	// convert yaml to a map (replaces IsKeyPresent functionality)
// 	if err = yaml.Unmarshal(dat, &configMap); err != nil {
// 		return errors.Wrapf(err, "Error parsing yaml file %s", yamlPath)
// 	}
//
// 	return nil
// }

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

// GetConfig gets the ecs-cli config object from the config file.
// map contains the keys that are present in the config file (maps string field name to string field value)
// map is type map[interface{}]interface{} to ensure fowards compatibility with changes that will
// cause certain keys to be mapped to maps of keys
func (rdwr *YAMLReadWriter) GetConfig() (*CLIConfig, map[interface{}]interface{}, error) {
	to := new(CLIConfig)
	configMap := make(map[interface{}]interface{})
	// read the raw bytes of the config file
	path := configPath(rdwr.destination)

	if errYAML := readYAML(path, configMap, to); errYAML == nil {
		// File is YAML
		return to, configMap, nil
	} else if to, configMap, errINI := readINI(rdwr.destination); errINI == nil {
		// File is INI
		return to, configMap, nil
	} else if errINI != nil {
		// Return the ini error
		// This will return a parsing error for ini (format error)
		return nil, nil, errINI
	} else {
		// If yaml through an error, but ini didn't throw an error- return the yaml error
		return nil, nil, errYAML
	}
}

// GetConfigs gets the ecs-cli config object from the config file(s).
// This function either reads the old single configuration file
// Or if the new files are present, it reads from them instead
func (rdwr *YAMLReadWriter) GetConfigs(clusterConfig string, profileConfig string) (*CLIConfig, map[interface{}]interface{}, error) {
	cliConfig := &CLIConfig{}
	configMap := make(map[interface{}]interface{})
	// read the raw bytes of the config file
	profilePath := credentialsFilePath(rdwr.destination)
	configPath := configFilePath(rdwr.destination)

	// 	If profile exists:
	// 	parse profile
	// 	parse config as YAML if it exists
	// 	return
	// Else if no YAML error on config
	// 	parse config
	// 	we know that profile is empty
	// 	return
	// Else if no INI Error
	// 	parse ini
	// 	return
	// Else if config exists
	// 	return YAML parse error
	// Else
	// 	return empty map and empty config object
	if _, err := os.Stat(profilePath); err != nil {
		// credentials file exists- so that means we are using the new style configs
	}

	if err == nil && yamlErr != nil { // file exists
		// old ini config
		iniReadWriter, err := NewINIReadWriter(rdwr.destination)
		if err != nil {
			return nil, nil, err
		}
		cliConfig, configMap, err = iniReadWriter.GetConfig()
		if err != nil {
			return nil, nil, err
		}

	} else {
		// If the ini file didn't exist, then we assume the yaml file exists
		// if it doesn't, then throw error
		// convert yaml to CliConfig
		clusterMap := make(map[interface{}]interface{})
		profileMap := make(map[interface{}]interface{})

		// read cluster file
		dat, err := ioutil.ReadFile(clusterPath)
		if err != nil {
			return nil, nil, err
		}

		// convert cluster yaml to a map (replaces IsKeyPresent functionality)
		if err = yaml.Unmarshal(dat, &clusterMap); err != nil {
			return nil, nil, err
		}

		// read profile file
		dat, err = ioutil.ReadFile(profilePath)
		if err != nil {
			return nil, nil, err
		}
		// convert profile yaml to a map (replaces IsKeyPresent functionality)
		if err = yaml.Unmarshal(dat, &profileMap); err != nil {
			return nil, nil, err
		}

		processProfileMap(profileConfig, profileMap, configMap, cliConfig)
		processClusterMap(clusterConfig, clusterMap, configMap, cliConfig)

	}
	return cliConfig, configMap, nil
}

func processProfileMap(profileKey string, profileMap map[interface{}]interface{}, configMap map[interface{}]interface{}, cliConfig *CLIConfig) error {
	if profileKey == "" {
		var ok bool
		profileKey, ok = profileMap["default"].(string)
		if !ok {
			return errors.New("Format issue with profile config file; expected key not found.")
		}
	}
	profile, ok := profileMap["ecs_profiles"].(map[interface{}]interface{})[profileKey].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}

	configMap[awsAccessKey] = profile[awsAccessKey]
	configMap[awsSecretKey] = profile[awsSecretKey]
	cliConfig.AWSAccessKey, ok = profile[awsAccessKey].(string)
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}
	cliConfig.AWSSecretKey, ok = profile[awsSecretKey].(string)
	if !ok {
		return errors.New("Format issue with profile config file; expected key not found.")
	}

	return nil

}

func processClusterMap(clusterConfigKey string, clusterMap map[interface{}]interface{}, configMap map[interface{}]interface{}, cliConfig *CLIConfig) error {
	if clusterConfigKey == "" {
		var ok bool
		clusterConfigKey, ok = clusterMap["default"].(string)
		if !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	cluster, ok := clusterMap["clusters"].(map[interface{}]interface{})[clusterConfigKey].(map[interface{}]interface{})
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}

	configMap[clusterKey] = cluster[clusterKey]
	configMap[regionKey] = cluster[regionKey]
	cliConfig.Cluster, ok = cluster[clusterKey].(string)
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}
	cliConfig.Region, ok = cluster[regionKey].(string)
	if !ok {
		return errors.New("Format issue with cluster config file; expected key not found.")
	}

	// Prefixes
	// ComposeProjectNamePrefix
	if _, ok := cluster[composeProjectNamePrefixKey]; ok {
		configMap[composeProjectNamePrefixKey] = cluster[composeProjectNamePrefixKey]
		cliConfig.ComposeProjectNamePrefix, ok = cluster[composeProjectNamePrefixKey].(string)
		if !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	// ComposeServiceNamePrefix
	if _, ok := cluster[composeServiceNamePrefixKey]; ok {
		configMap[composeServiceNamePrefixKey] = cluster[composeServiceNamePrefixKey]
		cliConfig.ComposeServiceNamePrefix, ok = cluster[composeServiceNamePrefixKey].(string)
		if !ok {
			return errors.New("Format issue with cluster config file; expected key not found.")
		}
	}
	// CFNStackNamePrefix
	if _, ok := cluster[cfnStackNamePrefixKey]; ok {
		configMap[cfnStackNamePrefixKey] = cluster[cfnStackNamePrefixKey]
		cliConfig.CFNStackNamePrefix, ok = cluster[cfnStackNamePrefixKey].(string)
		if !ok {
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
