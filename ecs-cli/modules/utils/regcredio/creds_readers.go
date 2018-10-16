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

package regcredio

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

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

// ReadCredsOutput parses an ECS creds output file into an RegistryCredsOutput struct
// TODO: use this to parse reg creds used with "compose" cmd
func ReadCredsOutput(filename string) (*ECSRegistryCredsOutput, error) {
	rawCredsOutput, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading file '%v'", filename)
	}

	credsOutput := &ECSRegistryCredsOutput{}
	if err = yaml.Unmarshal([]byte(rawCredsOutput), &credsOutput); err != nil {
		return nil, errors.Wrapf(err, "Error unmarshalling yaml data from registry credential ouput file: %s", filename)
	}

	return credsOutput, nil
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
