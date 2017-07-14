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
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/go-ini/ini"
)

const (
	configFileName = "config"
	configFileMode = os.FileMode(0600)
)

// ReadWriter interface has methods to read and write ecs-cli config to and from the config file.
type ReadWriter interface {
	Save(*CliConfig, *Destination) error
	IsInitialized() (bool, error)
	GetConfig() (*CliConfig, error)
	IsKeyPresent(string, string) bool
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
	*Destination
}

// NewReadWriter creates a new Parser object.
func NewReadWriter() (*YamlReadWriter, error) {
	dest, err := newDefaultDestination()
	if err != nil {
		return nil, err
	}

	return &YamlReadWriter{Destination: dest}, nil
}

// GetConfig gets the ecs-cli config object from the config file.
func (rdwr *YamlReadWriter) GetConfig() (*CliConfig, error) {
	to := new(CliConfig)
	err := rdwr.cfg.MapTo(to)
	if err != nil {
		return nil, err
	}

	return to, nil
}

// IniReadWriter implments the ReadWriter interfaces. It can be used to save and load
// ecs-cli config. Sample ecs-cli config:
// [ecs]
// cluster = test
// aws_profile =
// region = us-west-2
// aws_access_key_id =
// aws_secret_access_key =
// compose-project-name-prefix = ecscompose-
// compose-service-name-prefix =
// cfn-stack-name-prefix = ecs-cli-

type IniReadWriter struct {
	*Destination
	cfg *ini.File
}

// NewReadWriter creates a new Parser object.
func NewIniReadWriter() (*IniReadWriter, error) {
	dest, err := newDefaultDestination()
	if err != nil {
		return nil, err
	}

	iniCfg, err := newIniConfig(dest)
	if err != nil {
		return nil, err
	}

	return &IniReadWriter{Destination: dest, cfg: iniCfg}, nil
}

// GetConfig gets the ecs-cli config object from the config file.
func (rdwr *IniReadWriter) GetConfig() (*CliConfig, error) {
	to := new(CliConfig)
	err := rdwr.cfg.MapTo(to)
	if err != nil {
		return nil, err
	}

	return to, nil
}

// IsInitialized returns true if the config file can be read and if all the key config fields
// have been initialized.
func (rdwr *IniReadWriter) IsInitialized() (bool, error) {
	to := new(CliConfig)
	err := rdwr.cfg.MapTo(to)
	if err != nil {
		return false, err
	}

	if to.Cluster == "" {
		return false, nil
	}

	return true, nil
}

// IsKeyPresent returns true if the input key is present in the input section
func (rdwr *IniReadWriter) IsKeyPresent(section, key string) bool {
	return rdwr.cfg.Section(section).HasKey(key)
}

// ReadFrom initializes the ini object from an existing ecs-cli config object.
func (rdwr *IniReadWriter) ReadFrom(ecsConfig *CliConfig) error {
	return rdwr.cfg.ReflectFrom(ecsConfig)
}

// Save saves the config to a config file.
func (rdwr *IniReadWriter) Save(dest *Destination) error {
	destMode := dest.Mode
	err := os.MkdirAll(dest.Path, *destMode)
	if err != nil {
		return err
	}

	path := configPath(dest)

	// If config file exists, set permissions first, because we may be writing creds.
	if _, err := os.Stat(path); err == nil {
		err = os.Chmod(path, configFileMode)
		if err != nil {
			logrus.Errorf("Unable to chmod %s to mode %s", path, configFileMode)
			return err
		}
	}

	// Open the file, optionally creating it with our desired permissions.
	// This will let us pass it (as io.Writer) to go-ini but let us control the file.
	configFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, configFileMode)

	// Truncate the file in case the earlier contents are longer than the new
	// contents, so there will not be any trash at the end of the file
	configFile.Truncate(0)

	if err != nil {
		logrus.Errorf("Unable to open/create %s with mode %s", path, configFileMode)
		return err
	}
	defer configFile.Close()

	_, err = rdwr.cfg.WriteTo(configFile)
	if err != nil {
		logrus.Errorf("Unable to write config to %s", path)
		return err
	}

	return nil
}

func configPath(dest *Destination) string {
	return filepath.Join(dest.Path, configFileName)
}

func newIniConfig(dest *Destination) (*ini.File, error) {
	iniCfg := ini.Empty()
	path := configPath(dest)
	logrus.Debugf("using config file: %s", path)
	if _, err := os.Stat(path); err != nil {
		// TODO: handle os.isnotexist(path) and other errors differently
		// error reading config file, create empty config ini.
		logrus.Debugf("no config files found, initializing empty ini")
	} else {
		err = iniCfg.Append(path)
		if err != nil {
			return nil, err
		}
	}

	return iniCfg, nil
}
