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

package network

import (
	"time"

	"github.com/docker/docker/api/types"
	"golang.org/x/net/context"
)

// LocalEndpointsStopper groups the Docker NetworkInspect, ContainerStop, and NetworkRemove functions.
//
// These functions can be used together to remove a network once unwanted containers in the network are stopped.
type LocalEndpointsStopper interface {
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
	NetworkRemove(ctx context.Context, networkID string) error
}

// Teardown stops the Local Endpoints container and removes the Local network created by Setup.
// If there are other containers running in the network besides the endpoints container, this function does nothing.
//
// If there is any unexpected errors, we exit the program with a fatal log.
func Teardown(dockerClient *LocalEndpointsStopper) {

}
