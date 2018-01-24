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
		// amzn-ami-2017.09.g-amazon-ecs-optimized AMIs
		"ap-northeast-1": "ami-872c4ae1",
		"ap-northeast-2": "ami-c212b2ac",
		"ap-south-1":     "ami-00491f6f",
		"ap-southeast-1": "ami-910d72ed",
		"ap-southeast-2": "ami-58bb443a",
		"ca-central-1":   "ami-435bde27",
		"cn-north-1":     "ami-1300dd7e",
		"eu-central-1":   "ami-509a053f",
		"eu-west-1":      "ami-1d46df64",
		"eu-west-2":      "ami-67cbd003",
		"eu-west-3":      "ami-9aef59e7",
		"sa-east-1":      "ami-af521fc3",
		"us-east-1":      "ami-28456852",
		"us-east-2":      "ami-ce1c36ab",
		"us-west-1":      "ami-74262414",
		"us-west-2":      "ami-decc7fa6",
		"us-gov-west-1":  "ami-0a69e06b",
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
