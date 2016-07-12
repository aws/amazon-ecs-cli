// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package clients

import (
	"fmt"
	"runtime"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/version"
	"github.com/aws/aws-sdk-go/aws/request"
)

const UserAgentHeader = "User-Agent"

// CustomUserAgentHandler returns a http request handler that sets a custom user agent to all aws requests
func CustomUserAgentHandler() request.NamedHandler {
	return request.NamedHandler{
		Name: "ECSCLIUserAgentHandler",
		Fn: func(r *request.Request) {
			currentAgent := r.HTTPRequest.Header.Get(UserAgentHeader)
			r.HTTPRequest.Header.Set(UserAgentHeader,
				fmt.Sprintf("%s/%s (%s) %s", version.AppName, version.Version, runtime.GOOS, currentAgent))
		},
	}
}
