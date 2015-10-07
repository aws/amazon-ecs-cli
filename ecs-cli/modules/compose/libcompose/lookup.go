// This file is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found in the lookup package of libcompose :
// https://github.com/docker/libcompose/blob/master/lookup/file.go
// https://github.com/docker/libcompose/blob/master/lookup/simple_env.go
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
// Modifications are Copyright 2015 Amazon.com, Inc. or its affiliates. Licensed under the Apache License 2.0
package libcompose

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
)

// FileConfigLookup is a "bare" structure that implements the project.ConfigLookup interface
type FileConfigLookup struct {
}

// Lookup returns the content and the actual filename of the file that is "built" using the
// specified file and relativeTo string. file and relativeTo are supposed to be file path.
// If file starts with a slash ('/'), it tries to load it, otherwise it will build a
// filename using the folder part of relativeTo joined with file.
func (f *FileConfigLookup) Lookup(file, relativeTo string) ([]byte, string, error) {
	if strings.HasPrefix(file, "/") {
		logrus.Debugf("Reading file %s", file)
		bytes, err := ioutil.ReadFile(file)
		return bytes, file, err
	}

	fileName := path.Join(path.Dir(relativeTo), file)
	logrus.Debugf("Reading file %s relative to %s", fileName, relativeTo)
	bytes, err := ioutil.ReadFile(fileName)
	return bytes, fileName, err
}

// OsEnvLookup is a "bare" structure that implements the project.EnvironmentLookup interface
type OsEnvLookup struct {
}

// Lookup creates a string slice of string containing a "docker-friendly" environment string
// in the form of 'key=value'. It gets environment values using os.Getenv.
// If the os environment variable does not exists, the slice is empty. serviceName and config
// are not used at all in this implementation.
func (o *OsEnvLookup) Lookup(key, serviceName string, config *ServiceConfig) []string {
	ret := os.Getenv(key)
	if ret == "" {
		return []string{}
	} else {
		return []string{fmt.Sprintf("%s=%s", key, ret)}
	}
}
