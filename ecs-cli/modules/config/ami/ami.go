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
		// amzn-ami-2017.09.j-amazon-ecs-optimized AMIs
		"ap-northeast-1": "ami-bb5f13dd",
		"ap-northeast-2": "ami-3b19b455",
		"ap-south-1":     "ami-9e91cff1",
		"ap-southeast-1": "ami-f88ade84",
		"ap-southeast-2": "ami-a677b6c4",
		"ca-central-1":   "ami-db48cfbf",
		"cn-north-1":     "ami-ca508ca7",
		"eu-central-1":   "ami-3b7d1354",
		"eu-west-1":      "ami-64c4871d",
		"eu-west-2":      "ami-25f51242",
		"eu-west-3":      "ami-0356e07e",
		"sa-east-1":      "ami-da2c66b6",
		"us-east-1":      "ami-cad827b7",
		"us-east-2":      "ami-ef64528a",
		"us-west-1":      "ami-29b8b249",
		"us-west-2":      "ami-baa236c2",
		"us-gov-west-1":  "ami-cc3cb7ad",
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
