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
	// amzn-ami-2017.09.e-amazon-ecs-optimized AMIs
	regionToId["ap-northeast-1"] = "ami-af46dbc9"
	regionToId["ap-northeast-2"] = "ami-d6f454b8"
	regionToId["ap-south-1"] = "ami-c80b5fa7"
	regionToId["ap-southeast-1"] = "ami-fec3b482"
	regionToId["ap-southeast-2"] = "ami-b88e7cda"
	regionToId["ca-central-1"] = "ami-e8cb4e8c"
	regionToId["cn-north-1"] = "ami-f9a37e94"
	regionToId["eu-central-1"] = "ami-b378e8dc"
	regionToId["eu-west-1"] = "ami-7827b301"
	regionToId["eu-west-2"] = "ami-acd5cdc8"
	regionToId["eu-west-3"] = "ami-bd10a7c0"
	regionToId["sa-east-1"] = "ami-ca95d6a6"
	regionToId["us-east-1"] = "ami-13401669"
	regionToId["us-east-2"] = "ami-901338f5"
	regionToId["us-west-1"] = "ami-b3adacd3"
	regionToId["us-west-2"] = "ami-9a02a9e2"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
