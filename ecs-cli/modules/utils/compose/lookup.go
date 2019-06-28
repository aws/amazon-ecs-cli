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

package utils

import (
	"github.com/docker/libcompose/lookup"
)

// GetDefaultResourceLookup returns the default Lookup mechanism for resources.
// This implements a function to load a file relative to a given path. This is used to load
// files specified in env_file option, for example.
func GetDefaultResourceLookup() (*lookup.FileResourceLookup, error) {
	return &lookup.FileResourceLookup{}, nil
}
