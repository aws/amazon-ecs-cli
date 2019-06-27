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

// Package clients implements the secrets.SecretDecrypter interface for AWS clients.
package clients

import (
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/pkg/errors"
)

// region represents an AWS region.
type region string

// SSMDecrypter represents a SSM client that implements the secrets.SecretDecrypter interface.
type SSMDecrypter struct {
	ssmiface.SSMAPI

	// clients holds regional SSM clients.
	// The region "default" is always present and points to the same session as the user's default region.
	clients map[region]ssmiface.SSMAPI
}

// SecretsManagerDecrypter represents a SecretsManager client that implements the secrets.SecretDecrypter interface.
type SecretsManagerDecrypter struct {
	secretsmanageriface.SecretsManagerAPI
}

// DecryptSecret returns the decrypted parameter value from SSM.
//
// If the parameter is an ARN then the decrypted value is retrieved from the appropriate region.
// If the parameter is just the name of the parameter then the decrypted value is retrieved from the default region.
func (d *SSMDecrypter) DecryptSecret(arnOrName string) (string, error) {
	defer func() {
		// Reset the region of the client in case another SSM secret uses only the param name instead of full ARN.
		d.SSMAPI = d.getClient("default")
	}()

	// If the value is an ARN we need to retrieve the parameter name and update the region of the client.
	paramName := arnOrName
	if parsedARN, err := arn.Parse(arnOrName); err == nil {
		resource := strings.Split(parsedARN.Resource, "/") // Resource is formatted as parameter/{paramName}.
		paramName = strings.Join(resource[1:], "")
		d.SSMAPI = d.getClient(region(parsedARN.Region))
	}

	val, err := d.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to retrieve decrypted secret from %s due to %v", arnOrName, err)
	}
	return *val.Parameter.Value, nil
}

// getClient returns the SSM client for a given region.
// If there is no client available for that region, then creates and caches it.
func (d *SSMDecrypter) getClient(r region) ssmiface.SSMAPI {
	if c, ok := d.clients[r]; ok {
		return c
	}
	c := ssm.New(session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(string(r))},
	})))
	d.clients[r] = c
	return c
}

// DecryptSecret returns the decrypted secret value from Secrets Manager.
func (d *SecretsManagerDecrypter) DecryptSecret(arn string) (string, error) {
	val, err := d.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to retrieve decrypted secret from %s due to %v", arn, err)
	}
	return *val.SecretString, nil
}

// NewSSMDecrypter returns a new SSMDecrypter using the ECS CLI's default region.
func NewSSMDecrypter() (*SSMDecrypter, error) {
	sess, err := getDefaultSession()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create a new AWS session due to %v", err)
	}
	defaultClient := ssm.New(sess)

	clients := make(map[region]ssmiface.SSMAPI)
	clients["default"] = defaultClient
	clients[region(aws.StringValue(sess.Config.Region))] = defaultClient
	return &SSMDecrypter{
		SSMAPI:  defaultClient,
		clients: clients,
	}, nil
}

// NewSecretsManagerDecrypter returns a new SecretsManagerDecrypter using the ECS CLI's default region.
func NewSecretsManagerDecrypter() (*SecretsManagerDecrypter, error) {
	sess, err := getDefaultSession()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create a new AWS session due to %v", err)
	}
	return &SecretsManagerDecrypter{
		secretsmanager.New(sess),
	}, nil
}

// getDefaultSession returns a session for AWS clients where the region is set to the ECS CLI's default region.
// See https://github.com/aws/amazon-ecs-cli/blob/master/README.md#order-of-resolution-for-region
func getDefaultSession() (*session.Session, error) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		return nil, err
	}
	cmdConf, err := config.NewCommandConfig(nil, rdwr)
	if err != nil {
		return nil, err
	}
	return cmdConf.Session, nil
}
