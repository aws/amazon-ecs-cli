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
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/local/network/mock_network"
	"github.com/docker/docker/api/types"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

type mockStopperCalls func(mock *mock_network.MockLocalEndpointsStopper) *mock_network.MockLocalEndpointsStopper

func TestTeardown(t *testing.T) {
	tests := map[string]struct {
		configureCalls mockStopperCalls
	}{
		"network with no containers": {
			configureCalls: func(mock *mock_network.MockLocalEndpointsStopper) *mock_network.MockLocalEndpointsStopper {
				gomock.InOrder(
					mock.EXPECT().NetworkInspect(gomock.Any(), EcsLocalNetworkName, gomock.Any()).
						Return(types.NetworkResource{Containers: make(map[string]types.EndpointResource)}, nil),
					mock.EXPECT().ContainerStop(gomock.Any(), localEndpointsContainerName, nil).
						Return(errors.New(fmt.Sprintf("No such container: %s", localEndpointsContainerName))),
					mock.EXPECT().ContainerRemove(gomock.Any(), localEndpointsContainerName, gomock.Any()).
						Return(errors.New(fmt.Sprintf("No such container: %s", localEndpointsContainerName))),
					mock.EXPECT().NetworkRemove(gomock.Any(), EcsLocalNetworkName),
				)
				return mock
			},
		},
		"network with only local endpoints": {
			configureCalls: func(mock *mock_network.MockLocalEndpointsStopper) *mock_network.MockLocalEndpointsStopper {
				gomock.InOrder(
					mock.EXPECT().NetworkInspect(gomock.Any(), gomock.Eq(EcsLocalNetworkName), gomock.Any()).Return(
						types.NetworkResource{
							Containers: map[string]types.EndpointResource{
								localEndpointsContainerName: {
									Name: localEndpointsContainerName,
								},
							},
						}, nil),
					mock.EXPECT().ContainerStop(gomock.Any(), localEndpointsContainerName, nil),
					mock.EXPECT().ContainerRemove(gomock.Any(), localEndpointsContainerName, gomock.Any()),
					mock.EXPECT().NetworkRemove(gomock.Any(), EcsLocalNetworkName),
				)
				return mock
			},
		},
		"network with no running local endpoints": {
			configureCalls: func(mock *mock_network.MockLocalEndpointsStopper) *mock_network.MockLocalEndpointsStopper {
				gomock.InOrder(
					mock.EXPECT().NetworkInspect(gomock.Any(), gomock.Eq(EcsLocalNetworkName), gomock.Any()).Return(
						types.NetworkResource{
							Containers: map[string]types.EndpointResource{
								"some_container": {
									Name: "some_container",
								},
							},
						}, nil),

					// Should not be invoked
					mock.EXPECT().ContainerStop(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					mock.EXPECT().ContainerRemove(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					mock.EXPECT().NetworkRemove(gomock.Any(), gomock.Any()).Times(0),
				)
				return mock
			},
		},
		"network with running tasks": {
			configureCalls: func(mock *mock_network.MockLocalEndpointsStopper) *mock_network.MockLocalEndpointsStopper {
				gomock.InOrder(
					mock.EXPECT().NetworkInspect(gomock.Any(), gomock.Eq(EcsLocalNetworkName), gomock.Any()).Return(
						types.NetworkResource{
							Containers: map[string]types.EndpointResource{
								"some_container": {
									Name: "some_container",
								},
								localEndpointsContainerName: {
									Name: localEndpointsContainerName,
								},
							},
						}, nil),
					// Should not be invoked
					mock.EXPECT().ContainerStop(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					mock.EXPECT().ContainerRemove(gomock.Any(), gomock.Any(), gomock.Any()).Times(0),
					mock.EXPECT().NetworkRemove(gomock.Any(), gomock.Any()).Times(0),
				)
				return mock
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDocker := mock_network.NewMockLocalEndpointsStopper(ctrl)
			mockDocker = tc.configureCalls(mockDocker)

			Teardown(mockDocker)
		})
	}
}
