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

// Client ECR interface
type Client interface {
	GetAuthorizationToken(repositoryID string) (*Auth, error)
	CreateRepository(repositoryName string) (string, error)
	RepositoryExists(repositoryName string) bool
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
