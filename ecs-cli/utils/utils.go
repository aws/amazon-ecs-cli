// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package utils provides some utility functions. This is the kitchen sink.
package utils

import (
	"fmt"
	"os"
)

// GetHomeDir returns the file path of the user's home directory.
func GetHomeDir() (string, error) {
	// Can not use user.Current https://github.com/golang/go/issues/6376
	homeDir := os.Getenv("HOME") // *nix
	if homeDir == "" {           // Windows
		homeDir = os.Getenv("USERPROFILE")
	}

	if homeDir == "" {
		return "", fmt.Errorf("user home directory not found")
	}

	return homeDir, nil
}
