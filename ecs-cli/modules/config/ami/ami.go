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
	// amzn-ami-2015.09.f-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-43043329"
	regionToId["us-west-1"] = "ami-a77b0ac7"
	regionToId["us-west-2"] = "ami-02a24162"
	regionToId["eu-west-1"] = "ami-76e95b05"
	regionToId["eu-central-1"] = "ami-96b6adfa"
	regionToId["ap-northeast-1"] = "ami-18d8de76"
	regionToId["ap-southeast-1"] = "ami-9f60aefc"
	regionToId["ap-southeast-2"] = "ami-75a38416"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
