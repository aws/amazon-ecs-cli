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

package docker

import (
	"errors"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/docker/dockeriface/mock"
	"github.com/fsouza/go-dockerclient"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPushImage(t *testing.T) {
	mockDocker, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repository := "repository"
	tag := "tag"
	registry := "registry"

	mockDocker.EXPECT().PushImage(gomock.Any(), gomock.Any()).Do(func(opts, _ interface{}) {
		optsInput := opts.(docker.PushImageOptions)
		assert.Equal(t, repository, optsInput.Name, "Expected repository to match")
		assert.Equal(t, tag, optsInput.Tag, "Expected tag to match")
		assert.Equal(t, registry, optsInput.Registry, "Expected registry to match")
	}).Return(nil)

	err := client.PushImage(repository, tag, registry, docker.AuthConfiguration{})
	assert.NoError(t, err, "Push Image")
}

func TestPushImageErrorCase(t *testing.T) {
	mockDocker, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repository := "repository"
	tag := "tag"
	registry := "registry"

	mockDocker.EXPECT().PushImage(gomock.Any(), gomock.Any()).Return(errors.New("something failed"))

	err := client.PushImage(repository, tag, registry, docker.AuthConfiguration{})
	assert.Error(t, err, "Expected error while PushImage is called")
}

func TestTagImage(t *testing.T) {
	mockDocker, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repository := "repository"
	tag := "tag"
	image := "image"

	mockDocker.EXPECT().TagImage(gomock.Any(), gomock.Any()).Do(func(imageInput, opts interface{}) {
		assert.Equal(t, image, imageInput, "Expected image to match")
		optsInput := opts.(docker.TagImageOptions)
		assert.Equal(t, repository, optsInput.Repo, "Expected request.Repo to match")
		assert.Equal(t, tag, optsInput.Tag, "Expected request.Tag to match")
	}).Return(nil)

	err := client.TagImage(image, repository, tag)
	assert.NoError(t, err, "Tag Image")
}

func TestTagImageErrorCase(t *testing.T) {
	mockDocker, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repository := "repository"
	tag := "tag"
	image := "image"

	mockDocker.EXPECT().TagImage(gomock.Any(), gomock.Any()).Return(errors.New("something failed"))

	err := client.TagImage(image, repository, tag)
	assert.Error(t, err, "Expected error while TagImage is called")
}

func TestPullImage(t *testing.T) {
	mockDocker, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repository := "repository"
	tag := "tag"

	mockDocker.EXPECT().PullImage(gomock.Any(), gomock.Any()).Do(func(opts, auth interface{}) {
		optsInput := opts.(docker.PullImageOptions)
		assert.Equal(t, repository, optsInput.Repository, "Expected request.Repository to match")
		assert.Equal(t, tag, optsInput.Tag, "Expected request.Tag to match")
	}).Return(nil)

	err := client.PullImage(repository, tag, docker.AuthConfiguration{})
	assert.NoError(t, err, "Pull Image")
}

func TestPullImageErrorCase(t *testing.T) {
	mockDocker, client, ctrl := setupTestController(t)
	defer ctrl.Finish()

	repository := "repository"
	tag := "tag"

	mockDocker.EXPECT().PullImage(gomock.Any(), gomock.Any()).Return(errors.New("something failed"))

	err := client.PullImage(repository, tag, docker.AuthConfiguration{})
	assert.Error(t, err, "Expected error while PullImage is called")
}

func setupTestController(t *testing.T) (*mock_dockeriface.MockDockerAPI, Client, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockDocker := mock_dockeriface.NewMockDockerAPI(ctrl)
	client := newClient(mockDocker)

	return mockDocker, client, ctrl
}
