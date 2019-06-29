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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/converter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/docker"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/localproject"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/options"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/secrets"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/secrets/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Up creates a Compose file from an ECS task definition and runs it locally.
//
// The Amazon ECS Local Endpoints container needs to be running already for any local ECS task to work
// (see https://github.com/awslabs/amazon-ecs-local-container-endpoints).
// If the container is not running, this command creates a new network for all local ECS tasks to join
// and communicate with the Amazon ECS Local Endpoints container.
func Up(c *cli.Context) {
	if err := options.ValidateCombinations(c); err != nil {
		logrus.Fatalf(err.Error())
	}
	composePath, err := createComposeFile(c)
	if err != nil {
		logrus.Fatalf("Failed to get compose file path due to:\n%v", err)
	}
	runContainers(composePath)
}

// createComposeFile returns the path of the Compose file to start the containers from.
// The Compose file ordering priority is defined as:
// 1. Use the --task-def-compose flag if it exists
// 2. Use the default docker-compose.local.yml file if no flags were provided and the file exists
// 3. Otherwise, we need to create the Compose file.
func createComposeFile(c *cli.Context) (string, error) {
	if name := c.String(flags.TaskDefinitionCompose); name != "" {
		return filepath.Abs(name)
	}
	if shouldUseDefaultComposeFile(c) {
		return filepath.Abs(localproject.LocalOutDefaultFileName)
	}

	project := localproject.New(c)
	if err := createLocal(project); err != nil {
		return "", err
	}
	return filepath.Abs(project.LocalOutFileName())
}

func runContainers(composePath string) {
	network.Setup(docker.NewClient())

	logrus.Infof("Using %s to start containers", filepath.Base(composePath))
	config := readComposeFile(composePath)
	secrets := readSecrets(config)
	envVars := decryptSecrets(secrets)
	upComposeFile(config, envVars)
}

func readComposeFile(composePath string) *composeV3.Config {
	config, err := converter.UnmarshalComposeFile(composePath)
	if err != nil {
		logrus.Fatalf("Failed to unmarshal Compose file %s due to \n%v", composePath, err)
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
	ssmClient, err := clients.NewSSMDecrypter()
	secretsManagerClient, err := clients.NewSecretsManagerDecrypter()
	if err != nil {
		logrus.Fatalf("Failed to create clients to decrypt secrets due to \n%v", err)
	}

	envVars = make(map[string]string)
	for _, containerSecret := range containerSecrets {
		service, err := containerSecret.ServiceName()
		if err != nil {
			logrus.Fatalf("Failed to retrieve the service of the secret due to \n%v", err)
		}

		decrypted := ""
		err = nil
		switch service {
		case secretsmanager.ServiceName:
			decrypted, err = containerSecret.Decrypt(secretsManagerClient)
		case ssm.ServiceName:
			decrypted, err = containerSecret.Decrypt(ssmClient)
		default:
			err = errors.New(fmt.Sprintf("can't decrypt secret from service %s", service))
		}
		if err != nil {
			logrus.Fatalf("Failed to decrypt secret due to \n%v", err)
		}
		envVars[containerSecret.Name()] = decrypted
	}
	return
}

// upComposeFile starts the containers in the Compose config with the environment variables defined in envVars.
func upComposeFile(config *composeV3.Config, envVars map[string]string) {
	var envs []string
	for env, val := range envVars {
		envs = append(envs, fmt.Sprintf("%s=%s", env, val))
	}

	cmd := exec.Command("docker-compose", "-f", config.Filename, "up", "-d")
	cmd.Env = envs

	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Fatalf("Failed to run docker-compose up due to \n%v: %s", err, string(out))
	}
	fmt.Printf("Compose out: %s\n", string(out))
}

func shouldUseDefaultComposeFile(c *cli.Context) bool {
	for _, flagName := range c.FlagNames() {
		if c.IsSet(flagName) {
			return false
		}
	}
	if _, err := os.Stat(localproject.LocalOutDefaultFileName); err != nil {
		return false
	}
	return true
}
