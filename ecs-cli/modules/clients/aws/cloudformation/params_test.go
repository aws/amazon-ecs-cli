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

package cloudformation

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/stretchr/testify/assert"
)

const (
	parameterKeyAsgMaxSize = "AsgMaxSize"
	parameterKeyCluster    = "EcsCluster"
	parameterKeyAmiId      = "EcsAmiId"
)

func TestAddAndValidate(t *testing.T) {
	cfnParams := NewCfnStackParams([]string{parameterKeyCluster})

	err := cfnParams.Validate()
	if err == nil {
		t.Error("Expected validation error for empty parameter key map")
	}

	// Add AMI ID
	err = cfnParams.Add(parameterKeyAmiId, "ami-12345")
	if err != nil {
		t.Error("Error adding parameter: ", err)
	}
	err = cfnParams.Validate()
	if err == nil {
		t.Errorf("Expected validation error when %s is not specified", parameterKeyCluster)
	}

	// Add Cluster
	err = cfnParams.Add(parameterKeyCluster, "")
	if err != nil {
		t.Error("Error adding parameter: ", err)
	}
	err = cfnParams.Validate()
	if err == nil {
		t.Errorf("Expected validation error when %s is empty", parameterKeyCluster)
	}

	err = cfnParams.Add(parameterKeyCluster, "default")
	if err != nil {
		t.Error("Error adding parameter: ", err)
	}
	err = cfnParams.Validate()
	if err != nil {
		t.Error("Error validating parameter key", err)
	}

	paramsMap := cfnParams.Get()
	if len(paramsMap) != 2 { // 2 parameters have been added
		t.Errorf("Mismatch in number of keys in params map. Expected 2, found: %d", len(paramsMap))
	}

	clusterValue, exists := cfnParams.nameToKeys[parameterKeyCluster]
	if !exists {
		t.Errorf("Expected key %s does not exist", parameterKeyCluster)
	}

	if "default" != clusterValue {
		t.Errorf("Mismatch in cluster name. Expected [%s] Got [%s]", "default", clusterValue)
	}
}

func TestAddWithUsePreviousValue(t *testing.T) {
	existingParameters := []*cloudformation.Parameter{
		&cloudformation.Parameter{
			ParameterKey: aws.String("SomeParam1"),
		},
		&cloudformation.Parameter{
			ParameterKey: aws.String("SomeParam2"),
		},
	}
	cfnParams, err := NewCfnStackParamsForUpdate([]string{parameterKeyCluster}, existingParameters)
	assert.NoError(t, err, "Unexpected error getting New CFN Stack Params")

	params := cfnParams.Get()
	if 0 == len(params) {
		t.Error("Got empty params list")
	}

	for _, param := range params {
		usePrevious := param.UsePreviousValue
		paramName := aws.StringValue(param.ParameterKey)
		if usePrevious == nil {
			t.Fatalf("usePrevious is not set for '%s' in params map", paramName)
		}

		if !aws.BoolValue(usePrevious) {
			t.Errorf("usePrevious value for '%s' is false, expected to be true", paramName)
		}
	}

	err = cfnParams.AddWithUsePreviousValue(parameterKeyAsgMaxSize, false)
	if err != nil {
		t.Errorf("Error adding parameter with use previous value '%s': '%v'", parameterKeyAsgMaxSize, err)
	}

	size := "3"
	err = cfnParams.Add(parameterKeyAsgMaxSize, size)
	if err != nil {
		t.Errorf("Error adding parameter '%s': %v", parameterKeyAsgMaxSize, err)
	}

	param, err := cfnParams.GetParameter(parameterKeyAsgMaxSize)
	if err != nil {
		t.Errorf("Error getting parameter '%s': %v", parameterKeyAsgMaxSize, err)
	}
	usePrevious := param.UsePreviousValue
	if usePrevious == nil {
		t.Fatalf("usePrevious is not set for '%s' in params map", parameterKeyAsgMaxSize)
	}

	if aws.BoolValue(usePrevious) {
		t.Errorf("usePrevious value is true for '%s', expected false", parameterKeyAsgMaxSize)
	}

}
