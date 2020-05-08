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
// Source: github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ec2 (interfaces: EC2Client)

// Package mock_ec2 is a generated GoMock package.
package mock_ec2

import (
	reflect "reflect"

	ec2 "github.com/aws/aws-sdk-go/service/ec2"
	gomock "github.com/golang/mock/gomock"
)

// MockEC2Client is a mock of EC2Client interface
type MockEC2Client struct {
	ctrl     *gomock.Controller
	recorder *MockEC2ClientMockRecorder
}

// MockEC2ClientMockRecorder is the mock recorder for MockEC2Client
type MockEC2ClientMockRecorder struct {
	mock *MockEC2Client
}

// NewMockEC2Client creates a new mock instance
func NewMockEC2Client(ctrl *gomock.Controller) *MockEC2Client {
	mock := &MockEC2Client{ctrl: ctrl}
	mock.recorder = &MockEC2ClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEC2Client) EXPECT() *MockEC2ClientMockRecorder {
	return m.recorder
}

// DescribeInstanceTypeOfferings mocks base method
func (m *MockEC2Client) DescribeInstanceTypeOfferings(arg0 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeInstanceTypeOfferings", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeInstanceTypeOfferings indicates an expected call of DescribeInstanceTypeOfferings
func (mr *MockEC2ClientMockRecorder) DescribeInstanceTypeOfferings(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeInstanceTypeOfferings", reflect.TypeOf((*MockEC2Client)(nil).DescribeInstanceTypeOfferings), arg0)
}

// DescribeInstances mocks base method
func (m *MockEC2Client) DescribeInstances(arg0 []*string) (map[string]*ec2.Instance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeInstances", arg0)
	ret0, _ := ret[0].(map[string]*ec2.Instance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeInstances indicates an expected call of DescribeInstances
func (mr *MockEC2ClientMockRecorder) DescribeInstances(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeInstances", reflect.TypeOf((*MockEC2Client)(nil).DescribeInstances), arg0)
}

// DescribeNetworkInterfaces mocks base method
func (m *MockEC2Client) DescribeNetworkInterfaces(arg0 []*string) ([]*ec2.NetworkInterface, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DescribeNetworkInterfaces", arg0)
	ret0, _ := ret[0].([]*ec2.NetworkInterface)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DescribeNetworkInterfaces indicates an expected call of DescribeNetworkInterfaces
func (mr *MockEC2ClientMockRecorder) DescribeNetworkInterfaces(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DescribeNetworkInterfaces", reflect.TypeOf((*MockEC2Client)(nil).DescribeNetworkInterfaces), arg0)
}
