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
	"reflect"
	"sort"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/docker/libcompose/project"
	"github.com/docker/libcompose/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const testProjectName = "test-project"

func TestParseComposeForVersion1Files(t *testing.T) {
	// test data
	redisImage := "redis"
	cpuShares := int64(73)
	command := []string{"bundle exec thin -p 3000"}
	dnsServers := []string{"1.2.3.4"}
	dnsSearchDomains := []string{"search.example.com"}
	entryPoint := "/code/entrypoint.sh"
	env := []string{"RACK_ENV=development", "SESSION_PORT=session_port"}
	extraHosts := []string{"test.local:127.10.10.10"}
	hostname := "foobarbaz"
	labels := map[string]string{
		"label1":         "",
		"com.foo.label2": "value",
	}
	links := []string{"redis:redis"}
	logDriver := "json-file"
	logOpts := map[string]string{
		"max-file": "50",
		"max-size": "50k",
	}
	memLimit := int64(1000000000)
	ports := []string{"5000:5000", "127.0.0.1:8001:8001"}
	privileged := true
	readonly := true
	securityOpts := []string{"label:type:test_virt"}
	user := "user"
	volume := yaml.Volume{Destination: "./code"}
	volumes := yaml.Volumes{Volumes: []*yaml.Volume{&volume}}
	workingDir := "/var"

	composeFileString := `web:
  cpu_shares: 73
  command:
   - bundle exec thin -p 3000
  dns:
   - 1.2.3.4
  dns_search: search.example.com
  entrypoint: /code/entrypoint.sh
  environment:
    RACK_ENV: development
    SESSION_PORT: session_port
  extra_hosts:
   - test.local:127.10.10.10
  hostname: "foobarbaz"
  image: web
  labels:
   - label1
   - com.foo.label2=value
  links:
   - "redis:redis"
  log_driver: json-file
  log_opt:
    max-file: 50
    max-size: 50k
  mem_limit: 1000000000
  ports:
   - '5000:5000'
   - "127.0.0.1:8001:8001"
  privileged: true
  read_only: true
  security_opt:
   - label:type:test_virt
  ulimits:
    nofile: 1024
  user: user
  volumes:
   - ./code
  working_dir: /var
redis:
  image: redis`

	// setup project and parse
	composeBytes := [][]byte{}
	composeBytes = append(composeBytes, []byte(composeFileString))
	project := setupTestProject(t)
	project.context.ComposeBytes = composeBytes

	if err := project.parseCompose(); err != nil {
		t.Fatalf("Unexpected error parsing the compose string [%s]: %v", composeFileString, err)
	}

	if testProjectName != project.context.ProjectName {
		t.Errorf("ProjectName not overridden. Expected [%s] Got [%s]", testProjectName, project.context.ProjectName)
	}

	configs := project.ServiceConfigs()
	// verify redis ServiceConfig
	redis, _ := configs.Get("redis")
	if redis == nil || redis.Image != redisImage {
		t.Fatalf("Expected [%s] as a service with image [%s] but got configs [%v]", "redis", redisImage, configs)
	}

	// verify web ServiceConfig
	web, _ := configs.Get("web")
	if web == nil {
		t.Fatalf("Expected [%s] as a service but got configs [%v]", "web", configs)
	}
	if cpuShares != int64(web.CPUShares) {
		t.Errorf("Expected cpuShares to be [%d] but got [%d]", cpuShares, web.CPUShares)
	}
	if len(web.Command) != 1 || !reflect.DeepEqual(command[0], web.Command[0]) {
		t.Errorf("Expected command to be [%v] but got [%v]", command, web.Command)
	}
	if len(web.DNS) != 1 || !reflect.DeepEqual(dnsServers[0], web.DNS[0]) {
		t.Errorf("Expected dns to be [%v] but got [%v]", dnsServers, web.DNS)
	}
	if len(web.DNSSearch) != 1 || !reflect.DeepEqual(dnsSearchDomains[0], web.DNSSearch[0]) {
		t.Errorf("Expected dns search to be [%v] but got [%v]", dnsSearchDomains, web.DNSSearch)
	}
	if len(web.Entrypoint) != 1 || entryPoint != web.Entrypoint[0] {
		t.Errorf("Expected entryPoint to be [%s] but got [%s]", entryPoint, web.Entrypoint)
	}

	sort.Strings(env)
	webEnv := []string{}
	for _, val := range web.Environment {
		webEnv = append(webEnv, val)
	}
	sort.Strings(webEnv)
	if !reflect.DeepEqual(env, webEnv) {
		t.Errorf("Expected Environment to be [%v] but got [%v]", env, webEnv)
	}
	if !reflect.DeepEqual(extraHosts, web.ExtraHosts) {
		t.Errorf("Expected extraHosts to be [%v] but got [%v]", extraHosts, web.ExtraHosts)
	}
	if hostname != web.Hostname {
		t.Errorf("Expected Hostname to be [%s] but got [%s]", hostname, web.Hostname)
	}
	if len(labels) != len(web.Labels) ||
		labels["label1"] != web.Labels["label1"] || labels["com.foo.label2"] != web.Labels["com.foo.label2"] {
		t.Errorf("Expected labels to be [%v] but got [%v]", labels, web.Labels)
	}
	if len(web.Links) != 1 || !reflect.DeepEqual(links[0], web.Links[0]) {
		t.Errorf("Expected links to be [%v] but got [%v]", links, web.Links)
	}
	if logDriver != web.Logging.Driver {
		t.Errorf("Expected logDriver to be [%s] but got [%s]", logDriver, web.Logging.Driver)
	}
	if !reflect.DeepEqual(logOpts, web.Logging.Options) {
		t.Errorf("Expected logOpts to be [%v] but got [%v]", logOpts, web.Logging.Options)
	}
	if memLimit != int64(web.MemLimit) {
		t.Errorf("Expected memLimit to be [%d] but got [%d]", memLimit, web.MemLimit)
	}
	if !reflect.DeepEqual(ports, web.Ports) {
		t.Errorf("Expected ports to be [%v] but got [%v]", ports, web.Ports)
	}
	if privileged != web.Privileged {
		t.Errorf("Expected privileged to be [%t] but got [%t]", privileged, web.Privileged)
	}
	if readonly != web.ReadOnly {
		t.Errorf("Expected readonly to be [%t] but got [%t]", readonly, web.ReadOnly)
	}
	if !reflect.DeepEqual(securityOpts, web.SecurityOpt) {
		t.Errorf("Expected securityOpts to be [%v] but got [%v]", securityOpts, web.SecurityOpt)
	}
	if user != web.User {
		t.Errorf("Expected user to be [%s] but got [%s]", user, web.User)
	}
	if len(volumes.Volumes) != len(web.Volumes.Volumes) {
		t.Errorf("Expected len of volumes to be [%d] but got [%d]", len(volumes.Volumes), len(web.Volumes.Volumes))
	}
	if !reflect.DeepEqual(*volumes.Volumes[0], *web.Volumes.Volumes[0]) {
		t.Errorf("Expected volumes to be [%v] but got [%v]", volumes.Volumes[0], web.Volumes.Volumes[0])
	}
	if workingDir != web.WorkingDir {
		t.Errorf("Expected workingDir to be [%s] but got [%s]", user, web.WorkingDir)
	}

}

func TestParseComposeForVersion2Files(t *testing.T) {
	wordpressImage := "wordpress"
	mysqlImage := "mysql"
	ports := []string{"80:80"}
	memoryReservation := int64(500000000)

	composeFileString := `version: '2'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
    mem_reservation: 500000000
  mysql:
    image: mysql`

	// setup project and parse
	composeBytes := [][]byte{}
	composeBytes = append(composeBytes, []byte(composeFileString))
	project := setupTestProject(t)
	project.context.ComposeBytes = composeBytes

	if err := project.parseCompose(); err != nil {
		t.Fatalf("Unexpected error parsing the compose string [%s]: %v", composeFileString, err)
	}

	configs := project.ServiceConfigs()

	// verify wordpress ServiceConfig
	wordpress, ok := configs.Get("wordpress")
	if wordpress == nil || !ok || wordpress.Image != wordpressImage {
		t.Fatalf("Expected [%s] as a service with image [%s] but got configs [%v]", "redis", wordpressImage, configs)
	}
	if !reflect.DeepEqual(ports, wordpress.Ports) {
		t.Errorf("Expected ports to be [%v] but got [%v]", ports, wordpress.Ports)
	}

	assert.Equal(t, memoryReservation, int64(wordpress.MemReservation), "Expected memoryReservation to match")

	// verify mysql ServiceConfig
	mysql, ok := configs.Get("mysql")
	if mysql == nil || !ok || mysql.Image != mysqlImage {
		t.Fatalf("Expected [%s] as a service with image [%s] but got configs [%v]", "redis", mysqlImage, configs)
	}
}

func TestParseComposeForVersion1WithEnvFile(t *testing.T) {
	envKey := "rails_env"
	envValue := "development"
	envContents := []byte(envKey + "=" + envValue)

	envFile, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal("Error creating tmp file:", err)
	}
	defer os.Remove(envFile.Name()) // clean up
	if _, err := envFile.Write(envContents); err != nil {
		t.Fatal("Error writing to tmp file:", err)
	}

	webImage := "webapp"

	composeFileString := `web:
  image: webapp
  env_file:
  - ` + envFile.Name()

	// setup project and parse
	composeBytes := [][]byte{}
	composeBytes = append(composeBytes, []byte(composeFileString))
	project := setupTestProject(t)
	project.context.ComposeBytes = composeBytes

	if err := project.parseCompose(); err != nil {
		t.Fatalf("Unexpected error parsing the compose string [%s]: %v", composeFileString, err)
	}

	configs := project.ServiceConfigs()

	// verify wordpress ServiceConfig
	web, ok := configs.Get("web")
	if web == nil || !ok || web.Image != webImage {
		t.Fatalf("Expected [%s] as a service with image [%s] but got configs [%v]", "redis", webImage, configs)
	}

	// skips the second one if envKey2
	if web.Environment == nil || len(web.Environment) != 1 {
		t.Fatalf("Expected non empty Environment, but was [%v]", web.Environment)
	}
	if string(envContents) != web.Environment[0] {
		t.Errorf("Expected env [%s]=[%s] But was [%v]", envKey, envValue, web.Environment)
	}
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

	ecsParams := project.context.ECSParams
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

func TestParseECSParams_NoFile(t *testing.T) {
	project := setupTestProject(t)
	err := project.parseECSParams()
	if assert.NoError(t, err) {
		assert.Nil(t, project.context.ECSParams)
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
		ecsParams := project.context.ECSParams
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

func setupTestProject(t *testing.T) *ecsProject {
	return setupTestProjectWithEcsParams(t, "")
}

func setupTestProjectWithEcsParams(t *testing.T, ecsParamsFileName string) *ecsProject {
	envLookup, err := utils.GetDefaultEnvironmentLookup()
	if err != nil {
		t.Fatal("Unexpected error in setting up a project", err)
	}
	resourceLookup, err := utils.GetDefaultResourceLookup()
	if err != nil {
		t.Fatal("Unexpected error in setting up a project", err)
	}

	flagSet := flag.NewFlagSet("ecs-cli", 0)
	flagSet.String(flags.ProjectNameFlag, testProjectName, "")
	flagSet.String(flags.ECSParamsFileNameFlag, ecsParamsFileName, "")

	parentContext := cli.NewContext(nil, flagSet, nil)
	cliContext := cli.NewContext(nil, nil, parentContext)

	ecsContext := &context.Context{
		CLIContext: cliContext,
	}
	ecsContext.EnvironmentLookup = envLookup
	ecsContext.ResourceLookup = resourceLookup
	libcomposeProject := project.NewProject(&ecsContext.Context, nil, nil)

	return &ecsProject{
		context: ecsContext,
		Project: *libcomposeProject,
	}
}
