// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package sts

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/aws/aws-sdk-go/service/sts/stsiface"
)

//go:generate mockgen.sh github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/sts Client mock/$GOFILE
//go:generate mockgen.sh github.com/aws/aws-sdk-go/service/sts/stsiface STSAPI mock/sdk/stsiface_mock.go

// Client sts interface
type Client interface {
	GetAWSAccountID() (string, error)
}

// stsClient implements Client
type stsClient struct {
	client stsiface.STSAPI
	params *config.CliParams
}

// NewClient Creates a new sts client
func NewClient(params *config.CliParams) Client {
	client := sts.New(session.New(params.Session.Config))
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return newClient(params, client)
}

func newClient(params *config.CliParams, client stsiface.STSAPI) Client {
	return &stsClient{
		client: client,
		params: params,
	}
}

// GetAWSAccountID returns the accountId of the caller
func (c *stsClient) GetAWSAccountID() (string, error) {
	log.Info("Getting AWS account ID...")

	resp, err := c.client.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return "", err
	}
	return aws.StringValue(resp.Account), nil
}
