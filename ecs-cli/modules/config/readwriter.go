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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v2"

	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"

	"github.com/Sirupsen/logrus"
	"github.com/go-ini/ini"
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

	// Unfortunately we have to handle the old ini config
	if strings.HasPrefix(string(dat), "["+ecsSectionKey+"]") {
		// old ini config
		iniReadWriter, er := NewIniReadWriter()
		if er != nil {
			return nil, nil, err
		}
		to, err = iniReadWriter.GetConfig()
		if err != nil {
			return nil, nil, err
		}
		// we will need a map version to return, the yaml ReadWriter makes this Easy
		// but for ini we need to do it in a more annoying way
		// (converting the config value we read into yaml bytes)
		dat, err = yaml.Marshal(&to)
		if err != nil {
			return nil, nil, err
		}

	}

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

	configMap = configMap[ecsSectionKey].(map[interface{}]interface{})
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

//IsKeyPresent returns true if the input key is present in the input section
// func (rdwr *YamlReadWriter) IsKeyPresent(key string) (bool, error) {
// 	// read the raw bytes of the config file
// 	path := configPath(rdwr.destination)
// 	dat, err := ioutil.ReadFile(path)
// 	if err != nil {
// 		return false, err
// 	}
// 	// convert yaml to a map
// 	m := make(map[interface{}]interface{})
// 	err = yaml.Unmarshal(dat, &m)
// 	if err != nil {
// 		return false, err
// 	}
//
// 	_, ok := m[key]
//
// 	return ok, nil
//
// }

// IniReadWriter implments the ReadWriter interfaces.
// It can now only be used to load ecs-cli config.
// The config files is being migrated to use Yaml
// Sample ecs-cli config:
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

// oldCliConfig is the struct used to map to the ini config.
// This is to allow us to read old ini based config files
// CliConfig has been updated to use the yaml annotations
type oldCliConfig struct {
	*oldSectionKeys `ini:"ecs"`
}

// SectionKeys is the struct embedded in oldCliConfig. It groups all the keys in the 'ecs' section in the ini file.
type oldSectionKeys struct {
	Cluster                  string `ini:"cluster"`
	AwsProfile               string `ini:"aws_profile"`
	Region                   string `ini:"region"`
	AwsAccessKey             string `ini:"aws_access_key_id"`
	AwsSecretKey             string `ini:"aws_secret_access_key"`
	ComposeProjectNamePrefix string `ini:"compose-project-name-prefix"`
	ComposeServiceNamePrefix string `ini:"compose-service-name-prefix"`
	CFNStackNamePrefix       string `ini:"cfn-stack-name-prefix"`
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
func (rdwr *IniReadWriter) GetConfig() (*CliConfig, error) {
	//to := new(CliConfig)
	to := &CliConfig{SectionKeys: new(SectionKeys)}

	// read old ini formatted file
	oldFormat := new(oldCliConfig)
	err := rdwr.cfg.MapTo(oldFormat)
	if err != nil {
		return nil, err
	}

	// Convert to the new CliConfig
	// Unfortunately we have to handle putting in the default values here
	// thanfully, this code will eventually be removed one day
	// (when we no longer support the old ini configs)
	// If Prefixes not found, set to defaults.
	if !rdwr.IsKeyPresent(ecsSectionKey, composeProjectNamePrefixKey) {
		oldFormat.ComposeProjectNamePrefix = ecscli.ComposeProjectNamePrefixDefaultValue
	}
	if !rdwr.IsKeyPresent(ecsSectionKey, composeServiceNamePrefixKey) {
		oldFormat.ComposeServiceNamePrefix = ecscli.ComposeServiceNamePrefixDefaultValue
	}
	if !rdwr.IsKeyPresent(ecsSectionKey, cfnStackNamePrefixKey) {
		oldFormat.CFNStackNamePrefix = ecscli.CFNStackNamePrefixDefaultValue
	}
	to.Cluster = oldFormat.Cluster
	to.Region = oldFormat.Region
	to.AwsProfile = oldFormat.AwsProfile
	to.AwsAccessKey = oldFormat.AwsAccessKey
	to.AwsSecretKey = oldFormat.AwsSecretKey
	to.ComposeProjectNamePrefix = oldFormat.ComposeProjectNamePrefix
	to.ComposeServiceNamePrefix = oldFormat.ComposeServiceNamePrefix
	to.CFNStackNamePrefix = oldFormat.CFNStackNamePrefix
	logrus.Warn("to: " + fmt.Sprintf("%#v", to.SectionKeys) + " \noldFormat: " + fmt.Sprintf("%#v", oldFormat.oldSectionKeys))
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
// NOTE: This method should not be used anymore since we have moved to YAML.
// IniReadWriter is only used to read old ini formmatted config files; all
// writes should be to YAML formatted files.
func (rdwr *IniReadWriter) ReadFrom(ecsConfig *CliConfig) error {
	return rdwr.cfg.ReflectFrom(ecsConfig)
}

// Save saves the config to a config file.
// NOTE: This method should not be used anymore since we have moved to YAML.
// IniReadWriter is only used to read old ini formmatted config files; all
// writes should be to YAML formatted files.
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
