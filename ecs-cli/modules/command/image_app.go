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
	"io"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Sirupsen/logrus"
	ecrclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecr"
	stsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/sts"
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	dockerclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/docker"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	units "github.com/docker/go-units"
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
	stsClient := stsclient.NewClient(ecsParams)

	if err := pushImage(c, rdwr, dockerClient, ecrClient, stsClient); err != nil {
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
	stsClient := stsclient.NewClient(ecsParams)

	if err := pullImage(c, rdwr, dockerClient, ecrClient, stsClient); err != nil {
		logrus.Error("Error executing 'pull': ", err)
		return
	}
}

// ImageList lists images up to 1000 items from ECR repository
func ImageList(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error executing 'images': ", err)
		return
	}

	ecsParams, err := config.NewCliParams(c, rdwr)
	if err != nil {
		logrus.Error("Error executing 'images': ", err)
		return
	}

	ecrClient := ecrclient.NewClient(ecsParams)
	if err := getImages(c, rdwr, ecrClient); err != nil {
		logrus.Error("Error executing 'images': ", err)
		return
	}
}

func pushImage(c *cli.Context, rdwr config.ReadWriter, dockerClient dockerclient.Client, ecrClient ecrclient.Client, stsClient stsclient.Client) error {
	registryID := c.String(ecscli.RegistryIdFlag)
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

	registryID, err = getRegistryID(registryID, stsClient)
	if err != nil {
		return err
	}

	ecrAuth, err := ecrClient.GetAuthorizationToken(registryID)
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

func pullImage(c *cli.Context, rdwr config.ReadWriter, dockerClient dockerclient.Client, ecrClient ecrclient.Client, stsClient stsclient.Client) error {
	registryID := c.String(ecscli.RegistryIdFlag)
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

	registryID, err = getRegistryID(registryID, stsClient)
	if err != nil {
		return err
	}

	ecrAuth, err := ecrClient.GetAuthorizationToken(registryID)
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

type imageInfo struct {
	RepositoryName string
	Tag            string
	ImageDigest    string
	PushedAt       string
	Size           string
}

func getImages(c *cli.Context, rdwr config.ReadWriter, ecrClient ecrclient.Client) error {
	registryID := c.String(ecscli.RegistryIdFlag)
	args := c.Args() // repository names

	totalCount := 0
	w := tabwriter.NewWriter(os.Stdout, 20, 1, 3, ' ', 0)

	err := ecrClient.GetImages(aws.StringSlice(args), getTagStatus(c), registryID, func(imageDetails []*ecr.ImageDetail) error {
		// Prints all images in table
		for _, image := range imageDetails {
			info := imageInfo{
				RepositoryName: aws.StringValue(image.RepositoryName),
				ImageDigest:    aws.StringValue(image.ImageDigest),
			}
			info.PushedAt = units.HumanDuration(time.Now().UTC().Sub(time.Unix(image.ImagePushedAt.Unix(), 0))) + " ago"
			info.Size = units.HumanSizeWithPrecision(float64(aws.Int64Value(image.ImageSizeInBytes)), 3)
			if len(image.ImageTags) == 0 {
				info.Tag = "<none>"
				listImagesContent(w, info, totalCount)
				totalCount++
			}
			for _, tag := range image.ImageTags {
				info.Tag = aws.StringValue(tag)
				listImagesContent(w, info, totalCount)
				totalCount++
			}
		}
		return nil
	})
	w.Flush()
	if err != nil {
		return err
	}

	return nil
}

func listImagesContent(w *tabwriter.Writer, info imageInfo, count int) {
	if count%100 == 0 {
		w.Flush()
		fmt.Println()
		printImageRow(w, imageInfo{
			RepositoryName: "REPOSITORY NAME",
			Tag:            "TAG",
			ImageDigest:    "IMAGE DIGEST",
			PushedAt:       "PUSHED AT",
			Size:           "SIZE",
		})
	}
	printImageRow(w, info)
}

func printImageRow(w io.Writer, info imageInfo) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n",
		info.RepositoryName,
		info.Tag,
		info.ImageDigest,
		info.PushedAt,
		info.Size,
	)
}

func getTagStatus(c *cli.Context) string {
	if c.Bool(ecscli.TaggedFlag) && c.Bool(ecscli.UntaggedFlag) {
		return ""
	}

	if c.Bool(ecscli.TaggedFlag) {
		return ecr.TagStatusTagged
	}
	if c.Bool(ecscli.UntaggedFlag) {
		return ecr.TagStatusUntagged
	}

	return ""
}

func getRegistryID(registryID string, stsClient stsclient.Client) (string, error) {
	if registryID == "" {
		return stsClient.GetAWSAccountID()
	}
	return registryID, nil
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
