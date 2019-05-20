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
	"testing"

	"github.com/pkg/errors"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"

	"github.com/golang/mock/gomock"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network/mock_network"
)

type mockStarterCalls func(docker *mock_network.MockLocalEndpointsStarter) *mock_network.MockLocalEndpointsStarter

type notFoundErr struct{}

func (e notFoundErr) Error() string {
	return "Dummy not found error returned from Docker"
}

func (e notFoundErr) NotFound() bool {
	return true
}

func TestSetup(t *testing.T) {
	// We don't check the detailed configurations for the Local Network or the Local Container Endpoints.
	// The validation of whether those fields behave as expected should be captured in integration tests.
	// See https://github.com/aws/amazon-ecs-cli/issues/772
	tests := map[string]struct {
		configureCalls mockStarterCalls
	}{
		"new network and new container": {
			configureCalls: func(docker *mock_network.MockLocalEndpointsStarter) *mock_network.MockLocalEndpointsStarter {
				containerID := "1234"
				gomock.InOrder(
					// We expect to create the network if it doesn't exist already.
					docker.EXPECT().NetworkInspect(gomock.Any(), EcsLocalNetworkName, gomock.Any()).Return(types.NetworkResource{}, notFoundErr{}),
					docker.EXPECT().NetworkCreate(gomock.Any(), EcsLocalNetworkName, gomock.Any()).Return(types.NetworkCreateResponse{}, nil),

					// Don't fetch the ID of the Local Container endpoints if it's the first time we are creating it.
					docker.EXPECT().
						ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), localEndpointsContainerName).
						Return(container.ContainerCreateCreatedBody{
							ID: containerID,
						}, nil),
					docker.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Times(0),
					docker.EXPECT().ContainerStart(gomock.Any(), containerID, gomock.Any()),
				)
				return docker
			},
		},
		"existing network and existing container": {
			configureCalls: func(docker *mock_network.MockLocalEndpointsStarter) *mock_network.MockLocalEndpointsStarter {
				networkID := "abcd"
				containerID := "1234"
				gomock.InOrder(
					// Don't create the network if one already exists
					docker.EXPECT().NetworkInspect(gomock.Any(), EcsLocalNetworkName, gomock.Any()).
						Return(types.NetworkResource{ID: networkID}, nil),
					docker.EXPECT().NetworkCreate(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),

					// Retrieve the container ID if it already exists
					docker.EXPECT().
						ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), localEndpointsContainerName).
						Return(container.ContainerCreateCreatedBody{}, errors.New("Conflict. The container name is already in use.")),
					docker.EXPECT().ContainerList(gomock.Any(), gomock.Any()).Return([]types.Container{
						{
							ID: containerID,
						},
					}, nil),
					docker.EXPECT().ContainerStart(gomock.Any(), containerID, gomock.Any()),
				)
				return docker
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDocker := mock_network.NewMockLocalEndpointsStarter(ctrl)
			mockDocker = tc.configureCalls(mockDocker)

			Setup(mockDocker)
		})
	}
}
