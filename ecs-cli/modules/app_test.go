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
	"flag"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

func TestBeforeApp(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.Bool(flags.VerboseFlag, true, "")
	cliContext := cli.NewContext(nil, flagSet, nil)

	BeforeApp(cliContext)

	observedLogLevel := log.GetLevel()
	if log.DebugLevel != observedLogLevel {
		t.Errorf("Log level was supposed to be set to debug. Expected [%s] Got [%s]", log.DebugLevel, observedLogLevel)
	}
}
