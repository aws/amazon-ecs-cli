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

package logger

import (
	"reflect"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/value"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	log "github.com/sirupsen/logrus"
)

// supported fields/options from compose YAML file
var supportedComposeYamlOptions = []string{
	"cap_add",
	"cap_drop",
	"command",
	"cpu_shares",
	"dns",
	"dns_search",
	"entrypoint",
	"env_file",
	"environment",
	"extra_hosts",
	"hostname",
	"image",
	"labels",
	"links",
	"logging",
	"log_driver", // v1 only
	"log_opt",    // v1 only
	"mem_limit",
	"mem_reservation",
	"ports",
	"privileged",
	"read_only",
	"security_opt",
	"shm_size",
	"tmpfs",
	"ulimits",
	"user",
	"volumes", // v2
	"volumes_from",
	"working_dir",
}

var supportedComposeYamlOptionsMap = getSupportedComposeYamlOptionsMap()

// Create set of supported YAML fields in docker compose v2 ServiceConfigs
func getSupportedComposeYamlOptionsMap() map[string]bool {
	optionsMap := make(map[string]bool)
	for _, value := range supportedComposeYamlOptions {
		optionsMap[value] = true
	}
	return optionsMap
}

// LogUnsupportedServiceConfigFields logs a warning if there is an unsupported field specified in the docker-compose file
func LogUnsupportedServiceConfigFields(serviceName string, serviceConfig *config.ServiceConfig) {
	configValue := reflect.ValueOf(serviceConfig).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := configType.Field(i)
		// get the tag name (if any), defaults to fieldName
		tagName := fieldType.Name
		yamlTag := fieldType.Tag.Get("yaml") // Expected format `yaml:"tagName,omitempty"` // TODO, handle omitempty
		if yamlTag != "" {
			tags := strings.Split(yamlTag, ",")
			if len(tags) > 0 {
				tagName = tags[0]
			}
		}

		if tagName == "networks" && !validNetworksForService(serviceConfig) {
			log.WithFields(log.Fields{
				"option name":  tagName,
				"service name": serviceName,
			}).Warn("Skipping unsupported YAML option for service...")
		}

		zeroValue := value.IsZero(field)
		// if value is present for the field that is not in supportedYamlTags map, log a warning
		if tagName != "networks" && !zeroValue && !supportedComposeYamlOptionsMap[tagName] {
			log.WithFields(log.Fields{
				"option name":  tagName,
				"service name": serviceName,
			}).Warn("Skipping unsupported YAML option for service...")
		}
	}
}

// LogUnsupportedProjectFields adds a WARNING to the customer about the fields that are unused.
func LogUnsupportedProjectFields(project *project.Project) {
	// ecsProject#parseCompose, which calls the underlying libcompose.Project#Parse(),
	// always populates the project.NetworkConfig with one entry ("default").
	// See: https://github.com/docker/libcompose/blob/master/project/project.go#L277
	if project.NetworkConfigs != nil && len(project.NetworkConfigs) > 1 {
		log.WithFields(log.Fields{"option name": "networks"}).Warn("Skipping unsupported YAML option...")
	}
}

func validNetworksForService(config *config.ServiceConfig) bool {
	if config.Networks == nil {
		return false
	}
	if config.Networks.Networks == nil {
		return false
	}
	if len(config.Networks.Networks) != 1 {
		return false
	}

	return true
}
