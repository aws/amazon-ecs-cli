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

package ssm

import (
	"encoding/json"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

const (
	amazonLinux2RecommendedParameterName = "/aws/service/ecs/optimized-ami/amazon-linux-2/recommended"
)

type AMIMetadata struct {
	ImageID        string `json:"image_id"`
	OsName         string `json:"os"`
	AgentVersion   string `json:"ecs_agent_version"`
	RuntimeVersion string `json:"ecs_runtime_version"`
}

// Client defines methods to interact with the SSM API interface.
type Client interface {
	GetRecommendedECSLinuxAMI() (*AMIMetadata, error)
}

// ssmClient implements Client
type ssmClient struct {
	client ssmiface.SSMAPI
}

// NewSSMClient creates an instance of Client.
func NewSSMClient(commandConfig *config.CommandConfig) Client {
	client := ssm.New(commandConfig.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return &ssmClient{
		client: client,
	}
}

func (c *ssmClient) GetRecommendedECSLinuxAMI() (*AMIMetadata, error) {
	response, err := c.client.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(amazonLinux2RecommendedParameterName),
	})
	if err != nil {
		return nil, err
	}
	metadata := &AMIMetadata{}
	err = json.Unmarshal([]byte(aws.StringValue(response.Parameter.Value)), metadata)
	return metadata, err
}
