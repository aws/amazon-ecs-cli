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

package utils

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

func TestInSlice(t *testing.T) {
	list := []string{"apple", "banana", "coconut"}
	if !InSlice("apple", list) {
		t.Error("Failed to find string in list.")
	}
	if InSlice("orange", list) {
		t.Error("Incorrectly found non-member string in list.")
	}
}

func TestGetHomeDir(t *testing.T) {
	tempDirName := tempDir(t)
	defer os.Remove(tempDirName)
	os.Setenv("HOME", tempDirName)
	defer os.Unsetenv("HOME")

	_, err := GetHomeDir()
	assert.NoError(t, err, "Unexpected error getting home dir")
}

func TestGetHomeDirInWindows(t *testing.T) {
	tempDirName := tempDir(t)
	defer os.Remove(tempDirName)
	os.Setenv("USERPROFILE", tempDirName)
	defer os.Unsetenv("USERPROFILE")

	_, err := GetHomeDir()
	assert.NoError(t, err, "Unexpected error getting home dir")
}

func TestParseTags(t *testing.T) {
	actualTags := make([]*ecs.Tag, 0)
	expectedTags := []*ecs.Tag{
		&ecs.Tag{
			Key:   aws.String("Pink"),
			Value: aws.String("Floyd"),
		},
		&ecs.Tag{
			Key:   aws.String("Tame"),
			Value: aws.String("Impala"),
		},
	}

	var err error
	actualTags, err = ParseTags("Pink=Floyd,Tame=Impala", actualTags)
	assert.NoError(t, err, "Unexpected error calling ParseTags")
	assert.ElementsMatch(t, actualTags, expectedTags, "Expected tags to match")

}

func TestParseTagsEmptyValue(t *testing.T) {
	actualTags := make([]*ecs.Tag, 0)
	expectedTags := []*ecs.Tag{
		&ecs.Tag{
			Key:   aws.String("thecheese"),
			Value: aws.String(""),
		},
		&ecs.Tag{
			Key:   aws.String("standsalone"),
			Value: aws.String(""),
		},
	}

	var err error
	actualTags, err = ParseTags("thecheese=,standsalone=", actualTags)
	assert.NoError(t, err, "Unexpected error calling ParseTags")
	assert.ElementsMatch(t, actualTags, expectedTags, "Expected tags to match")

}

func TestParseTagInvalidFormat(t *testing.T) {
	actualTags := make([]*ecs.Tag, 0)

	var err error
	_, err = ParseTags("incorrectly=formatted,tags", actualTags)
	assert.Error(t, err, "Expected error calling ParseTags")
}

func TestGetPartition(t *testing.T) {
	var partitionTests = []struct {
		region    string
		partition string
	}{
		{"us-east-1", "aws"},
		{"us-east-2", "aws"},
		{"us-west-1", "aws"},
		{"us-west-2", "aws"},
		{"ap-south-1", "aws"},
		{"ap-northeast-1", "aws"},
		{"ap-northeast-2", "aws"},
		{"ap-southeast-1", "aws"},
		{"ap-southeast-2", "aws"},
		{"ca-central-1", "aws"},
		{"sa-east-1", "aws"},
		{"eu-central-1", "aws"},
		{"eu-north-1", "aws"},
		{"eu-west-1", "aws"},
		{"eu-west-2", "aws"},
		{"eu-west-3", "aws"},
		{"cn-north-1", "aws-cn"},
		{"cn-northwest-1", "aws-cn"},
		{"us-gov-east-1", "aws-us-gov"},
		{"us-gov-west-1", "aws-us-gov"},
	}

	for _, test := range partitionTests {
		t.Run(test.region, func(t *testing.T) {
			assert.Equal(t, test.partition, GetPartition(test.region))
		})
	}
}

func tempDir(t *testing.T) string {
	// Create a temprorary directory for the dummy ecs config
	tempDirName, err := ioutil.TempDir("", "test")
	assert.NoError(t, err, "Unexpected error while creating the dummy ecs config directory")
	return tempDirName
}
