// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package secretsmanager

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
)

// SMClient defines methods for interacting with the SecretsManagerAPI interface
type SMClient interface {
	CreateSecret(secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error)
	DescribeSecret(secretID string) (*secretsmanager.DescribeSecretOutput, error)
	ListSecrets(*string) (*secretsmanager.ListSecretsOutput, error)
	PutSecretValue(input secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error)
	GetSecretValue(string) (string, error)
}

type secretsManagerClient struct {
	client secretsmanageriface.SecretsManagerAPI
}

// NewSecretsManagerClient creates an instance of an secretsManagerClient
func NewSecretsManagerClient(config *config.CommandConfig) SMClient {
	client := secretsmanager.New(config.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())

	return newClient(client)
}

func newClient(client secretsmanageriface.SecretsManagerAPI) SMClient {
	return &secretsManagerClient{
		client: client,
	}
}

func (c *secretsManagerClient) CreateSecret(input secretsmanager.CreateSecretInput) (*secretsmanager.CreateSecretOutput, error) {
	output, err := c.client.CreateSecret(&input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *secretsManagerClient) DescribeSecret(secretID string) (*secretsmanager.DescribeSecretOutput, error) {
	request := secretsmanager.DescribeSecretInput{}
	request.SetSecretId(secretID)

	output, err := c.client.DescribeSecret(&request)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *secretsManagerClient) ListSecrets(nextToken *string) (*secretsmanager.ListSecretsOutput, error) {
	request := secretsmanager.ListSecretsInput{
		NextToken: nextToken,
	}
	output, err := c.client.ListSecrets(&request)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *secretsManagerClient) PutSecretValue(input secretsmanager.PutSecretValueInput) (*secretsmanager.PutSecretValueOutput, error) {
	output, err := c.client.PutSecretValue(&input)

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *secretsManagerClient) GetSecretValue(secretName string) (string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(secretName),
		VersionStage: aws.String("AWSCURRENT"), // VersionStage defaults to AWSCURRENT if unspecified
	}
	output, err := c.client.GetSecretValue(input)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeDecryptionFailure:
				// Secrets Manager can't decrypt the protected secret text using the provided KMS key.
				fmt.Println(secretsmanager.ErrCodeDecryptionFailure, aerr.Error())

			case secretsmanager.ErrCodeInternalServiceError:
				// An error occurred on the server side.
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())

			case secretsmanager.ErrCodeInvalidParameterException:
				// You provided an invalid value for a parameter.
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())

			case secretsmanager.ErrCodeInvalidRequestException:
				// You provided a parameter value that is not valid for the current state of the resource.
				fmt.Println(secretsmanager.ErrCodeInvalidRequestException, aerr.Error())

			case secretsmanager.ErrCodeResourceNotFoundException:
				// We can't find the resource that you asked for.
				fmt.Println(secretsmanager.ErrCodeResourceNotFoundException, aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			return "", err
		}
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secretString, decodedBinarySecret string

	if output.SecretString != nil {
		secretString = *output.SecretString
	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(output.SecretBinary)))
		len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, output.SecretBinary)
		if err != nil {
			return "", fmt.Errorf("Base64 Decode Error: %s", err)
		}
		decodedBinarySecret = string(decodedBinarySecretBytes[:len])
	}

	if secretString != "" {
		return secretString, nil
	}

	return decodedBinarySecret, nil
}
