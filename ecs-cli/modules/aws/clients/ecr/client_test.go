// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package ecr

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecr/mock/sdk"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	clusterName = "test"
)

// mockReadWriter implements ReadWriter interface to return just the cluster
// field when performing read.
type mockReadWriter struct{}

func (rdwr *mockReadWriter) GetConfig() (*config.CliConfig, error) {
	return config.NewCliConfig(clusterName), nil
}

func (rdwr *mockReadWriter) ReadFrom(ecsConfig *config.CliConfig) error {
	return nil
}

func (rdwr *mockReadWriter) IsInitialized() (bool, error) {
	return true, nil
}

func (rdwr *mockReadWriter) Save(dest *config.Destination) error {
	return nil
}

func (rdwr *mockReadWriter) IsKeyPresent(section, key string) bool {
	return true
}

func TestGetAuthorizationToken(t *testing.T) {
	mockEcr, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	username := "username"
	password := "password"
	authorizationToken := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	proxyEndpoint := "https://proxyEndpoint"
	mockEcr.EXPECT().GetAuthorizationToken(gomock.Any()).Return(&ecr.GetAuthorizationTokenOutput{
		AuthorizationData: []*ecr.AuthorizationData{
			&ecr.AuthorizationData{
				AuthorizationToken: aws.String(authorizationToken),
				ProxyEndpoint:      aws.String(proxyEndpoint),
			},
		},
	}, nil)

	ecrAuth, err := client.GetAuthorizationToken("")
	assert.NoError(t, err, "GetAuthorizationToken")
	assert.Equal(t, username, ecrAuth.Username, "Expected username to match")
	assert.Equal(t, password, ecrAuth.Password, "Expected password to match")
	assert.Equal(t, proxyEndpoint, ecrAuth.ProxyEndpoint, "Expected proxyEndpoint to match")
}

func TestGetAuthorizationTokenErrorCase(t *testing.T) {
	mockEcr, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	mockEcr.EXPECT().GetAuthorizationToken(gomock.Any()).Return(nil, errors.New("something failed"))

	_, err := client.GetAuthorizationToken("")
	assert.Error(t, err, "Expected error while GetAuthorizationToken is called")
}

func TestRepositoryExists(t *testing.T) {
	mockEcr, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryName := "repositoryName"
	mockEcr.EXPECT().DescribeRepositories(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecr.DescribeRepositoriesInput)
		assert.Equal(t, repositoryName, aws.StringValue(req.RepositoryNames[0]), "Expected repositoryName to match")
	}).Return(nil, errors.New("something failed"))

	exists := client.RepositoryExists(repositoryName)
	assert.False(t, exists, "RepositoryExists should return false")
}

func TestRepositoryDoesNotExists(t *testing.T) {
	mockEcr, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryName := "repositoryName"

	mockEcr.EXPECT().DescribeRepositories(gomock.Any()).Return(&ecr.DescribeRepositoriesOutput{}, nil)

	exists := client.RepositoryExists(repositoryName)
	assert.True(t, exists, "RepositoryExists should return true")
}

func TestCreateRepository(t *testing.T) {
	mockEcr, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryName := "repositoryName"

	mockEcr.EXPECT().CreateRepository(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecr.CreateRepositoryInput)
		assert.Equal(t, repositoryName, aws.StringValue(req.RepositoryName), "Expected repositoryName to match")
	}).Return(&ecr.CreateRepositoryOutput{Repository: &ecr.Repository{RepositoryName: &repositoryName}}, nil)

	_, err := client.CreateRepository(repositoryName)
	assert.NoError(t, err, "Create Repository should not fail")
}

func TestCreateRepositoryErrorCase(t *testing.T) {
	mockEcr, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryName := "repositoryName"

	mockEcr.EXPECT().CreateRepository(gomock.Any()).Return(nil, errors.New("something failed"))

	_, err := client.CreateRepository(repositoryName)
	assert.Error(t, err, "Expected error while CreateRepository is called")
}

func setupTestController(t *testing.T) (*mock_ecriface.MockECRAPI, Client, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockEcr := mock_ecriface.NewMockECRAPI(ctrl)
	mockSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	client := newClient(&config.CliParams{Session: mockSession}, mockEcr)

	return mockEcr, client, ctrl
}
