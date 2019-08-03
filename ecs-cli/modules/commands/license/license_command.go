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

// Package licenseCommand defines the command for displaying license information
package licenseCommand

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/license"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/usage"
	"github.com/urfave/cli"
)

// LicenseCommand prints the license
func LicenseCommand() cli.Command {
	return cli.Command{
		Name:         "license",
		Usage:        usage.License,
		Action:       license.PrintLicense,
		OnUsageError: flags.UsageErrorFactory("license"),
	}
}
