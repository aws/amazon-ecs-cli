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
	regionToId := map[string]string{
		// amzn-ami-2017.09.k-amazon-ecs-optimized AMIs
		"ap-northeast-1": "ami-5add893c",
		"ap-northeast-2": "ami-ba74d8d4",
		"ap-south-1":     "ami-2149114e",
		"ap-southeast-1": "ami-acbcefd0",
		"ap-southeast-2": "ami-4cc5072e",
		"ca-central-1":   "ami-a535b2c1",
		"cn-north-1":     "ami-dc934cb1",
		"eu-central-1":   "ami-ac055447",
		"eu-west-1":      "ami-bfb5fec6",
		"eu-west-2":      "ami-a48d6bc3",
		"eu-west-3":      "ami-914afcec",
		"sa-east-1":      "ami-d3bce9bf",
		"us-east-1":      "ami-cb17d8b6",
		"us-east-2":      "ami-1b90a67e",
		"us-gov-west-1":  "ami-6546cc04",
		"us-west-1":      "ami-9cbbaffc",
		"us-west-2":      "ami-05b5277d",
	}

	return &staticAmiIds{regionToId: regionToId}
}

func (c *staticAmiIds) Get(region string) (string, error) {
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}
