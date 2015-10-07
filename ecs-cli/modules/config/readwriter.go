// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// ReadWriter interface has methods to read and write ecs-cli config to and from the config file.
type ReadWriter interface {
	Save(*Destination) error
	IsInitialized() (bool, error)
	ReadFrom(*CliConfig) error
	GetConfig() (*CliConfig, error)
}

// IniReadWriter implments the ReadWriter interfaces. It can be used to save and load
// ecs-cli config.
type IniReadWriter struct {
	*Destination
	cfg *ini.File
}

// NewReadWriter creates a new Parser object.
func NewReadWriter() (*IniReadWriter, error) {
	dest, err := newDefaultDestionation()
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

// ReadFrom initializes the ini object from an existing ecs-cli config object.
func (rdwr *IniReadWriter) ReadFrom(ecsConfig *CliConfig) error {
	return rdwr.cfg.ReflectFrom(ecsConfig)
}

// Save saves the config to a config file.
func (rdwr *IniReadWriter) Save(dest *Destination) error {
	mode := dest.Mode
	err := os.MkdirAll(dest.Path, *mode)
	if err != nil {
		return err
	}
	// TODO: Move to const.
	return rdwr.cfg.SaveTo(filepath.Join(dest.Path, "config"))
}

func newIniConfig(dest *Destination) (*ini.File, error) {
	iniCfg := ini.Empty()
	// TODO: Move to const.
	filename := filepath.Join(dest.Path, "config")
	logrus.Debugf("using config file: %s", filename)
	if _, err := os.Stat(filename); err != nil {
		// TODO: handle os.isnotexist(filename) and other errors differently
		// error reading config file, create empty config ini.
		logrus.Debugf("no config files found, initializing empty ini")
	} else {
		err = iniCfg.Append(filename)
		if err != nil {
			return nil, err
		}
	}

	return iniCfg, nil
}
