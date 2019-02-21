// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

// Package utils provides some utility functions. This is the kitchen sink.
package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	// ECSCLIResourcePrefix is prepended to the names of resources created through the ecs-cli
	ECSCLIResourcePrefix = "amazon-ecs-cli-setup-"
)

// InSlice checks if the given string exists in the given slice:
func InSlice(str string, list []string) bool {
	for _, s := range list {
		if s == str {
			return true
		}
	}
	return false
}

// GetHomeDir returns the file path of the user's home directory.
func GetHomeDir() (string, error) {
	// Can not use user.Current https://github.com/golang/go/issues/6376
	homeDir := os.Getenv("HOME") // *nix
	if homeDir == "" {           // Windows
		homeDir = os.Getenv("USERPROFILE")
	}

	if homeDir == "" {
		return "", fmt.Errorf("user home directory not found")
	}

	return homeDir, nil
}

// EntityAlreadyExists returns true if an error indicates that the AWS resource already exists
func EntityAlreadyExists(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		return awsErr.Code() == "EntityAlreadyExists"
	}
	return false
}

// ParseTags parses AWS Resource tags from the flag value
// users specify tags in this format: key1=value1,key2=value2,key3=value3
func ParseTags(flagValue string, tags []*ecs.Tag) ([]*ecs.Tag, error) {
	keyValPairs := strings.Split(flagValue, ",")
	for _, kv := range keyValPairs {
		pair := strings.Split(kv, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("Tag input not formatted correctly: %s", kv)
		}
		tags = append(tags, &ecs.Tag{
			Key:   aws.String(pair[0]),
			Value: aws.String(pair[1]),
		})
	}
	return tags, nil
}
