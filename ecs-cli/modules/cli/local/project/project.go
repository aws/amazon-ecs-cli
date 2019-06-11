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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/secretsmanager"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/urfave/cli"
)

const (
	LocalOutDefaultFileName = "./docker-compose.local.yml"
	LocalOutFileMode        = os.FileMode(0644) // Owner=read/write, Other=readonly
	LocalInFileName         = "./task-definition.json"
)

// Interface for a local project, holding data needed to convert an ECS Task Definition to a Docker Compose file
type LocalProject interface {
	ReadTaskDefinition() error
	Convert() error
	Write() error
	LocalOutFileName() string
}

type localProject struct {
	context          *cli.Context
	taskDefinition   *ecs.TaskDefinition
	localBytes       []byte
	localOutFileName string
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

// ReadTaskDefinition reads an ECS Task Definition either from a local file
// or from retrieving one from ECS and stores it on the local project
func (p *localProject) ReadTaskDefinition() error {
	arn := p.context.String(flags.TaskDefinitionArnFlag)
	filename := p.context.String(flags.TaskDefinitionFileFlag)

	if arn != "" && filename != "" {
		return fmt.Errorf("Cannot specify both --%s and --%s flags.", flags.TaskDefinitionArnFlag, flags.TaskDefinitionFileFlag)
	}

	var taskDefinition *ecs.TaskDefinition
	var err error

	if arn != "" {
		taskDefinition, err = p.readTaskDefinitionFromArn(arn)
		if err != nil {
			return err
		}
	} else if filename != "" {
		taskDefinition, err = p.readTaskDefinitionFromFile(filename)
		if err != nil {
			return err
		}
	} else if _, err := os.Stat(LocalInFileName); err == nil {
		// Try reading local task-definition.json file by default
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

func (p *localProject) readTaskDefinitionFromFile(filename string) (*ecs.TaskDefinition, error) {
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

var newCommandConfig = func(context *cli.Context, rdwr config.ReadWriter) (*config.CommandConfig, error) {
	return config.NewCommandConfig(context, rdwr)
}

// FIXME: NOTE this will actually read from either ARN or Task Definition family name
func (p *localProject) readTaskDefinitionFromArn(arn string) (*ecs.TaskDefinition, error) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		return nil, err
	}

	commandConfig, err := newCommandConfig(p.context, rdwr)
	if err != nil {
		return nil, err
	}

	ecsClient := ecsclient.NewECSClient(commandConfig)

	return ecsClient.DescribeTaskDefinition(arn)
}

// Convert translates an ECS Task Definition into a Compose V3 schema and
// stores the data on the project
func (p *localProject) Convert() error {
	// FIXME get secrets here, pass to converter?
	data, err := converter.ConvertToDockerCompose(p.taskDefinition)

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
	if fileName := p.context.String(flags.LocalOutputFlag); fileName != "" {
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

// Get secret value stored in AWS Secrets Manager
// TODO apply to each container
func (p *localProject) getSecret(secretName string) (string, error) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		return "", err
	}

	commandConfig, err := newCommandConfig(p.context, rdwr)
	if err != nil {
		return "", err
	}

	client := secretsmanager.NewSecretsManagerClient(commandConfig)

	secret, err := client.GetSecretValue(secretName)
	if err != nil {
		return "", err
	}

	return secret, nil
}
