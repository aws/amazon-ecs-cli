// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package tagging

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	taggingSDK "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi/resourcegroupstaggingapiiface"
)

// Client defines methods to interact with the SSM API interface.
type Client interface {
	TagResources(*resourcegroupstaggingapi.TagResourcesInput) (*resourcegroupstaggingapi.TagResourcesOutput, error)
}

// taggingClient implements Client
type taggingClient struct {
	client taggingSDK.ResourceGroupsTaggingAPIAPI
}

// NewTaggingClient creates an instance of Client.
func NewTaggingClient(commandConfig *config.CommandConfig) Client {
	client := resourcegroupstaggingapi.New(commandConfig.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return &taggingClient{
		client: client,
	}
}

func (c *taggingClient) TagResources(input *resourcegroupstaggingapi.TagResourcesInput) (*resourcegroupstaggingapi.TagResourcesOutput, error) {
	return c.client.TagResources(input)
}
