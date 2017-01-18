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

package docker

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/docker/dockeriface"
	"github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
)

//go:generate mockgen.sh github.com/aws/amazon-ecs-cli/ecs-cli/modules/docker Client mock/$GOFILE
//go:generate mockgen.sh github.com/aws/amazon-ecs-cli/ecs-cli/modules/docker/dockeriface DockerAPI dockeriface/mock/dockeriface_mock.go

const (
	DockerVersion_1_17 = "1.17"
)

// Client is an interface specifying the subset of
// github.com/fsouza/go-dockerclient.DockerClient that the agent uses.
type Client interface {
	PullImage(repository, tag string, auth docker.AuthConfiguration) error
	PushImage(repository, tag, registry string, auth docker.AuthConfiguration) error
	TagImage(image, repository, tag string) error
}

type dockerClient struct {
	client dockeriface.DockerAPI
}

// NewClient creates a new docker client
// TODO create interface for docker.NewVersionedClientFromEnv for testing
func NewClient() (Client, error) {
	client, err := docker.NewVersionedClientFromEnv(string(DockerVersion_1_17))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create docker client")
	}
	return newClient(client), nil
}

func newClient(client dockeriface.DockerAPI) Client {
	return &dockerClient{
		client: client,
	}
}

func (c *dockerClient) PushImage(repository, tag, registry string, auth docker.AuthConfiguration) error {
	log.WithFields(log.Fields{
		"repository": repository,
		"tag":        tag,
	}).Info("Pushing image")

	opts := docker.PushImageOptions{
		Name:     repository,
		Tag:      tag,
		Registry: registry,
	}

	if err := c.client.PushImage(opts, auth); err != nil {
		return errors.Wrap(err, "unable to push image")
	}
	log.Info("Image pushed")
	return nil
}

// Tags repository[:tag] to local docker image
func (c *dockerClient) TagImage(sourceImage, repository, tag string) error {
	log.WithFields(log.Fields{
		"source-image": sourceImage,
		"repository":   repository,
		"tag":          tag,
	}).Info("Tagging image")

	opts := docker.TagImageOptions{
		Repo: repository,
		Tag:  tag,
	}

	if err := c.client.TagImage(sourceImage, opts); err != nil {
		return errors.Wrap(err, "unable to tag image")
	}
	log.Info("Image tagged")
	return nil
}

func (c *dockerClient) PullImage(repository, tag string, auth docker.AuthConfiguration) error {
	log.WithFields(log.Fields{
		"repository": repository,
		"tag":        tag,
	}).Info("Pulling image")

	opts := docker.PullImageOptions{
		Repository: repository,
		Tag:        tag,
	}

	if err := c.client.PullImage(opts, auth); err != nil {
		return errors.Wrap(err, "unable to pull image")
	}
	log.Infof("Image pulled")
	return nil
}
