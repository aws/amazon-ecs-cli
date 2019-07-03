// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package converter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	composeV3 "github.com/docker/cli/cli/compose/types"
	"github.com/pkg/errors"
)

const (
	// jsonFileLogDriver is the default Docker logger.
	jsonFileLogDriver = "json-file"
)

// ConvertToComposeOverride returns a Docker Compose object to be used to override containers defined
// in the task definition.
//
// Overrides the AWS_CONTAINER_CREDENTIALS_RELATIVE_URI environment variable to "/creds" for every service.
// Overrides the logging driver to "json-file" for every service.
func ConvertToComposeOverride(taskDefinition *ecs.TaskDefinition) (*composeV3.Config, error) {
	if taskDefinition == nil {
		return nil, errors.New("task definition cannot be nil")
	}
	if len(taskDefinition.ContainerDefinitions) == 0 {
		return nil, errors.New("task definition needs to have container definitions")
	}

	var services []composeV3.ServiceConfig
	for _, container := range taskDefinition.ContainerDefinitions {
		conf := composeV3.ServiceConfig{
			Name: aws.StringValue(container.Name),
			Environment: composeV3.MappingWithEquals{
				ecsCredsProviderEnvName: aws.String(endpointsTempCredsPath),
			},
			Logging: &composeV3.LoggingConfig{
				Driver: jsonFileLogDriver,
			},
		}
		services = append(services, conf)
	}

	return &composeV3.Config{
		Version:  composeVersion,
		Services: services,
	}, nil
}
