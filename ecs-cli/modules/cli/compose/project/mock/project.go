// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/project (interfaces: Project)

package mock_project

import (
	adapter "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	context "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	entity "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	config "github.com/docker/libcompose/config"
	project "github.com/docker/libcompose/project"
	gomock "github.com/golang/mock/gomock"
)

// Mock of Project interface
type MockProject struct {
	ctrl     *gomock.Controller
	recorder *_MockProjectRecorder
}

// Recorder for MockProject (not exported)
type _MockProjectRecorder struct {
	mock *MockProject
}

func NewMockProject(ctrl *gomock.Controller) *MockProject {
	mock := &MockProject{ctrl: ctrl}
	mock.recorder = &_MockProjectRecorder{mock}
	return mock
}

func (_m *MockProject) EXPECT() *_MockProjectRecorder {
	return _m.recorder
}

func (_m *MockProject) ContainerConfigs() []adapter.ContainerConfig {
	ret := _m.ctrl.Call(_m, "ContainerConfigs")
	ret0, _ := ret[0].([]adapter.ContainerConfig)
	return ret0
}

func (_mr *_MockProjectRecorder) ContainerConfigs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ContainerConfigs")
}

func (_m *MockProject) Context() *context.ECSContext {
	ret := _m.ctrl.Call(_m, "Context")
	ret0, _ := ret[0].(*context.ECSContext)
	return ret0
}

func (_mr *_MockProjectRecorder) Context() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Context")
}

func (_m *MockProject) Create() error {
	ret := _m.ctrl.Call(_m, "Create")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Create() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Create")
}

func (_m *MockProject) Down() error {
	ret := _m.ctrl.Call(_m, "Down")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Down() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Down")
}

func (_m *MockProject) Entity() entity.ProjectEntity {
	ret := _m.ctrl.Call(_m, "Entity")
	ret0, _ := ret[0].(entity.ProjectEntity)
	return ret0
}

func (_mr *_MockProjectRecorder) Entity() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Entity")
}

func (_m *MockProject) Info() (project.InfoSet, error) {
	ret := _m.ctrl.Call(_m, "Info")
	ret0, _ := ret[0].(project.InfoSet)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (_mr *_MockProjectRecorder) Info() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Info")
}

func (_m *MockProject) Name() string {
	ret := _m.ctrl.Call(_m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockProjectRecorder) Name() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Name")
}

func (_m *MockProject) Parse() error {
	ret := _m.ctrl.Call(_m, "Parse")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Parse() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Parse")
}

func (_m *MockProject) Run(_param0 map[string][]string) error {
	ret := _m.ctrl.Call(_m, "Run", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Run(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Run", arg0)
}

func (_m *MockProject) Scale(_param0 int) error {
	ret := _m.ctrl.Call(_m, "Scale", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Scale(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Scale", arg0)
}

func (_m *MockProject) ServiceConfigs() *config.ServiceConfigs {
	ret := _m.ctrl.Call(_m, "ServiceConfigs")
	ret0, _ := ret[0].(*config.ServiceConfigs)
	return ret0
}

func (_mr *_MockProjectRecorder) ServiceConfigs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "ServiceConfigs")
}

func (_m *MockProject) Start() error {
	ret := _m.ctrl.Call(_m, "Start")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Start() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Start")
}

func (_m *MockProject) Stop() error {
	ret := _m.ctrl.Call(_m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Stop() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Stop")
}

func (_m *MockProject) Up() error {
	ret := _m.ctrl.Call(_m, "Up")
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockProjectRecorder) Up() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Up")
}

func (_m *MockProject) VolumeConfigs() *adapter.Volumes {
	ret := _m.ctrl.Call(_m, "VolumeConfigs")
	ret0, _ := ret[0].(*adapter.Volumes)
	return ret0
}

func (_mr *_MockProjectRecorder) VolumeConfigs() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "VolumeConfigs")
}
