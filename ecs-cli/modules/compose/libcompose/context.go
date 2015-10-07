// This file is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found :
// https://github.com/docker/libcompose/blob/master/project/context.go
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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
)

var projectRegexp = regexp.MustCompile("[^a-zA-Z0-9_.-]")

// Context encapsulates the configurations needed while setting up/running containers
type Context struct {
	Timeout             int
	Log                 bool
	Signal              string
	ComposeFile         string
	ComposeBytes        []byte
	ProjectName         string
	isOpen              bool
	EnvironmentLookup   EnvironmentLookup
	ConfigLookup        ConfigLookup
	IgnoreMissingConfig bool
}

func (c *Context) readComposeFile() error {
	if c.ComposeBytes != nil {
		return nil
	}

	logrus.Debugf("Opening compose file: %s", c.ComposeFile)

	if c.ComposeFile == "-" {
		if composeBytes, err := ioutil.ReadAll(os.Stdin); err != nil {
			logrus.Errorf("Failed to read compose file from stdin: %v", err)
			return err
		} else {
			c.ComposeBytes = composeBytes
		}
	} else if c.ComposeFile != "" {
		if composeBytes, err := ioutil.ReadFile(c.ComposeFile); os.IsNotExist(err) {
			if c.IgnoreMissingConfig {
				return nil
			}
			logrus.Errorf("Failed to find %s", c.ComposeFile)
			return err
		} else if err != nil {
			logrus.Errorf("Failed to open %s", c.ComposeFile)
			return err
		} else {
			c.ComposeBytes = composeBytes
		}
	}

	return nil
}

func (c *Context) determineProject() error {
	name, err := c.lookupProjectName()
	if err != nil {
		return err
	}

	c.ProjectName = projectRegexp.ReplaceAllString(strings.ToLower(name), "-")

	if c.ProjectName == "" {
		return fmt.Errorf("Falied to determine project name")
	}

	if strings.ContainsAny(c.ProjectName[0:1], "_.-") {
		c.ProjectName = "x" + c.ProjectName
	}

	return nil
}

func (c *Context) lookupProjectName() (string, error) {
	if c.ProjectName != "" {
		return c.ProjectName, nil
	}

	if envProject := os.Getenv("COMPOSE_PROJECT_NAME"); envProject != "" {
		return envProject, nil
	}

	f, err := filepath.Abs(c.ComposeFile)
	if err != nil {
		logrus.Errorf("Failed to get absolute directory for: %s", c.ComposeFile)
		return "", err
	}

	// NOTE: Commenting out the conversion to UnixPath to enable its functioning in Windows
	// f = toUnixPath(f)

	parent := path.Base(path.Dir(f))
	if parent != "" && parent != "." {
		return parent, nil
	} else if wd, err := os.Getwd(); err != nil {
		return "", err
	} else {
		return path.Base(toUnixPath(wd)), nil
	}
}

func toUnixPath(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}

func (c *Context) Open() error {
	if c.isOpen {
		return nil
	}

	if err := c.readComposeFile(); err != nil {
		return err
	}

	if err := c.determineProject(); err != nil {
		return err
	}

	c.isOpen = true
	return nil
}
