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

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
)

const (
	yamlConfigFileName = "config.yml"
	configFileMode     = os.FileMode(0600)
)

// ReadWriter interface has methods to read and write ecs-cli config to and from the config file.
type ReadWriter interface {
	Save(*CliConfig) error
	GetConfig() (*CliConfig, map[interface{}]interface{}, error)
}

// YamlReadWriter implments the ReadWriter interfaces. It can be used to save and load
// ecs-cli config. Sample ecs-cli config:
// cluster: test
// aws_profile:
// region: us-west-2
// aws_access_key_id:
// aws_secret_access_key:
// compose-project-name-prefix: ecscompose-
// compose-service-name-prefix:
// cfn-stack-name-prefix: ecs-cli-
type YamlReadWriter struct {
	destination *Destination
}

// NewReadWriter creates a new Parser object.
func NewReadWriter() (*YamlReadWriter, error) {
	dest, err := newDefaultDestination()
	if err != nil {
		return nil, err
	}

	return &YamlReadWriter{destination: dest}, nil
}

func readYaml(yamlPath string, configMap map[interface{}]interface{}, cliConfig *CliConfig) error {
	// convert yaml to CliConfig
	dat, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(dat, cliConfig); err != nil {
		return err
	}

	// convert yaml to a map (replaces IsKeyPresent functionality)
	if err = yaml.Unmarshal(dat, &configMap); err != nil {
		return err
	}

	return nil
}

// GetConfig gets the ecs-cli config object from the config file.
func (rdwr *YamlReadWriter) GetConfig() (*CliConfig, map[interface{}]interface{}, error) {
	to := new(CliConfig)
	configMap := make(map[interface{}]interface{})
	// read the raw bytes of the config file
	iniPath := iniConfigPath(rdwr.destination)
	yamlPath := yamlConfigPath(rdwr.destination)

	_, iniErr := os.Stat(iniPath)
	_, yamlErr := os.Stat(yamlPath)
	if yamlErr == nil {
		readYaml(yamlPath, configMap, to)
	} else if iniErr == nil { // file exists
		// old ini config
		iniReadWriter, err := NewIniReadWriter(rdwr.destination)
		if err != nil {
			return nil, nil, err
		}
		to, configMap, err = iniReadWriter.GetConfig()
		if err != nil {
			return nil, nil, err
		}

	} else {
		// if neither file existed we return the yaml error
		return nil, nil, yamlErr
	}
	return to, configMap, nil
}

func (rdwr *YamlReadWriter) Save(cliConfig *CliConfig) error {
	destMode := rdwr.destination.Mode
	if err := os.MkdirAll(rdwr.destination.Path, *destMode); err != nil {
		return err
	}
	// set version
	cliConfig.Version = configVersion

	path := yamlConfigPath(rdwr.destination)

	// Warn the user if ini path also exists
	iniPath := iniConfigPath(rdwr.destination)
	if _, iniErr := os.Stat(iniPath); iniErr == nil {
		logrus.Warnf("Writing yaml formatted config to %s/.ecs/%s.\nIni formatted config still exists in %s/.ecs/%s.", os.Getenv("HOME"), yamlConfigFileName, os.Getenv("HOME"), iniConfigFileName)
	}

	// If config file exists, set permissions first, because we may be writing creds.
	if _, err := os.Stat(path); err == nil {
		if err = os.Chmod(path, configFileMode); err != nil {
			logrus.Errorf("Unable to chmod %s to mode %s", path, configFileMode)
			return err
		}
	}

	data, err := yaml.Marshal(cliConfig)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(path, data, configFileMode.Perm()); err != nil {
		logrus.Errorf("Unable to write config to %s", path)
		return err
	}

	return nil
}

func yamlConfigPath(dest *Destination) string {
	return filepath.Join(dest.Path, yamlConfigFileName)
}
