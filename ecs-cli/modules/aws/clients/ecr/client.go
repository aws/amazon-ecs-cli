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
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	login "github.com/awslabs/amazon-ecr-credential-helper/ecr-login/api"
	"github.com/pkg/errors"
)

//go:generate mockgen.sh github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecr Client mock/$GOFILE
//go:generate mockgen.sh github.com/awslabs/amazon-ecr-credential-helper/ecr-login/api Client mock/credential-helper/login_mock.go
//go:generate mockgen.sh github.com/aws/aws-sdk-go/service/ecr/ecriface ECRAPI mock/sdk/ecriface_mock.go

const (
	CacheDir = "~/.ecs"
)
// ProcessImageDetails callback function for describe images
type ProcessImageDetails func(images []*ecr.ImageDetail) error

// ProcessRepositories callback function for describe repositories
type ProcessRepositories func(repositories []*string) error

// Client ECR interface
type Client interface {
	GetAuthorizationToken(repositoryID string) (*Auth, error)
	CreateRepository(repositoryName string) (string, error)
	RepositoryExists(repositoryName string) bool
	GetImages(repositoryNames []*string, tagStatus string, registryID string, processFn ProcessImageDetails) error
}

// ecrClient implements Client
type ecrClient struct {
	client      ecriface.ECRAPI
	loginClient login.Client
	params      *config.CliParams
	auth        *Auth
}

// NewClient Creates a new ECR client
func NewClient(params *config.CliParams) Client {
	client := ecr.New(params.Session, params.Session.Config)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	loginClient := login.DefaultClientFactory{}.NewClientWithOptions(login.Options{
		Session:  params.Session,
		Config:   params.Session.Config,
		CacheDir: CacheDir,
	})
	return newClient(params, client, loginClient)
}

func newClient(params *config.CliParams, client ecriface.ECRAPI, loginClient login.Client) Client {
	return &ecrClient{
		client:      client,
		loginClient: loginClient,
		params:      params,
	}
}

// Auth keeps track of the ECR auth
type Auth struct {
	ProxyEndpoint string
	Registry      string
	Username      string
	Password      string
}

func (c *ecrClient) GetAuthorizationToken(registryID string) (*Auth, error) {
	log.Debug("Getting authorization token...")
	auth, err := c.loginClient.GetCredentialsByRegistryID(registryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to serialize authorization token")
	}

	return &Auth{
		Username:      auth.Username,
		Password:      auth.Password,
		ProxyEndpoint: auth.ProxyEndpoint,
		Registry:      strings.Replace(auth.ProxyEndpoint, "https://", "", -1),
	}, nil
}

func (c *ecrClient) RepositoryExists(repositoryName string) bool {
	_, err := c.client.DescribeRepositories(&ecr.DescribeRepositoriesInput{RepositoryNames: []*string{&repositoryName}})
	log.WithFields(log.Fields{
		"repository": repositoryName,
	}).Debug("Check if repository exists")
	return err == nil
}

func (c *ecrClient) CreateRepository(repositoryName string) (string, error) {
	log.WithFields(log.Fields{
		"repository": repositoryName,
	}).Info("Creating repository")

	resp, err := c.client.CreateRepository(
		&ecr.CreateRepositoryInput{RepositoryName: aws.String(repositoryName)})
	if err != nil {
		return "", errors.Wrap(err, "unable to create repository")
	}
	if resp == nil || resp.Repository == nil {
		return "", errors.New("create repository response is empty")
	}

	log.Info("Repository created")
	return aws.StringValue(resp.Repository.RepositoryName), nil
}

func (c *ecrClient) GetImages(repositoryNames []*string, tagStatus string, registryID string, processFn ProcessImageDetails) error {
	log.Debug("Getting images")
	numOfImageCalls := 0
	err := c.describeRepositories(repositoryNames, registryID, func(repositories []*string) error {
		for _, repository := range repositories {
			err := c.describeImages(aws.StringValue(repository), tagStatus, registryID, processFn, numOfImageCalls)
			numOfImageCalls++
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}

func (c *ecrClient) describeRepositories(repositoryNames []*string, registryID string, outputFn ProcessRepositories) error {
	var outErr error

	input := &ecr.DescribeRepositoriesInput{}

	// Skip DescribeRepositories calls if repositoryNames are specified
	if repositoryNames != nil && len(repositoryNames) > 0 {
		if err := outputFn(repositoryNames); err != nil {
			return err
		}
		return nil
	}

	if registryID != "" {
		input.SetRegistryId(registryID)
	}

	err := c.client.DescribeRepositoriesPages(input, func(resp *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		repositoryNames = []*string{}
		for _, repository := range resp.Repositories {
			repositoryNames = append(repositoryNames, repository.RepositoryName)
		}
		if err := outputFn(repositoryNames); err != nil {
			outErr = err
			return false
		}
		return !lastPage
	})

	if err != nil {
		return err
	}
	return outErr
}

func (c *ecrClient) describeImages(repositoryName string, tagStatus string, registryID string, outputFn ProcessImageDetails, numOfCalls int) error {
	var outErr error

	filter := &ecr.DescribeImagesFilter{}
	if tagStatus != "" {
		filter.SetTagStatus(tagStatus)
	}

	input := &ecr.DescribeImagesInput{
		RepositoryName: aws.String(repositoryName),
		Filter:         filter,
	}

	if registryID != "" {
		input.SetRegistryId(registryID)
	}

	err := c.client.DescribeImagesPages(input, func(resp *ecr.DescribeImagesOutput, lastPage bool) bool {
		if err := outputFn(resp.ImageDetails); err != nil {
			outErr = err
			return false
		}
		if numOfCalls > 50 {
			outErr = errors.New("please specify the repository name if you wish to see more")
			return false
		}
		return !lastPage
	})

	if err != nil {
		return err
	}
	return outErr
}
