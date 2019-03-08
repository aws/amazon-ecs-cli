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

// Package amimetadata provides AMI metadata given an instance type.
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
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

// SSM parameter names to retrieve ECS optimized AMI.
// See: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/retrieve-ecs-optimized_AMI.html
const (
	amazonLinux2X86RecommendedParameterName    = "/aws/service/ecs/optimized-ami/amazon-linux-2/recommended"
	amazonLinux2ARM64RecommendedParameterName  = "/aws/service/ecs/optimized-ami/amazon-linux-2/arm64/recommended"
	amazonLinux2X86GPURecommendedParameterName = "/aws/service/ecs/optimized-ami/amazon-linux-2/gpu/recommended"
)

// AMIMetadata is returned through ssm:GetParameters and can be used to retrieve the ImageId
// while launching instances.
//
// See: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/retrieve-ecs-optimized_AMI.html
// See: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-as-launchconfig.html#cfn-as-launchconfig-imageid
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

// metadataClient implements Client.
type metadataClient struct {
	client ssmiface.SSMAPI
	region string
}

// NewMetadataClient creates an instance of Client.
func NewMetadataClient(commandConfig *config.CommandConfig) Client {
	client := ssm.New(commandConfig.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return &metadataClient{
		client: client,
		region: aws.StringValue(commandConfig.Session.Config.Region),
	}
}

// GetRecommendedECSLinuxAMI returns the recommended Amazon ECS-Optimized AMI Metadata given the instance type.
func (c *metadataClient) GetRecommendedECSLinuxAMI(instanceType string) (*AMIMetadata, error) {
	if isARM64Instance(instanceType) {
		logrus.Infof("Using Arm ecs-optimized AMI because instance type was %s", instanceType)
		return c.parameterValueFor(amazonLinux2ARM64RecommendedParameterName)
	}
	if isGPUInstance(instanceType) {
		logrus.Infof("Using GPU ecs-optimized AMI because instance type was %s", instanceType)
		return c.parameterValueFor(amazonLinux2X86GPURecommendedParameterName)
	}
	return c.parameterValueFor(amazonLinux2X86RecommendedParameterName)
}

func (c *metadataClient) parameterValueFor(ssmParamName string) (*AMIMetadata, error) {
	response, err := c.client.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(ssmParamName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			if aerr.Code() == ssm.ErrCodeParameterNotFound {
				// Added for AMIs which are only supported in some regions
				return nil, errors.Wrapf(err,
					"Could not find Recommended Amazon Linux 2 AMI %s in %s; the AMI may not be supported in this region",
					ssmParamName,
					c.region)
			}
		}
		return nil, err
	}
	metadata := &AMIMetadata{}
	err = json.Unmarshal([]byte(aws.StringValue(response.Parameter.Value)), metadata)
	return metadata, err
}

func isARM64Instance(instanceType string) bool {
	r := regexp.MustCompile("a1\\.(medium|\\d*x?large)")
	if r.MatchString(instanceType) {
		return true
	}
	return false
}

func isGPUInstance(instanceType string) bool {
	if strings.HasPrefix(instanceType, "p2.") {
		return true
	}
	if strings.HasPrefix(instanceType, "p3.") {
		return true
	}
	if strings.HasPrefix(instanceType, "p3dn.") {
		return true
	}
	return false
}
