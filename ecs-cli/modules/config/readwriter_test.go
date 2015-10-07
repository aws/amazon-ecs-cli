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
	"io/ioutil"
	"os"
	"testing"
)

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

func TestNewConfigReadWriter(t *testing.T) {
	dest, err := newMockDestination()
	if err != nil {
		t.Fatal("Error creating mock config destination:", err)
	}
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

	if initialized {
		t.Errorf("Unexpected state: initialized set to true for empty config")
	}

	err = os.MkdirAll(dest.Path, *dest.Mode)
	if err != nil {
		t.Fatalf("Could not create config directory: ", err)
	}
	defer os.RemoveAll(dest.Path)

	clusterName := "test-cluster"
	// Craete a new config file
	newConfig := &CliConfig{&SectionKeys{Cluster: clusterName}}
	err = parser.ReadFrom(newConfig)
	if err != nil {
		t.Fatalf("Could not create config from struct", err)
	}

	err = parser.Save(dest)
	if err != nil {
		t.Fatalf("Could not save config file", err)
	}

	// Reinitialize from the writtern file.
	iniCfg, err = newIniConfig(dest)
	if err != nil {
		t.Fatal("Error creating config ini", err)
	}
	parser = &IniReadWriter{Destination: dest, cfg: iniCfg}
	initialized, err = parser.IsInitialized()
	if err != nil {
		t.Errorf("Error getting if initialized from ini", err)
	}

	if !initialized {
		t.Errorf("Unexpected state: initialized set to false for non-empty config")
	}

	readConfig, err := parser.GetConfig()
	if err != nil {
		t.Errorf("Error reading config:", err)
	}

	if clusterName != readConfig.Cluster {
		t.Errorf("Cluster name mismatch in config. Expected [%s] Got [%s]", clusterName, readConfig.Cluster)
	}
}
