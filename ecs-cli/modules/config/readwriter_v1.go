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
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	"github.com/Sirupsen/logrus"
)

const (
	configFileName = "config"
	configFileMode = os.FileMode(0600)
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

// GetConfig gets the ecs-cli config object from the config file.
func (rdwr *YamlReadWriter) GetConfig() (*CliConfig, map[interface{}]interface{}, error) {
	to := new(CliConfig)
	configMap := make(map[interface{}]interface{})
	// read the raw bytes of the config file
	path := configPath(rdwr.destination)
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	// Handle the case where the old ini config is still there
	if strings.HasPrefix(string(dat), "["+ecsSectionKey+"]") {
		// old ini config
		iniReadWriter, err := NewIniReadWriter()
		if err != nil {
			return nil, nil, err
		}
		to, configMap, err = iniReadWriter.GetConfig()
		if err != nil {
			return nil, nil, err
		}

	} else {
		// convert yaml to CliConfig
		err = yaml.Unmarshal(dat, to)
		if err != nil {
			return nil, nil, err
		}

		// convert yaml to a map (replaces IsKeyPresent functionality)
		err = yaml.Unmarshal(dat, &configMap)
		if err != nil {
			return nil, nil, err
		}
		tmpMap, ok := configMap[ecsVersionKey].(map[interface{}]interface{})
		if !ok {
			return nil, nil, errors.New("Interface conversion panic; config file may not be the right version.")
		}
		configMap = tmpMap

	}
	return to, configMap, nil
}

func (rdwr *YamlReadWriter) Save(cliConfig *CliConfig) error {
	destMode := rdwr.destination.Mode
	err := os.MkdirAll(rdwr.destination.Path, *destMode)
	if err != nil {
		return err
	}

	path := configPath(rdwr.destination)

	// If config file exists, set permissions first, because we may be writing creds.
	if _, err := os.Stat(path); err == nil {
		err = os.Chmod(path, configFileMode)
		if err != nil {
			logrus.Errorf("Unable to chmod %s to mode %s", path, configFileMode)
			return err
		}
	}

	if err != nil {
		logrus.Errorf("Unable to open/create %s with mode %s", path, configFileMode)
		return err
	}

	data, err := yaml.Marshal(cliConfig)
	err = ioutil.WriteFile(path, data, configFileMode.Perm())
	if err != nil {
		logrus.Errorf("Unable to write config to %s", path)
		return err
	}

	return nil
}

func configPath(dest *Destination) string {
	return filepath.Join(dest.Path, configFileName)
}
