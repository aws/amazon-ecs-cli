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
	iniConfigFileName = "config"
)

// INIReadWriter
// NOTE: DEPRECATED. These functions are only left here so that
// we can read old ini based config files for customers who have
// still been using older versions of the CLI. All new config files
// will be written in the YAML format. INIReadWriter can now only
// be used to load ecs-cli config.
// Sample old ini ecs-cli config:
// [ecs]
// cluster = test
// aws_profile =
// region = us-west-2
// aws_access_key_id =
// aws_secret_access_key =
// compose-project-name-prefix = ecscompose-
// compose-service-name-prefix =
// cfn-stack-name-prefix = ecs-cli-
type INIReadWriter struct {
	*Destination
	cfg *ini.File
}

// NewINIReadWriter creates a new Ini Parser object for the old ini configs
func NewINIReadWriter(dest *Destination) (*INIReadWriter, error) {

	iniCfg, err := newINIConfig(dest)
	if err != nil {
		return nil, err
	}

	return &INIReadWriter{Destination: dest, cfg: iniCfg}, nil
}

// GetConfig gets the ecs-cli config object from the config file.
// map contains the keys that are present in the config file (maps string field name to string field value)
// map is type map[interface{}]interface{} to ensure fowards compatibility with changes that will
// cause certain keys to be mapped to maps of keys
func (rdwr *INIReadWriter) GetConfig() (*CLIConfig, map[interface{}]interface{}, error) {
	configMap := make(map[interface{}]interface{})
	to := &CLIConfig{}

	// read old ini formatted file
	oldFormat := &iniCLIConfig{iniSectionKeys: new(iniSectionKeys)}
	err := rdwr.cfg.MapTo(oldFormat)
	if err != nil {
		return nil, nil, err
	}

	// Create the configMap
	if rdwr.IsKeyPresent(ecsSectionKey, "cluster") {
		configMap["cluster"] = oldFormat.Cluster
	}
	if rdwr.IsKeyPresent(ecsSectionKey, "aws_profile") {
		configMap["aws_profile"] = oldFormat.AwsProfile
	}
	if rdwr.IsKeyPresent(ecsSectionKey, "region") {
		configMap["region"] = oldFormat.Region
	}
	if rdwr.IsKeyPresent(ecsSectionKey, "aws_access_key_id") {
		configMap["aws_access_key_id"] = oldFormat.AwsAccessKey
	}
	if rdwr.IsKeyPresent(ecsSectionKey, "aws_secret_access_key") {
		configMap["aws_secret_access_key"] = oldFormat.AwsSecretKey
	}
	if rdwr.IsKeyPresent(ecsSectionKey, "compose-project-name-prefix") {
		configMap["compose-project-name-prefix"] = oldFormat.ComposeProjectNamePrefix
	}
	if rdwr.IsKeyPresent(ecsSectionKey, "compose-service-name-prefix") {
		configMap["compose-service-name-prefix"] = oldFormat.ComposeServiceNamePrefix
	}
	if rdwr.IsKeyPresent(ecsSectionKey, "cfn-stack-name-prefix") {
		configMap["cfn-stack-name-prefix"] = oldFormat.CFNStackNamePrefix
	}

	// Convert to the new CliConfig
	to.Cluster = oldFormat.Cluster
	to.Region = oldFormat.Region
	to.AWSProfile = oldFormat.AwsProfile
	to.AWSAccessKey = oldFormat.AwsAccessKey
	to.AWSSecretKey = oldFormat.AwsSecretKey
	to.ComposeProjectNamePrefix = oldFormat.ComposeProjectNamePrefix
	to.ComposeServiceNamePrefix = oldFormat.ComposeServiceNamePrefix
	to.CFNStackNamePrefix = oldFormat.CFNStackNamePrefix
	return to, configMap, nil
}

// IsInitialized returns true if the config file can be read and if all the key config fields
// have been initialized.
func (rdwr *INIReadWriter) IsInitialized() (bool, error) {
	to := new(CLIConfig)
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
func (rdwr *INIReadWriter) IsKeyPresent(section, key string) bool {
	return rdwr.cfg.Section(section).HasKey(key)
}

func newINIConfig(dest *Destination) (*ini.File, error) {
	iniCfg := ini.Empty()
	path := iniConfigPath(dest)
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

func iniConfigPath(dest *Destination) string {
	return filepath.Join(dest.Path, iniConfigFileName)
}
