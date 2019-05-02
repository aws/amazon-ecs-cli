// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Package cfn contains validation functions against a CloudFormation stack.
package cfn

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/require"
)

// TestStackNameExists fails if there is no CloudFormation stack with the name stackName.
func TestStackNameExists(t *testing.T, stackName string) {
	client := newClient(t)
	resp, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	require.NoError(t, err, "Unexpected CloudFormation error during DescribeStacks")
	require.NotNil(t, resp.Stacks, "CloudFormation stacks should not be nil")
	require.Equalf(t, 1, len(resp.Stacks), "Expected to receive only 1 stack", "got %v", resp.Stacks)
	require.Equalf(t, stackName, *resp.Stacks[0].StackName, "Unexpected stack name", "wanted %s, got %s", stackName, *resp.Stacks[0].StackName)
}

// TestNoStackName fails if there is a CloudFormation stack with the name stackName.
func TestNoStackName(t *testing.T, stackName string) {
	client := newClient(t)
	resp, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "ValidationError":
				t.Logf("Success: stack %s does not exist", stackName)
				return
			default:
				require.NoError(t, err, "Unexpected CloudFormation error during DescribeStacks")
			}
		}
	}

	require.Equalf(t, 0, len(resp.Stacks), "No stack with the name %s should exist", stackName)
}

func newClient(t *testing.T) *cloudformation.CloudFormation {
	sess, err := session.NewSession()
	require.NoError(t, err, "failed to create new session for CloudFormation")
	conf := aws.NewConfig()
	return cloudformation.New(sess, conf)
}
