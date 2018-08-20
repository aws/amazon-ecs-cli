// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package regcreds

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcreds"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Up creates or updates registry credential secrets and an ECS task execution role needed to use them in a task def
func Up(c *cli.Context) {

	ecsCredsInputFile := c.String(flags.ComposeFileNameFlag)

	credsInput, err := readers.ReadCredsInput(ecsCredsInputFile)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}
	log.Infof("Read creds input: %+v", credsInput) // remove after SDK calls added

	//TODO: create secrets, create role, produce output
}
