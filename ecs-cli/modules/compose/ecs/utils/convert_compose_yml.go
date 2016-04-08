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
	"fmt"

	libcompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/libcompose"
	"github.com/kylelemons/go-gypsy/yaml"
)

// supported fields/options from compose YAML file
var supportedComposeYamlOptions map[string]bool

// initializes the supportedComposeYamlOptions map
func composeOptionsInit() {
	// TODO, extract constants from libcompose tagNames
	supportedComposeYamlOptions = map[string]bool{
		"cpu_shares":   true,
		"command":      true,
		"dns":          true,
		"dns_search":   true,
		"entrypoint":   true,
		"env_file":     true,
		"environment":  true,
		"extra_hosts":  true,
		"hostname":     true,
		"image":        true,
		"labels":       true,
		"links":        true,
		"log_driver":   true,
		"log_opt":      true,
		"mem_limit":    true,
		"ports":        true,
		"privileged":   true,
		"read_only":    true,
		"security_opt": true,
		"ulimits":      true,
		"user":         true,
		"volumes":      true,
		"volumes_from": true,
		"working_dir":  true,
	}
}

// UnmarshalComposeConfig Deserializes the document to yaml Node structure
func UnmarshalComposeConfig(context libcompose.Context) (configs map[string]*libcompose.ServiceConfig, retErr error) {
	// recover panic errors from yaml package
	defer func() {
		if err := recover(); err != nil {
			retErr = fmt.Errorf("Unable to unmarshal compose config. Error: %v", err)
		}
	}()

	yamlConfig := parseComposeConfig(context)
	yamlServices, err := nodeToMap(yamlConfig)
	if err != nil {
		return nil, err
	}

	composeOptionsInit()
	configs, retErr = convertToServiceConfigs(yamlServices)
	return
}

// convertToServiceConfigs transforms the yaml structure to a map of libcompose.ServiceConfigs
func convertToServiceConfigs(services yaml.Map) (map[string]*libcompose.ServiceConfig, error) {
	configs := make(map[string]*libcompose.ServiceConfig)

	for name, data := range services {
		config := libcompose.ServiceConfig{}
		// TODO, while unmarshaling skip unsupported fields and report warnings
		err := unmarshal(data, &config, supportedComposeYamlOptions)
		if err != nil {
			LogError(err, "Unable to convert yaml to service configs")
			return nil, err
		}
		configs[name] = &config
	}
	return configs, nil
}

// parseComposeConfig transforms the compose file/bytes into readable yaml structure
func parseComposeConfig(context libcompose.Context) yaml.Node {
	return parseYamlString(string(context.ComposeBytes))
}
