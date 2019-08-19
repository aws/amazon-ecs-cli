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

// Package localproject defines LocalProject interface and implements them on LocalProject

package localproject

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/converter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/urfave/cli"
)

const (
	taskDefArn = "arn:aws:ecs:us-west-2:123412341234:task-definition/myTaskDef:1"
)

func mockTaskDef() *ecs.TaskDefinition {
	taskDef := &ecs.TaskDefinition{}
	taskDef.SetTaskDefinitionArn(taskDefArn)
	return taskDef
}

func TestLocalOutFileName(t *testing.T) {
	testCases := map[string]struct {
		inputContext *cli.Context
		wantedName   string
	}{
		"without output flag": {
			inputContext: func() *cli.Context {
				flagSet := flag.NewFlagSet("ecs-cli", 0)
				return cli.NewContext(nil, flagSet, nil)
			}(),
			wantedName: LocalOutDefaultFileName,
		},
		"with output flag": {
			inputContext: func() *cli.Context {
				flagSet := flag.NewFlagSet("ecs-cli", 0)
				flagSet.String(flags.Output, "hello", "")
				return cli.NewContext(nil, flagSet, nil)
			}(),
			wantedName: "hello",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			p := New(tc.inputContext)

			// WHEN
			actualName := p.LocalOutFileName()

			// THEN
			require.Equal(t, tc.wantedName, actualName)
		})
	}
}

func TestOverrideFileName(t *testing.T) {
	testCases := map[string]struct {
		inputContext *cli.Context
		wantedName   string
	}{
		"without output flag": {
			inputContext: func() *cli.Context {
				flagSet := flag.NewFlagSet("ecs-cli", 0)
				return cli.NewContext(nil, flagSet, nil)
			}(),
			wantedName: "docker-compose.ecs-local.override.yml",
		},
		"with output flag": {
			inputContext: func() *cli.Context {
				flagSet := flag.NewFlagSet("ecs-cli", 0)
				flagSet.String(flags.Output, "hello", "")
				return cli.NewContext(nil, flagSet, nil)
			}(),
			wantedName: "hello.override.yml",
		},
		"with extension": {
			inputContext: func() *cli.Context {
				flagSet := flag.NewFlagSet("ecs-cli", 0)
				flagSet.String(flags.Output, "hello.yml", "")
				return cli.NewContext(nil, flagSet, nil)
			}(),
			wantedName: "hello.override.yml",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			p := New(tc.inputContext)

			// WHEN
			actualName := p.OverrideFileName()

			// THEN
			require.Equal(t, tc.wantedName, actualName)
		})
	}
}

func TestReadTaskDefinition_FromRemote(t *testing.T) {
	// GIVEN
	taskDefName := "myTaskDef"
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.TaskDefinitionRemote, taskDefName, "")
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	expectedTaskDef := mockTaskDef()
	expectedMetadata := &converter.LocalCreateMetadata{
		InputType: RemoteTaskDefType,
		Value:     taskDefName,
	}

	oldRead := readTaskDefFromRemote
	readTaskDefFromRemote = func(remote string, p *LocalProject) (*ecs.TaskDefinition, error) {
		return mockTaskDef(), nil
	}
	defer func() { readTaskDefFromRemote = oldRead }()

	// WHEN
	err := project.ReadTaskDefinition()

	// THEN
	assert.NoError(t, err, "Unexpected error reading task definition")
	assert.Equal(t, expectedTaskDef, project.TaskDefinition())
	assert.Equal(t, expectedMetadata, project.InputMetadata())
}

func TestReadTaskDefinition_FromLocal(t *testing.T) {
	// GIVEN
	taskDefFile := "some-file.json"
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.TaskDefinitionFile, taskDefFile, "")
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	expectedTaskDef := mockTaskDef()
	expectedLabelValue, _ := filepath.Abs(taskDefFile)
	expectedMetadata := &converter.LocalCreateMetadata{
		InputType: LocalTaskDefType,
		Value:     expectedLabelValue,
	}

	oldRead := readTaskDefFromLocal
	readTaskDefFromLocal = func(filename string) (*ecs.TaskDefinition, error) {
		return mockTaskDef(), nil
	}
	defer func() { readTaskDefFromLocal = oldRead }()

	// WHEN
	err := project.ReadTaskDefinition()

	// THEN
	assert.NoError(t, err, "Unexpected error reading task definition")
	assert.Equal(t, expectedTaskDef, project.TaskDefinition())
	assert.Equal(t, expectedMetadata, project.InputMetadata())
}

func TestReadTaskDefinition_FromLocalDefault(t *testing.T) {
	// GIVEN
	flagSet := flag.NewFlagSet("ecs-cli", 0) // No flags specified
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	oldDefaultInputExists := defaultInputExists
	defaultInputExists = func() bool {
		return true
	}
	defer func() { defaultInputExists = oldDefaultInputExists }()

	expectedTaskDef := mockTaskDef()
	expectedLabelValue, _ := filepath.Abs(LocalInFileName)
	expectedMetadata := &converter.LocalCreateMetadata{
		InputType: LocalTaskDefType,
		Value:     expectedLabelValue,
	}

	oldRead := readTaskDefFromLocal
	readTaskDefFromLocal = func(filename string) (*ecs.TaskDefinition, error) {
		return mockTaskDef(), nil
	}
	defer func() { readTaskDefFromLocal = oldRead }()

	// WHEN
	err := project.ReadTaskDefinition()

	// THEN
	assert.NoError(t, err, "Unexpected error reading task definition")
	assert.Equal(t, expectedTaskDef, project.TaskDefinition())
	assert.Equal(t, expectedMetadata, project.InputMetadata())
}

func TestReadTaskDefinition_ErrorIfTwoInputsSpecified(t *testing.T) {
	// GIVEN
	taskDefName := "myTaskDef"
	taskDefFile := "some-file.json"
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.TaskDefinitionRemote, taskDefName, "")
	flagSet.String(flags.TaskDefinitionFile, taskDefFile, "")
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	// WHEN
	err := project.ReadTaskDefinition()

	// THEN
	assert.Error(t, err, "Expected error reading task definition")
}

func TestWrite_New(t *testing.T) {
	// GIVEN
	flagSet := flag.NewFlagSet("ecs-cli", 0) // No flags specified
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	oldFileExists := fileExists
	fileExists = func(filename string) bool {
		return false
	}
	defer func() { fileExists = oldFileExists }()

	oldWrite := write
	write = func(filename string, content []byte) error {
		tmpfile, err := ioutil.TempFile("", filename)
		assert.NoError(t, err, "Unexpected error in creating temp compose file")
		defer os.Remove(tmpfile.Name())
		return nil
	}

	defer func() { write = oldWrite }()

	// WHEN
	err := project.Write()

	// THEN
	assert.NoError(t, err, "Unexpected error in writing test local compose file")
	assert.Equal(t, LocalOutDefaultFileName, project.LocalOutFileName())
}

func TestWrite_New_WithOutputFlag(t *testing.T) {
	// GIVEN
	expectedOutputFile := "foo.yml"
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.Output, expectedOutputFile, "")
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	oldFileExists := fileExists
	fileExists = func(filename string) bool {
		return false
	}
	defer func() { fileExists = oldFileExists }()

	oldWrite := write
	write = func(filename string, content []byte) error {
		tmpfile, err := ioutil.TempFile("", filename)
		assert.NoError(t, err, "Unexpected error in creating temp compose file")

		_, err = tmpfile.Write(content)
		assert.NoError(t, err, "Unexpected error in writing temp compose file")

		defer os.Remove(tmpfile.Name())
		return nil
	}

	defer func() { write = oldWrite }()

	// WHEN
	err := project.Write()

	// THEN
	assert.NoError(t, err, "Unexpected error in writing local compose file")
	assert.Equal(t, expectedOutputFile, project.LocalOutFileName())
}

func TestWrite_Existing(t *testing.T) {
	// GIVEN
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	oldFileExists := fileExists
	fileExists = func(filename string) bool {
		return true
	}
	defer func() { fileExists = oldFileExists }()

	called := false
	oldOverwrite := overwriteFile
	overwriteFile = func(filename string, content []byte) error {
		called = true
		return nil
	}

	defer func() { overwriteFile = oldOverwrite }()

	// WHEN
	err := project.Write()

	// THEN
	assert.NoError(t, err, "Unexpected error in writing local compose file")
	assert.Equal(t, LocalOutDefaultFileName, project.LocalOutFileName())
	assert.True(t, called)
}

func TestWrite_ExistingFile_WithForce(t *testing.T) {
	// GIVEN
	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.Bool(flags.ForceFlag, true, "")
	context := cli.NewContext(nil, flagSet, nil)
	project := New(context)

	oldFileExists := fileExists
	fileExists = func(filename string) bool {
		return true
	}
	defer func() { fileExists = oldFileExists }()

	called := false
	oldOverwrite := overwriteFile
	overwriteFile = func(filename string, content []byte) error {
		called = true
		return nil
	}

	defer func() { overwriteFile = oldOverwrite }()

	oldWrite := write
	write = func(filename string, content []byte) error {
		tmpfile, err := ioutil.TempFile("", filename)
		assert.NoError(t, err, "Unexpected error in creating temp compose file")

		_, err = tmpfile.Write(content)
		assert.NoError(t, err, "Unexpected error in writing temp compose file")

		defer os.Remove(tmpfile.Name())
		return nil
	}

	defer func() { write = oldWrite }()

	// WHEN
	err := project.Write()

	// THEN
	assert.NoError(t, err, "Unexpected error in writing local compose file")
	assert.Equal(t, LocalOutDefaultFileName, project.LocalOutFileName())
	assert.False(t, called)
}
