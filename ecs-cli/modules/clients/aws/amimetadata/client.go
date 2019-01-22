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

package amimetadata

import (
	"encoding/json"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pkg/errors"
)

const (
	amazonLinux2X86RecommendedParameterName   = "/aws/service/ecs/optimized-ami/amazon-linux-2/recommended"
	amazonLinux2ARM64RecommendedParameterName = "/aws/service/ecs/optimized-ami/amazon-linux-2/arm64/recommended"
)

const (
	ArchitectureTypeARM64 = "arm64"
	ArchitectureTypeX86   = "x86"
)

type AMIMetadata struct {
	ImageID        string `json:"image_id"`
	OsName         string `json:"os"`
	AgentVersion   string `json:"ecs_agent_version"`
	RuntimeVersion string `json:"ecs_runtime_version"`
}

// Client defines methods to interact with the SSM API interface.
type Client interface {
	GetRecommendedECSLinuxAMI(string) (*AMIMetadata, error)
}

// ssmClient implements Client
type metadataClient struct {
	client ssmiface.SSMAPI
	region string
}

// NewSSMClient creates an instance of Client.
func NewMetadataClient(commandConfig *config.CommandConfig) Client {
	client := ssm.New(commandConfig.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return &metadataClient{
		client: client,
		region: aws.StringValue(commandConfig.Session.Config.Region),
	}
}

func (c *metadataClient) GetRecommendedECSLinuxAMI(architecture string) (*AMIMetadata, error) {
	ssmParam := amazonLinux2X86RecommendedParameterName
	if architecture == ArchitectureTypeARM64 {
		ssmParam = amazonLinux2ARM64RecommendedParameterName
	}

	response, err := c.client.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(ssmParam),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == ssm.ErrCodeParameterNotFound {
				// Added for arm AMIs which are only supported in some regions
				return nil, errors.Wrapf(err, "Could not find Recommended Amazon Linux 2 AMI in %s with architecture %s; the AMI may not be supported in this region", c.region, architecture)
			}
		}
		return nil, err
	}
	metadata := &AMIMetadata{}
	err = json.Unmarshal([]byte(aws.StringValue(response.Parameter.Value)), metadata)
	return metadata, err
}
