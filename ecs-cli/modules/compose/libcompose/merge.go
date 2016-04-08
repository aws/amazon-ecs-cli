// This file is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found :
// https://github.com/docker/libcompose/blob/master/project/merge.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Modifications are Copyright 2015-2016 Amazon.com, Inc. or its affiliates. Licensed under the Apache License 2.0
// - Extracted local variables
// - Continued to use "ConfigLookup" instead of the new "ResourceLookup"
// - Instead of introducing a new struct RawService, passed the necessary config

package libcompose

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// GetEnvVarsFromConfig reads the environment variables from the compose file keys : environment and env_file
func GetEnvVarsFromConfig(context Context, inputCfg *ServiceConfig) ([]string, error) {

	envVars := inputCfg.Environment.Slice()
	envFiles := inputCfg.EnvFile.Slice()
	if len(envFiles) == 0 {
		return envVars, nil
	}

	composeFile := context.ComposeFile
	if context.ConfigLookup == nil {
		return nil, fmt.Errorf("Can not use env_file in file %s no mechanism provided to load files", composeFile)
	}

	for i := len(envFiles) - 1; i >= 0; i-- {
		envFile := envFiles[i]
		// Lookup envFile relative to the compose file path
		content, _, err := context.ConfigLookup.Lookup(envFile, composeFile)
		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(bytes.NewBuffer(content))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			key := strings.SplitAfter(line, "=")[0]

			found := false
			for _, v := range envVars {
				if strings.HasPrefix(v, key) {
					found = true
					break
				}
			}

			if !found {
				envVars = append(envVars, line)
			}
		}

		if scanner.Err() != nil {
			return nil, scanner.Err()
		}
	}

	return envVars, nil
}
