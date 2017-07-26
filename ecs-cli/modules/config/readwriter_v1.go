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
	configFileName = "config"
	configFileMode = os.FileMode(0600) // Owner=read/write, Other=None
)

// ReadWriter interface has methods to read and write ecs-cli config to and from the config file.
type ReadWriter interface {
	Save(*CLIConfig) error
	GetConfig() (*CLIConfig, map[interface{}]interface{}, error)
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

func readYAML(yamlPath string, configMap map[interface{}]interface{}, cliConfig *CLIConfig) error {
	// convert yaml to CliConfig
	dat, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return errors.Wrapf(err, "Error reading yaml file %s", yamlPath)
	}
	if err = yaml.Unmarshal(dat, cliConfig); err != nil {
		return errors.Wrapf(err, "Error parsing yaml file %s", yamlPath)
	}

	// convert yaml to a map (replaces IsKeyPresent functionality)
	if err = yaml.Unmarshal(dat, &configMap); err != nil {
		return errors.Wrapf(err, "Error parsing yaml file %s", yamlPath)
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

func configPath(dest *Destination) string {
	return filepath.Join(dest.Path, configFileName)
}
