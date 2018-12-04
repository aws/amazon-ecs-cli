// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package cloudwatchlogs

import (
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
)

func TestNewCloudWatchLogsClientLeavesRegionIntact(t *testing.T) {
	// Test that it doesn't mess up command config session and re-set the region
	correctRegion := "eu-west-1"
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(correctRegion),
	})
	assert.NoError(t, err)
	params := &config.CommandConfig{
		Session: sess,
	}

	NewCloudWatchLogsClient(params, "us-east-2")

	assert.Equal(t, correctRegion, aws.StringValue(params.Session.Config.Region), "Expected configured region to remain unchanged after call to NewCloudWatchLogsClient()")

}
