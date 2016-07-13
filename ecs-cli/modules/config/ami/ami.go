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
	// amzn-ami-2016.03.e-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-55870742"
	regionToId["us-west-1"] = "ami-07713767"
	regionToId["us-west-2"] = "ami-241bd844"
	regionToId["eu-west-1"] = "ami-c74127b4"
	regionToId["eu-central-1"] = "ami-3b54be54"
	regionToId["ap-northeast-1"] = "ami-2b08f44a"
	regionToId["ap-southeast-1"] = "ami-6b61bc08"
	regionToId["ap-southeast-2"] = "ami-d5b59eb6"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
