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
	regionToId["us-east-1"] = "ami-71ef560b"
	regionToId["us-east-2"] = "ami-1b8ca37e"
	regionToId["us-west-1"] = "ami-e5cdf385"
	regionToId["us-west-2"] = "ami-a64d9ade"
	regionToId["ca-central-1"] = "ami-c802baac"
	regionToId["cn-north-1"] = "ami-8c68bbe1"
	regionToId["eu-central-1"] = "ami-4255d32d"
	regionToId["eu-west-1"] = "ami-014ae578"
	regionToId["eu-west-2"] = "ami-4f8d912b"
	regionToId["ap-northeast-1"] = "ami-3405af52"
	regionToId["ap-northeast-2"] = "ami-502c883e"
	regionToId["ap-southeast-1"] = "ami-134e0670"
	regionToId["ap-southeast-2"] = "ami-2ab95148"

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
