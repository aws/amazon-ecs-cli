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
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/cli/cli/compose/types"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalComposeFile(t *testing.T) {
	testCases := map[string]struct {
		composeContent string
		wantedConfig   types.Config
	}{
		"simple": {
			composeContent: `
version: "3.0"
services:
  nginx:
    environment:
      API_KEY: 1234
    image: nginx
    labels:
      ecs-local.secret.TestSecret1: arn:aws:secretsmanager:us-east-1:11111111111:secret:alpha/efe/local`,
			wantedConfig: types.Config{
				Version: "3.0",
				Services: []types.ServiceConfig{
					{
						Name:  "nginx",
						Image: "nginx",
						Environment: types.MappingWithEquals{
							"API_KEY": aws.String("1234"),
						},
						Labels: types.Labels{
							"ecs-local.secret.TestSecret1": "arn:aws:secretsmanager:us-east-1:11111111111:secret:alpha/efe/local",
						},
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			fname := fmt.Sprintf("unmarshal-test-%s", name)
			tmpfile, err := ioutil.TempFile("", fname)
			require.NoError(t, err, "Unexpected error in creating temp compose file")
			defer os.Remove(tmpfile.Name())

			if _, err := tmpfile.Write([]byte(tc.composeContent)); err != nil {
				require.NoError(t, err, "Unexpected error while writing temp compose file")
			}

			// When
			actualConfig, err := UnmarshalComposeFile(tmpfile.Name())

			// Then
			require.NoError(t, err, "Unmarshalling the Compose file should not have failed")
			require.Equal(t, tc.wantedConfig.Services, actualConfig.Services)
		})
	}
}
