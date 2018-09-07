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

package readers

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ECSRegCredsInput contains registry cred entries for creation and/or use in a task execution role
type ECSRegCredsInput struct {
	Version             string
	RegistryCredentials RegistryCreds `yaml:"registry_credentials"`
}

// RegistryCreds is a map of registry names to RegCredEntry structs
type RegistryCreds map[string]RegistryCredEntry

// RegistryCredEntry contains info needed to create an AWS Secrets Manager secret and match it to an ECS container(s)
type RegistryCredEntry struct {
	SecretManagerARN string   `yaml:"secret_manager_arn"`
	Username         string   `yaml:"username"`
	Password         string   `yaml:"password"`
	KmsKeyID         string   `yaml:"kms_key_id"`
	ContainerNames   []string `yaml:"container_names"`
}

// ReadCredsInput parses 'registry-creds up' input into an ECSRegCredsInput struct
func ReadCredsInput(filename string) (*ECSRegCredsInput, error) {

	rawCredsInput, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading file '%v'", filename)
	}
	credsInput := &ECSRegCredsInput{}

	if err = yaml.Unmarshal([]byte(rawCredsInput), &credsInput); err != nil {
		return nil, errors.Wrapf(err, "Error unmarshalling yaml data from credential input file: %s", filename)
	}

	expandedCredsInput := RegistryCreds{}
	for regName, credEntry := range credsInput.RegistryCredentials {
		expandedCredEntry := expandCredEntry(credEntry)
		expandedCredsInput[regName] = expandedCredEntry
	}

	credsInput.RegistryCredentials = expandedCredsInput

	return credsInput, nil
}

// expandCredEntry checks if individual fields are env vars and if so, retrieves & sets that value
func expandCredEntry(credEntry RegistryCredEntry) RegistryCredEntry {
	expandedSecretARN := getValueOrEnvVar(credEntry.SecretManagerARN)
	expandedUsername := getValueOrEnvVar(credEntry.Username)
	expandedPassword := getValueOrEnvVar(credEntry.Password)
	expandedKmsKeyID := getValueOrEnvVar(credEntry.KmsKeyID)
	//TODO: look for env vars in container names?

	expandedCredEntry := RegistryCredEntry{
		SecretManagerARN: expandedSecretARN,
		Username:         expandedUsername,
		Password:         expandedPassword,
		KmsKeyID:         expandedKmsKeyID,
		ContainerNames:   credEntry.ContainerNames,
	}
	return expandedCredEntry
}

// selectively runs ExpandEnv() to avoid indescriminant replacement of substrings with '$'
// e.g., password='c00l$tuff2018' -> return same; password='${MY_PASSWORD}' -> return env value.
func getValueOrEnvVar(s string) string {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		return os.ExpandEnv(s)
	}
	return s
}
