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

package cloudwatchlogs

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
)

// Client defines methods to interact with the CloudWatch API interface.
type Client interface {
	FilterAllLogEvents(*cloudwatchlogs.FilterLogEventsInput, func([]*cloudwatchlogs.FilteredLogEvent)) error
}

// ec2Client implements EC2Client
type cwLogsClient struct {
	client cloudwatchlogsiface.CloudWatchLogsAPI
}

// NewCloudWatchLogsClient creates an instance of ec2Client object.
func NewCloudWatchLogsClient(params *config.CLIParams, logRegion string) Client {
	session := params.Session
	session.Config = session.Config.WithRegion(logRegion)
	client := cloudwatchlogs.New(session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return &cwLogsClient{
		client: client,
	}
}

func (c *cwLogsClient) FilterAllLogEvents(input *cloudwatchlogs.FilterLogEventsInput, action func([]*cloudwatchlogs.FilteredLogEvent)) error {
	err := c.client.FilterLogEventsPages(input,
		func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
			action(page.Events)
			return !lastPage
		})
	return err
}
