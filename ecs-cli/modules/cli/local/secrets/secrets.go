// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Package secrets implements functions to decrypt container secrets defined in a ECS task definition.
package secrets

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
)

// SecretDecrypter wraps the DecryptSecret function.
//
// DecryptSecret returns the decrypted value of a secret given its ARN.
type SecretDecrypter interface {
	DecryptSecret(arn string) (string, error)
}

// A ContainerSecret labels an ECS secret with the container name it belongs to.
type ContainerSecret struct {
	containerName string
	secret        ecs.Secret
}

// NewContainerSecret returns a new container secret.
func NewContainerSecret(containerName, secretName, secretValue string) *ContainerSecret {
	return &ContainerSecret{
		containerName: containerName,
		secret: ecs.Secret{
			Name:      aws.String(secretName),
			ValueFrom: aws.String(secretValue),
		},
	}
}

// Decrypt returns the decrypted secret value.
func (cs *ContainerSecret) Decrypt(sd SecretDecrypter) (string, error) {
	return sd.DecryptSecret(aws.StringValue(cs.secret.ValueFrom))
}

// ServiceName returns whether the secret belongs to SSM or Secrets Manager.
// If it can't determine the service or the ARN belongs to a different service then it returns an error.
func (cs *ContainerSecret) ServiceName() (string, error) {
	secretARN := aws.StringValue(cs.secret.ValueFrom)
	parsedARN, err := arn.Parse(secretARN)
	if err != nil {
		if strings.Contains(err.Error(), "arn: invalid prefix") {
			// If the Systems Manager Parameter Store parameter exists in the same Region,
			// then you can use either the full ARN or name of the parameter.
			// See https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data.html#secrets-logconfig.
			return ssm.ServiceName, nil
		}
		return "", errors.Wrapf(err, "Could not determine the service name of %s", secretARN)
	}
	if parsedARN.Service == secretsmanager.ServiceName {
		return secretsmanager.ServiceName, nil
	}
	if parsedARN.Service == ssm.ServiceName {
		return ssm.ServiceName, nil
	}
	return "", errors.New(fmt.Sprintf("Unexpected service %s for secret %s", parsedARN.Service, secretARN))
}

// Name returns a unique name describing the container secret. The format of the name is "containerName_secretName".
func (cs *ContainerSecret) Name() string {
	return fmt.Sprintf("%s_%s", cs.containerName, aws.StringValue(cs.secret.Name))
}
