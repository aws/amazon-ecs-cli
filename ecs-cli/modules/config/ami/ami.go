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
	// amzn-ami-2017.09.f-amazon-ecs-optimized AMIs
	regionToId["ap-northeast-1"] = "ami-72f36a14"
	regionToId["ap-northeast-2"] = "ami-59b71737"
	regionToId["ap-south-1"] = "ami-f4db8f9b"
	regionToId["ap-southeast-1"] = "ami-e782f29b"
	regionToId["ap-southeast-2"] = "ami-7aa15c18"
	regionToId["ca-central-1"] = "ami-9afc79fe"
	regionToId["cn-north-1"] = "ami-e4c81589"
	regionToId["eu-central-1"] = "ami-eacf5d85"
	regionToId["eu-west-1"] = "ami-acb020d5"
	regionToId["eu-west-2"] = "ami-4d809829"
	regionToId["eu-west-3"] = "ami-5e02b523"
	regionToId["sa-east-1"] = "ami-49256725"
	regionToId["us-east-1"] = "ami-ba722dc0"
	regionToId["us-east-2"] = "ami-13af8476"
	regionToId["us-west-1"] = "ami-9df0f0fd"
	regionToId["us-west-2"] = "ami-c9c87cb1"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
