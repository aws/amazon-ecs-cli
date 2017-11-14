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
			Usage: "[Required] Specify the task id of the task from which to find logs.",
		},
		cli.StringFlag{
			Name:  flags.TaskDefinitionFlag,
			Usage: "Task definition of the task for which you want to view logs. Required with Task ID if the task has been stopped already. Format: family:revision, or the full ARN.",
		},
		cli.BoolFlag{
			Name:  flags.FollowLogsFlag,
			Usage: "Continuously poll for new logs.",
		},
		cli.StringFlag{
			Name:  flags.FilterPatternFlag,
			Usage: "Substring to search for within the logs.",
		},
		cli.StringFlag{
			Name:  flags.ContainerNameFlag,
			Usage: "Filter logs for a given container definition. Required if all the Container Definitions in your Task Definition do not use the same log group.",
		},
		cli.IntFlag{
			Name:  flags.SinceFlag,
			Usage: fmt.Sprintf("Filter logs in the last X minutes. Can not be used with %s", flags.StartTimeFlag),
		},
		cli.StringFlag{
			Name:  flags.StartTimeFlag,
			Usage: fmt.Sprintf("Return logs after this time. Can not be used with %s", flags.SinceFlag),
		},
		cli.StringFlag{
			Name:  flags.EndTimeFlag,
			Usage: fmt.Sprintf("Return logs before this time. Can not be used with %s", flags.FollowLogsFlag),
		},
		cli.BoolFlag{
			Name:  flags.TimeStampsFlag + ",t",
			Usage: "View timestamps with the logs",
		},
	}
}
