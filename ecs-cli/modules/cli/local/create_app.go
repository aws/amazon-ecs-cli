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

// Package local implements the subcommands to run ECS task definitions locally
// (See: https://github.com/aws/containers-roadmap/issues/180).
package local

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/aws-sdk-go/service/ecs"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	LocalOutFileName = "./docker-compose.local.yml"
	LocalOutFileMode = os.FileMode(0644) // Owner=read/write, Other=readonly
	LocalInFileName = "./task-definition.json"
)

func Create(c *cli.Context) {
	err := createLocal(c)
	if err != nil {
		log.Fatalf("Error with local create: %s", err.Error())
	}

	fmt.Printf("Successfully created %s\n", LocalOutFileName)
}

func createLocal(c *cli.Context) error {
	// Read task definition (from file or ARN)
	// returns ecs.TaskDefinition
	taskDefinition, err := readTaskDefinition(c)
	fmt.Printf("TASK DEF THAT I READ: %+v\n", taskDefinition)
	if err != nil {
		return err
	}

	// Convert to docker compose
	data, err := convertLocal(taskDefinition)
	if err != nil {
		return err
	}

	// Write to docker-compose.local.yml file
	err = writeLocal(data)
	if err != nil {
		return err
	}

	return nil
}
func readTaskDefinitionFromFile(filename string) (*ecs.TaskDefinition, error) {
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
func readTaskDefinitionFromArn(arn string, c *cli.Context) (*ecs.TaskDefinition, error) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		return nil, err
	}
	commandConfig, err := newCommandConfig(c, rdwr)
	if err != nil {
		return nil, err
	}

	ecsClient := ecsclient.NewECSClient(commandConfig)
	return ecsClient.DescribeTaskDefinition(arn)
}

func readTaskDefinition(c *cli.Context) (*ecs.TaskDefinition, error) {
	arn := c.String(flags.TaskDefinitionArnFlag)
	filename := c.String(flags.TaskDefinitionFileFlag)

	if arn != "" && filename != "" {
		return nil, fmt.Errorf("Cannot specify both --%s and --%s flags.", flags.TaskDefinitionArnFlag, flags.TaskDefinitionFileFlag)
	}

	if arn != "" {
		return readTaskDefinitionFromArn(arn, c)
	}

	if filename != "" {
		return readTaskDefinitionFromFile(filename)
	}

	// Try reading local task-definition.json file
	if _, err := os.Stat(LocalInFileName); err == nil {
		return readTaskDefinitionFromFile(LocalInFileName)
	}

	return nil, fmt.Errorf("Could not detect valid Task Definition")
}

// FIXME placeholder
func convertLocal(taskDefinition *ecs.TaskDefinition) ([]byte, error) {
	data := []byte("taskDefinition")
	return data, nil
}

func writeLocal(data []byte) error {
	// Will error if the file already exists, otherwise create
	out, err := os.OpenFile(LocalOutFileName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, LocalOutFileMode)
	defer out.Close()

	if err != nil {
		fmt.Println("docker-compose.local.yml file already exists. Do you want to write over this file? [y/N]")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("Error reading input: %s", err.Error())
		}

		formattedInput := strings.ToLower(strings.TrimSpace(input))

		if formattedInput != "yes" && formattedInput != "y" {
			return fmt.Errorf("Aborted writing compose file. To retry, rename or move %s", LocalOutFileName) // TODO add force flag
		}

		// Overwrite local compose file
		err = ioutil.WriteFile(LocalOutFileName, data, LocalOutFileMode)
		return err
	}

	_, err = out.Write(data)

	return err
}
