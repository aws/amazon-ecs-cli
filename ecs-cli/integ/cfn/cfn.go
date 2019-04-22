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
	"github.com/stretchr/testify/assert"
)

// TestStackNameExists fails if there is no CloudFormation stack with the name stackName.
func TestStackNameExists(t *testing.T, stackName string) {
	client := newClient(t)
	resp, err := client.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		assert.FailNowf(t, "unexpected CloudFormation error during DescribeStacks", "wanted no errors, got %v", err)
	}
	if resp.Stacks == nil {
		assert.FailNow(t, "stacks should not be nil")
	}
	if len(resp.Stacks) != 1 {
		assert.FailNowf(t, "did not receive only 1 stack", "wanted only one stack, got %d", len(resp.Stacks))
	}
	if *resp.Stacks[0].StackName != stackName {
		assert.FailNowf(t, "unexpected stack name", "wanted %s, got %s", stackName, *resp.Stacks[0].StackName)
	}
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
				t.Logf("Stack %s does not exist as expected", stackName)
				return
			default:
				assert.NoError(t, err, "unexpected CloudFormation error during DescribeStacks")
			}
		}
	}

	assert.Equalf(t, 0, len(resp.Stacks), "No stack with the name %s should exist", stackName)
}

func newClient(t *testing.T) *cloudformation.CloudFormation {
	sess, err := session.NewSession()
	if err != nil {
		// Fail the upTest immediately if we won't be able to evaluate it
		assert.FailNowf(t, "failed to create new session for upTest clients", "%v", err)
	}

	conf := aws.NewConfig()
	return cloudformation.New(sess, conf)
}
