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

package ecs

import (
	"flag"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
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
	volumes := []string{".:/code"}
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
   - .:/code
  working_dir: /var
redis:
  image: redis`

	// setup project and parse
	composeBytes := [][]byte{}
	composeBytes = append(composeBytes, []byte(composeFileString))
	project := setupTestProject(t)
	project.context.ComposeBytes = composeBytes

	if err := project.parseCompose(); err != nil {
		t.Fatalf("Unexpected error parsing the compose string [%s]", composeFileString, err)
	}

	if testProjectName != project.context.ProjectName {
		t.Errorf("ProjectName not overriden. Expected [%s] Got [%s]", testProjectName, project.context.ProjectName)
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
	if cpuShares != web.CPUShares {
		t.Errorf("Expected cpuShares to be [%s] but got [%s]", cpuShares, web.CPUShares)
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
	if memLimit != web.MemLimit {
		t.Errorf("Expected memLimit to be [%s] but got [%s]", memLimit, web.MemLimit)
	}
	if !reflect.DeepEqual(ports, web.Ports) {
		t.Errorf("Expected ports to be [%v] but got [%v]", ports, web.Ports)
	}
	if privileged != web.Privileged {
		t.Errorf("Expected privileged to be [%s] but got [%s]", privileged, web.Privileged)
	}
	if readonly != web.ReadOnly {
		t.Errorf("Expected readonly to be [%s] but got [%s]", readonly, web.ReadOnly)
	}
	if !reflect.DeepEqual(securityOpts, web.SecurityOpt) {
		t.Errorf("Expected securityOpts to be [%v] but got [%v]", securityOpts, web.SecurityOpt)
	}
	if user != web.User {
		t.Errorf("Expected user to be [%s] but got [%s]", user, web.User)
	}
	if !reflect.DeepEqual(volumes, web.Volumes) {
		t.Errorf("Expected volumes to be [%v] but got [%v]", volumes, web.Volumes)
	}
	if workingDir != web.WorkingDir {
		t.Errorf("Expected workingDir to be [%s] but got [%s]", user, web.WorkingDir)
	}

}

func TestParseComposeForVersion2Files(t *testing.T) {
	wordpressImage := "wordpress"
	mysqlImage := "mysql"
	ports := []string{"80:80"}

	composeFileString := `version: '2'
services:
  wordpress:
    image: wordpress
    ports: ["80:80"]
  mysql:
    image: mysql`

	// setup project and parse
	composeBytes := [][]byte{}
	composeBytes = append(composeBytes, []byte(composeFileString))
	project := setupTestProject(t)
	project.context.ComposeBytes = composeBytes

	if err := project.parseCompose(); err != nil {
		t.Fatalf("Unexpected error parsing the compose string [%s]", composeFileString, err)
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
		t.Fatalf("Unexpected error parsing the compose string [%s]", composeFileString, err)
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

func setupTestProject(t *testing.T) *ecsProject {
	envLookup, err := utils.GetDefaultEnvironmentLookup()
	if err != nil {
		t.Fatal("Unexpected error in setting up a project", err)
	}
	resourceLookup, err := utils.GetDefaultResourceLookup()
	if err != nil {
		t.Fatal("Unexpected error in setting up a project", err)
	}

	composeContext := flag.NewFlagSet("ecs-cli", 0)
	composeContext.String(ProjectNameFlag, testProjectName, "")
	parentContext := cli.NewContext(nil, composeContext, nil)
	cliContext := cli.NewContext(nil, nil, parentContext)

	ecsContext := &Context{
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
