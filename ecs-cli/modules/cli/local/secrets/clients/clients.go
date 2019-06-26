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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/pkg/errors"
)

// SSMDecrypter represents a SSM client that implements the secrets.SecretDecrypter interface.
type SSMDecrypter struct {
	*ssm.SSM
}

// SecretsManagerDecrypter represents a SecretsManager client that implements the secrets.SecretDecrypter interface.
type SecretsManagerDecrypter struct {
	*secretsmanager.SecretsManager
}

// DecryptSecret returns the decrypted parameter value from SSM.
//
// If the parameter is an ARN then the decrypted value is retrieved from the appropriate region.
// If the parameter is just the name of the parameter then the decrypted value is retrieved from the default region.
func (d *SSMDecrypter) DecryptSecret(arnOrName string) (string, error) {
	defer func() {
		// Reset the region of the client in case another SSM secret uses only the param name instead of full ARN.
		d.SSM = ssm.New(session.Must(session.NewSessionWithOptions(session.Options{})))
	}()

	// If the value is an ARN we need to retrieve the parameter name and update the region of the client.
	paramName := arnOrName
	if parsedARN, err := arn.Parse(arnOrName); err == nil {
		resource := strings.Split(parsedARN.Resource, "/") // Resource is formatted as parameter/{paramName}.
		paramName = strings.Join(resource[1:], "")
		d.SSM = ssm.New(session.Must(session.NewSessionWithOptions(session.Options{
			Config: aws.Config{Region: aws.String(parsedARN.Region)},
		})))
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

// NewSSMDecrypter returns a new SSMDecrypter using the default region.
func NewSSMDecrypter() (*SSMDecrypter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create a new AWS session due to %v", err)
	}
	return &SSMDecrypter{
		ssm.New(sess),
	}, nil
}

// NewSecretsManagerDecrypter returns a new SecretsManagerDecrypter using the default region.
func NewSecretsManagerDecrypter() (*SecretsManagerDecrypter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create a new AWS session due to %v", err)
	}
	return &SecretsManagerDecrypter{
		secretsmanager.New(sess),
	}, nil
}
