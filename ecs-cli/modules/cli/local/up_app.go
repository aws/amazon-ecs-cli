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
	"bytes"
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
// The Amazon ECS Local Endpoints container needs to be running already for any
// local ECS task to work (see
// https://github.com/awslabs/amazon-ecs-local-container-endpoints). If the
// container is not running, this command creates a new network for all local
// ECS tasks to join and communicate with the Amazon ECS Local Endpoints
// container.
func Up(c *cli.Context) {
	if err := options.ValidateFlagPairs(c); err != nil {
		logrus.Fatalf(err.Error())
	}
	basePath, err := composeProjectPath(c)
	if err != nil {
		logrus.Fatalf("Failed to create Compose files due to:\n%v", err)
	}
	overridePaths, err := composeOverridePaths(basePath, c.StringSlice(flags.ComposeOverride))
	if err != nil {
		logrus.Fatalf("Failed to get the path of override Compose files due to:\n%v", err)
	}

	runContainers(basePath, overridePaths)
}

// composeProjectPath creates Compose files if necessary and returns the path of the base Compose file.
//
// If the user set the TaskDefinitionCompose flag, then return that Compose
// file path.  If the user doesn't have any flags set, and doesn't have
// LocalInFileName but has a LocalOutDefaultFileName, then we use the
// LocalOutDefaultFileName file.  Otherwise, we create a new Compose file from
// the user's flags and return its path.
func composeProjectPath(c *cli.Context) (string, error) {
	if c.IsSet(flags.TaskDefinitionFile) {
		return createNewComposeProject(c)
	}
	if c.IsSet(flags.TaskDefinitionRemote) {
		return createNewComposeProject(c)
	}
	if c.IsSet(flags.TaskDefinitionCompose) {
		return filepath.Abs(c.String(flags.TaskDefinitionCompose))
	}

	// No input flags were provided, prioritize LocalInFileName over LocalOutDefaultFileName.
	if _, err := os.Stat(localproject.LocalInFileName); err == nil {
		return createNewComposeProject(c)
	} else if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "could not check if file %s exists", localproject.LocalInFileName)
	}
	if _, err := os.Stat(localproject.LocalOutDefaultFileName); err == nil {
		return filepath.Abs(localproject.LocalOutDefaultFileName)
	} else if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "could not check if file %s exists", localproject.LocalOutDefaultFileName)
	}
	return "", errors.New(fmt.Sprintf("need to provide one of %s or %s", localproject.LocalInFileName, localproject.LocalOutDefaultFileName))
}

func createNewComposeProject(c *cli.Context) (string, error) {
	project := localproject.New(c)
	if err := createLocal(project); err != nil {
		return "", err
	}
	return project.LocalOutFileFullPath()
}

func composeOverridePaths(basePath string, additionalRelPaths []string) ([]string, error) {
	defaultPath := basePath[:len(basePath)-len(filepath.Ext(basePath))] + ".override.yml"
	paths := []string{defaultPath}
	if len(additionalRelPaths) > 0 {
		for _, relPath := range additionalRelPaths {
			p, err := filepath.Abs(relPath)
			if err != nil {
				return nil, errors.Wrapf(err, "cannot get absolute path of %s", relPath)
			}
			paths = append(paths, p)
		}
	}

	// Prune paths that don't exist
	var overridePaths []string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			overridePaths = append(overridePaths, p)
		} else {
			logrus.Warnf("Skipping Compose file %s due to:\n%v", filepath.Base(p), err)
		}
	}
	return overridePaths, nil
}

func runContainers(basePath string, overridePaths []string) {
	network.Setup(docker.NewClient())

	config := readComposeFile(basePath)
	secrets := readSecrets(config)
	envVars := decryptSecrets(secrets)
	upCompose(envVars, basePath, overridePaths)
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

// upCompose starts the containers in the Compose files with the environment variables defined in envVars.
func upCompose(envVars map[string]string, basePath string, overridePaths []string) {
	// Gather environment variables
	var envs []string
	for env, val := range envVars {
		envs = append(envs, fmt.Sprintf("%s=%s", env, val))
	}
	// Need to add $PATH because of --build, see https://stackoverflow.com/a/55371721/1201381
	envs = append(envs, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))

	// Disable orphaned containers checking
	envs = append(envs, fmt.Sprint("COMPOSE_IGNORE_ORPHANS=true"))

	// Gather command arguments
	var b bytes.Buffer
	b.WriteString(filepath.Base(basePath))
	args := []string{"-f", basePath}
	for _, p := range overridePaths {
		b.WriteString(fmt.Sprintf(", %s", filepath.Base(p)))
		args = append(args, "-f", p)
	}
	args = append(args, "up", "--build", "-d")

	// Run the command with the environment variables and arguments
	logrus.Infof("Using %s files to start containers", b.String())
	cmd := exec.Command("docker-compose", args...)
	cmd.Env = envs

	out, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Fatalf("Failed to run docker-compose up due to \n%v: %s", err, string(out))
	}
	fmt.Printf("Compose out: %s\n", string(out))
}
