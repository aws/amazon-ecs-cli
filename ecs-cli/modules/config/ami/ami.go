// Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package ami

import "fmt"

// ECSAmiIds interface is used to get the ami id for a specified region.
type ECSAmiIds interface {
	Get(string) (string, error)
}

// staticAmiIds impmenets the ECSAmiIds interface to get the AMI id for
// a region using a hardcoded map of values.
type staticAmiIds struct {
	regionToId map[string]string
}

func NewStaticAmiIds() ECSAmiIds {
	regionToId := make(map[string]string)
	// amzn-ami-2015.09.e-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-cb2305a1"
	regionToId["us-west-1"] = "ami-bdafdbdd"
	regionToId["us-west-2"] = "ami-ec75908c"
	regionToId["eu-west-1"] = "ami-13f84d60"
	regionToId["eu-central-1"] = "ami-c3253caf"
	regionToId["ap-northeast-1"] = "ami-e9724c87"
	regionToId["ap-southeast-1"] = "ami-5f31fd3c"
	regionToId["ap-southeast-2"] = "ami-83af8ae0"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
