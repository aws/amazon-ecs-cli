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
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	// ECSCredFileTimeFmt is the timestamp format to use on 'registry-creds up' outputs
	ECSCredFileTimeFmt = "20060102T150405Z"
	// ECSCredFileBaseName is the base name of any private registry cred file produced or read by the ecs-cli
	ECSCredFileBaseName = "ecs-registry-creds"
)

// GenerateCredsOutput marshals credential output JSON into YAML and outputs it to a file
func GenerateCredsOutput(creds map[string]CredsOutputEntry, roleName, outputDir string, policyCreatTime *time.Time) error {
	outputResources := CredResources{
		ContainerCredentials: creds,
		TaskExecutionRole:    roleName,
	}
	regOutput := ECSRegistryCredsOutput{
		Version:             "1",
		CredentialResources: outputResources,
	}
	credBytes, err := yaml.Marshal(regOutput)
	if err != nil {
		return err
	}

	outputFileDir := outputDir
	if outputFileDir == "" {
		wdir, err := os.Getwd()
		if err != nil {
			return err
		}
		outputFileDir = wdir
	}

	timeStamp := time.Now().UTC()
	if policyCreatTime != nil {
		timeStamp = *policyCreatTime
	}
	timestampedSuffix := fmt.Sprintf("_%s.yml", timeStamp.Format(ECSCredFileTimeFmt))

	file, err := os.Create(outputFileDir + string(os.PathSeparator) + ECSCredFileBaseName + timestampedSuffix)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Info("Writing registry credential output to new file " + file.Name())
	_, err = file.Write(credBytes)
	if err != nil {
		return err
	}

	return nil
}

// BuildOutputEntry returns a CredsOutputEntry with the provided parameters
func BuildOutputEntry(arn string, key string, containers []string) CredsOutputEntry {
	return CredsOutputEntry{
		CredentialARN:  arn,
		KMSKeyID:       key,
		ContainerNames: containers,
	}
}
