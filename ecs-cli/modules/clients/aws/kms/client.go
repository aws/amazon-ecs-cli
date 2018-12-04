// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package kms

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

// Client defines methods for interacting with KMS
type Client interface {
	DescribeKey(keyID string) (*kms.DescribeKeyOutput, error)
	GetValidKeyARN(keyID string) (string, error)
}

type kmsClient struct {
	client kmsiface.KMSAPI
}

// NewKMSClient creates an instance of a kmsClient
func NewKMSClient(config *config.CommandConfig) Client {
	client := kms.New(config.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())

	return newClient(client)
}

func newClient(client kmsiface.KMSAPI) Client {
	return &kmsClient{
		client: client,
	}
}

func (c *kmsClient) DescribeKey(keyID string) (*kms.DescribeKeyOutput, error) {
	request := kms.DescribeKeyInput{
		KeyId: aws.String(keyID),
	}

	output, err := c.client.DescribeKey(&request)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *kmsClient) GetValidKeyARN(keyID string) (string, error) {
	ARNString := ""

	ARNobj, err := arn.Parse(keyID)

	if err == nil {
		ARNString = ARNobj.String()
	} else {
		keyResult, err := c.DescribeKey(keyID)
		if err != nil {
			return "", err
		}
		keyMetadata := *keyResult.KeyMetadata
		ARNString = *keyMetadata.Arn
	}
	return ARNString, nil
}
