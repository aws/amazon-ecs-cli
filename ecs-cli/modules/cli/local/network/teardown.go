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
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// LocalEndpointsStopper groups the Docker NetworkInspect, ContainerStop, ContainerRemove, and NetworkRemove functions.
//
// These functions can be used together to remove a network once unwanted containers are stopped.
type LocalEndpointsStopper interface {
	networkInspector
	containerStopper
	containerRemover
	networkRemover
}

type networkInspector interface {
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
}

type containerStopper interface {
	ContainerStop(ctx context.Context, containerID string, timeout *time.Duration) error
}

type containerRemover interface {
	ContainerRemove(ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
}

type networkRemover interface {
	NetworkRemove(ctx context.Context, networkID string) error
}

// Teardown removes both the Local Endpoints container and the Local network created by Setup.
// If there are other containers running in the network besides the endpoints container, this function does nothing.
//
// If there is any unexpected errors, we exit the program with a fatal log.
func Teardown(dockerClient LocalEndpointsStopper) {
	if hasRunningTasksInNetwork(dockerClient) {
		return
	}
	logrus.Infof("The network %s has no more running tasks, stopping the endpoints containers...", EcsLocalNetworkName)

	stopEndpointsContainer(dockerClient)
	removeEndpointsContainer(dockerClient)
	removeLocalNetwork(dockerClient)
}

// hasRunningTasksInNetwork returns true if there are other containers besides the
// endpoints container running in the local network, false otherwise.
func hasRunningTasksInNetwork(d networkInspector) bool {
	ctx, cancel := context.WithTimeout(context.Background(), dockerTimeout)
	defer cancel()

	resp, err := d.NetworkInspect(ctx, EcsLocalNetworkName, types.NetworkInspectOptions{})
	if err != nil {
		logrus.Fatalf("Failed to inspect network %s due to %v", EcsLocalNetworkName, err)
	}

	if len(resp.Containers) > 1 {
		// Has other containers running in the network
		logrus.Infof("%d other task(s) running locally, skipping network removal.", len(resp.Containers)-1)
		return true
	}

	for _, container := range resp.Containers {
		if container.Name != localEndpointsContainerName {
			// The only container running in the network is a task without the endpoints container.
			// This scenario should not happen unless the user themselves stopped the endpoints container.
			logrus.Warnf("The %s container is running in the %s network without the %s container, please stop it first",
				container.Name, EcsLocalNetworkName, localEndpointsContainerName)
			return true
		}
	}

	return false
}

func stopEndpointsContainer(d containerStopper) {
	ctx, cancel := context.WithTimeout(context.Background(), dockerTimeout)
	defer cancel()

	err := d.ContainerStop(ctx, localEndpointsContainerName, nil)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such container") {
			// The containers in the network were already stopped by the user using "docker stop", do nothing.
			return
		}
		logrus.Fatalf("Failed to stop %s container due to %v", localEndpointsContainerName, err)
	}
	logrus.Infof("Stopped the %s container successfully, removing it...", localEndpointsContainerName)
}

// removeEndpointsContainer removes the endpoints container.
//
// If we do not remove the container, then the user will receive a "network not found" error on using "local up".
// Here is a sample scenario:
// 1) User runs "local up" and creates a new local network with an endpoints container.
// 2) User runs "local down" and stops the endpoints container but does not remove it, however the network is removed.
// 3) User runs "local up" again and creates a new local network but re-starts the old endpoints container.
// The old endpoints container tries to connect to the network created in step 1) and fails.
func removeEndpointsContainer(d containerRemover) {
	ctx, cancel := context.WithTimeout(context.Background(), dockerTimeout)
	defer cancel()

	err := d.ContainerRemove(ctx, localEndpointsContainerName, types.ContainerRemoveOptions{})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such container") {
			// The containers in the network were already removed by the user using "docker rm", do nothing.
			return
		}
		logrus.Fatalf("Failed to remove %s container due to %v", localEndpointsContainerName, err)
	}
	logrus.Infof("Removed the %s container successfully, removing the %s network...",
		localEndpointsContainerName, EcsLocalNetworkName)
}

func removeLocalNetwork(d networkRemover) {
	ctx, cancel := context.WithTimeout(context.Background(), dockerTimeout)
	defer cancel()

	err := d.NetworkRemove(ctx, EcsLocalNetworkName)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no such network") {
			// The network was removed, do nothing.
			return
		}
		logrus.Fatalf("Failed to remove %s network due to %v", EcsLocalNetworkName, err)
	}
	logrus.Infof("Removed the %s network successfully", EcsLocalNetworkName)
}
