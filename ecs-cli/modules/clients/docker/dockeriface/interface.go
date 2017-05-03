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

// Package dockeriface contains an interface for go-dockerclient
package dockeriface

import "github.com/fsouza/go-dockerclient"

// DockerAPI is an interface specifying the subset of
// github.com/fsouza/go-dockerclient.Client
type DockerAPI interface {
	PushImage(opts docker.PushImageOptions, auth docker.AuthConfiguration) error
	PullImage(opts docker.PullImageOptions, auth docker.AuthConfiguration) error
	TagImage(name string, opts docker.TagImageOptions) error
}
