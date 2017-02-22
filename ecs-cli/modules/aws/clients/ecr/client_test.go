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
	"errors"
	"testing"

	mock_login "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecr/mock/credential-helper"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecr/mock/sdk"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	login "github.com/awslabs/amazon-ecr-credential-helper/ecr-login/api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	registryID = "123456789012"
)

func TestGetAuthorizationToken(t *testing.T) {
	_, mockLogin, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	username := "username"
	password := "password"
	registry := "proxyEndpoint"
	proxyEndpoint := "https://" + registry
	mockLogin.EXPECT().GetCredentialsByRegistryID(gomock.Any()).Return(&login.Auth{
		Username:      username,
		Password:      password,
		ProxyEndpoint: proxyEndpoint,
	}, nil)

	ecrAuth, err := client.GetAuthorizationToken(registryID)
	assert.NoError(t, err, "GetAuthorizationToken")
	assert.Equal(t, username, ecrAuth.Username, "Expected username to match")
	assert.Equal(t, password, ecrAuth.Password, "Expected password to match")
	assert.Equal(t, proxyEndpoint, ecrAuth.ProxyEndpoint, "Expected proxyEndpoint to match")
	assert.Equal(t, registry, ecrAuth.Registry, "Expected registry to match")
}

func TestGetAuthorizationTokenErrorCase(t *testing.T) {
	_, mockLogin, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	mockLogin.EXPECT().GetCredentialsByRegistryID(gomock.Any()).Return(nil, errors.New("something failed"))

	_, err := client.GetAuthorizationToken(registryID)
	assert.Error(t, err, "Expected error while GetAuthorizationToken is called")
}

func TestRepositoryExists(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
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
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryName := "repositoryName"

	mockEcr.EXPECT().DescribeRepositories(gomock.Any()).Return(&ecr.DescribeRepositoriesOutput{}, nil)

	exists := client.RepositoryExists(repositoryName)
	assert.True(t, exists, "RepositoryExists should return true")
}

func TestCreateRepository(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
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
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryName := "repositoryName"

	mockEcr.EXPECT().CreateRepository(gomock.Any()).Return(nil, errors.New("something failed"))

	_, err := client.CreateRepository(repositoryName)
	assert.Error(t, err, "Expected error while CreateRepository is called")
}

func setupTestController(t *testing.T) (*mock_ecriface.MockECRAPI, *mock_login.MockClient, Client, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockEcr := mock_ecriface.NewMockECRAPI(ctrl)
	mockLogin := mock_login.NewMockClient(ctrl)
	mockSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	client := newClient(&config.CliParams{Session: mockSession}, mockEcr, mockLogin)

	return mockEcr, mockLogin, client, ctrl
}
