// NOTE: DEPRECATED. These functions are only left here so that
// we can read old ini based config files for customers who have
// still been using older versions of the CLI. All new config files
// will be written in the YAML format.

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

	"github.com/Sirupsen/logrus"
	"github.com/go-ini/ini"
)

// IniReadWriter
// It can now only be used to load ecs-cli config.
// The config files is being migrated to use Yaml
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

type IniReadWriter struct {
	*Destination
	cfg *ini.File
}

// NewIniReadWriter creates a new Ini Parser object for the old ini configs
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
func (rdwr *IniReadWriter) GetConfig() (*CliConfig, map[interface{}]interface{}, error) {
	configMap := make(map[interface{}]interface{})
	to := &CliConfig{SectionKeys: new(SectionKeys)}

	// read old ini formatted file
	oldFormat := &oldCliConfig{oldSectionKeys: new(oldSectionKeys)}
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
	to.AwsProfile = oldFormat.AwsProfile
	to.AwsAccessKey = oldFormat.AwsAccessKey
	to.AwsSecretKey = oldFormat.AwsSecretKey
	to.ComposeProjectNamePrefix = oldFormat.ComposeProjectNamePrefix
	to.ComposeServiceNamePrefix = oldFormat.ComposeServiceNamePrefix
	to.CFNStackNamePrefix = oldFormat.CFNStackNamePrefix
	return to, configMap, nil
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
