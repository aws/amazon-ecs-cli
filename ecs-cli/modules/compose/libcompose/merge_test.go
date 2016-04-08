// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package libcompose

import (
	"io/ioutil"
	"os"
	"sort"
	"testing"
)

func TestGetEnvVarsFromConfigWhenNoEnvVars(t *testing.T) {
	observedEnvVars, err := GetEnvVarsFromConfig(Context{}, &ServiceConfig{})
	if err != nil {
		t.Fatal("Unexpected error while GetEnvVarsFromConfig when no env vars exist")
	}
	if len(observedEnvVars) != 0 {
		t.Errorf("Expected GetEnvVarsFromConfig to return empty env vars but got [%v]", observedEnvVars)
	}
}

func TestGetEnvVarsFromConfigWhenNoEnvFiles(t *testing.T) {
	envKey := "rails_env"
	envValue := "development"
	env := envKey + "=" + envValue
	serviceConfig := &ServiceConfig{
		Environment: NewMaporEqualSlice([]string{env}),
	}
	observedEnvVars, err := GetEnvVarsFromConfig(Context{}, serviceConfig)
	if err != nil {
		t.Fatal("Unexpected error while GetEnvVarsFromConfig when no env files exist")
	}
	if len(observedEnvVars) != 1 || observedEnvVars[0] != env {
		t.Errorf("Expected GetEnvVarsFromConfig to return env vars [%s] but got [%v]", env, observedEnvVars)
	}
}

func TestGetEnvVarsFromConfigWhenEnvFilesButNoLookup(t *testing.T) {
	envKey := "rails_env"
	envValue := "development"
	env := envKey + "=" + envValue
	envFileName := "envFile"
	serviceConfig := &ServiceConfig{
		Environment: NewMaporEqualSlice([]string{env}),
		EnvFile:     NewStringorslice(envFileName),
	}
	_, err := GetEnvVarsFromConfig(Context{}, serviceConfig)
	if err == nil {
		t.Fatal("Expected error while GetEnvVarsFromConfig when no config was supplied")
	}
}

func TestGetEnvVarsFromConfigWhenEnvFilesButNoEnvVars(t *testing.T) {
	envKey := "rails_env"
	envValue := "development"
	env := envKey + "=" + envValue
	envContents := []byte(env)

	envFile := writeToTmpFile(t, envContents)
	defer os.Remove(envFile.Name()) // clean up

	serviceConfig := &ServiceConfig{
		EnvFile: NewStringorslice(envFile.Name()),
	}
	context := Context{
		ConfigLookup: &FileConfigLookup{},
	}
	observedEnvVars, err := GetEnvVarsFromConfig(context, serviceConfig)

	if err != nil {
		t.Fatal("Unexpected error while GetEnvVarsFromConfig when env file is supplied")
	}
	if len(observedEnvVars) != 1 || observedEnvVars[0] != env {
		t.Errorf("Expected GetEnvVarsFromConfig to return env vars [%s] but got [%v]", env, observedEnvVars)
	}
}

func TestGetEnvVarsFromConfigWhenEnvVarsAndFiles(t *testing.T) {
	envVar := "envKey1=envValue1"

	envFile1var1 := "envKey2=envValue2"
	envFile1var2 := "envKey1=envValue3" // repeat the same key as env var. Should not be selected.
	envFile1var3 := "envKey4=envKey4"   // 2nd file has the same variable. Should not be selected.

	envFile2var1 := "envKey3=envValue3"
	envFile2var2 := "envKey4=envValue4first" // repeat the same key as previous file

	envFile1 := writeToTmpFile(t, []byte(envFile1var1+"\n"+envFile1var2+"\n"+envFile1var3))
	defer os.Remove(envFile1.Name())

	envFile2 := writeToTmpFile(t, []byte(envFile2var1+"\n"+envFile2var2))
	defer os.Remove(envFile2.Name())

	serviceConfig := &ServiceConfig{
		Environment: NewMaporEqualSlice([]string{envVar}),
		EnvFile:     NewStringorslice(envFile1.Name(), envFile2.Name()),
	}
	context := Context{
		ConfigLookup: &FileConfigLookup{},
	}
	observedEnvVars, err := GetEnvVarsFromConfig(context, serviceConfig)

	if err != nil {
		t.Fatal("Unexpected error while GetEnvVarsFromConfig when env file is supplied")
	}
	if len(observedEnvVars) != 4 {
		t.Errorf("Expected GetEnvVarsFromConfig to return env vars count=[%d] but got=[%d]", 4, len(observedEnvVars))
	}

	sort.Strings(observedEnvVars)
	verifyEnvVar(t, "envKey1=envValue1", observedEnvVars[0])
	verifyEnvVar(t, "envKey2=envValue2", observedEnvVars[1])
	verifyEnvVar(t, "envKey3=envValue3", observedEnvVars[2])
	verifyEnvVar(t, "envKey4=envValue4first", observedEnvVars[3])
}

func verifyEnvVar(t *testing.T, expected, observed string) {
	if expected != observed {
		t.Errorf("Expected GetEnvVarsFromConfig to return env var [%s] but got [%s]", expected, observed)
	}
}

func writeToTmpFile(t *testing.T, contents []byte) *os.File {
	envFile, err := ioutil.TempFile("", "envfile")
	if err != nil {
		t.Fatal("Error creating tmp file:", err)
	}
	if _, err := envFile.Write(contents); err != nil {
		t.Fatal("Error writing to tmp file:", err)
	}
	return envFile
}
