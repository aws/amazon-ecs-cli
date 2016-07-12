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

package utils

import (
	"os"
	"path/filepath"

	lConfig "github.com/docker/libcompose/config"
	"github.com/docker/libcompose/lookup"
)

// GetDefaultEnvironmentLookup returns the default Lookup mechanism for environment variables.
// Order of resolution:
// 1. Environment values specified (in the form of 'key=value') in a '.env' file in the current working directory.
// 2. Environment values specified in the shell (using os.Getenv). If the os environment variable does not exists,
//    the slice is empty, and the environment variable is skipped.
func GetDefaultEnvironmentLookup() (*lookup.ComposableEnvLookup, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return &lookup.ComposableEnvLookup{
		Lookups: []lConfig.EnvironmentLookup{
			&lookup.EnvfileLookup{
				Path: filepath.Join(cwd, ".env"),
			},
			&lookup.OsEnvLookup{},
		},
	}, nil
}

// GetDefaultResourceLookup returns the default Lookup mechanism for resources.
// This implements a function to load a file relative to a given path. This is used to load
// files specified in env_file option, for example.
func GetDefaultResourceLookup() (*lookup.FileConfigLookup, error) {
	return &lookup.FileConfigLookup{}, nil
}
