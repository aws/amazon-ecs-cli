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

package local

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/converter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/docker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/localproject"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type containerSecret struct {
	containerName string
	ecs.Secret
}

// Up creates a Compose file from an ECS task definition and runs it locally.
//
// The Amazon ECS Local Endpoints container needs to be running already for any local ECS task to work
// (see https://github.com/awslabs/amazon-ecs-local-container-endpoints).
// If the container is not running, this command creates a new network for all local ECS tasks to join
// and communicate with the Amazon ECS Local Endpoints container.
func Up(c *cli.Context) {
	// TODO When we don't provide any flags, Create() should just check if a "./docker-compose.local.yml" already exists.
	// If so then Create() should do nothing else, otherwise it should error.
	Create(c)

	network.Setup(docker.NewClient())

	config := readComposeFile(c)
	secrets := readSecrets(config)
	envVars := decryptSecrets(secrets)
	upComposeFile(config, envVars)
}

func readComposeFile(c *cli.Context) *composeV3.Config {
	filename := localproject.LocalOutDefaultFileName
	if c.String(flags.LocalOutputFlag) != "" {
		filename = c.String(flags.LocalOutputFlag)
	}
	config, err := converter.UnmarshalComposeFile(filename)
	if err != nil {
		logrus.Fatalf("Failed to unmarshal Compose file %s due to %v", filename, err)
	}
	return config
}

func readSecrets(config *composeV3.Config) []containerSecret {
	var secrets []containerSecret
	for _, service := range config.Services {
		for label, arn := range service.Labels {
			if !strings.HasPrefix(label, converter.SecretLabelPrefix) {
				continue
			}
			namespaces := strings.Split(label, ".")
			name := namespaces[len(namespaces)-1]
			secrets = append(secrets, containerSecret{
				containerName: service.Name,
				Secret: ecs.Secret{
					Name:      aws.String(name),
					ValueFrom: aws.String(arn),
				},
			})
		}
	}
	return secrets
}

func decryptSecrets(secrets []containerSecret) (envVars map[string]string) {
	sess, err := session.NewSession()
	if err != nil {
		logrus.Fatalf("Failed to create a new AWS session due to %v", err)
	}
	ssmClient := ssm.New(sess)
	secretsManagerClient := secretsmanager.New(sess)

	envVars = make(map[string]string)
	for _, secret := range secrets {
		service, err := serviceNameOf(aws.StringValue(secret.ValueFrom))
		if err != nil {
			logrus.Fatalf("Failed to retrieve the service of the secret due to %v", err)
		}

		// See https://github.com/aws/amazon-ecs-cli/issues/797
		name := fmt.Sprintf("%s_%s", secret.containerName, *secret.Name)
		if service == secretsmanager.ServiceName {
			val, err := secretsManagerClient.GetSecretValue(&secretsmanager.GetSecretValueInput{
				SecretId: secret.ValueFrom,
			})
			if err != nil {
				logrus.Fatalf("Failed to retrieve secret value with ARN %s due to %v", *secret.ValueFrom, err)
			}
			envVars[name] = *val.SecretString
		}
		if service == ssm.ServiceName {
			val, err := ssmClient.GetParameter(&ssm.GetParameterInput{
				Name:           secret.ValueFrom,
				WithDecryption: aws.Bool(true),
			})
			if err != nil {
				logrus.Fatalf("Failed to retrieve parameter value with ARN %s due to %v", *secret.ValueFrom, err)
			}
			envVars[name] = *val.Parameter.Value
		}
	}
	return
}

// serviceNameOf returns the service name of the secret based on its ARN value.
// It can be from either from SSM or Secrets Manager service.
func serviceNameOf(value string) (string, error) {
	parsedARN, err := arn.Parse(value)
	if err != nil {
		if strings.Contains(err.Error(), "arn: invalid prefix") {
			// If the Systems Manager Parameter Store parameter exists in the same Region,
			// then you can use either the full ARN or name of the parameter.
			return ssm.ServiceName, nil
		}
		return "", errors.Wrapf(err, "Could not determine the service name of %s", value)
	}
	if parsedARN.Service == secretsmanager.ServiceName {
		return secretsmanager.ServiceName, nil
	}
	if parsedARN.Service == ssm.ServiceName {
		return ssm.ServiceName, nil
	}
	return "", errors.Wrapf(err, "Unexpected service %s for secret %s", parsedARN.Service, value)
}

func upComposeFile(config *composeV3.Config, envVars map[string]string) {
	var envs []string
	for env, val := range envVars {
		envs = append(envs, fmt.Sprintf("%s=%s", env, val))
	}

	cmd := exec.Command("docker-compose", "-f", config.Filename, "up", "-d")
	cmd.Env = envs

	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Fatalf("Failed to run docker-compose up due to %v", err)
	}
	fmt.Printf("Compose out: %s\n", string(out)) // TODO logrus?
}
