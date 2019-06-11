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

// Package network provides functionality to setup and teardown the ECS local network.
package network

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// LocalEndpointsStarter groups Docker functions to create the ECS local network and
// the Local Container Endpoints container if they don't exist.
type LocalEndpointsStarter interface {
	networkCreator
	imagePuller
	containerStarter
}

type networkCreator interface {
	NetworkInspect(ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
	NetworkCreate(ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)
}

type imagePuller interface {
	ImageList(ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	ImagePull(ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
}

type containerStarter interface {
	ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	ContainerCreate(ctx context.Context, config *container.Config,
		hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig,
		containerName string) (container.ContainerCreateCreatedBody, error)
	ContainerStart(ctx context.Context, containerID string, options types.ContainerStartOptions) error
}

// Configuration for the ECS local network.
const (
	// EcsLocalNetworkName is the name of the network created for local ECS tasks to join.
	EcsLocalNetworkName = "ecs-local-network"

	// Range of IP addresses that containers in the network get assigned.
	ecsLocalNetworkSubnet = "169.254.170.0/24"
)

// Configuration for the Local Endpoints container.
const (
	// Name of the image used to pull from DockerHub, see https://hub.docker.com/r/amazon/amazon-ecs-local-container-endpoints
	localEndpointsImageName = "amazon/amazon-ecs-local-container-endpoints"

	// Reserved IP address that the container listens to answer requests on metadata, creds, and stats.
	localEndpointsContainerIpAddr = "169.254.170.2"

	// Name of the container, we need to give it a name so that we don't re-create a container every time we setup.
	localEndpointsContainerName = "amazon-ecs-local-container-endpoints"
)

// Setup creates a user-defined bridge network with a running Local Container Endpoints container. If the network
// already exists or the container is already running then this function does nothing.
//
// If there is any unexpected errors, we exit the program with a fatal log.
func Setup(dockerClient LocalEndpointsStarter) {
	setupLocalNetwork(dockerClient)
	setupLocalEndpointsImage(dockerClient)
	setupLocalEndpointsContainer(dockerClient)
}

func setupLocalNetwork(dockerClient networkCreator) {
	if localNetworkExists(dockerClient) {
		logrus.Infof("The network %s already exists", EcsLocalNetworkName)
		return
	}
	createLocalNetwork(dockerClient)
}

func localNetworkExists(dockerClient networkCreator) bool {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	_, err := dockerClient.NetworkInspect(ctx, EcsLocalNetworkName, types.NetworkInspectOptions{})
	if err != nil {
		if client.IsErrNotFound(err) {
			return false
		}
		// Unexpected error while inspecting docker networks, we want to crash the app.
		logrus.Fatalf("Failed to inspect docker network %s due to %v", EcsLocalNetworkName, err)
	}
	return true
}

func createLocalNetwork(dockerClient networkCreator) {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	logrus.Infof("Creating network: %s...", EcsLocalNetworkName)
	resp, err := dockerClient.NetworkCreate(ctx, EcsLocalNetworkName, types.NetworkCreate{
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{
				{
					Subnet: ecsLocalNetworkSubnet,
				},
			},
		},
	})
	if err != nil {
		logrus.Fatalf("Failed to create network %s with subnet %s due to %v", EcsLocalNetworkName, ecsLocalNetworkSubnet, err)
	}
	logrus.Infof("Created network %s with ID %s", EcsLocalNetworkName, resp.ID)
}

func setupLocalEndpointsImage(dockerClient imagePuller) {
	if localEndpointsImageExists(dockerClient) {
		return
	}
	pullLocalEndpointsImage(dockerClient)
}

func localEndpointsImageExists(dockerClient imagePuller) bool {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	args := filters.NewArgs(filters.Arg("reference", localEndpointsImageName))
	imgs, err := dockerClient.ImageList(ctx, types.ImageListOptions{
		Filters: args,
	})
	if err != nil {
		logrus.Fatalf("Failed to list images with filters %v due to %v", args, err)
	}

	for _, img := range imgs {
		for _, repotag := range img.RepoTags {
			if strings.HasPrefix(repotag, localEndpointsImageName) {
				return true
			}
		}
	}
	return false
}

func pullLocalEndpointsImage(dockerClient imagePuller) {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	logrus.Infof("Pulling image %s", localEndpointsImageName)
	rc, err := dockerClient.ImagePull(ctx, localEndpointsImageName, types.ImagePullOptions{})
	if err != nil {
		logrus.Fatalf("Failed to pull image with err %v", err)
	}
	defer rc.Close()

	ioutil.ReadAll(rc)
	logrus.Infof("Pulled image %s", localEndpointsImageName)
}

func setupLocalEndpointsContainer(docker containerStarter) {
	containerID := createLocalEndpointsContainer(docker)
	startContainer(docker, containerID)
}

// createLocalEndpointsContainer returns the ID of the newly created container.
// If the container already exists, returns the ID of the existing container.
func createLocalEndpointsContainer(dockerClient containerStarter) string {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	// See First Scenario in https://aws.amazon.com/blogs/compute/a-guide-to-locally-testing-containers-with-amazon-ecs-local-endpoints-and-docker-compose/
	// for an explanation of these fields.
	resp, err := dockerClient.ContainerCreate(ctx,
		&container.Config{
			Image: localEndpointsImageName,
			Env: []string{
				"AWS_PROFILE=default",
				"HOME=/home",
			},
		},
		&container.HostConfig{
			Binds: []string{
				"/var/run:/var/run",
				fmt.Sprintf("%s/.aws/:/home/.aws/", os.Getenv("HOME")),
			},
		},
		&network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				EcsLocalNetworkName: {
					NetworkID: EcsLocalNetworkName,
					IPAMConfig: &network.EndpointIPAMConfig{
						IPv4Address: localEndpointsContainerIpAddr,
					},
				},
			},
		},
		localEndpointsContainerName,
	)
	if err != nil {
		if strings.Contains(err.Error(), "Conflict") {
			// We already created this container before, fetch its ID and return it.
			containerID := localEndpointsContainerID(dockerClient)
			logrus.Infof("The %s container already exists with ID %s", localEndpointsContainerName, containerID)
			return containerID
		}
		logrus.Fatalf("Failed to create container %s due to %v", localEndpointsContainerName, err)
	}

	logrus.Infof("Created the %s container with ID %s", localEndpointsContainerName, resp.ID)
	return resp.ID
}

func localEndpointsContainerID(dockerClient containerStarter) string {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	resp, err := dockerClient.ContainerList(ctx, types.ContainerListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("name", localEndpointsContainerName),
		),
	})
	if err != nil {
		logrus.Fatalf("Failed to list containers with name %s due to %v", localEndpointsContainerName, err)
	}
	if len(resp) != 1 {
		logrus.Fatalf("Expected to find one container named %s but found %d", localEndpointsContainerName, len(resp))
	}
	return resp[0].ID
}

func startContainer(dockerClient containerStarter, containerID string) {
	ctx, cancel := context.WithTimeout(context.Background(), docker.TimeoutInS)
	defer cancel()

	// If the container is already running, Docker does not return an error response.
	if err := dockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); err != nil {
		logrus.Fatalf("Failed to start container with ID %s due to %v", containerID, err)
	}
	logrus.Infof("Started container with ID %s", containerID)
}
