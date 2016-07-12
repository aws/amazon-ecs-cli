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

package cache

import (
	"encoding/gob"
	"os"
	"path/filepath"

	"github.com/aws/amazon-ecs-cli/ecs-cli/utils"
)

// available for injecting mocks for testing
var osOpen = os.Open
var osCreate = os.Create
var osMkdirAll = os.MkdirAll

// rw only for the user in-line with how most programs setup their cache
// directories. Also, this directory could container sensitive-ish files due to
// environment variables, so keep it user-only.
const cacheDirMode = 0700
const cachePrefix = "ecs-cli"

// cacheDir is a helper function to return the 'cache' directory for a given
// application name
func cacheDir(name string) (string, error) {
	homedir, err := utils.GetHomeDir()
	if err != nil {
		return "", err
	}
	// TODO, speciailize for windows, possibly OS X
	return filepath.Join(homedir, ".cache", cachePrefix, name), nil
}

type fsCache struct {
	name     string
	cacheDir string
}

// NewFSCache returns a new cache backed by the filesystem. The 'name' value
// should be constant in order to access the same data between uses.
// The cache will be namespaced under this project
func NewFSCache(name string) (Cache, error) {
	dir, err := cacheDir(name)
	if err != nil {
		return nil, err
	}

	err = osMkdirAll(dir, cacheDirMode)
	if err != nil {
		return nil, err
	}

	return &fsCache{
		name:     name,
		cacheDir: dir,
	}, nil
}

func (self *fsCache) Put(key string, val interface{}) (retErr error) {
	file, err := osCreate(filepath.Join(self.cacheDir, key))
	if err != nil {
		return err
	}
	defer func() {
		closeErr := file.Close()
		// Avoid masking the 'gob.Encode' error, it's earlier and probably more specific
		if closeErr != nil && retErr == nil {
			// named return error
			retErr = closeErr
		}
	}()
	valEnc := gob.NewEncoder(file)
	return valEnc.Encode(val)
}

func (self *fsCache) Get(key string, i interface{}) error {
	file, err := osOpen(filepath.Join(self.cacheDir, key))
	if err != nil {
		return err
	}
	defer file.Close()
	valDec := gob.NewDecoder(file)
	return valDec.Decode(i)
}
