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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/value"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/libcompose/config"
	"github.com/docker/libcompose/project"
	log "github.com/sirupsen/logrus"
)

// supported fields/options from compose 1/2 YAML file
var supportedComposeV1V2YamlOptions = []string{
	"cap_add",
	"cap_drop",
	"command",
	"cpu_shares",
	"devices",
	"dns",
	"dns_search",
	"entrypoint",
	"env_file",
	"environment",
	"extends",
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

// supported fields/options from compose 3 YAML file
var supportedFieldsInV3 = map[string]bool{
	"CapAdd":      true,
	"CapDrop":     true,
	"Command":     true,
	"Devices":     true,
	"DNS":         true,
	"DNSSearch":   true,
	"Entrypoint":  true,
	"Environment": true,
	"EnvFile":     true,
	"ExtraHosts":  true,
	"Hostname":    true,
	"HealthCheck": true,
	"Image":       true,
	"Labels":      true,
	"Links":       true,
	"Logging":     true,
	"Name":        true,
	"Ports":       true,
	"Privileged":  true,
	"ReadOnly":    true,
	"SecurityOpt": true,
	"Tmpfs":       true,
	"Ulimits":     true,
	"User":        true,
	"Volumes":     true,
	"WorkingDir":  true,
}

var supportedComposeV1V2YamlOptionsMap = getSupportedComposeV1V2YamlOptionsMap()

// Create set of supported YAML fields in docker compose v2 ServiceConfigs
func getSupportedComposeV1V2YamlOptionsMap() map[string]bool {
	optionsMap := make(map[string]bool)
	for _, value := range supportedComposeV1V2YamlOptions {
		optionsMap[value] = true
	}
	return optionsMap
}

// LogUnsupportedV1V2ServiceConfigFields logs a warning if there is an unsupported field specified in the docker-compose file
func LogUnsupportedV1V2ServiceConfigFields(serviceName string, serviceConfig *config.ServiceConfig) {
	configValue := reflect.ValueOf(serviceConfig).Elem()
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := configType.Field(i)
		// get the tag name (if any), defaults to fieldName
		tagName := fieldType.Name
		yamlTag := fieldType.Tag.Get("yaml") // Expected format `yaml:"tagName,omitempty"`
		if yamlTag != "" {
			tags := strings.Split(yamlTag, ",")
			if tags[0] != "" {
				tagName = tags[0]
			}
		}

		if tagName == "networks" && !validNetworksForService(serviceConfig) {
			logWarningForUnsupportedServiceOption(tagName, serviceName)
		}

		zeroValue := value.IsZero(field)
		// if value is present for the field that is not in supportedYamlTags map, log a warning
		if tagName != "networks" && !zeroValue && !supportedComposeV1V2YamlOptionsMap[tagName] {
			logWarningForUnsupportedServiceOption(tagName, serviceName)
		}
	}
}

// LogUnsupportedV3ServiceConfigFields logs a warning if there is an unsupported field specified in the docker-compose file
func LogUnsupportedV3ServiceConfigFields(servConfig types.ServiceConfig) {
	configValue := reflect.ValueOf(servConfig)
	configType := configValue.Type()

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Field(i)
		fieldType := configType.Field(i)

		if supportedFieldsInV3[fieldType.Name] == false && !value.IsZero(field) {
			// convert field name so it more closely resembles option in yaml file
			optionName := adapter.ConvertCamelCaseToUnderScore(fieldType.Name)
			logWarningForUnsupportedServiceOption(optionName, servConfig.Name)
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

func logWarningForUnsupportedServiceOption(tagName, serviceName string) {
	log.WithFields(log.Fields{
		"option name":  tagName,
		"service name": serviceName,
	}).Warn("Skipping unsupported YAML option for service...")
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
