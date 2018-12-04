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

package iam

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

// Client defines methods for interacting with the IAMAPI interface
type Client interface {
	AttachRolePolicy(policyArn, roleName string) (*iam.AttachRolePolicyOutput, error)
	CreateRole(iam.CreateRoleInput) (*iam.CreateRoleOutput, error)
	CreatePolicy(iam.CreatePolicyInput) (*iam.CreatePolicyOutput, error)
	CreateOrFindRole(string, string, string) (string, error)
}

type iamClient struct {
	client iamiface.IAMAPI
}

// NewIAMClient creates an instance of iamClient
func NewIAMClient(config *config.CommandConfig) Client {
	client := iam.New(config.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())

	return newClient(client)
}

func newClient(client iamiface.IAMAPI) Client {
	return &iamClient{
		client: client,
	}
}

func (c *iamClient) AttachRolePolicy(policyArn, roleName string) (*iam.AttachRolePolicyOutput, error) {
	request := iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(roleName),
	}

	output, err := c.client.AttachRolePolicy(&request)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *iamClient) CreateRole(input iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	output, err := c.client.CreateRole(&input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *iamClient) CreatePolicy(input iam.CreatePolicyInput) (*iam.CreatePolicyOutput, error) {
	output, err := c.client.CreatePolicy(&input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// CreateOrFindRole returns a new role ARN or an empty string if role already exists
func (c *iamClient) CreateOrFindRole(roleName, roleDescription, assumeRolePolicyDoc string) (string, error) {
	createRoleRequest := iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(assumeRolePolicyDoc),
		Description:              aws.String(roleDescription),
		RoleName:                 aws.String(roleName),
	}
	roleResult, err := c.CreateRole(createRoleRequest)
	// if err is b/c role already exists, OK to continue
	if err != nil && !utils.EntityAlreadyExists(err) {
		return "", err
	}
	// TODO: validate AssumeRolePolicyDocument of existing role?

	newRoleString := ""
	if roleResult != nil {
		newRole := *roleResult.Role
		newRoleString = *newRole.Arn
	}
	// TODO, maybe: find & return existing role ARN, plus bool to indicate whether new or not?

	return newRoleString, nil
}
