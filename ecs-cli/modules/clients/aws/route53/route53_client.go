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

// Package route53 contains functions for working with the route53 APIs
// that back ECS Service Discovery
package route53

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/service/route53"
)

// Private Route53 Client that can be mocked in unit tests
// The SDK's route53 client implements this interface
type route53Client interface {
	GetHostedZone(input *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error)
}

// factory function to create clients
func newRoute53Client(config *config.CommandConfig) route53Client {
	r53Client := route53.New(config.Session)
	r53Client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return r53Client
}
