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

// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/docker (interfaces: Client)

// Package mock_docker is a generated GoMock package.
package mock_docker

import (
	reflect "reflect"

	go_dockerclient "github.com/fsouza/go-dockerclient"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// PullImage mocks base method.
func (m *MockClient) PullImage(arg0, arg1 string, arg2 go_dockerclient.AuthConfiguration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PullImage", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// PullImage indicates an expected call of PullImage.
func (mr *MockClientMockRecorder) PullImage(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PullImage", reflect.TypeOf((*MockClient)(nil).PullImage), arg0, arg1, arg2)
}

// PushImage mocks base method.
func (m *MockClient) PushImage(arg0, arg1, arg2 string, arg3 go_dockerclient.AuthConfiguration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PushImage", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(error)
	return ret0
}

// PushImage indicates an expected call of PushImage.
func (mr *MockClientMockRecorder) PushImage(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PushImage", reflect.TypeOf((*MockClient)(nil).PushImage), arg0, arg1, arg2, arg3)
}

// TagImage mocks base method.
func (m *MockClient) TagImage(arg0, arg1, arg2 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TagImage", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// TagImage indicates an expected call of TagImage.
func (mr *MockClientMockRecorder) TagImage(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TagImage", reflect.TypeOf((*MockClient)(nil).TagImage), arg0, arg1, arg2)
}
