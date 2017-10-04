// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//      http://aws.amazon.com/apache2.0/
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
	// amzn-ami-2017.03.g-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-ec33cc96"
	regionToId["us-east-2"] = "ami-34032e51"
	regionToId["us-west-1"] = "ami-d5d0e0b5"
	regionToId["us-west-2"] = "ami-29f80351"
	regionToId["ca-central-1"] = "ami-9b54edff"
	regionToId["cn-north-1"] = "ami-dba87bb6"
	regionToId["eu-central-1"] = "ami-40d5672f"
	regionToId["eu-west-1"] = "ami-13f7226a"
	regionToId["eu-west-2"] = "ami-eb62708f"
	regionToId["ap-northeast-1"] = "ami-21815747"
	regionToId["ap-northeast-2"] = "ami-7ee13b10"
	regionToId["ap-southeast-1"] = "ami-99f588fa"
	regionToId["ap-southeast-2"] = "ami-4f08e82d"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
