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

package image

import (
	"errors"
	"flag"
	"os"
	"testing"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecr"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecr/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/sts/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/docker/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	ecrApi "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/fsouza/go-dockerclient"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	repository    = "repository"
	tag           = "tag-v0.1.0"
	image         = repository + ":" + tag
	registry      = "registry"
	registryID    = "123456789"
	repositoryURI = registry + "/" + repository
	clusterName   = "defaultCluster"
)

type mockReadWriter struct {
	clusterName string
}

func (rdwr *mockReadWriter) Get(cluster string, profile string) (*config.CLIConfig, error) {
	return config.NewCLIConfig(rdwr.clusterName), nil
}

func (rdwr *mockReadWriter) SaveProfile(configName string, profile *config.Profile) error {
	return nil
}

func (rdwr *mockReadWriter) SaveCluster(configName string, cluster *config.Cluster) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultProfile(configName string) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultCluster(configName string) error {
	return nil
}

func newMockReadWriter() *mockReadWriter {
	return &mockReadWriter{clusterName: clusterName}
}

func TestImagePush(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		mockDocker.EXPECT().TagImage(image, repositoryURI, tag).Return(nil),
		mockECR.EXPECT().RepositoryExists(repository).Return(false),
		mockECR.EXPECT().CreateRepository(repository).Return(repository, nil),
		mockDocker.EXPECT().PushImage(repositoryURI, tag, registry,
			docker.AuthConfiguration{}).Return(nil),
	)

	context := setAllPushImageFlags()
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.NoError(t, err, "Error pushing image")
}

func TestImagePushWithURI(t *testing.T) {
	repositoryWithURI := "012345678912.dkr.ecr.us-east-1.amazonaws.com/" + image

	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		// Skips GetAWSAccountID
		mockECR.EXPECT().GetAuthorizationToken(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		// Skips TagImage
		mockECR.EXPECT().RepositoryExists(repository).Return(false),
		mockECR.EXPECT().CreateRepository(repository).Return(repository, nil),
		mockDocker.EXPECT().PushImage(repositoryURI, tag, registry,
			docker.AuthConfiguration{}).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-push", 0)
	flagSet.Parse([]string{repositoryWithURI})
	context := cli.NewContext(nil, flagSet, nil)
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.NoError(t, err, "Error pushing image")
}

func TestImagePushWhenRepositoryExists(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		mockDocker.EXPECT().TagImage(image, repositoryURI, tag).Return(nil),
		mockECR.EXPECT().RepositoryExists(repository).Return(true),
		// Skips CreateRepository
		mockDocker.EXPECT().PushImage(repositoryURI, tag, registry,
			docker.AuthConfiguration{}).Return(nil),
	)

	context := setAllPushImageFlags()
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.NoError(t, err, "Error pushing image")
}

func TestImagePushWithNoArguments(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	flagSet := flag.NewFlagSet("ecs-cli-push", 0)
	context := cli.NewContext(nil, flagSet, nil)
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expect error pushing image")
}

func TestImagePushWithTooManyArguments(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	flagSet := flag.NewFlagSet("ecs-cli-push", 0)
	flagSet.Parse([]string{repository, image})
	context := cli.NewContext(nil, flagSet, nil)
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expect error pushing image")
}

func TestImagePushWhenGethAuthorizationTokenFail(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(nil, errors.New("something failed")),
	)

	context := setAllPushImageFlags()
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expect error pushing image")
}

func TestImagePushWhenTagImageFail(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		mockDocker.EXPECT().TagImage(image, repositoryURI, tag).Return(errors.New("something failed")),
	)

	context := setAllPushImageFlags()
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expect error pushing image")
}

func TestImagePushWhenCreateRepositoryFail(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		mockDocker.EXPECT().TagImage(image, repositoryURI, tag).Return(nil),
		mockECR.EXPECT().RepositoryExists(repository).Return(false),
		mockECR.EXPECT().CreateRepository(repository).Return("", errors.New("something failed")),
	)

	context := setAllPushImageFlags()
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expect error pushing image")
}

func TestImagePushFail(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		mockDocker.EXPECT().TagImage(image, repositoryURI, tag).Return(nil),
		mockECR.EXPECT().RepositoryExists(repository).Return(false),
		mockECR.EXPECT().CreateRepository(repository).Return(repository, nil),
		mockDocker.EXPECT().PushImage(repositoryURI, tag, registry,
			docker.AuthConfiguration{}).Return(errors.New("something failed")),
	)

	context := setAllPushImageFlags()
	err := pushImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expect error pushing image")
}

func TestImagePull(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		mockDocker.EXPECT().PullImage(repositoryURI, tag,
			docker.AuthConfiguration{}).Return(nil),
	)

	context := setAllPullImageFlags()
	err := pullImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.NoError(t, err, "Error pulling image")
}

func TestImagePullWithoutImage(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	flagSet := flag.NewFlagSet("ecs-cli-pull", 0)
	context := cli.NewContext(nil, flagSet, nil)
	err := pullImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expected error pulling image")
}

func TestImagePullWhenGetAuthorizationTokenFail(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(nil, errors.New("something failed")),
	)

	context := setAllPullImageFlags()
	err := pullImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expected error pulling image")
}

func TestImagePullFail(t *testing.T) {
	mockECR, mockDocker, mockSTS := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockSTS.EXPECT().GetAWSAccountID().Return(registryID, nil),
		mockECR.EXPECT().GetAuthorizationTokenByID(gomock.Any()).Return(&ecr.Auth{
			Registry: registry,
		}, nil),
		mockDocker.EXPECT().PullImage(repositoryURI, tag,
			docker.AuthConfiguration{}).Return(errors.New("something failed")),
	)

	context := setAllPullImageFlags()
	err := pullImage(context, newMockReadWriter(), mockDocker, mockECR, mockSTS)
	assert.Error(t, err, "Expected error pulling image")
}

func TestImageList(t *testing.T) {
	mockECR, _, _ := setupTestController(t)
	setupEnvironmentVar()

	imageDigest := "sha:2561234567"
	repositoryName := "repo-name"
	pushedAt := time.Unix(1489687380, 0)
	size := int64(1024)
	tags := aws.StringSlice([]string{"tag1", "tag2"})
	gomock.InOrder(
		mockECR.EXPECT().GetImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_, _, _, x interface{}) {
			funct := x.(ecr.ProcessImageDetails)
			funct([]*ecrApi.ImageDetail{&ecrApi.ImageDetail{
				ImageDigest:      aws.String(imageDigest),
				RepositoryName:   aws.String(repositoryName),
				ImagePushedAt:    &pushedAt,
				ImageSizeInBytes: aws.Int64(size),
				ImageTags:        tags,
			}})
		}).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-images", 0)
	context := cli.NewContext(nil, flagSet, nil)
	err := getImages(context, newMockReadWriter(), mockECR)
	assert.NoError(t, err, "Error listing images")
}

func TestImageListFail(t *testing.T) {
	mockECR, _, _ := setupTestController(t)
	setupEnvironmentVar()

	gomock.InOrder(
		mockECR.EXPECT().GetImages(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("something failed")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-images", 0)
	context := cli.NewContext(nil, flagSet, nil)
	err := getImages(context, newMockReadWriter(), mockECR)
	assert.Error(t, err, "Expected error listing images")
}

func TestSplitImageName(t *testing.T) {
	observedRegistryURI, observedRepo, observedTag, err := splitImageName(image, "[:]", "format")

	assert.Empty(t, observedRegistryURI, "RegistryURI should be empty")
	assert.Equal(t, repository, observedRepo, "Repository should match")
	assert.Equal(t, tag, observedTag, "Tag should match")
	assert.NoError(t, err, "Error splitting image name")
}

func TestSplitImageNameWithSha256(t *testing.T) {
	sha := "sha256:0b3787ac21ffb4edbd6710e0e60f991d5ded8d8a4f558209ef5987f73db4211a"
	expectedImage := repository + "@" + sha
	observedRegistryURI, observedRepo, observedTag, err := splitImageName(expectedImage, "[:|@]", "format")

	assert.Empty(t, observedRegistryURI, "RegistryURI should be empty")
	assert.Equal(t, repository, observedRepo, "Repository should match")
	assert.Equal(t, sha, observedTag, "Tag should match")
	assert.NoError(t, err, "Error splitting image name")
}

func TestSplitImageNameWithURI(t *testing.T) {
	uri := "012345678912.dkr.ecr.us-east-1.amazonaws.com"
	expectedImage := uri + "/" + repository
	observedRegistryURI, observedRepo, observedTag, err := splitImageName(expectedImage, "[:|@]", "format")

	assert.Equal(t, uri, observedRegistryURI, "RegistryURI should match")
	assert.Equal(t, repository, observedRepo, "Repository should match")
	assert.Empty(t, observedTag, "Tag should be empty")
	assert.NoError(t, err, "Error splitting image name")
}

func TestSplitImageNameErrorCase(t *testing.T) {
	invalidImage := "rep@sha256:0b3787ac21ffb4edbd6710e0e60f991d5ded8d8a4f558209ef5987f73db4211a"
	_, _, _, err := splitImageName(invalidImage, "[:]", "format")

	assert.Error(t, err, "Expected error splitting image name")
}

func setupTestController(t *testing.T) (*mock_ecr.MockClient, *mock_docker.MockClient, *mock_sts.MockClient) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECR := mock_ecr.NewMockClient(ctrl)
	mockDocker := mock_docker.NewMockClient(ctrl)
	mockSTS := mock_sts.NewMockClient(ctrl)

	return mockECR, mockDocker, mockSTS
}

func setupEnvironmentVar() {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()
}

func setAllPushImageFlags() *cli.Context {
	flagSet := flag.NewFlagSet("ecs-cli-push", 0)
	flagSet.Parse([]string{image})
	return cli.NewContext(nil, flagSet, nil)
}

func setAllPullImageFlags() *cli.Context {
	flagSet := flag.NewFlagSet("ecs-cli-pull", 0)
	flagSet.Parse([]string{image})
	return cli.NewContext(nil, flagSet, nil)
}
