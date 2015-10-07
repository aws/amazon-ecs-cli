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
	memLimit := int64(1000000000)
	ports := []string{"5000:5000", "127.0.0.1:8001:8001"}
	links := []string{"redis:redis"}
	volumes := []string{".:/code"}
	command := []string{"bundle exec thin -p 3000"}
	entryPoint := "/code/entrypoint.sh"
	env := []string{"RACK_ENV=development", "SESSION_SECRET=session_secret"}

	composeFileString := `web:
  cpu_shares: 73
  mem_limit: 1000000000
  entrypoint: /code/entrypoint.sh
  command: 
   - bundle exec thin -p 3000
  ports:
   - "5000:5000"
   - "127.0.0.1:8001:8001"
  volumes:
   - .:/code
  environment:
    RACK_ENV: development
    SESSION_SECRET: session_secret
  links:
   - "redis:redis"
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
	if memLimit != web.MemLimit {
		t.Errorf("Expected memLimit to be [%s] but got [%s]", memLimit, web.MemLimit)
	}
	if !reflect.DeepEqual(ports, web.Ports) {
		t.Errorf("Expected ports to be [%v] but got [%v]", ports, web.Ports)
	}
	if !reflect.DeepEqual(volumes, web.Volumes) {
		t.Errorf("Expected volumes to be [%v] but got [%v]", volumes, web.Volumes)
	}

	if !reflect.DeepEqual(command, web.Command.Slice()) {
		t.Errorf("Expected command to be [%v] but got [%v]", command, web.Command.Slice())
	}
	if len(web.Entrypoint.Slice()) != 1 || entryPoint != web.Entrypoint.ToString() {
		t.Errorf("Expected entryPoint to be [%s] but got [%s]", entryPoint, web.Entrypoint.ToString())
	}
	if !reflect.DeepEqual(links, web.Links.Slice()) {
		t.Errorf("Expected links to be [%v] but got [%v]", links, web.Links.Slice())
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
