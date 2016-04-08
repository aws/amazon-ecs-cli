// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package utils

import (
	"reflect"
	"sort"
	"testing"

	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
)

func TestUnmarshalComposeConfig(t *testing.T) {
	redisImage := "redis"
	cpuShares := int64(73)
	command := []string{"bundle exec thin -p 3000"}
	dnsServers := []string{"1.2.3.4"}
	dnsSearchDomains := []string{"search.example.com"}
	entryPoint := "/code/entrypoint.sh"
	envFiles := []string{"./common.env", "/opt/prod.env"}
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
	ulimits := []string{"nofile=1024"}
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
  env_file:
   - ./common.env
   - /opt/prod.env
  environment:
    RACK_ENV: development
    SESSION_PORT: session_port
  extra_hosts:
   - test.local:127.10.10.10
  hostname: "foobarbaz"
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
   - "5000:5000"
   - "127.0.0.1:8001:8001"
  privileged: true
  read_only: true
  security_opt:
   - label:type:test_virt
  ulimits:
   - nofile=1024
  user: user
  volumes:
   - .:/code
  volumes_from:
  working_dir: /var
redis:
  image: redis`

	context := libcompose.Context{
		ComposeBytes: []byte(composeFileString),
	}
	configs, err := UnmarshalComposeConfig(context)
	if err != nil {
		t.Fatalf("Unable to unmarshall compose string [%s]", composeFileString)
	}

	// verify redis ServiceConfig
	redis := configs["redis"]
	if redis == nil || redis.Image != redisImage {
		t.Fatalf("Expected [%s] as a service with image [%s] but got configs [%v]", "redis", redisImage, configs)
	}

	// verify web ServiceConfig
	web := configs["web"]
	if web == nil {
		t.Fatalf("Expected [%s] as a service but got configs [%v]", "web", configs)
	}
	if cpuShares != web.CpuShares {
		t.Errorf("Expected cpuShares to be [%s] but got [%s]", cpuShares, web.CpuShares)
	}
	if !reflect.DeepEqual(command, web.Command.Slice()) {
		t.Errorf("Expected command to be [%v] but got [%v]", command, web.Command.Slice())
	}
	if !reflect.DeepEqual(dnsServers, web.DNS.Slice()) {
		t.Errorf("Expected dns to be [%v] but got [%v]", dnsServers, web.DNS.Slice())
	}
	if !reflect.DeepEqual(dnsSearchDomains, web.DNSSearch.Slice()) {
		t.Errorf("Expected dns search to be [%v] but got [%v]", dnsSearchDomains, web.DNSSearch.Slice())
	}
	if len(web.Entrypoint.Slice()) != 1 || entryPoint != web.Entrypoint.ToString() {
		t.Errorf("Expected entryPoint to be [%s] but got [%s]", entryPoint, web.Entrypoint.ToString())
	}

	sort.Strings(env)
	webEnv := []string{}
	for _, val := range web.Environment.Slice() {
		webEnv = append(webEnv, val)
	}
	sort.Strings(webEnv)
	if !reflect.DeepEqual(env, webEnv) {
		t.Errorf("Expected Environment to be [%v] but got [%v]", env, webEnv)
	}
	if !reflect.DeepEqual(envFiles, web.EnvFile.Slice()) {
		t.Errorf("Expected env_file to be [%v] but got [%v]", envFiles, web.EnvFile.Slice())
	}

	if !reflect.DeepEqual(extraHosts, web.ExtraHosts) {
		t.Errorf("Expected extraHosts to be [%v] but got [%v]", extraHosts, web.ExtraHosts)
	}
	if hostname != web.Hostname {
		t.Errorf("Expected Hostname to be [%s] but got [%s]", hostname, web.Hostname)
	}
	if !reflect.DeepEqual(labels, web.Labels.MapParts()) {
		t.Errorf("Expected labels to be [%v] but got [%v]", labels, web.Labels.MapParts())
	}
	if !reflect.DeepEqual(links, web.Links.Slice()) {
		t.Errorf("Expected links to be [%v] but got [%v]", links, web.Links.Slice())
	}
	if logDriver != web.LogDriver {
		t.Errorf("Expected logDriver to be [%s] but got [%s]", logDriver, web.LogDriver)
	}
	if !reflect.DeepEqual(logOpts, web.LogOpt) {
		t.Errorf("Expected logOpts to be [%v] but got [%v]", logOpts, web.LogOpt)
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
	if !reflect.DeepEqual(ulimits, web.ULimits) {
		t.Errorf("Expected ulimits to be [%v] but got [%v]", ulimits, web.ULimits)
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

func TestUnmarshalComposeConfigSkipUnsupportedYaml(t *testing.T) {
	composeFileString := `web:
  external_links:
    - /tmp/compose.yml`

	context := libcompose.Context{
		ComposeBytes: []byte(composeFileString),
	}
	configs, err := UnmarshalComposeConfig(context)
	if err != nil {
		t.Fatalf("Unable to unmarshall compose string [%s]", composeFileString)
	}

	// verify web ServiceConfig
	web := configs["web"]
	if web == nil {
		t.Fatalf("Expected [%s] as a service but got configs [%v]", "web", configs)
	}
	if len(web.ExternalLinks) != 0 {
		t.Errorf("Expected external links to be empty but got [%s]", web.ExternalLinks)
	}
}

func TestUnmarshalComposeConfigRecoverFromPanic(t *testing.T) {
	// panic: interface conversion: yaml.Node is yaml.Map, not yaml.List
	composeFileString := `root:
  child1:
  - child2:`

	context := libcompose.Context{
		ComposeBytes: []byte(composeFileString),
	}
	_, err := UnmarshalComposeConfig(context)
	if err == nil {
		t.Fatalf("Should have failed while parsing the config [%s]", composeFileString)
	}
}
