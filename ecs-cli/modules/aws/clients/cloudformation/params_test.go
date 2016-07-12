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

package cloudformation

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestAddAndValidate(t *testing.T) {
	cfnParams := NewCfnStackParams()

	err := cfnParams.Validate()
	if err == nil {
		t.Error("Expected validation error for empty parameter key map")
	}

	err = cfnParams.Add(ParameterKeyKeyPairName, "default")
	if err != nil {
		t.Error("Error adding parameter: ", err)
	}
	err = cfnParams.Validate()
	if err == nil {
		t.Errorf("Expected validation error when only %s is specified", ParameterKeyKeyPairName)
	}

	err = cfnParams.Add(ParameterKeyAmiId, "ami-12345")
	if err != nil {
		t.Error("Error adding parameter: ", err)
	}
	err = cfnParams.Validate()
	if err == nil {
		t.Errorf("Expected validation error when %s is not specified", ParameterKeyCluster)
	}

	err = cfnParams.Add(ParameterKeyCluster, "")
	if err != nil {
		t.Error("Error adding parameter: ", err)
	}
	err = cfnParams.Validate()
	if err == nil {
		t.Errorf("Expected validation error when %s is empty", ParameterKeyCluster)
	}

	err = cfnParams.Add(ParameterKeyCluster, "default")
	if err != nil {
		t.Error("Error adding parameter: ", err)
	}
	err = cfnParams.Validate()
	if err != nil {
		t.Error("Error validating parameter key", err)
	}

	paramsMap := cfnParams.Get()
	if len(requiredParameterNames) != len(paramsMap) {
		t.Errorf("Mismatch in number of keys in params map. %d != %d", len(requiredParameterNames), len(paramsMap))
	}

	clusterValue, exists := cfnParams.nameToKeys[ParameterKeyCluster]
	if !exists {
		t.Errorf("Expcted key %s does not exist", ParameterKeyCluster)
	}

	if "default" != clusterValue {
		t.Errorf("Mismtach in cluster name. Expected [%s] Got [%s]", "default", clusterValue)
	}
}

func TestAddWithUsePreviousValueAndValidate(t *testing.T) {
	cfnParams := NewCfnStackParamsForUpdate()
	err := cfnParams.Validate()
	if err != nil {
		t.Error("Error validating params for update: ", err)
	}

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

	err = cfnParams.AddWithUsePreviousValue(ParameterKeyAsgMaxSize, false)
	if err != nil {
		t.Errorf("Error adding parameter with use previous value '%s': '%v'", ParameterKeyAsgMaxSize, err)
	}
	err = cfnParams.Validate()
	if err == nil {
		t.Errorf("Expected error for param '%s' when usePrevious is false and value is not set", ParameterKeyAsgMaxSize)
	}

	size := "3"
	err = cfnParams.Add(ParameterKeyAsgMaxSize, size)
	if err != nil {
		t.Errorf("Error adding parameter '%s': %v", ParameterKeyAsgMaxSize, err)
	}

	param, err := cfnParams.GetParameter(ParameterKeyAsgMaxSize)
	if err != nil {
		t.Errorf("Error getting parameter '%s': %v", ParameterKeyAsgMaxSize, err)
	}
	usePrevious := param.UsePreviousValue
	if usePrevious == nil {
		t.Fatalf("usePrevious is not set for '%s' in params map", ParameterKeyAsgMaxSize)
	}

	if aws.BoolValue(usePrevious) {
		t.Errorf("usePrevious value is true for '%s', expected false", ParameterKeyAsgMaxSize)
	}

}
