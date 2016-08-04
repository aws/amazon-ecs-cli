// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"testing"
)

const testClusterName = "test-cluster"

func newMockDestination() (*Destination, error) {
	tmpPath, err := ioutil.TempDir(os.TempDir(), "ecs-cli-test-")
	if err != nil {
		return nil, err
	}

	mode, err := GetFilePermissions(tmpPath)
	if err != nil {
		return nil, err
	}

	return &Destination{Path: tmpPath, Mode: mode}, nil
}

func setupParser(t *testing.T, dest *Destination, shouldBeInitialized bool) *IniReadWriter {
	iniCfg, err := newIniConfig(dest)
	if err != nil {
		t.Fatal("Error creating config ini", err)
	}
	parser := &IniReadWriter{Destination: dest, cfg: iniCfg}

	// Test when unitialized.
	initialized, err := parser.IsInitialized()
	if err != nil {
		t.Errorf("Error getting if initialized from ini", err)
	}

	if shouldBeInitialized != initialized {
		t.Error("Unexpected state during parser initialization. Expected initialized to be [%s] but found [%s]", shouldBeInitialized, initialized)
	}

	return parser
}

func createConfig(t *testing.T, parser *IniReadWriter, dest *Destination) {
	// Create a new config file
	newConfig := &CliConfig{&SectionKeys{Cluster: testClusterName}}
	err := parser.ReadFrom(newConfig)
	if err != nil {
		t.Fatalf("Could not create config from struct", err)
	}

	err = parser.Save(dest)
	if err != nil {
		t.Fatalf("Could not save config file", err)
	}
}

func TestConfigPermissions(t *testing.T) {
	dest, err := newMockDestination()
	if err != nil {
		t.Fatal("Error creating mock config destination:", err)
	}
	parser := setupParser(t, dest, false)

	err = os.MkdirAll(dest.Path, *dest.Mode)
	if err != nil {
		t.Fatalf("Could not create config directory: ", err)
	}
	defer os.RemoveAll(dest.Path)

	// Create config file and confirm it has expected initial permissions
	createConfig(t, parser, dest)

	path := configPath(dest)
	confirmConfigMode(t, path, configFileMode)

	// Now set the config mode to something bad
	badMode := os.FileMode(0777)
	err = os.Chmod(path, badMode)
	if err != nil {
		t.Fatalf("Unable to change mode of new config %v", path)
	}
	confirmConfigMode(t, path, badMode)

	// Save the config and confirm it's fixed again
	err = parser.Save(dest)
	if err != nil {
		t.Fatalf("Unable to save to new config %v", path)
	}
	confirmConfigMode(t, path, configFileMode)
}

func confirmConfigMode(t *testing.T, path string, expected os.FileMode) {
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Unable to stat config file %s", path)
	}

	mode := info.Mode()
	if mode != expected {
		t.Fatalf("Mode of config %v not expected %v", mode, expected)
	}
}

func TestNewConfigReadWriter(t *testing.T) {
	dest, err := newMockDestination()
	if err != nil {
		t.Fatal("Error creating mock config destination:", err)
	}
	parser := setupParser(t, dest, false)

	err = os.MkdirAll(dest.Path, *dest.Mode)
	if err != nil {
		t.Fatalf("Could not create config directory: ", err)
	}
	defer os.RemoveAll(dest.Path)

	createConfig(t, parser, dest)

	// Reinitialize from the written file.
	parser = setupParser(t, dest, true)

	readConfig, err := parser.GetConfig()
	if err != nil {
		t.Errorf("Error reading config:", err)
	}

	if testClusterName != readConfig.Cluster {
		t.Errorf("Cluster name mismatch in config. Expected [%s] Got [%s]", testClusterName, readConfig.Cluster)
	}
	if !parser.IsKeyPresent(ecsSectionKey, composeProjectNamePrefixKey) || readConfig.ComposeProjectNamePrefix != "" {
		t.Errorf("Compose Project prefix name mismatch in config. Expected empty string Got [%s]", readConfig.ComposeProjectNamePrefix)
	}
	if !parser.IsKeyPresent(ecsSectionKey, composeServiceNamePrefixKey) || readConfig.ComposeServiceNamePrefix != "" {
		t.Errorf("Compose service name prefix mismatch in config. Expected empty string Got [%s]", readConfig.ComposeServiceNamePrefix)
	}
	if !parser.IsKeyPresent(ecsSectionKey, cfnStackNamePrefixKey) || readConfig.CFNStackNamePrefix != "" {
		t.Errorf("CFNStackNamePrefix mismatch in config. Expected empty string Got [%s]", readConfig.CFNStackNamePrefix)
	}
}

func TestMissingPrefixes(t *testing.T) {
	configContentsNoPrefixes := `[ecs]
cluster = test
aws_profile =
region = us-west-2
aws_access_key_id =
aws_secret_access_key =
`
	dest, err := newMockDestination()
	if err != nil {
		t.Fatal("Error creating mock config destination:", err)
	}
	err = os.MkdirAll(dest.Path, *dest.Mode)
	if err != nil {
		t.Fatalf("Could not create config directory: ", err)
	}
	defer os.RemoveAll(dest.Path)

	if err = ioutil.WriteFile(dest.Path+"/"+configFileName, []byte(configContentsNoPrefixes), *dest.Mode); err != nil {
		t.Fatal(err)
	}

	parser := setupParser(t, dest, true)
	readConfig, err := parser.GetConfig()
	if err != nil {
		t.Errorf("Error reading config:", err)
	}

	if parser.IsKeyPresent(ecsSectionKey, cfnStackNamePrefixKey) {
		t.Errorf("Expected key [%s] not to be present. Got value=[%s]", cfnStackNamePrefixKey, readConfig.CFNStackNamePrefix)
	}
	if parser.IsKeyPresent(ecsSectionKey, composeServiceNamePrefixKey) {
		t.Errorf("Expected key [%s] not to be present. Got value=[%s]", composeServiceNamePrefixKey, readConfig.ComposeServiceNamePrefix)
	}
	if parser.IsKeyPresent(ecsSectionKey, composeProjectNamePrefixKey) {
		t.Errorf("Expected key [%s] not to be present. Got value=[%s]", composeProjectNamePrefixKey, readConfig.ComposeProjectNamePrefix)
	}

}
