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
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils"
)

// Destination stores the config destination path to write to and the permissions to create the
// ecs config directory if none exists.
type Destination struct {
	Path string
	Mode *os.FileMode
}

// GetFilePermissions is a utility method that gets permissions of a file.
func GetFilePermissions(fileName string) (*os.FileMode, error) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		logrus.Warnf("Error getting permissions of file: %s", fileName)
		return nil, err
	}

	mode := fileInfo.Mode()
	return &mode, nil
}

// newDefaultDestination creates a new Destination object.
func newDefaultDestination() (*Destination, error) {
	homeDir, err := utils.GetHomeDir()
	if err != nil {
		return nil, err
	}
	mode, err := GetFilePermissions(homeDir)
	if err != nil {
		return nil, err
	}

	// TODO: Move to const.
	return &Destination{Path: filepath.Join(homeDir, ".ecs"), Mode: mode}, nil
}
