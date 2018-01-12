// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package app

import (
	log "github.com/sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

// BeforeApp is an action that is executed before any cli command.
func BeforeApp(c *cli.Context) error {
	if c.GlobalBool(flags.VerboseFlag) || c.Bool(flags.VerboseFlag) {
		log.SetLevel(log.DebugLevel)
	}
	return nil
}
