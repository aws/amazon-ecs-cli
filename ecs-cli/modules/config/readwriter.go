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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
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
func (rdwr *INIReadWriter) GetConfig(cliConfig *CLIConfig) error {

	// read old ini formatted file
	iniFormat := &iniCLIConfig{iniSectionKeys: new(iniSectionKeys)}
	if err := rdwr.cfg.MapTo(iniFormat); err != nil {
		return err
	}

	// If Prefixes not found, set to defaults.
	if !rdwr.IsKeyPresent(ecsSectionKey, composeProjectNamePrefixKey) {
		iniFormat.ComposeProjectNamePrefix = flags.ComposeProjectNamePrefixDefaultValue
	}
	if !rdwr.IsKeyPresent(ecsSectionKey, composeServiceNamePrefixKey) {
		iniFormat.ComposeServiceNamePrefix = flags.ComposeServiceNamePrefixDefaultValue
	}
	if !rdwr.IsKeyPresent(ecsSectionKey, cfnStackNamePrefixKey) {
		iniFormat.CFNStackNamePrefix = flags.CFNStackNamePrefixDefaultValue
	}

	// Convert to the new CliConfig
	cliConfig.Version = iniConfigVersion
	cliConfig.Cluster = iniFormat.Cluster
	cliConfig.Region = iniFormat.Region
	cliConfig.AWSProfile = iniFormat.AwsProfile
	cliConfig.AWSAccessKey = iniFormat.AWSAccessKey
	cliConfig.AWSSecretKey = iniFormat.AWSSecretKey
	cliConfig.ComposeServiceNamePrefix = iniFormat.ComposeServiceNamePrefix
	cliConfig.ComposeProjectNamePrefix = iniFormat.ComposeProjectNamePrefix
	cliConfig.CFNStackNamePrefix = iniFormat.CFNStackNamePrefix

	return nil
}

// IsKeyPresent returns true if the input key is present in the input section
func (rdwr *INIReadWriter) IsKeyPresent(section, key string) bool {
	return rdwr.cfg.Section(section).HasKey(key)
}

func newINIConfig(dest *Destination) (*ini.File, error) {
	iniCfg := ini.Empty()
	path := ConfigFilePath(dest)
	if _, err := os.Stat(path); err == nil {
		if err = iniCfg.Append(path); err != nil {
			return nil, errors.Wrap(err, "Failed to initialize ini reader")
		}
	}

	return iniCfg, nil
}
