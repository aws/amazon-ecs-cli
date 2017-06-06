// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	// amzn-ami-2017.03.b-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-0e297018"
	regionToId["us-east-2"] = "ami-43d0f626"
	regionToId["us-west-1"] = "ami-ac5849cf"
	regionToId["us-west-2"] = "ami-596d6520"
	regionToId["ca-central-1"] = "ami-8cfb44e8"
	regionToId["eu-central-1"] = "ami-25a4004a"
	regionToId["eu-west-1"] = "ami-5ae4f83c"
	regionToId["eu-west-2"] = "ami-ada6b1c9"
	regionToId["ap-northeast-1"] = "ami-3a000e5d"
	regionToId["ap-southeast-1"] = "ami-2428ab47"
	regionToId["ap-southeast-2"] = "ami-ac5849cf"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
