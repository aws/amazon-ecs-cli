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
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/localproject"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// TaskReadConvertWriter is the interface that groups the ReadTaskDefinition, Convert, and Write methods.
type TaskReadConvertWriter interface {
	ReadTaskDefinition() error
	Convert() error
	Write() error
}

// Create reads in an ECS task definition, converts and writes it to a local
// Docker Compose file
func Create(c *cli.Context) {
	project := localproject.New(c)

	err := createLocal(project)
	if err != nil {
		log.Fatalf("Error with local create: %s", err.Error())
	}

	fmt.Printf("Successfully wrote %s\n", project.LocalOutFileName())
}

func createLocal(project TaskReadConvertWriter) error {
	// Reads task definition and loads it onto project
	err := project.ReadTaskDefinition()
	if err != nil {
		return err
	}

	// Converts Task definition and loads onto project
	err = project.Convert()
	if err != nil {
		return err
	}

	// Write to local output file
	err = project.Write()
	if err != nil {
		return err
	}

	return nil
}
