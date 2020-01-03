// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package converter

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestConvertToCompose(t *testing.T) {
	testCases := map[string]struct {
		input         *ecs.TaskDefinition
		wantedCompose *composeV3.Config
		wantedError   error
	}{
		"nil task definition": {
			input:         nil,
			wantedCompose: nil,
			wantedError:   errors.New("task definition cannot be nil"),
		},
		"no containers": {
			input: &ecs.TaskDefinition{
				ContainerDefinitions: []*ecs.ContainerDefinition{},
			},
			wantedCompose: nil,
			wantedError:   errors.New("task definition needs to have container definitions"),
		},
		"single container definition": {
			input: &ecs.TaskDefinition{
				ContainerDefinitions: []*ecs.ContainerDefinition{
					{
						Name: aws.String("app"),
					},
				},
			},
			wantedCompose: &composeV3.Config{
				Version: composeVersion,
				Services: composeV3.Services{
					{
						Name: "app",
						Logging: &composeV3.LoggingConfig{
							Driver: jsonFileLogDriver,
						},
					},
				},
			},
			wantedError: nil,
		},
		"multiple container definitions": {
			input: &ecs.TaskDefinition{
				ContainerDefinitions: []*ecs.ContainerDefinition{
					{
						Name: aws.String("app"),
					},
					{
						Name: aws.String("envoyproxy"),
					},
				},
			},
			wantedCompose: &composeV3.Config{
				Version: composeVersion,
				Services: composeV3.Services{
					{
						Name: "app",
						Logging: &composeV3.LoggingConfig{
							Driver: jsonFileLogDriver,
						},
					},
					{
						Name: "envoyproxy",
						Logging: &composeV3.LoggingConfig{
							Driver: jsonFileLogDriver,
						},
					},
				},
			},
			wantedError: nil,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			actualCompose, actualErr := ConvertToComposeOverride(tc.input)

			if tc.wantedError != nil {
				require.EqualError(t, actualErr, tc.wantedError.Error())
			} else {
				require.Equal(t, tc.wantedCompose, actualCompose)
			}
		})
	}
}
