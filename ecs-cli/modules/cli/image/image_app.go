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
	"fmt"
	"io"
	"os"
	"regexp"
	"text/tabwriter"
	"time"

	ecrclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecr"
	stsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/sts"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/tagging"
	dockerclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/docker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecr"
	taggingSDK "github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	units "github.com/docker/go-units"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// const symbols and widths
const (
	MinWidth    = 20
	TabWidth    = 1
	Padding     = 3
	PaddingChar = ' '
	NumOfFlags  = 0
	PageSize    = 100

	// const formats
	PushImageFormat = "ECR_REPOSITORY[:TAG]"
	PullImageFormat = "ECR_REPOSITORY[:TAG|@DIGEST]"
	ListImageFormat = "[ECR_REPOSITORY]"
)

// ImagePush does ecr login, tag image, and push image to ECR repository
func ImagePush(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'push': ", err)
	}

	commandConfig, err := config.NewCommandConfig(c, rdwr)
	if err != nil {
		logrus.Fatal("Error executing 'push': ", err)
	}

	dockerClient, err := dockerclient.NewClient()
	if err != nil {
		logrus.Fatal("Error executing 'push': ", err)
	}

	ecrClient := getECRClient(c, commandConfig)
	stsClient := stsclient.NewClient(commandConfig)
	taggingClient := tagging.NewTaggingClient(commandConfig)

	region := aws.StringValue(commandConfig.Session.Config.Region)

	if err := pushImage(c, region, dockerClient, ecrClient, stsClient, taggingClient); err != nil {
		logrus.Fatal("Error executing 'push': ", err)
	}
}

// ImagePull does ecr login and pulls from ECR repository
func ImagePull(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'pull': ", err)
	}

	commandConfig, err := config.NewCommandConfig(c, rdwr)
	if err != nil {
		logrus.Fatal("Error executing 'pull': ", err)
	}

	dockerClient, err := dockerclient.NewClient()
	if err != nil {
		logrus.Fatal("Error executing 'pull': ", err)
	}

	ecrClient := getECRClient(c, commandConfig)
	stsClient := stsclient.NewClient(commandConfig)

	if err := pullImage(c, rdwr, dockerClient, ecrClient, stsClient); err != nil {
		logrus.Fatal("Error executing 'pull': ", err)
	}
}

// ImageList lists images up to 1000 items from ECR repository
func ImageList(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'images': ", err)
	}

	commandConfig, err := config.NewCommandConfig(c, rdwr)
	if err != nil {
		logrus.Fatal("Error executing 'images': ", err)
	}

	ecrClient := getECRClient(c, commandConfig)

	if err := getImages(c, rdwr, ecrClient); err != nil {
		logrus.Fatal("Error executing 'images': ", err)
		return
	}
}

func getECRClient(c *cli.Context, commandConfig *config.CommandConfig) ecrclient.Client {
	ecrClient := ecrclient.NewClient(commandConfig)

	useFips := c.Bool(flags.UseFIPSFlag)

	if useFips {
		fipsClient, err := ecrclient.NewFipsClient(commandConfig)
		if err != nil {
			logrus.Fatal("Error creating FIPS client: ", err)
		}
		ecrClient = fipsClient
	}
	return ecrClient
}

func pushImage(c *cli.Context, region string, dockerClient dockerclient.Client, ecrClient ecrclient.Client, stsClient stsclient.Client, taggingClient tagging.Client) error {
	registryID := c.String(flags.RegistryIdFlag)

	args := c.Args()

	if len(args) != 1 {
		return fmt.Errorf("ecs-cli push requires exactly 1 argument")
	}

	image := args[0]

	registryURI, repository, tag, err := splitImageName(image, "[:]", PushImageFormat)
	if err != nil {
		return err
	}

	// For tagging (need the full ARN) and ECR auth, we need the registry ID
	// We can get this either from the registry URI or from STS
	if registryURI == "" {
		registryID, err = getRegistryID(registryID, stsClient)
		if err != nil {
			return err
		}
	} else {
		registryID, err = getRegistryIDFromURI(registryURI)
		if err != nil {
			return err
		}
	}

	ecrAuth, err := getECRAuth(registryURI, registryID, stsClient, ecrClient)
	if err != nil {
		return err
	}

	repositoryURI := ecrAuth.Registry + "/" + repository

	// Tag image to ECR uri
	if registryURI == "" {
		if err := dockerClient.TagImage(image, repositoryURI, tag); err != nil {
			return err
		}
	}

	// Check if repo exists, create if not present
	if !ecrClient.RepositoryExists(repository) {
		if _, err := ecrClient.CreateRepository(repository); err != nil {
			return err
		}
	}

	if tagVal := c.String(flags.ResourceTagsFlag); tagVal != "" {
		tags, err := utils.GetTagsMap(tagVal)
		if err != nil {
			return err
		}
		logrus.WithField("repository", repository).Info("Tagging repository...")
		repoARN := getRepoARN(region, registryID, repository)
		logrus.Debugf("Using repository ARN: %s", repoARN)
		err = tagRepo(repoARN, tags, taggingClient)
		if err != nil {
			return err
		}
	}

	// Push Image to ECR
	dockerAuth := docker.AuthConfiguration{
		Username:      ecrAuth.Username,
		Password:      ecrAuth.Password,
		ServerAddress: ecrAuth.ProxyEndpoint,
	}

	err = dockerClient.PushImage(repositoryURI, tag, ecrAuth.Registry, dockerAuth)
	return err
}

func pullImage(c *cli.Context, rdwr config.ReadWriter, dockerClient dockerclient.Client, ecrClient ecrclient.Client, stsClient stsclient.Client) error {
	registryID := c.String(flags.RegistryIdFlag)
	args := c.Args()
	if len(args) != 1 {
		return fmt.Errorf("ecs-cli pull requires exactly 1 argument")
	}
	image := args[0]

	registryURI, repository, tag, err := splitImageName(image, "[:|@]", PullImageFormat)
	if err != nil {
		return err
	}

	registryID, err = getRegistryID(registryID, stsClient)
	if err != nil {
		return err
	}

	ecrAuth, err := getECRAuth(registryURI, registryID, stsClient, ecrClient)
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
	err = dockerClient.PullImage(repositoryURI, tag, dockerAuth)
	return err
}

type imageInfo struct {
	RepositoryName string
	Tag            string
	ImageDigest    string
	PushedAt       string
	Size           string
}

func getImages(c *cli.Context, rdwr config.ReadWriter, ecrClient ecrclient.Client) error {
	registryID := c.String(flags.RegistryIdFlag)
	args := c.Args() // repository names

	totalCount := 0

	w := tabwriter.NewWriter(os.Stdout, MinWidth, TabWidth, Padding, PaddingChar, NumOfFlags)

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
	return err
}

func listImagesContent(w *tabwriter.Writer, info imageInfo, count int) {
	if count%PageSize == 0 {
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
	if c.Bool(flags.TaggedFlag) && c.Bool(flags.UntaggedFlag) {
		return ""
	}

	if c.Bool(flags.TaggedFlag) {
		return ecr.TagStatusTagged
	}
	if c.Bool(flags.UntaggedFlag) {
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

func getECRAuth(registryURI string, registryID string,
	stsClient stsclient.Client, ecrClient ecrclient.Client) (*ecrclient.Auth, error) {

	if registryURI == "" {
		return ecrClient.GetAuthorizationTokenByID(registryID)
	}

	return ecrClient.GetAuthorizationToken(registryURI)
}

func splitImageName(image string, seperatorRegExp string, format string) (registry string, repository string, tag string, err error) {
	re := regexp.MustCompile(
		`^(?:((?:[a-zA-Z0-9][a-zA-Z0-9-_]*)\.dkr\.ecr(\-fips)?\.[a-zA-Z0-9\-_]+\.amazonaws\.com(?:\.cn)?)/)?` + // registry uri (Optional)
			`([0-9a-z\-_/]+)` + // repository
			`(?:` + seperatorRegExp + `([0-9A-Za-z_.\-:]+))?$`) // tag or sha (Optional)
	matches := re.FindStringSubmatch(image)
	if len(matches) == 0 {
		return "", "", "", fmt.Errorf("Please specify the image name in the correct format [%s]", format)
	}

	return matches[1], matches[3], matches[4], nil
}

func getRegistryIDFromURI(registryURI string) (string, error) {
	re := regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-_]*)\.dkr\.ecr(\-fips)?\.[a-zA-Z0-9\-_]+\.amazonaws(\.com)?(\.cn)?`)
	matches := re.FindStringSubmatch(registryURI)
	if len(matches) < 2 {
		return "", fmt.Errorf("Could not parse account ID from registry URI; URI=%s", registryURI)
	}
	return matches[1], nil
}

func tagRepo(repoARN string, tags map[string]*string, taggingClient tagging.Client) error {
	input := &taggingSDK.TagResourcesInput{
		ResourceARNList: aws.StringSlice([]string{
			repoARN,
		}),
		Tags: tags,
	}
	output, err := taggingClient.TagResources(input)
	if err != nil {
		return err
	}

	for resource, info := range output.FailedResourcesMap {
		return fmt.Errorf("Failed to tag respository %s; error=%s", resource, *info.ErrorMessage)
	}
	return nil
}

// arn:aws:ecr:region:account-id:repository/repository-name
func getRepoARN(region, accountID, repository string) string {
	repoARN := arn.ARN{
		Partition: utils.GetPartition(region),
		Service:   "ecr",
		Region:    region,
		AccountID: accountID,
		Resource:  "repository/" + repository,
	}
	return repoARN.String()
}
