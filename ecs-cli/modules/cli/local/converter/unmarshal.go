// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package converter

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
)

// UnmarshalComposeFile decodes a Docker Compose file.
func UnmarshalComposeFile(filename string) (*types.Config, error) {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read Compose file %s", filename)
	}

	parsed, err := loader.ParseYAML(raw)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse bytes from %s", filename)
	}

	wd, _ := os.Getwd()
	details := types.ConfigDetails{
		WorkingDir: wd,
		ConfigFiles: []types.ConfigFile{
			{
				Filename: filename,
				Config:   parsed,
			},
		},
		Environment: getEnvironment(),
	}
	compose, err := loader.Load(details)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load Compose file with details %v", details)
	}
	return compose, nil
}

func getEnvironment() map[string]string {
	envs := os.Environ()
	m := make(map[string]string, len(envs))
	for _, env := range envs {
		parts := strings.SplitN(env, "=", 2)
		m[parts[0]] = parts[1]
	}
	return m
}
