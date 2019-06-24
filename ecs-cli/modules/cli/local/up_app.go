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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/secrets"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

type ssmDecrypter struct {
	client *ssm.SSM
}

type secretsManagerDecrypter struct {
	client *secretsmanager.SecretsManager
}

// DecryptSecret returns the decrypted parameter value from SSM.
func (d *ssmDecrypter) DecryptSecret(arnOrName string) (string, error) {
	val, err := d.client.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(arnOrName),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to retrieve decrypted secret from %s due to %v", arnOrName, err)
	}
	return *val.Parameter.Value, nil
}

// DecryptSecret returns the decrypter secret value from Secrets Manager.
func (d *secretsManagerDecrypter) DecryptSecret(arn string) (string, error) {
	val, err := d.client.GetSecretValue(&secretsmanager.GetSecretValueInput{
		SecretId: aws.String(arn),
	})
	if err != nil {
		return "", errors.Wrapf(err, "Failed to retrieve decrypted secret from %s due to %v", arn, err)
	}
	return *val.SecretString, nil
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

func readSecrets(config *composeV3.Config) []*secrets.ContainerSecret {
	var containerSecrets []*secrets.ContainerSecret
	for _, service := range config.Services {
		for label, secretARN := range service.Labels {
			if !strings.HasPrefix(label, converter.SecretLabelPrefix) {
				continue
			}
			namespaces := strings.Split(label, ".")
			secretName := namespaces[len(namespaces)-1]

			containerSecrets = append(containerSecrets, secrets.NewContainerSecret(service.Name, secretName, secretARN))
		}
	}
	return containerSecrets
}

func decryptSecrets(containerSecrets []*secrets.ContainerSecret) (envVars map[string]string) {
	ssmClient, err := newSSMDecrypter()
	secretsManagerClient, err := newSecretsManagerDecrypter()
	if err != nil {
		logrus.Fatal(err.Error())
	}

	envVars = make(map[string]string)
	for _, containerSecret := range containerSecrets {
		service, err := containerSecret.ServiceName()
		if err != nil {
			logrus.Fatalf("Failed to retrieve the service of the secret due to %v", err)
		}

		decrypted := ""
		err = nil
		if service == secretsmanager.ServiceName {
			decrypted, err = containerSecret.Decrypt(secretsManagerClient)
		}
		if service == ssm.ServiceName {
			decrypted, err = containerSecret.Decrypt(ssmClient)
		}
		if err != nil {
			logrus.Fatal(err.Error())
		}
		envVars[containerSecret.Name()] = decrypted
	}
	return
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

func newSSMDecrypter() (*ssmDecrypter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create a new AWS session due to %v", err)
	}
	return &ssmDecrypter{
		client: ssm.New(sess),
	}, nil
}

func newSecretsManagerDecrypter() (*secretsManagerDecrypter, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to create a new AWS session due to %v", err)
	}
	return &secretsManagerDecrypter{
		client: secretsmanager.New(sess),
	}, nil
}
