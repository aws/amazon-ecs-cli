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

	"github.com/aws/aws-sdk-go/service/ecs"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	LocalOutFileName = "./docker-compose.local.yml"
	LocalOutFileMode = os.FileMode(0644) // Owner=read/write, Other=readonly
)

func Create(c *cli.Context) {
	// 1. Read in task definition (from file or ARN)
	filename := "./task-definition.json" // FIXME defaults

	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error reading task definition from %v", filename)
	}

	// 2. Parse task def into go object
	taskDefinition := ecs.TaskDefinition{}
	json.Unmarshal(bytes, &taskDefinition)

	// 3. Convert to docker compose
	// fmt.Printf("TASK DEF: %+v", taskDefinition)

	// 4. Write to docker-compose.local.yml file

	data := []byte("taskDefinition")

	err = writeLocal(data)
	if err != nil {
		log.Fatalf("Error with local create: %s", err.Error())
	}

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
