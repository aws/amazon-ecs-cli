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

package secrets

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestName(t *testing.T) {
	testCases := map[string]struct {
		input  *ContainerSecret
		wanted string
	}{
		"simple": {
			input:  NewContainerSecret("mongodb", "DB_PASSWORD", "arn:aws:secretsmanager:us-east-1:11111111111:secret:alpha/efe/local"),
			wanted: "mongodb_DB_PASSWORD",
		},
		"complex": {
			input:  NewContainerSecret("mongodb", "DB_PASSWORD", "arn:aws:secretsmanager:us-east-1:11111111111:secret:alpha/efe/local/mongo/aws"),
			wanted: "mongodb_DB_PASSWORD",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.input.Name()
			require.Equal(t, tc.wanted, got)
		})
	}
}

func TestServiceName(t *testing.T) {
	testCases := map[string]struct {
		input             *ContainerSecret
		wantedServiceName string
		wantedError       error
	}{
		"ssm if not an ARN": {
			input:             NewContainerSecret("mongodb", "DB_PASSWORD", "DB_PASSWORD_PARAM"),
			wantedServiceName: ssm.ServiceName,
			wantedError:       nil,
		},
		"unexpected parsing error": {
			input:             NewContainerSecret("mongodb", "DB_PASSWORD", "arn:aws:INVALID"),
			wantedServiceName: "",
			wantedError:       errors.New("some error"),
		},
		"secretsmanager ARN": {
			input:             NewContainerSecret("mongodb", "DB_PASSWORD", "arn:aws:secretsmanager:us-east-1:11111111111:secret:alpha/efe/local"),
			wantedServiceName: secretsmanager.ServiceName,
			wantedError:       nil,
		},
		"ssm ARN": {
			input:             NewContainerSecret("mongodb", "DB_PASSWORD", "arn:aws:ssm:us-east-1:11111111111:parameter/TEST_DB_PASSWORD"),
			wantedServiceName: ssm.ServiceName,
			wantedError:       nil,
		},
		"other service ARN": {
			input:             NewContainerSecret("mongodb", "DB_PASSWORD", "arn:aws:ecs:us-east-1:11111111111:task-definition/ABCD:1"),
			wantedServiceName: "",
			wantedError:       errors.New("some error"),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			nameGotten, errorGotten := tc.input.ServiceName()

			require.Equal(t, tc.wantedServiceName, nameGotten)
			if tc.wantedError != nil {
				require.Errorf(t, errorGotten, "Expected the fn to error but instead got the name %s", nameGotten)
			} else {
				require.NoError(t, tc.wantedError)
			}
		})
	}
}
