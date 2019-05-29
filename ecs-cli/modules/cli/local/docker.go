// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package local

import (
	"os"

	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
)

const (
	// minDockerAPIVersion is the minimum Docker API version that supports
	// both the Local Endpoints container and the Docker API operations used by "local" sub-commands.
	// See https://github.com/awslabs/amazon-ecs-local-container-endpoints/blob/3417a48b676c5b215fb9583bcbdc8a0b0e23aa8e/local-container-endpoints/clients/docker/client.go#L30.
	minDockerAPIVersion = "1.27"
)

func newDockerClient() *client.Client {
	if os.Getenv("DOCKER_API_VERSION") == "" {
		// If the user does not explicitly set the API version, then the SDK can choose
		// an API version that's too new for the user's Docker engine.
		_ = os.Setenv("DOCKER_API_VERSION", minDockerAPIVersion)
	}

	client, err := client.NewEnvClient()
	if err != nil {
		logrus.Fatalf("Could not create a docker client due to %v", err)
	}
	return client
}
