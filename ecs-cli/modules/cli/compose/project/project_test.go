// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package project

import (
	"flag"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/regcredio"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const testProjectName = "test-project"

func TestParseCompose_V2(t *testing.T) {
	// Setup docker-compose file
	composeFileString := `version: '2'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
  mysql:
    image: mysql`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeFileString))
	assert.NoError(t, err, "Unexpected error in writing to test file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// Set up project
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// verify container configs are populated
	err = project.parseCompose()
	assert.NoError(t, err, "Unexpected error parsing file")
	assert.NotEmpty(t, project.ContainerConfigs(), "Expected container configs to not be empty")

	// verify project name is set
	assert.Equal(t, testProjectName, project.ecsContext.ProjectName, "Expected ProjectName to be overridden.")

	// verify top-level volumes are empty
	assert.Empty(t, project.VolumeConfigs().VolumeWithHost, "Expected volume configs to be empty")
	assert.Empty(t, project.VolumeConfigs().VolumeEmptyHost, "Expected volume configs to be empty")
}

func TestParseCompose_V2_WithVolumeConfigs(t *testing.T) {
	// Setup docker-compose file
	composeFileString := `version: '2'
services:
  wordpress:
    image: wordpress
  mysql:
    image: mysql
    volumes:
      - banana:/tmp/cache
      - :/tmp/cache
      - ./cache:/tmp/cache:ro
volumes:
  banana:`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeFileString))
	assert.NoError(t, err, "Unexpected error in writing to test file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// Set up project
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	err = project.parseCompose()
	expectedNamedVolumes := []string{"banana", "volume-1"}
	expectedHosts := map[string]string{"./cache": "volume-2"}

	// verify VolumeConfigs are populated
	assert.Equal(t, expectedHosts, project.VolumeConfigs().VolumeWithHost, "Expected volume configs to match")
	assert.Equal(t, expectedNamedVolumes, project.VolumeConfigs().VolumeEmptyHost, "Expected volume configs to match")
}

func TestParseECSParams(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: arn:aws:iam::123456789012:role/my_role
  services:
    mysql:
      essential: false

run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: [subnet-feedface, subnet-deadbeef]
      security_groups:
        - sg-bafff1ed
        - sg-c0ffeefe`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	project := setupTestProjectWithEcsParams(t, ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	if err := project.parseECSParams(); err != nil {
		t.Fatalf("Unexpected error parsing the ecs-params data [%s]: %v", ecsParamsString, err)
	}

	ecsParams := project.ecsContext.ECSParams
	assert.NotNil(t, ecsParams, "Expected ecsParams to be set on project")
	assert.Equal(t, "1", ecsParams.Version, "Expected Version to match")

	td := ecsParams.TaskDefinition

	assert.Equal(t, "host", td.NetworkMode, "Expected NetworkMode to match")
	assert.Equal(t, "arn:aws:iam::123456789012:role/my_role", td.TaskRoleArn, "Expected TaskRoleArn to match")

	networkConfigs := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration
	assert.Equal(t, []string{"subnet-feedface", "subnet-deadbeef"}, networkConfigs.Subnets, "Expected Subnets to match")
	assert.Equal(t, []string{"sg-bafff1ed", "sg-c0ffeefe"}, networkConfigs.SecurityGroups, "Expected SecurityGroups to match")

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")
}
func TestParseECSParamsWithEnvironment(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  task_size:
    mem_limit: ${MEM_LIMIT}
    cpu_limit: $CPU_LIMIT`

	os.Setenv("MEM_LIMIT", "1000")
	os.Setenv("CPU_LIMIT", "200")

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	project := setupTestProjectWithEcsParams(t, ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = project.parseECSParams()
	if assert.NoError(t, err) {
		ecsParams := project.ecsContext.ECSParams
		ts := ecsParams.TaskDefinition.TaskSize
		assert.Equal(t, "200", ts.Cpu, "Expected CPU to match")
		assert.Equal(t, "1000", ts.Memory, "Expected CPU to match")
	}

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")
}

func TestParseECSParams_NoFile(t *testing.T) {
	project := setupTestProject(t)
	err := project.parseECSParams()
	if assert.NoError(t, err) {
		assert.Nil(t, project.ecsContext.ECSParams)
	}
}

func TestParseECSParams_WithFargateParams(t *testing.T) {
	ecsParamsString := `version: 1
task_definition:
  ecs_network_mode: awsvpc
  task_execution_role: arn:aws:iam::123456789012:role/fargate_role
  task_size:
    mem_limit: 1000
    cpu_limit: 200

run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: [subnet-feedface, subnet-deadbeef]
      security_groups:
        - sg-bafff1ed
        - sg-c0ffeefe
      assign_public_ip: ENABLED`

	content := []byte(ecsParamsString)

	tmpfile, err := ioutil.TempFile("", "ecs-params")
	assert.NoError(t, err, "Could not create ecs fields tempfile")

	ecsParamsFileName := tmpfile.Name()
	defer os.Remove(ecsParamsFileName)

	project := setupTestProjectWithEcsParams(t, ecsParamsFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs fields tempfile")

	err = project.parseECSParams()
	if assert.NoError(t, err) {
		ecsParams := project.ecsContext.ECSParams
		assert.NotNil(t, ecsParams, "Expected ecsParams to be set on project")
		assert.Equal(t, "1", ecsParams.Version, "Expected Version to match")

		td := ecsParams.TaskDefinition
		assert.Equal(t, "awsvpc", td.NetworkMode, "Expected NetworkMode to match")
		assert.Equal(t, "arn:aws:iam::123456789012:role/fargate_role", td.ExecutionRole, "Expected ExecutionRole to match")

		ts := td.TaskSize
		assert.Equal(t, "200", ts.Cpu, "Expected CPU to match")
		assert.Equal(t, "1000", ts.Memory, "Expected CPU to match")

		networkConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration
		assert.Equal(t, []string{"subnet-feedface", "subnet-deadbeef"}, networkConfig.Subnets, "Expected Subnets to match")
		assert.Equal(t, []string{"sg-bafff1ed", "sg-c0ffeefe"}, networkConfig.SecurityGroups, "Expected SecurityGroups to match")
		assert.Equal(t, utils.Enabled, networkConfig.AssignPublicIp, "Expected AssignPublicIp to match")

	}

	err = tmpfile.Close()
	assert.NoError(t, err, "Could not close tempfile")
}

func TestThrowErrorForUnsupportedComposeVersion(t *testing.T) {
	unsupportedVersion := "4"
	composeFileString := `version: '` + unsupportedVersion + `'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
    mem_reservation: 500000000
  mysql:
    image: mysql`

	// set up compose file
	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal("Unexpected error in creating test file", err)
	}
	defer os.Remove(tmpfile.Name())
	if _, err := tmpfile.Write([]byte(composeFileString)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}

	// set up project and parse
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())
	observedError := project.parseCompose()

	expectedError := "Unsupported Docker Compose version found: " + unsupportedVersion

	if assert.Error(t, observedError) {
		assert.Equal(t, expectedError, observedError.Error())
	}
}

func setupTestProject(t *testing.T) *ecsProject {
	return setupTestProjectWithEcsParams(t, "")
}

func setupTestProjectWithEcsParams(t *testing.T, ecsParamsFileName string) *ecsProject {
	return setupTestProjectWithECSRegistryCreds(t, ecsParamsFileName, "")
}

// TODO: refactor into all-purpose 'setupTestProject' func
func setupTestProjectWithECSRegistryCreds(t *testing.T, ecsParamsFileName, credFileName string) *ecsProject {
	envLookup, err := utils.GetDefaultEnvironmentLookup()
	assert.NoError(t, err, "Unexpected error setting up environment lookup")

	resourceLookup, err := utils.GetDefaultResourceLookup()
	assert.NoError(t, err, "Unexpected error setting up resource lookup")

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.ProjectNameFlag, testProjectName, "")
	flagSet.String(flags.ECSParamsFileNameFlag, ecsParamsFileName, "")
	flagSet.String(flags.RegistryCredsFileNameFlag, credFileName, "")

	parentContext := cli.NewContext(nil, flagSet, nil)
	cliContext := cli.NewContext(nil, nil, parentContext)

	ecsContext := &context.ECSContext{
		CLIContext: cliContext,
	}
	ecsContext.EnvironmentLookup = envLookup
	ecsContext.ResourceLookup = resourceLookup

	return &ecsProject{
		ecsContext: ecsContext,
	}
}

func TestParseECSRegistryCreds(t *testing.T) {
	credsInputString := `version: "1"
registry_credential_outputs:
  task_execution_role: someTestRole
  container_credentials:
    my.example.registry.net:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net
      container_names:
      - web
    another.example.io:
      credentials_parameter: arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-another.example.io
      kms_key_id: arn:aws:kms::key/some-key-57yrt
      container_names:
      - test`

	content := []byte(credsInputString)

	tmpfile, err := ioutil.TempFile("", regcredio.ECSCredFileBaseName)
	assert.NoError(t, err, "Could not create ecs registry creds tempfile")

	credFileName := tmpfile.Name()
	defer os.Remove(credFileName)

	project := setupTestProjectWithECSRegistryCreds(t, "", credFileName)

	_, err = tmpfile.Write(content)
	assert.NoError(t, err, "Could not write data to ecs registry creds tempfile")

	if err := project.parseECSRegistryCreds(); err != nil {
		t.Fatalf("Unexpected error parsing the "+regcredio.ECSCredFileBaseName+" data [%s]: %v", credsInputString, err)
	}

	ecsRegCreds := project.ecsRegistryCreds
	assert.NotNil(t, ecsRegCreds, "Expected "+regcredio.ECSCredFileBaseName+" to be set on project")
	assert.Equal(t, "1", ecsRegCreds.Version, "Expected Version to match")

	credResources := ecsRegCreds.CredentialResources
	assert.NotNil(t, credResources, "Expected credential resources to be non-nil")
	assert.Equal(t, "someTestRole", credResources.TaskExecutionRole)
	assert.NotNil(t, credResources.ContainerCredentials, "Expected ContainerCredentials to be non-nil")

	firstOutputEntry := credResources.ContainerCredentials["my.example.registry.net"]
	assert.NotEmpty(t, firstOutputEntry)
	assert.Equal(t, "arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-my.example.registry.net", firstOutputEntry.CredentialARN)
	assert.Equal(t, "", firstOutputEntry.KMSKeyID)
	assert.ElementsMatch(t, []string{"web"}, firstOutputEntry.ContainerNames)

	secondOutputEntry := credResources.ContainerCredentials["another.example.io"]
	assert.NotEmpty(t, secondOutputEntry)
	assert.Equal(t, "arn:aws:secretsmanager::secret:amazon-ecs-cli-setup-another.example.io", secondOutputEntry.CredentialARN)
	assert.Equal(t, "arn:aws:kms::key/some-key-57yrt", secondOutputEntry.KMSKeyID)
	assert.ElementsMatch(t, []string{"test"}, secondOutputEntry.ContainerNames)
}

func TestParseECSRegistryCreds_NoFile(t *testing.T) {
	project := setupTestProject(t)
	err := project.parseECSRegistryCreds()
	if assert.NoError(t, err) {
		assert.Nil(t, project.ecsRegistryCreds)
	}
}
