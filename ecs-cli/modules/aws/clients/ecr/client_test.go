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
	registryID     = "123456789012"
	repositoryName = "repo-name"
	imageDigest    = "sha:256"
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

	mockEcr.EXPECT().DescribeRepositories(gomock.Any()).Return(&ecr.DescribeRepositoriesOutput{}, nil)

	exists := client.RepositoryExists(repositoryName)
	assert.True(t, exists, "RepositoryExists should return true")
}

func TestCreateRepository(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	mockEcr.EXPECT().CreateRepository(gomock.Any()).Do(func(input interface{}) {
		req := input.(*ecr.CreateRepositoryInput)
		assert.Equal(t, repositoryName, aws.StringValue(req.RepositoryName), "Expected repositoryName to match")
	}).Return(&ecr.CreateRepositoryOutput{Repository: &ecr.Repository{RepositoryName: aws.String(repositoryName)}}, nil)

	_, err := client.CreateRepository(repositoryName)
	assert.NoError(t, err, "Create Repository should not fail")
}

func TestCreateRepositoryErrorCase(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	mockEcr.EXPECT().CreateRepository(gomock.Any()).Return(nil, errors.New("something failed"))

	_, err := client.CreateRepository(repositoryName)
	assert.Error(t, err, "Expected error while CreateRepository is called")
}

func TestGetImages(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryNames := []*string{aws.String(repositoryName)}
	imageDetail := &ecr.ImageDetail{RepositoryName: aws.String(repositoryName), ImageDigest: aws.String(imageDigest)}
	tagStatus := ecr.TagStatusTagged

	// Does not call describeRepositories when repositoryNames are specified
	mockEcr.EXPECT().DescribeImagesPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		req := x.(*ecr.DescribeImagesInput)
		assert.Equal(t, repositoryName, aws.StringValue(req.RepositoryName), "Expected repository name to match")
		assert.NotNil(t, req.Filter, "Expected filter not to be empty")
		assert.Equal(t, tagStatus, aws.StringValue(req.Filter.TagStatus), "Expected tag status to match")
		assert.Equal(t, registryID, aws.StringValue(req.RegistryId), "Expected registry id to match")

		funct := y.(func(output *ecr.DescribeImagesOutput, lastPage bool) bool)
		funct(&ecr.DescribeImagesOutput{ImageDetails: []*ecr.ImageDetail{imageDetail}}, true)
	}).Return(nil)

	err := client.GetImages(repositoryNames, tagStatus, registryID, func(imageDetails []*ecr.ImageDetail) error {
		assert.Equal(t, 1, len(imageDetails), "Expected imageDetails to be 1")
		assert.Equal(t, imageDetail.ImageDigest, imageDetails[0].ImageDigest, "Expected image digest to match")
		assert.Equal(t, imageDetail.RepositoryName, imageDetails[0].RepositoryName, "Expected repository name to match")
		return nil
	})
	assert.NoError(t, err, "Get Images should not fail")
}

func TestGetImagesPagination(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repos := []*ecr.Repository{
		&ecr.Repository{RepositoryName: aws.String("repo1")},
		&ecr.Repository{RepositoryName: aws.String("repo2")},
	}
	images := []*ecr.ImageDetail{
		&ecr.ImageDetail{ImageDigest: aws.String("sha:256first")},
		&ecr.ImageDetail{ImageDigest: aws.String("sha:256second")},
	}

	// Returns 2 repositories
	mockEcr.EXPECT().DescribeRepositoriesPages(gomock.Any(), gomock.Any()).Do(func(_, y interface{}) {
		funct := y.(func(output *ecr.DescribeRepositoriesOutput, lastPage bool) bool)
		callNextPage := funct(&ecr.DescribeRepositoriesOutput{Repositories: repos}, false)
		assert.True(t, callNextPage, "Expected to call next page")
	}).Return(nil)
	// Returns 1 image without pagination (lastPage=true)
	mockEcr.EXPECT().DescribeImagesPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		// Expects to be the first repo returned
		req := x.(*ecr.DescribeImagesInput)
		assert.Equal(t, repos[0].RepositoryName, req.RepositoryName, "Expected repository name to match")

		funct := y.(func(output *ecr.DescribeImagesOutput, lastPage bool) bool)
		funct(&ecr.DescribeImagesOutput{
			ImageDetails: []*ecr.ImageDetail{images[0]},
		}, true)
	}).Return(nil)
	// Returns 1 image with pagination (lastPage=false)
	mockEcr.EXPECT().DescribeImagesPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		// Expects to be the second repo returned
		req := x.(*ecr.DescribeImagesInput)
		assert.Equal(t, repos[1].RepositoryName, req.RepositoryName, "Expected repository name to match")

		funct := y.(func(output *ecr.DescribeImagesOutput, lastPage bool) bool)
		callNextPage := funct(&ecr.DescribeImagesOutput{
			ImageDetails: []*ecr.ImageDetail{images[1]},
		}, false)
		assert.True(t, callNextPage, "Expected to call next page")
	}).Return(nil)

	count := 0
	err := client.GetImages(nil, "", "", func(imageDetails []*ecr.ImageDetail) error {
		assert.Equal(t, 1, len(imageDetails), "Expected imageDetails to be 1")
		assert.Equal(t, images[count].ImageDigest, imageDetails[0].ImageDigest, "Expected image digest to match")
		count++
		return nil
	})
	assert.NoError(t, err, "Get Images should not fail")
}

func TestGetImagesErrorCase(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryNames := []*string{aws.String(repositoryName)}
	repositoryDetail := &ecr.Repository{RepositoryName: repositoryNames[0], RegistryId: aws.String(registryID)}
	imageDetail := &ecr.ImageDetail{RepositoryName: aws.String(repositoryName), ImageDigest: aws.String(imageDigest)}

	mockEcr.EXPECT().DescribeRepositoriesPages(gomock.Any(), gomock.Any()).Do(func(_, y interface{}) {
		funct := y.(func(output *ecr.DescribeRepositoriesOutput, lastPage bool) bool)
		funct(&ecr.DescribeRepositoriesOutput{Repositories: []*ecr.Repository{repositoryDetail}}, true)
	}).Return(nil)

	mockEcr.EXPECT().DescribeImagesPages(gomock.Any(), gomock.Any()).Do(func(_, y interface{}) {
		funct := y.(func(output *ecr.DescribeImagesOutput, lastPage bool) bool)
		funct(&ecr.DescribeImagesOutput{ImageDetails: []*ecr.ImageDetail{imageDetail}}, true)
	}).Return(nil)

	err := client.GetImages(nil, "", "", func(imageDetails []*ecr.ImageDetail) error {
		return errors.New("something failed")
	})
	assert.Error(t, err, "Get Images should not fail")
}

func TestDescribeRepositoriesErrorCase(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	mockEcr.EXPECT().DescribeRepositoriesPages(gomock.Any(), gomock.Any()).Return(errors.New("something failed"))

	err := client.GetImages(nil, "", "", nil)
	assert.Error(t, err, "Get Images should fail")
}

func TestDescribeImageErrorCase(t *testing.T) {
	mockEcr, _, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repositoryNames := []*string{aws.String(repositoryName)}
	repositoryDetail := &ecr.Repository{RepositoryName: repositoryNames[0], RegistryId: aws.String(registryID)}

	mockEcr.EXPECT().DescribeRepositoriesPages(gomock.Any(), gomock.Any()).Do(func(_, y interface{}) {
		funct := y.(func(output *ecr.DescribeRepositoriesOutput, lastPage bool) bool)
		funct(&ecr.DescribeRepositoriesOutput{Repositories: []*ecr.Repository{repositoryDetail}}, true)
	}).Return(nil)
	mockEcr.EXPECT().DescribeImagesPages(gomock.Any(), gomock.Any()).Return(errors.New("something failed"))

	err := client.GetImages(nil, "", "", nil)
	assert.Error(t, err, "Get Images should fail")
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
