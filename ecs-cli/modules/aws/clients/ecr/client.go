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
	"encoding/base64"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecr/ecriface"
	"github.com/pkg/errors"
)

//go:generate mockgen.sh github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecr Client mock/$GOFILE
//go:generate mockgen.sh github.com/aws/aws-sdk-go/service/ecr/ecriface ECRAPI mock/sdk/ecriface_mock.go

// Client ECR interface
type Client interface {
	GetAuthorizationToken(repositoryID string) (*Auth, error)
	CreateRepository(repositoryName string) (string, error)
	RepositoryExists(repositoryName string) bool
}

// ecrClient implements Client
type ecrClient struct {
	client ecriface.ECRAPI
	params *config.CliParams
	auth   *Auth
}

// NewClient Creates a new ECR client
func NewClient(params *config.CliParams) Client {
	client := ecr.New(session.New(params.Session.Config))
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return newClient(params, client)
}

func newClient(params *config.CliParams, client ecriface.ECRAPI) Client {
	return &ecrClient{
		client: client,
		params: params,
	}
}

// Auth keeps track of the ECR auth
type Auth struct {
	AuthorizationToken string
	ProxyEndpoint      string
	Username           string
	Password           string
	Registry           string
}

func (c *ecrClient) GetAuthorizationToken(repositoryID string) (*Auth, error) {
	log.Info("Getting authorization token...")
	var repositoryIDs []*string
	if repositoryID != "" {
		repositoryIDs = []*string{aws.String(repositoryID)}
	}
	resp, err := c.client.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{RegistryIds: repositoryIDs})
	if err != nil {
		return nil, errors.Wrap(err, "ecr: failed to get authorization token")
	}

	if len(resp.AuthorizationData) < 1 {
		return nil, errors.New("ecr: no authorization token received")
	}
	ecrAuth := Auth{
		AuthorizationToken: aws.StringValue(resp.AuthorizationData[0].AuthorizationToken),
		ProxyEndpoint:      aws.StringValue(resp.AuthorizationData[0].ProxyEndpoint),
	}

	token, err := base64.StdEncoding.DecodeString(ecrAuth.AuthorizationToken)
	if err != nil {
		return nil, errors.Wrap(err, "ecr: failed to serialize authorization token")
	}
	auth := strings.SplitN(string(token), ":", 2)
	if len(auth) < 2 {
		return nil, errors.Wrap(err, "ecr: failed to serialize authorization token")
	}
	ecrAuth.Username = auth[0]
	ecrAuth.Password = auth[1]
	ecrAuth.Registry = strings.Replace(ecrAuth.ProxyEndpoint, "https://", "", -1)

	return &ecrAuth, nil
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
