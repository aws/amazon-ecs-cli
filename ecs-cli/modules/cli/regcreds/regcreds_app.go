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

package regcreds

import (
	"fmt"

	secretsClient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/secretsmanager"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcreds"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// CredsOutputEntry contains the credential ARN and associated container names
// TODO: use & move to output_reader once implemented?
type CredsOutputEntry struct {
	CredentialARN  string
	ContainerNames []string
}

// Up creates or updates registry credential secrets and an ECS task execution role needed to use them in a task def
func Up(c *cli.Context) {
	args := c.Args()

	if len(args) != 1 {
		log.Fatal("Exactly 1 credential file is required. Found: ", len(args))
	}

	credsInput, err := readers.ReadCredsInput(args[0])
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	err = validateCredsInput(*credsInput)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	commandConfig := getNewCommandConfig(c)

	_, err = getOrCreateRegistryCredentials(credsInput.RegistryCredentials, commandConfig, c)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	//TODO: create role, produce output
}

func getOrCreateRegistryCredentials(entryMap readers.RegistryCreds, cmdConfig *config.CommandConfig, c *cli.Context) (*map[string]CredsOutputEntry, error) {
	registryResults := make(map[string]CredsOutputEntry)
	updateAllowed := c.Bool(flags.UpdateExistingSecretsFlag)

	smClient := secretsClient.NewSecretsManagerClient(cmdConfig)

	for registryName, credentialEntry := range entryMap {
		hasCredPair := hasCredPair(credentialEntry)
		hasSecretARN := false
		if credentialEntry.SecretManagerARN != "" {
			hasSecretARN = true
		}

		log.Infof("Processing credentials for registry %s...", registryName)

		if hasCredPair && hasSecretARN {
			arn, err := updateOrWarnForExistingSecret(credentialEntry, updateAllowed, smClient)
			if err != nil {
				return nil, err
			}
			registryResults[registryName] = buildOutputEntry(arn, credentialEntry.ContainerNames)

		} else if hasSecretARN {
			registryResults[registryName] = buildOutputEntry(credentialEntry.SecretManagerARN, credentialEntry.ContainerNames)
			log.Infof("Using existing secret %s.", registryName)

		} else {
			arn, err := createNewRegistrySecret(registryName, credentialEntry, smClient)
			if err != nil {
				return nil, err
			}
			registryResults[registryName] = buildOutputEntry(arn, credentialEntry.ContainerNames)
		}
	}

	log.Infof("\n up results: %v", registryResults)

	return &registryResults, nil
}

func createNewRegistrySecret(registryName string, credEntry readers.RegistryCredEntry, smClient secretsClient.SMClient) (string, error) {

	secretName := generateSecretName(registryName)

	existingSecret, _ := smClient.DescribeSecret(secretName)
	if existingSecret != nil {
		log.Infof("Existing credential secret found, using %s", *existingSecret.ARN)

		return *existingSecret.ARN, nil
	}

	secretString := generateSecretString(credEntry.Username, credEntry.Password)

	createSecretRequest := secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(secretString),
		Description:  aws.String(fmt.Sprintf("Created with the ECS CLI for use with registry %s", registryName)),
	}
	if credEntry.KmsKeyID != "" {
		createSecretRequest.SetKmsKeyId(credEntry.KmsKeyID)
	}

	output, err := smClient.CreateSecret(createSecretRequest)
	if err != nil {
		return "", err
	}
	log.Infof("New credential secret created: %s", *output.ARN)

	return *output.ARN, nil
}

func updateOrWarnForExistingSecret(credEntry readers.RegistryCredEntry, updateAllowed bool, smClient secretsClient.SMClient) (string, error) {
	secretArn := credEntry.SecretManagerARN

	if updateAllowed {
		updatedSecretString := generateSecretString(credEntry.Username, credEntry.Password)
		putSecretValueRequest := secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretArn),
			SecretString: aws.String(updatedSecretString),
		}

		_, err := smClient.PutSecretValue(putSecretValueRequest)
		if err != nil {
			return "", err
		}

		log.Infof("Updated existing secret %s with new value", secretArn)

	} else {
		log.Warnf("'username' and 'password' found but ignored for existing secret %s. To update existing secrets with new values, use '--update-existing-secrets' flag.", secretArn)
	}

	return secretArn, nil
}

func validateCredsInput(input readers.ECSRegCredsInput) error {
	// TODO: validate version?

	inputRegCreds := input.RegistryCredentials

	if len(inputRegCreds) == 0 {
		return errors.New("provided credentials must contain at least one registry")
	}

	for registryName, credentialEntry := range inputRegCreds {
		if !hasRequiredFields(credentialEntry) {
			return fmt.Errorf("missing required field(s) for registry %s; registry credentials should contain existing secret ARN or username + password", registryName)
		}
	}
	return nil
}

func hasRequiredFields(entry readers.RegistryCredEntry) bool {
	if (entry.SecretManagerARN != "") || hasCredPair(entry) {
		return true
	}
	return false
}

func hasCredPair(entry readers.RegistryCredEntry) bool {
	if entry.Username != "" && entry.Password != "" {
		return true
	}
	return false
}

func getNewCommandConfig(c *cli.Context) *config.CommandConfig {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}
	commandConfig, err := config.NewCommandConfig(c, rdwr)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	return commandConfig
}

func generateSecretName(regName string) string {
	return "amazon-ecs-cli-setup-" + regName
}

func generateSecretString(username, password string) string {
	return `{"username":"` + username + `"},{"password":"` + password + `"}`
}

func buildOutputEntry(arn string, containers []string) CredsOutputEntry {
	return CredsOutputEntry{
		CredentialARN:  arn,
		ContainerNames: containers,
	}
}
