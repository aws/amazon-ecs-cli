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
	"strings"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/iam"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/kms"
	secretsClient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/secretsmanager"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcredio"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	maxContainersPerTaskDef = 10
)

// Up creates or updates registry credential secrets and an ECS task execution role needed to use them in a task def
func Up(c *cli.Context) {
	args := c.Args()

	if len(args) != 1 {
		log.Fatal("Exactly 1 credential file is required. Found: ", len(args))
	}

	// create clients
	commandConfig := getNewCommandConfig(c)

	smClient := secretsClient.NewSecretsManagerClient(commandConfig)
	kmsClient := kms.NewKMSClient(commandConfig)
	iamClient := iam.NewIAMClient(commandConfig)

	// validate provided values before creating any resources
	credsInput, err := regcredio.ReadCredsInput(args[0])
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	validatedRegCreds, err := validateCredsInput(*credsInput, kmsClient)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	roleName := c.String(flags.RoleNameFlag)
	skipRole := c.Bool(flags.NoRoleFlag)

	err = validateRoleDetails(roleName, skipRole)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	outputDir := c.String(flags.OutputDirFlag)
	skipOutput := c.Bool(flags.NoOutputFileFlag)

	err = validateOutputOptions(outputDir, skipOutput)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	// find or create secrets, role
	updateAllowed := c.Bool(flags.UpdateExistingSecretsFlag)

	credentialOutput, err := getOrCreateRegistryCredentials(validatedRegCreds, smClient, updateAllowed)
	if err != nil {
		log.Fatal("Error executing 'up': ", err)
	}

	var policyCreateTime *time.Time
	if !skipRole {
		region := commandConfig.Session.Config.Region

		roleParams := executionRoleParams{
			CredEntries: credentialOutput,
			RoleName:    roleName,
			Region:      *region,
		}

		policyCreateTime, err = createTaskExecutionRole(roleParams, iamClient, kmsClient)
		if err != nil {
			log.Fatal("Error executing 'up': ", err)
		}
	} else {
		log.Info("Skipping role creation.")
	}

	// produce output file
	if !skipOutput {
		regcredio.GenerateCredsOutput(credentialOutput, roleName, outputDir, policyCreateTime)
	} else {
		log.Info("Skipping generation of registry credentials output file.")
	}

	log.Info("\nIf your input file contains sensitive information, make sure that you delete it after use.")
}

func getOrCreateRegistryCredentials(entryMap regcredio.RegistryCreds, smClient secretsClient.SMClient, updateAllowed bool) (map[string]regcredio.CredsOutputEntry, error) {
	registryResults := make(map[string]regcredio.CredsOutputEntry)

	for registryName, credentialEntry := range entryMap {
		log.Infof("Processing credentials for registry %s...", registryName)

		arn := credentialEntry.SecretManagerARN
		var keyForSecret *string
		if arn == "" {
			newSecretARN, key, err := findOrCreateRegistrySecret(registryName, credentialEntry, smClient)
			if err != nil {
				return nil, err
			}
			arn = newSecretARN
			keyForSecret = &key
		} else if credentialEntry.HasCredPair() {
			if err := updateOrWarnForExistingSecret(credentialEntry, updateAllowed, smClient); err != nil {
				return nil, err
			}
		} else {
			log.Infof("Using existing secret %s.", arn)
		}

		if keyForSecret == nil {
			keyForSecret = &credentialEntry.KmsKeyID
		}
		registryResults[registryName] = regcredio.BuildOutputEntry(arn, *keyForSecret, credentialEntry.ContainerNames)
	}

	return registryResults, nil
}

// returns the ARN of a new or existing registry secret (and, if applicable, the KMS key associated with that secret)
func findOrCreateRegistrySecret(registryName string, credEntry regcredio.RegistryCredEntry, smClient secretsClient.SMClient) (string, string, error) {

	secretName := generateECSResourceName(registryName)

	existingSecret, _ := smClient.DescribeSecret(*secretName)
	if existingSecret != nil {
		log.Infof("Existing credential secret found, using %s", *existingSecret.ARN)

		if existingSecret.KmsKeyId != nil {
			return *existingSecret.ARN, *existingSecret.KmsKeyId, nil
		}

		return *existingSecret.ARN, "", nil
	}

	secretString := generateSecretString(credEntry.Username, credEntry.Password)

	createSecretRequest := secretsmanager.CreateSecretInput{
		Name:         secretName,
		SecretString: secretString,
		Description:  generateSecretDescription(registryName),
	}

	kmsKey := credEntry.KmsKeyID
	if kmsKey != "" {
		createSecretRequest.SetKmsKeyId(kmsKey)
	}

	output, err := smClient.CreateSecret(createSecretRequest)
	if err != nil {
		return "", "", err
	}
	log.Infof("New credential secret created: %s", *output.ARN)

	return *output.ARN, kmsKey, nil
}

func updateOrWarnForExistingSecret(credEntry regcredio.RegistryCredEntry, updateAllowed bool, smClient secretsClient.SMClient) error {
	secretARN := credEntry.SecretManagerARN

	if updateAllowed {
		updatedSecretString := generateSecretString(credEntry.Username, credEntry.Password)
		putSecretValueRequest := secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretARN),
			SecretString: updatedSecretString,
		}

		_, err := smClient.PutSecretValue(putSecretValueRequest)
		if err != nil {
			return err
		}

		log.Infof("Updated existing secret %s with new value", secretARN)

	} else {
		log.Warnf("'username' and 'password' found but ignored for existing secret %s. To update existing secrets with new values, use '--update-existing-secrets' flag.", secretARN)
	}
	return nil
}

func validateCredsInput(input regcredio.ECSRegCredsInput, kmsClient kms.Client) (map[string]regcredio.RegistryCredEntry, error) {
	// TODO: validate version?

	inputRegCreds := input.RegistryCredentials

	if len(inputRegCreds) == 0 {
		return nil, errors.New("provided credentials must contain at least one registry")
	}
	if len(inputRegCreds) > maxContainersPerTaskDef {
		return nil, errors.New("no more than" + string(maxContainersPerTaskDef) + "registry credential entries can be created at one time")
	}

	namedContainers := make(map[string]bool)
	outputRegCreds := make(map[string]regcredio.RegistryCredEntry)

	for registryName, credentialEntry := range inputRegCreds {
		if !credentialEntry.HasRequiredFields() {
			return nil, fmt.Errorf("missing required field(s) for registry %s; registry credentials should contain an existing secret ARN or username + password", registryName)
		}
		if len(credentialEntry.ContainerNames) > 0 {
			for _, container := range credentialEntry.ContainerNames {
				if namedContainers[container] {
					return nil, fmt.Errorf("container '%s' appears in more than one registry; container names must be unique across given registry credentials", container)
				}
				namedContainers[container] = true
			}
		}
		if len(credentialEntry.ContainerNames) == 0 {
			log.Warnf("No container names given for registry '%s'; output cannot be incorporated into a task definition when running 'compose' command", registryName)
		}
		if credentialEntry.SecretManagerARN != "" && !isARN(credentialEntry.SecretManagerARN) {
			return nil, fmt.Errorf("invalid secrets_manager_arn for registry %s", registryName)
		}
		// if key specified as ID or alias, validate & get ARN
		if credentialEntry.KmsKeyID != "" {
			keyARN, err := kmsClient.GetValidKeyARN(credentialEntry.KmsKeyID)
			if err != nil {
				return nil, err
			}
			credentialEntry.KmsKeyID = keyARN
		}
		// if both present, validate secret ARN & key are in same region
		if credentialEntry.SecretManagerARN != "" && credentialEntry.KmsKeyID != "" {
			secretRegion := strings.Split(credentialEntry.SecretManagerARN, ":")[3]
			keyRegion := strings.Split(credentialEntry.KmsKeyID, ":")[3]

			if secretRegion != keyRegion {
				return nil, fmt.Errorf("region of 'secrets_manager_arn'(%s) and 'kms_key_id'(%s) for registry %s do not match; secret and encryption key must be in same region", secretRegion, keyRegion, registryName)
			}
		}
		outputRegCreds[registryName] = credentialEntry
	}
	return outputRegCreds, nil
}

func getValidKeyARN(keyID string, kmsClient kms.Client) (string, error) {
	arn := ""

	if isARN(keyID) {
		arn = keyID
	} else {
		keyResult, err := kmsClient.DescribeKey(keyID)
		if err != nil {
			return "", err
		}
		keyMetadata := *keyResult.KeyMetadata
		arn = *keyMetadata.Arn
	}
	return arn, nil
}

func validateRoleDetails(roleName string, noRole bool) error {
	if noRole && roleName != "" {
		return fmt.Errorf("both role name ('%s') and '--no-role' specified; please specify either a role name or the '--no-role' flag", roleName)
	}
	if !noRole && roleName == "" {
		return errors.New("no value specified for '--role-name'; please specify either a role name or the '--no-role' flag")
	}
	return nil
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

func validateOutputOptions(outputDir string, skipOutput bool) error {
	if outputDir != "" && skipOutput {
		return fmt.Errorf("Only one of '--"+flags.OutputDirFlag+"' (value '%s') and '--"+flags.NoOutputFileFlag+"' can be specified but both are present", outputDir)
	}
	return nil
}
