// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package utils

import (
	"testing"
)

func TestGetIdFromArn(t *testing.T) {
	arnPrefix := "arn:aws:ecs:us-west-2:accountId:task-definition/"

	// arn with 2 parts delimited by /
	expectedId := "task-definition-name:1"
	arn := arnPrefix + expectedId
	observedId := GetIdFromArn(arn)
	if expectedId != observedId {
		t.Errorf("Expected id to be [%s] but got [%s]", expectedId, observedId)
	}

	// arn with 3 parts delimited by /
	expectedId = "testing/testing"
	arn = arnPrefix + expectedId
	observedId = GetIdFromArn(arn)
	if expectedId != observedId {
		t.Errorf("Expected id to be [%s] but got [%s]", expectedId, observedId)
	}

	// arn with 1 parts delimited by /
	expectedId = ""
	arn = "testing"
	observedId = GetIdFromArn(arn)
	if expectedId != observedId {
		t.Errorf("Expected id to be [%s] but got [%s]", expectedId, observedId)
	}
}
