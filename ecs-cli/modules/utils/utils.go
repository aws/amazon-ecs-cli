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
	"strconv"
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
		pair := strings.SplitN(kv, "=", 2)
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

// GetTagsMap parses AWS Resource tags from the flag value
// users specify tags in this format: key1=value1,key2=value2,key3=value3
// Returns tags in the format used by the standalone resource tagging API
func GetTagsMap(flagValue string) (map[string]*string, error) {
	tags := make(map[string]*string)
	keyValPairs := strings.Split(flagValue, ",")
	for _, pair := range keyValPairs {
		split := strings.SplitN(pair, "=", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("Tag input not formatted correctly: %s", pair)
		}
		tags[split[0]] = aws.String(split[1])
	}
	return tags, nil
}

// GetPartition returns the partition for a given region
// This is meant to be used when constructing ARNs
func GetPartition(region string) string {
	if strings.HasPrefix(region, "cn") {
		return "aws-cn"
	} else if strings.HasPrefix(region, "us-gov") {
		return "aws-us-gov"
	} else {
		return "aws"
	}
}

// ParseLoadBalancers parses a StringSlice array into an array of load balancers struct
// Input: ["targetGroupArn="...",containerName="...",containerPort=80","targetGroupArn="...",containerName="...",containerPort=40"]
func ParseLoadBalancers(flagValues []string) ([]*ecs.LoadBalancer, error) {
	var list []*ecs.LoadBalancer

	for _, flagValue := range flagValues {
		m := make(map[string]string)
		elbv2 := false
		keyValPairs := strings.Split(flagValue, ",")

		for _, kv := range keyValPairs {
			pair := strings.SplitN(kv, "=", -1)
			if len(pair) > 2 {
				return nil, fmt.Errorf("Only include one = to indicate your value in your %s", pair[0])
			}

			m[pair[0]] = pair[1]
			if pair[0] == "targetGroupArn" {
				elbv2 = true
			}
		}
		containerPort, err := strconv.ParseInt(m["containerPort"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("Fail to parse container port %s for container %s", m["containerPort"], m["containerName"])
		}
		if elbv2 {
			list = append(list, &ecs.LoadBalancer{
				TargetGroupArn: aws.String(m["targetGroupArn"]),
				ContainerName:  aws.String(m["containerName"]),
				ContainerPort:  aws.Int64((containerPort)),
			})
		} else {
			list = append(list, &ecs.LoadBalancer{
				LoadBalancerName: aws.String(m["loadBalancerName"]),
				ContainerName:    aws.String(m["containerName"]),
				ContainerPort:    aws.Int64((containerPort)),
			})
		}
	}
	return list, nil
}
