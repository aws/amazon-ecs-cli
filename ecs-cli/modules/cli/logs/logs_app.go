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

package logs

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	cwlogsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudwatchlogs"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

const (
	followLogsWaitTime = 30
)

type logConfiguration struct {
	logGroup  *string
	logRegion *string
	logPrefix *string
}

type logInfo struct {
	logGroup  *string
	logRegion *string
	prefixes  map[*string]*string
}

// Logs is the action for logsCommand. It retrieves container logs for a task from CloudWatch
func Logs(c *cli.Context) {
	err := validateLogFlags(c)
	if err != nil {
		logrus.Fatal("Error executing 'logs': ", err)
	}
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'logs': ", err)
	}
	ecsParams, err := config.NewCliParams(c, rdwr)
	if err != nil {
		logrus.Fatal("Error executing 'logs': ", err)
	}

	ecsClient := ecsclient.NewECSClient()
	ecsClient.Initialize(ecsParams)
	request, logRegion, err := logsRequest(c, ecsClient, ecsParams)
	if err != nil {
		logrus.Fatal("Error executing 'logs': ", err)
	}

	cwLogsClient := cwlogsclient.NewCloudWatchLogsClient(ecsParams, logRegion)

	printLogEvents(c, request, cwLogsClient)
}

func logsRequest(context *cli.Context, ecsClient ecsclient.ECSClient, params *config.CliParams) (*cloudwatchlogs.FilterLogEventsInput, string, error) {
	taskID := context.String(command.TaskIDFlag)
	taskDefIdentifier := context.String(command.TaskDefinitionFlag)

	var err error
	if taskDefIdentifier == "" {
		taskDefIdentifier, err = getTaskDefArn(context, ecsClient, params)
		if err != nil {
			return nil, "", err
		}
	}

	taskDef, err := ecsClient.DescribeTaskDefinition(taskDefIdentifier)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("Failed to Describe TaskDefinition; try using --%s to specify the Task Defintion.", command.TaskDefinitionFlag))
	}

	containerName := context.String(command.ContainerNameFlag)
	logConfig, err := getLogConfiguration(taskDef, taskID, containerName)

	if err != nil {
		return nil, "", errors.Wrap(err, "Failed to get log configuration")
	}

	streams := logStreams(logConfig.prefixes, taskID)

	request, err := filterLogEventsInputFromContext(context)
	if err != nil {
		return nil, "", errors.Wrap(err, "Failed to create FilterLogEvents request")
	}
	request.SetLogGroupName(aws.StringValue(logConfig.logGroup))
	request.SetLogStreamNames(aws.StringSlice(streams))

	return request, aws.StringValue(logConfig.logRegion), nil
}

func getTaskDefArn(context *cli.Context, ecsClient ecsclient.ECSClient, params *config.CliParams) (string, error) {
	var taskIDs []*string
	taskID := context.String(command.TaskIDFlag)
	taskIDs = append(taskIDs, aws.String(taskID))
	tasks, err := ecsClient.DescribeTasks(taskIDs)
	if err != nil {
		return "", errors.Wrap(err, "Failed to Describe Task")
	}
	if len(tasks) == 0 {
		return "", fmt.Errorf("Failed to describe Task: Could Not Find Task %s in cluster %s in region %s. If the task has been stopped, use --%s to specify the Task Definition.", taskID, params.Cluster, aws.StringValue(params.Session.Config.Region), command.TaskDefinitionFlag)
	}

	return aws.StringValue(tasks[0].TaskDefinitionArn), nil
}

func printLogEvents(context *cli.Context, input *cloudwatchlogs.FilterLogEventsInput, cwLogsClient cwlogsclient.Client) {
	var lastEvent *cloudwatchlogs.FilteredLogEvent
	cwLogsClient.FilterAllLogEvents(input, func(events []*cloudwatchlogs.FilteredLogEvent) {
		for _, event := range events {
			lastEvent = event
			if context.Bool(command.TimeStampsFlag) {
				timeStamp := time.Unix(0, aws.Int64Value(event.Timestamp)*int64(time.Millisecond))
				fmt.Printf("%s\t%s\n", timeStamp.Format(time.RFC3339), aws.StringValue(event.Message))
			} else {
				fmt.Println(aws.StringValue(event.Message))
			}
			fmt.Println()
		}
	})

	for context.Bool(command.FollowLogsFlag) && lastEvent != nil {
		time.Sleep(followLogsWaitTime * time.Second)
		input.SetStartTime(aws.Int64Value(lastEvent.Timestamp) + 1)
		printLogEvents(context, input, cwLogsClient)
	}
}

// validateLogFlags ensures that conflicting flags are not used
func validateLogFlags(context *cli.Context) error {
	if taskID := context.String(command.TaskIDFlag); taskID == "" {
		return fmt.Errorf("TaskID must be specified with the --%s flag", command.TaskIDFlag)
	}

	startTime := context.String(command.StartTimeFlag)
	endTime := context.String(command.EndTimeFlag)
	since := context.Int(command.SinceFlag)

	if since > 0 && startTime != "" {
		return fmt.Errorf("--%s can not be used with --%s", command.SinceFlag, command.StartTimeFlag)
	}

	if context.Bool(command.FollowLogsFlag) && endTime != "" {
		return fmt.Errorf("--%s can not be used with --%s", command.FollowLogsFlag, command.EndTimeFlag)
	}
	return nil
}

func cwTimestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// filterLogEventsInputFromContext takes the command line flags and builds a FilterLogEventsInput object
// Does not handle validation of flags
func filterLogEventsInputFromContext(context *cli.Context) (*cloudwatchlogs.FilterLogEventsInput, error) {
	input := &cloudwatchlogs.FilterLogEventsInput{}
	if pattern := context.String(command.FilterPatternFlag); pattern != "" {
		input.SetFilterPattern(pattern)
	}

	if startTime := context.String(command.StartTimeFlag); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			return nil, err
		}
		input.SetStartTime(cwTimestamp(t))
	}

	if endTime := context.String(command.EndTimeFlag); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			return nil, err
		}
		input.SetEndTime(cwTimestamp(t))
	}

	if since := context.Int(command.SinceFlag); since > 0 {
		now := time.Now()
		then := now.Add(time.Duration(-since) * time.Minute)
		input.SetStartTime(cwTimestamp(then))
		input.SetEndTime(cwTimestamp(now))
	}

	return input, nil
}

func logStreams(prefixes map[*string]*string, taskID string) []string {
	var streams []string
	for containerName, prefix := range prefixes {
		streams = append(streams, aws.StringValue(prefix)+"/"+aws.StringValue(containerName)+"/"+taskID)
	}

	return streams
}

func getLogConfiguration(taskDef *ecs.TaskDefinition, taskID string, containerName string) (*logInfo, error) {
	logConfig := &logInfo{}
	logConfig.prefixes = make(map[*string]*string)

	if containerName != "" {
		var container *ecs.ContainerDefinition
		for _, containerDef := range taskDef.ContainerDefinitions {
			if aws.StringValue(containerDef.Name) == containerName {
				container = containerDef
				break
			}
		}
		info, err := getContainerLogConfig(container)
		if err != nil {
			return nil, err
		}
		logConfig.prefixes[container.Name] = info.logPrefix
		logConfig.logGroup = info.logGroup
		logConfig.logRegion = info.logRegion
	} else {
		info, err := getContainerLogConfig(taskDef.ContainerDefinitions[0])
		if err != nil {
			return nil, err
		}
		logConfig.logGroup = info.logGroup
		logConfig.logRegion = info.logRegion
		logConfig.prefixes[taskDef.ContainerDefinitions[0].Name] = info.logPrefix
		for _, containerDef := range taskDef.ContainerDefinitions {
			info, err := getContainerLogConfig(containerDef)
			if err != nil {
				return nil, err
			}
			if aws.StringValue(info.logGroup) != aws.StringValue(logConfig.logGroup) {
				return nil, logConfigMisMatchError(taskDef, "awslogs-group")
			}
			if aws.StringValue(info.logRegion) != aws.StringValue(logConfig.logRegion) {
				return nil, logConfigMisMatchError(taskDef, "awslogs-region")
			}
			logConfig.prefixes[containerDef.Name] = info.logPrefix
		}
	}
	return logConfig, nil
}

func getContainerLogConfig(containerDef *ecs.ContainerDefinition) (*logConfiguration, error) {
	logConfig := &logConfiguration{}
	if aws.StringValue(containerDef.LogConfiguration.LogDriver) != "awslogs" {
		return nil, fmt.Errorf("Container: Must specify log driver as awslogs")
	}

	var ok bool
	logConfig.logPrefix, ok = containerDef.LogConfiguration.Options["awslogs-stream-prefix"]
	if !ok || aws.StringValue(logConfig.logPrefix) == "" {
		return nil, fmt.Errorf("Container Definition %s: Log String Prefix (awslogs-stream-prefix) must be specified in each container definition in order to retrieve logs", aws.StringValue(containerDef.Name))
	}

	logConfig.logGroup = containerDef.LogConfiguration.Options["awslogs-group"]

	logConfig.logRegion = containerDef.LogConfiguration.Options["awslogs-region"]

	return logConfig, nil

}

func logConfigMisMatchError(taskDef *ecs.TaskDefinition, fieldName string) error {
	return fmt.Errorf("Log Configuration Field %s mismatches in at least one container definition in %s. Use the --%s option to query logs for an individual container.", fieldName, aws.StringValue(taskDef.TaskDefinitionArn), command.ContainerNameFlag)
}
