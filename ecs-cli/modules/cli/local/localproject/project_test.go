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

// Package localproject defines LocalProject interface and implements them on localProject

package localproject

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestWrite(t *testing.T) {
	// GIVEN
	flagSet := flag.NewFlagSet("ecs-cli", 0) // No flags specified
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	oldOpenFile := openFile
	openFile = func(filename string) (*os.File, error) {
		tmpfile, err := ioutil.TempFile("", filename)
		assert.NoError(t, err, "Unexpected error in creating temp compose file")
		defer os.Remove(tmpfile.Name())

		return tmpfile, nil
	}
	defer func() { openFile = oldOpenFile }()

	// WHEN
	err := project.Write()

	// THEN
	assert.NoError(t, err, "Unexpected error in writing local compose file")
	assert.Equal(t, LocalOutDefaultFileName, project.LocalOutFileName())
}

func TestWrite_WithOutputFlag(t *testing.T) {
	// GIVEN
	expectedOutputFile := "foo.yml"
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.LocalOutputFlag, expectedOutputFile, "")
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	oldOpenFile := openFile
	openFile = func(filename string) (*os.File, error) {
		tmpfile, err := ioutil.TempFile("", filename)
		assert.NoError(t, err, "Unexpected error in creating temp compose file")
		defer os.Remove(tmpfile.Name())

		return tmpfile, nil
	}

	defer func() { openFile = oldOpenFile }()

	// WHEN
	err := project.Write()

	// THEN
	assert.NoError(t, err, "Unexpected error in writing local compose file")
	assert.Equal(t, expectedOutputFile, project.LocalOutFileName())
}
