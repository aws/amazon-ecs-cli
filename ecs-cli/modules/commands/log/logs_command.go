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

package logsCommand

import (
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/logs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/urfave/cli"
)

// LogCommand Retrieves container logs from CloudWatch.
func LogCommand() cli.Command {
	return cli.Command{
		Name:         "logs",
		Usage:        "Retrieves container logs from CloudWatch logs. Assumes your Task Definition uses the awslogs driver and has a log stream prefix specified.",
		Flags:        append(flags.OptionalConfigFlags(), logFlags()...),
		Action:       logs.Logs,
		OnUsageError: flags.UsageErrorFactory("logs"),
	}
}

func logFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  flags.TaskIDFlag,
			Usage: "Print the logs for this ECS Task.",
		},
		cli.StringFlag{
			Name:  flags.TaskDefinitionFlag,
			Usage: "[Optional] Specifies the name or full Amazon Resource Name (ARN) of the ECS Task Definition associated with the Task ID. This is only needed if the Task is using an inactive Task Definition.",
		},
		cli.BoolFlag{
			Name:  flags.FollowLogsFlag,
			Usage: "[Optional] Specifies if the logs should be streamed.",
		},
		cli.StringFlag{
			Name:  flags.FilterPatternFlag,
			Usage: "[Optional] Substring to search for within the logs.",
		},
		cli.StringFlag{
			Name:  flags.ContainerNameFlag,
			Usage: "[Optional] Prints the logs for the given container. Required if containers in the Task use different log groups",
		},
		cli.IntFlag{
			Name:  flags.SinceFlag,
			Usage: fmt.Sprintf("[Optional] Returns logs newer than a relative duration in minutes. Can not be used with --%s", flags.StartTimeFlag),
		},
		cli.StringFlag{
			Name:  flags.StartTimeFlag,
			Usage: fmt.Sprintf("[Optional] Returns logs after a specific date (format: RFC 3339. Example: 2006-01-02T15:04:05+07:00). Cannot be used with --%s flag", flags.SinceFlag),
		},
		cli.StringFlag{
			Name:  flags.EndTimeFlag,
			Usage: fmt.Sprintf("[Optional] Returns logs before a specific date (format: RFC 3339. Example: 2006-01-02T15:04:05+07:00). Cannot be used with --%s", flags.FollowLogsFlag),
		},
		cli.BoolFlag{
			Name:  flags.TimeStampsFlag + ",t",
			Usage: "[Optional] Shows timestamps on each line in the log output.",
		},
	}
}
