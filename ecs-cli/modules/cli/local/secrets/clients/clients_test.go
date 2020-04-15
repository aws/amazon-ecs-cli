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

package clients

import (
	"testing"

	mock_ssmiface "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/amimetadata/mock/sdk"
	mock_secretsmanageriface "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/secretsmanager/mock/sdk"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestSSMDecrypter_DecryptSecret(t *testing.T) {
	testCases := map[string]struct {
		input          string
		wantedSecret   string
		setupDecrypter func(ctrl *gomock.Controller) *SSMDecrypter
	}{
		"with ARN in default region": {
			input:        "arn:aws:ssm:us-east-1:11111111111:parameter/TEST_DB_PASSWORD",
			wantedSecret: "1234",
			setupDecrypter: func(ctrl *gomock.Controller) *SSMDecrypter {
				defaultClient := mock_ssmiface.NewMockSSMAPI(ctrl)

				m := make(map[region]ssmiface.SSMAPI)
				m["default"] = defaultClient
				m["us-east-1"] = defaultClient

				gomock.InOrder(
					defaultClient.EXPECT().GetParameter(&ssm.GetParameterInput{
						Name:           aws.String("TEST_DB_PASSWORD"),
						WithDecryption: aws.Bool(true),
					}).Return(&ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String("1234"),
						},
					}, nil),
				)

				return &SSMDecrypter{
					SSMAPI:  defaultClient,
					clients: m,
				}
			},
		},
		"with ARN in different region": {
			input:        "arn:aws:ssm:us-west-2:11111111111:parameter/TEST_DB_PASSWORD",
			wantedSecret: "what??",
			setupDecrypter: func(ctrl *gomock.Controller) *SSMDecrypter {
				iadClient := mock_ssmiface.NewMockSSMAPI(ctrl)
				pdxClient := mock_ssmiface.NewMockSSMAPI(ctrl)

				m := make(map[region]ssmiface.SSMAPI)
				m["default"] = iadClient
				m["us-east-1"] = iadClient
				m["us-west-2"] = pdxClient

				gomock.InOrder(
					pdxClient.EXPECT().GetParameter(&ssm.GetParameterInput{
						Name:           aws.String("TEST_DB_PASSWORD"),
						WithDecryption: aws.Bool(true),
					}).Return(&ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String("what??"),
						},
					}, nil),

					iadClient.EXPECT().GetParameter(gomock.Any()).Times(0), // Should not have called IAD
				)

				return &SSMDecrypter{
					SSMAPI:  iadClient,
					clients: m,
				}
			},
		},
		"with ARN and forward slashes": {
			input:        "arn:aws:ssm:us-west-2:11111111111:parameter/TEST/DB/PASSWORD",
			wantedSecret: "ponies",
			setupDecrypter: func(ctrl *gomock.Controller) *SSMDecrypter {
				iadClient := mock_ssmiface.NewMockSSMAPI(ctrl)
				pdxClient := mock_ssmiface.NewMockSSMAPI(ctrl)

				m := make(map[region]ssmiface.SSMAPI)
				m["default"] = iadClient
				m["us-east-1"] = iadClient
				m["us-west-2"] = pdxClient

				gomock.InOrder(
					pdxClient.EXPECT().GetParameter(&ssm.GetParameterInput{
						Name:           aws.String("/TEST/DB/PASSWORD"),
						WithDecryption: aws.Bool(true),
					}).Return(&ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String("what??"),
						},
					}, nil),

					iadClient.EXPECT().GetParameter(gomock.Any()).Times(0), // Should not have called IAD
				)

				return &SSMDecrypter{
					SSMAPI:  iadClient,
					clients: m,
				}
			},
		},
		"without ARN": {
			input:        "TEST_DB_PASSWORD",
			wantedSecret: "hello",
			setupDecrypter: func(ctrl *gomock.Controller) *SSMDecrypter {
				iadClient := mock_ssmiface.NewMockSSMAPI(ctrl)
				pdxClient := mock_ssmiface.NewMockSSMAPI(ctrl)

				m := make(map[region]ssmiface.SSMAPI)
				m["default"] = iadClient
				m["us-east-1"] = iadClient
				m["us-west-2"] = pdxClient

				gomock.InOrder(
					iadClient.EXPECT().GetParameter(&ssm.GetParameterInput{
						Name:           aws.String("TEST_DB_PASSWORD"),
						WithDecryption: aws.Bool(true),
					}).Return(&ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String("hello"),
						},
					}, nil),

					pdxClient.EXPECT().GetParameter(gomock.Any()).Times(0), // Should not have called PDX
				)

				return &SSMDecrypter{
					SSMAPI:  iadClient,
					clients: m,
				}
			},
		},
		"with forward slash": {
			input:        "/TEST/DB/PASSWORD",
			wantedSecret: "hello",
			setupDecrypter: func(ctrl *gomock.Controller) *SSMDecrypter {
				iadClient := mock_ssmiface.NewMockSSMAPI(ctrl)
				pdxClient := mock_ssmiface.NewMockSSMAPI(ctrl)

				m := make(map[region]ssmiface.SSMAPI)
				m["default"] = iadClient
				m["us-east-1"] = iadClient
				m["us-west-2"] = pdxClient

				gomock.InOrder(
					iadClient.EXPECT().GetParameter(&ssm.GetParameterInput{
						Name:           aws.String("/TEST/DB/PASSWORD"),
						WithDecryption: aws.Bool(true),
					}).Return(&ssm.GetParameterOutput{
						Parameter: &ssm.Parameter{
							Value: aws.String("hello"),
						},
					}, nil),

					pdxClient.EXPECT().GetParameter(gomock.Any()).Times(0), // Should not have called PDX
				)

				return &SSMDecrypter{
					SSMAPI:  iadClient,
					clients: m,
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			decrypter := tc.setupDecrypter(ctrl)

			// When
			got, err := decrypter.DecryptSecret(tc.input)

			// Then
			require.Equal(t, tc.wantedSecret, got)
			require.NoError(t, err)
			require.Equal(t, decrypter.clients["default"], decrypter.SSMAPI, "Expected DecryptSecret() to reset the client after each call")
		})
	}
}

func TestSecretsManagerDecrypter_DecryptSecret(t *testing.T) {
	testCases := map[string]struct {
		input          string
		wantedSecret   string
		setupDecrypter func(ctrl *gomock.Controller) *SecretsManagerDecrypter
	}{
		"with ARN in IAD": {
			input:        "arn:aws:secretsmanager:us-east-1:11111111111:secret:alpha/efe/local-j0gCbT",
			wantedSecret: "verysafe",
			setupDecrypter: func(ctrl *gomock.Controller) *SecretsManagerDecrypter {
				iadClient := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)
				pdxClient := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)

				m := make(map[region]secretsmanageriface.SecretsManagerAPI)
				m["default"] = iadClient
				m["us-east-1"] = iadClient
				m["us-west-2"] = pdxClient

				gomock.InOrder(
					iadClient.EXPECT().GetSecretValue(&secretsmanager.GetSecretValueInput{
						SecretId: aws.String("arn:aws:secretsmanager:us-east-1:11111111111:secret:alpha/efe/local-j0gCbT"),
					}).Return(&secretsmanager.GetSecretValueOutput{
						SecretString: aws.String("verysafe"),
					}, nil),

					pdxClient.EXPECT().GetSecretValue(gomock.Any()).Times(0),
				)

				return &SecretsManagerDecrypter{
					SecretsManagerAPI: iadClient,
					clients:           m,
				}
			},
		},
		"with ARN in PDX": {
			input:        "arn:aws:secretsmanager:us-west-2:11111111111:secret:alpha/efe/local-j0gCbT",
			wantedSecret: "veryverysafe",
			setupDecrypter: func(ctrl *gomock.Controller) *SecretsManagerDecrypter {
				iadClient := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)
				pdxClient := mock_secretsmanageriface.NewMockSecretsManagerAPI(ctrl)

				m := make(map[region]secretsmanageriface.SecretsManagerAPI)
				m["default"] = iadClient
				m["us-east-1"] = iadClient
				m["us-west-2"] = pdxClient

				gomock.InOrder(
					pdxClient.EXPECT().GetSecretValue(&secretsmanager.GetSecretValueInput{
						SecretId: aws.String("arn:aws:secretsmanager:us-west-2:11111111111:secret:alpha/efe/local-j0gCbT"),
					}).Return(&secretsmanager.GetSecretValueOutput{
						SecretString: aws.String("veryverysafe"),
					}, nil),

					iadClient.EXPECT().GetSecretValue(gomock.Any()).Times(0), // Should not have called IAD
				)

				return &SecretsManagerDecrypter{
					SecretsManagerAPI: iadClient,
					clients:           m,
				}
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// Given
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			decrypter := tc.setupDecrypter(ctrl)

			// When
			got, err := decrypter.DecryptSecret(tc.input)

			// Then
			require.Equal(t, tc.wantedSecret, got)
			require.NoError(t, err)
		})
	}
}
