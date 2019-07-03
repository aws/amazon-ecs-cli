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

package converter

import (
	composeV3 "github.com/docker/cli/cli/compose/types"
	"gopkg.in/yaml.v2"
)

// MarshalComposeConfig serializes a Docker Compose object into a YAML document.
func MarshalComposeConfig(conf composeV3.Config, filename string) ([]byte, error) {
	conf.Filename = filename
	return yaml.Marshal(conf)
}
