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
// Source: github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/iam (interfaces: Client)

// Package mock_iam is a generated GoMock package.
package mock_iam

import (
	reflect "reflect"

	iam "github.com/aws/aws-sdk-go/service/iam"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// AttachRolePolicy mocks base method
func (m *MockClient) AttachRolePolicy(arg0, arg1 string) (*iam.AttachRolePolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AttachRolePolicy", arg0, arg1)
	ret0, _ := ret[0].(*iam.AttachRolePolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AttachRolePolicy indicates an expected call of AttachRolePolicy
func (mr *MockClientMockRecorder) AttachRolePolicy(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AttachRolePolicy", reflect.TypeOf((*MockClient)(nil).AttachRolePolicy), arg0, arg1)
}

// CreateOrFindRole mocks base method
func (m *MockClient) CreateOrFindRole(arg0, arg1, arg2 string, arg3 []*iam.Tag) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrFindRole", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrFindRole indicates an expected call of CreateOrFindRole
func (mr *MockClientMockRecorder) CreateOrFindRole(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrFindRole", reflect.TypeOf((*MockClient)(nil).CreateOrFindRole), arg0, arg1, arg2, arg3)
}

// CreatePolicy mocks base method
func (m *MockClient) CreatePolicy(arg0 iam.CreatePolicyInput) (*iam.CreatePolicyOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePolicy", arg0)
	ret0, _ := ret[0].(*iam.CreatePolicyOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePolicy indicates an expected call of CreatePolicy
func (mr *MockClientMockRecorder) CreatePolicy(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePolicy", reflect.TypeOf((*MockClient)(nil).CreatePolicy), arg0)
}

// CreateRole mocks base method
func (m *MockClient) CreateRole(arg0 iam.CreateRoleInput) (*iam.CreateRoleOutput, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRole", arg0)
	ret0, _ := ret[0].(*iam.CreateRoleOutput)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateRole indicates an expected call of CreateRole
func (mr *MockClientMockRecorder) CreateRole(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRole", reflect.TypeOf((*MockClient)(nil).CreateRole), arg0)
}
