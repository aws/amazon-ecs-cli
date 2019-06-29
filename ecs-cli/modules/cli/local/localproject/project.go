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

// Package localproject defines LocalProject interface and implements them on localProject
package localproject

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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
	LocalOutDefaultFileName = "docker-compose.local.yml"

	// LocalOutFileMode represents the file can be read/write by its owner.
	LocalOutFileMode = os.FileMode(0600) // Owner=read/write, Other=none

	// LocalInFileName represents the default local file name for task definition JSON.
	LocalInFileName = "task-definition.json"
)

// Interface for a local project, holding data needed to convert an ECS Task Definition to a Docker Compose file
type LocalProject interface {
	ReadTaskDefinition() error
	Convert() error
	Write() error
	LocalOutFileName() string
	TaskDefinition() *ecs.TaskDefinition
	InputMetadata() *converter.LocalCreateMetadata
}

type localProject struct {
	context          *cli.Context
	taskDefinition   *ecs.TaskDefinition
	localBytes       []byte
	localOutFileName string
	inputMetadata    *converter.LocalCreateMetadata
}

// New instantiates a new Local Project
func New(context *cli.Context) LocalProject {
	p := &localProject{context: context}
	return p
}

// TaskDefinition returns the ECS task definition to be converted
func (p *localProject) TaskDefinition() *ecs.TaskDefinition {
	return p.taskDefinition
}

// LocalOutFileName returns name of compose file output by local.Create
func (p *localProject) LocalOutFileName() string {
	return p.localOutFileName
}

// InputMetadata returns the metadata on the task definition used to create the docker compose file
func (p *localProject) InputMetadata() *converter.LocalCreateMetadata {
	return p.inputMetadata
}

// ReadTaskDefinition reads an ECS Task Definition either from a local file
// or from retrieving one from ECS and stores it on the local project
func (p *localProject) ReadTaskDefinition() error {
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
		taskDefinition, err = p.readTaskDefinitionFromFile(filename)
		if err != nil {
			return err
		}

	} else if defaultInputExists() {
		taskDefinition, err = p.readTaskDefinitionFromFile(LocalInFileName)
		if err != nil {
			return err
		}
	}

	if taskDefinition == nil {
		return fmt.Errorf("Could not detect valid Task Definition")
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

func (p *localProject) readTaskDefinitionFromFile(filename string) (*ecs.TaskDefinition, error) {
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

func (p *localProject) readTaskDefinitionFromRemote(remote string) (*ecs.TaskDefinition, error) {
	p.inputMetadata = &converter.LocalCreateMetadata{
		InputType: RemoteTaskDefType,
		Value:     remote,
	}
	return readTaskDefFromRemote(remote, p)
}

var readTaskDefFromRemote = func(remote string, p *localProject) (*ecs.TaskDefinition, error) {
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
func (p *localProject) Convert() error {
	data, err := converter.ConvertToDockerCompose(p.taskDefinition, p.inputMetadata)

	if err != nil {
		return err
	}

	p.localBytes = data

	return nil
}

// Write writes the compose data to a local compose file. The output filename is stored on the project
func (p *localProject) Write() error {
	// Will error if the file already exists, otherwise create
	p.localOutFileName = LocalOutDefaultFileName

	if fileName := p.context.String(flags.Output); fileName != "" {
		p.localOutFileName = fileName
	}

	return p.writeFile()
}

func (p *localProject) writeFile() error {
	out, err := openFile(p.localOutFileName)
	defer out.Close()

	// File already exists
	if err != nil {
		return p.overwriteFile()
	}

	_, err = out.Write(p.localBytes)

	return err
}

// Facilitates test mocking
var openFile func(filename string) (*os.File, error) = func(filename string) (*os.File, error) {
	return os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, LocalOutFileMode)
}

func (p *localProject) overwriteFile() error {
	filename := p.localOutFileName

	fmt.Printf("%s file already exists. Do you want to write over this file? [y/N]\n", filename)

	reader := bufio.NewReader(os.Stdin)
	stdin, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("Error reading stdin: %s", err.Error())
	}

	input := strings.ToLower(strings.TrimSpace(stdin))

	if input != "yes" && input != "y" {
		return fmt.Errorf("Aborted writing compose file. To retry, rename or move %s", filename) // TODO add force flag
	}

	// Overwrite local compose file
	return ioutil.WriteFile(filename, p.localBytes, LocalOutFileMode)
}
