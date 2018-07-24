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
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

var ParameterNotFoundError = errors.New("Parameter not found")

// CfnStackParams is used to create cloudformation parameters used to create the stack.
type CfnStackParams struct {
	nameToKeys         map[string]string
	params             []*cloudformation.Parameter
	requiredParameters map[string]bool
}

// NewCfnStackParams creates an object of CfnStackParams struct,
func NewCfnStackParams(requiredParameterNames []string) *CfnStackParams {
	params := &CfnStackParams{
		params:             make([]*cloudformation.Parameter, 0),
		nameToKeys:         make(map[string]string),
		requiredParameters: make(map[string]bool),
	}
	for _, s := range requiredParameterNames {
		params.requiredParameters[s] = true
	}
	return params
}

// NewCfnStackParamsForUpdate creates a new object of CfnStackParams struct and populates it for updating the stack.
func NewCfnStackParamsForUpdate(requiredParameterNames []string, existingParams []*cloudformation.Parameter) (*CfnStackParams, error) {
	params := NewCfnStackParams(requiredParameterNames)
	for _, param := range existingParams {
		// Set UsePreviousValue = true for all the stack parameters.
		err := params.AddWithUsePreviousValue(aws.StringValue(param.ParameterKey), true)
		if err != nil {
			return nil, err
		}
	}
	return params, nil
}

// Add adds a key and the value for the same to the cloudformation parameters. If the key already
// exists, the value is overwritten,
func (s *CfnStackParams) Add(key string, value string) error {
	param, err := s.GetParameter(key)
	if err != nil {
		// either new parameter or bad state.
		if err != ParameterNotFoundError {
			// bad state.
			return err
		}
		// UsePreviousValue is false since we are explicitly setting the value for a parameter here.
		s.params = append(s.params, &cloudformation.Parameter{
			ParameterKey:     aws.String(key),
			ParameterValue:   aws.String(value),
			UsePreviousValue: aws.Bool(false),
		})
		s.nameToKeys[key] = value

	} else {
		// parameter found.
		param.ParameterValue = aws.String(value)
		// UsePreviousValue is false since we are explicitly setting the value for a parameter here.
		param.UsePreviousValue = aws.Bool(false)
		s.nameToKeys[key] = value
	}
	return nil
}

// AddWithUsePreviousValue adds a key to the stack parameters with UsePreviousValue set to the specified
// boolean value. This is used while creating parameter list required by the UpdateStack method.
func (s *CfnStackParams) AddWithUsePreviousValue(key string, usePreviousValue bool) error {
	param, err := s.GetParameter(key)
	if err != nil {
		// either new parameter or bad state.
		if err != ParameterNotFoundError {
			// bad state.
			return err
		}
		s.params = append(s.params, &cloudformation.Parameter{
			ParameterKey:     aws.String(key),
			UsePreviousValue: aws.Bool(usePreviousValue),
		})
		s.nameToKeys[key] = ""

	} else {
		// parameter found.
		param.UsePreviousValue = aws.Bool(usePreviousValue)
		s.nameToKeys[key] = ""
	}
	return nil
}

// Get gets the cloudformation parameters from the CfnStackParams object,
func (s *CfnStackParams) Get() []*cloudformation.Parameter {
	return s.params
}

// GetParameter gets the cloudformation parameter for a given key name. Returns an error if not found.
func (s *CfnStackParams) GetParameter(key string) (*cloudformation.Parameter, error) {
	_, exists := s.nameToKeys[key]
	if !exists {
		return nil, ParameterNotFoundError
	}

	for _, param := range s.params {
		if key == aws.StringValue(param.ParameterKey) {
			return param, nil
		}
	}

	return nil, fmt.Errorf("Invalid state: Could not find parameter key for %s", key)
}

// Validate validates that the cloudformation parameters contain all the required keys and that the values for these keys
// are not empty.
func (s *CfnStackParams) Validate() error {
	// TODO: Additional validation for fields. Example: are vpcAzs comma delimited? valid characters in cidr etc.

	// Validate if all the required parameters are present.
	validatedParams := make(map[string]bool)
	for key := range s.requiredParameters {
		param, err := s.GetParameter(key)
		if err != nil {
			return err
		}
		if err := validateParam(param, key); err != nil {
			return err
		}
		validatedParams[key] = true
	}

	// Validate if fields are set correctly for rest of the parameters.
	for _, param := range s.Get() {
		key := aws.StringValue(param.ParameterKey)
		if _, exists := validatedParams[key]; !exists {
			if err := validateParam(param, key); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateParams validates if a cloudformation Parameter is properly set.
func validateParam(param *cloudformation.Parameter, key string) error {
	val := aws.StringValue(param.ParameterValue)
	if val == "" {
		// If value is not set, we expect UsePreviousValue to be set.
		usePrevious := param.UsePreviousValue
		if !aws.BoolValue(usePrevious) {
			// aws.BoolValue does the nil check for us.
			return fmt.Errorf("ParameterValue and UsePreviousValue not set for parameter key '%s'", key)
		}
	}

	return nil
}
