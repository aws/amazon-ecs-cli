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

package command

import (
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
	ecrclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecr"
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	dockerclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/docker"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/urfave/cli"
)

const (
	SEPERATOR_AT    = "@"
	SEPERATOR_COLON = ":"
)

// ImagePush does ecr login, tag image, and push image to ECR repository
func ImagePush(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error executing 'push': ", err)
		return
	}

	ecsParams, err := config.NewCliParams(c, rdwr)
	if err != nil {
		logrus.Error("Error executing 'push': ", err)
		return
	}

	dockerClient, err := dockerclient.NewClient()
	if err != nil {
		logrus.Error("Error executing 'push': ", err)
		return
	}
	ecrClient := ecrclient.NewClient(ecsParams)

	if err := pushImage(c, rdwr, dockerClient, ecrClient); err != nil {
		logrus.Error("Error executing 'push': ", err)
		return
	}
}

// ImagePull does ecr login and pulls from ECR repository
func ImagePull(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error executing 'pull': ", err)
		return
	}

	ecsParams, err := config.NewCliParams(c, rdwr)
	if err != nil {
		logrus.Error("Error executing 'pull': ", err)
		return
	}

	dockerClient, err := dockerclient.NewClient()
	if err != nil {
		logrus.Error("Error executing 'pull': ", err)
		return
	}
	ecrClient := ecrclient.NewClient(ecsParams)

	if err := pullImage(c, rdwr, dockerClient, ecrClient); err != nil {
		logrus.Error("Error executing 'pull': ", err)
		return
	}
}

func pushImage(c *cli.Context, rdwr config.ReadWriter, dockerClient dockerclient.Client, ecrClient ecrclient.Client) error {
	registryId := c.String(ecscli.RegistryIdFlag)
	targetImage := c.String(ecscli.ToFlag)
	sourceImage := c.String(ecscli.FromFlag)
	args := c.Args()

	if len(args) == 0 && (sourceImage == "" || targetImage == "") {
		return fmt.Errorf("ecs-cli push requires exactly 1 argument or [--%s] and [--%s] flags", ecscli.FromFlag, ecscli.ToFlag)
	}

	if len(args) == 1 && (sourceImage != "" || targetImage != "") {
		return fmt.Errorf("ecs-cli push does not allow [--%s] and [--%s] flags to be used with arguments", ecscli.FromFlag, ecscli.ToFlag)
	}

	if len(args) > 1 {
		return fmt.Errorf("ecs-cli push requires exactly 1 argument")
	}

	if (sourceImage == "" && targetImage != "") || (sourceImage != "" && targetImage == "") {
		return fmt.Errorf("ecs-cli push requires [--%s] and [--%s] flags", ecscli.FromFlag, ecscli.ToFlag)
	}

	if len(args) == 1 {
		sourceImage = args[0]
		targetImage = args[0]
	}

	repository, tag, err := splitImageName(targetImage, SEPERATOR_COLON, PUSH_IMAGE_FORMAT)
	if err != nil {
		return err
	}

	ecrAuth, err := ecrClient.GetAuthorizationToken(registryId)
	if err != nil {
		return err
	}

	repositoryURI := ecrAuth.Registry + "/" + repository

	// Tag image to ECR uri
	if err := dockerClient.TagImage(sourceImage, repositoryURI, tag); err != nil {
		return err
	}

	// Check if repo exists, create if not present
	if !ecrClient.RepositoryExists(repository) {
		if _, err := ecrClient.CreateRepository(repository); err != nil {
			return err
		}
	}

	// Push Image to ECR
	dockerAuth := docker.AuthConfiguration{
		Username:      ecrAuth.Username,
		Password:      ecrAuth.Password,
		ServerAddress: ecrAuth.ProxyEndpoint,
	}
	if err := dockerClient.PushImage(repositoryURI, tag, ecrAuth.Registry, dockerAuth); err != nil {
		return err
	}

	return nil
}

func pullImage(c *cli.Context, rdwr config.ReadWriter, dockerClient dockerclient.Client, ecrClient ecrclient.Client) error {
	registryId := c.String(ecscli.RegistryIdFlag)
	args := c.Args()
	if len(args) != 1 {
		return fmt.Errorf("ecs-cli pull requires exactly 1 argument")
	}
	image := args[0]

	seperator := SEPERATOR_COLON
	if strings.Contains(image, SEPERATOR_AT) {
		seperator = SEPERATOR_AT
	}
	repository, tag, err := splitImageName(image, seperator, PULL_IMAGE_FORMAT)
	if err != nil {
		return err
	}

	ecrAuth, err := ecrClient.GetAuthorizationToken(registryId)
	if err != nil {
		return err
	}

	repositoryURI := ecrAuth.Registry + "/" + repository

	dockerAuth := docker.AuthConfiguration{
		Username:      ecrAuth.Username,
		Password:      ecrAuth.Password,
		ServerAddress: ecrAuth.ProxyEndpoint,
	}

	// Pull Image
	if err := dockerClient.PullImage(repositoryURI, tag, dockerAuth); err != nil {
		return err
	}

	return nil
}

func splitImageName(image string, seperator string, format string) (string, string, error) {
	s := strings.Split(image, seperator)
	if len(s) > 2 {
		return "", "", fmt.Errorf("Please specify the image name in the correct format [%s]", format)
	}
	repository := s[0]
	tag := ""
	if len(s) == 2 {
		tag = s[1]
	}
	return repository, tag, nil
}
