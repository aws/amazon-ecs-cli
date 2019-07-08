// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Package localproject provides functionality to retrieve a task definition, convert it to a Docker Compose config, and write it to a file.
package localproject

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/converter"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	arnParser "github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	// LocalTaskDefType indicates if the task definition is read from a local file
	LocalTaskDefType = "local"

	// RemoteTaskDefType indicates if the task definition is retrieved from ECS via ARN or name
	RemoteTaskDefType = "remote"
)

const (
	// LocalOutDefaultFileName represents the default name for the output Docker
	// Compose file.
	LocalOutDefaultFileName = "docker-compose.ecs-local.yml"

	// LocalOutFileMode represents the file can be read/write by its owner.
	LocalOutFileMode = os.FileMode(0600) // Owner=read/write, Other=none

	// LocalInFileName represents the default local file name for task definition JSON.
	LocalInFileName = "task-definition.json"
)

// Indexes of Compose files
const (
	baseComposeIndex = iota
	overrideComposeIndex
)

// LocalProject holds data needed to convert an ECS Task Definition to Docker Compose files.
type LocalProject struct {
	context        *cli.Context
	taskDefinition *ecs.TaskDefinition
	composeBytes   [][]byte
	inputMetadata  *converter.LocalCreateMetadata
}

// New instantiates a new Local Project
func New(context *cli.Context) *LocalProject {
	return &LocalProject{
		context: context,
		composeBytes: [][]byte{
			nil,
			nil,
		},
	}
}

// TaskDefinition returns the ECS task definition to be converted
func (p *LocalProject) TaskDefinition() *ecs.TaskDefinition {
	return p.taskDefinition
}

// LocalOutFileName returns name of compose file output by local.Create
func (p *LocalProject) LocalOutFileName() string {
	if customName := p.context.String(flags.Output); customName != "" {
		return customName
	}
	return LocalOutDefaultFileName
}

func (p *LocalProject) LocalOutFileFullPath() (string, error) {
	return filepath.Abs(p.LocalOutFileName())
}

// OverrideFileName returns the name of the override Compose file.
func (p *LocalProject) OverrideFileName() string {
	baseName := p.LocalOutFileName()
	baseExt := filepath.Ext(baseName)
	return baseName[:len(baseName)-len(baseExt)] + ".override.yml"
}

// InputMetadata returns the metadata on the task definition used to create the docker compose file
func (p *LocalProject) InputMetadata() *converter.LocalCreateMetadata {
	return p.inputMetadata
}

// ReadTaskDefinition reads an ECS Task Definition either from a local file
// or from retrieving one from ECS and stores it on the local project
func (p *LocalProject) ReadTaskDefinition() error {
	remote := p.context.String(flags.TaskDefinitionRemote)
	filename := p.context.String(flags.TaskDefinitionFile)

	if remote != "" && filename != "" {
		return fmt.Errorf("cannot specify both --%s and --%s flags", flags.TaskDefinitionRemote, flags.TaskDefinitionFile)
	}

	var taskDefinition *ecs.TaskDefinition
	var err error

	if remote != "" {
		taskDefinition, err = p.readTaskDefinitionFromRemote(remote)
		logrus.Infof("Reading task definition from %s:%v\n", aws.StringValue(taskDefinition.Family), aws.Int64Value(taskDefinition.Revision))
		if err != nil {
			return err
		}
	} else if filename != "" {
		filename, err = filepath.Abs(filename)
		if err != nil {
			return err
		}

		taskDefinition, err = p.readTaskDefinitionFromFile(filename)
		if err != nil {
			return err
		}

	} else if defaultInputExists() {
		filename, err = filepath.Abs(LocalInFileName)
		if err != nil {
			return err
		}

		taskDefinition, err = p.readTaskDefinitionFromFile(filename)
		if err != nil {
			return err
		}
	}

	if taskDefinition == nil {
		return fmt.Errorf(fmt.Sprintf("could not detect valid task definition.\nEither set one of --%s or --%s flags, or define a %s file.",
			flags.TaskDefinitionFile, flags.TaskDefinitionRemote, LocalInFileName))
	}

	p.taskDefinition = taskDefinition
	return nil
}

var defaultInputExists = func() bool {
	if _, err := os.Stat(LocalInFileName); err == nil {
		return true
	}
	return false
}

func (p *LocalProject) readTaskDefinitionFromFile(filename string) (*ecs.TaskDefinition, error) {
	p.inputMetadata = &converter.LocalCreateMetadata{
		InputType: LocalTaskDefType,
		Value:     filename,
	}
	return readTaskDefFromLocal(filename)
}

var readTaskDefFromLocal = func(filename string) (*ecs.TaskDefinition, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("Error reading task definition from %s: %s", filename, err.Error())
	}

	taskDefinition := ecs.TaskDefinition{}
	err = json.Unmarshal(bytes, &taskDefinition)
	if err != nil {
		return nil, fmt.Errorf("Error parsing task definition JSON: %s", err.Error())
	}

	return &taskDefinition, nil
}

var newCommandConfig = func(context *cli.Context, rdwr config.ReadWriter, region string) (*config.CommandConfig, error) {
	if region != "" {
		return config.NewCommandConfigWithRegion(context, rdwr, region)
	}
	return config.NewCommandConfig(context, rdwr)
}

func (p *LocalProject) readTaskDefinitionFromRemote(remote string) (*ecs.TaskDefinition, error) {
	p.inputMetadata = &converter.LocalCreateMetadata{
		InputType: RemoteTaskDefType,
		Value:     remote,
	}
	return readTaskDefFromRemote(remote, p)
}

var readTaskDefFromRemote = func(remote string, p *LocalProject) (*ecs.TaskDefinition, error) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		return nil, err
	}

	region := ""
	if parsedArn, err := arnParser.Parse(remote); err == nil {
		region = parsedArn.Region
	}

	commandConfig, err := newCommandConfig(p.context, rdwr, region)
	if err != nil {
		return nil, err
	}

	ecsClient := ecsclient.NewECSClient(commandConfig)

	return ecsClient.DescribeTaskDefinition(remote)
}

// Convert translates an ECS Task Definition into a Compose V3 schema and
// stores the data on the project
func (p *LocalProject) Convert() error {
	if err := p.convertBaseCompose(); err != nil {
		return err
	}
	if err := p.convertOverrideCompose(); err != nil {
		return err
	}
	return nil
}

func (p *LocalProject) convertBaseCompose() error {
	conf, err := converter.ConvertToComposeConfig(p.taskDefinition, p.inputMetadata)
	if err != nil {
		return err
	}
	data, err := converter.MarshalComposeConfig(*conf, p.LocalOutFileName())
	if err != nil {
		return err
	}
	p.composeBytes[baseComposeIndex] = data
	return nil
}

func (p *LocalProject) convertOverrideCompose() error {
	conf, err := converter.ConvertToComposeOverride(p.taskDefinition)
	if err != nil {
		return err
	}
	data, err := converter.MarshalComposeConfig(*conf, p.OverrideFileName())
	if err != nil {
		return err
	}
	p.composeBytes[overrideComposeIndex] = data
	return nil
}

// Write writes the compose data to a local compose file. The output filename is stored on the project
func (p *LocalProject) Write() error {
	if err := writeFile(p.LocalOutFileName(), p.composeBytes[baseComposeIndex]); err != nil {
		return err
	}

	if err := writeFileIfNotExist(p.OverrideFileName(), p.composeBytes[overrideComposeIndex]); err != nil {
		return err
	}

	return nil
}

func writeFile(filename string, content []byte) error {
	f, err := openFile(filename)
	defer f.Close()

	// File already exists
	if err != nil {
		return overwriteFile(filename, content)
	}

	_, err = f.Write(content)
	logrus.Infof("Successfully wrote %s", filename)
	return err
}

func writeFileIfNotExist(filename string, content []byte) error {
	f, err := openFile(filename)
	defer f.Close()

	if err != nil {
		// File already exists, skip writing it.
		logrus.Infof("%s already exists, skipping write.", filename)
		return nil
	}
	_, err = f.Write(content)
	logrus.Infof("Successfully wrote %s", filename)
	return err
}

// Facilitates test mocking
var openFile = func(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, LocalOutFileMode)
}

func overwriteFile(filename string, content []byte) error {
	fmt.Printf("%s file already exists. Do you want to write over this file? [y/N]\n", filename)

	reader := bufio.NewReader(os.Stdin)
	stdin, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed reading stdin: %s", err.Error())
	}

	input := strings.ToLower(strings.TrimSpace(stdin))

	if input != "yes" && input != "y" {
		return fmt.Errorf("aborted writing compose file. To retry, rename or move %s", filename) // TODO add force flag
	}

	// Overwrite local compose file
	return ioutil.WriteFile(filename, content, LocalOutFileMode)
}
