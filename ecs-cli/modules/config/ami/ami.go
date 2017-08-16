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
	// amzn-ami-2017.03.e-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-d61027ad"
	regionToId["us-east-2"] = "ami-bb8eaede"
	regionToId["us-west-1"] = "ami-514e6431"
	regionToId["us-west-2"] = "ami-c6f81abe"
	regionToId["ca-central-1"] = "ami-32bb0556"
	regionToId["cn-north-1"] = "ami-49d80824"
	regionToId["eu-central-1"] = "ami-f15ff69e"
	regionToId["eu-west-1"] = "ami-bd7e8dc4"
	regionToId["eu-west-2"] = "ami-0a85946e"
	regionToId["ap-northeast-1"] = "ami-ab5ea9cd"
	regionToId["ap-southeast-1"] = "ami-ae0b91cd"
	regionToId["ap-southeast-2"] = "ami-c3233ba0"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
