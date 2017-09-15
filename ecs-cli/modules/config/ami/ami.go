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

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	mainBucketS3URL     = "https://s3.amazonaws.com/ecs-ami-id/"
	cnNorth1BucketS3URL = "https://s3.cn-north-1.amazonaws.com.cn/ecs-ami-id/"
	// time out in milliseconds to wait to pull the bucket
	requestTimeout = 1000

	maxRetries = 2
)

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
	// amzn-ami-2017.03.f-amazon-ecs-optimized AMIs
	regionToId["us-east-1"] = "ami-9eb4b1e5"
	regionToId["us-east-2"] = "ami-1c002379"
	regionToId["us-west-1"] = "ami-4a2c192a"
	regionToId["us-west-2"] = "ami-1d668865"
	regionToId["ca-central-1"] = "ami-b677c9d2"
	regionToId["cn-north-1"] = "ami-28e63645"
	regionToId["eu-central-1"] = "ami-0460cb6b"
	regionToId["eu-west-1"] = "ami-8fcc32f6"
	regionToId["eu-west-2"] = "ami-cb1101af"
	regionToId["ap-northeast-1"] = "ami-b743bed1"
	regionToId["ap-southeast-1"] = "ami-9d1f7efe"
	regionToId["ap-southeast-2"] = "ami-c1a6bda2"

	return &staticAmiIds{regionToId: regionToId}
}

// Get returns the AMI ID for a given region
// It returns the AMI ID that it finds in:
// 	1. The S3 bucket
// 	2. The cn-north-1 mirror of the main S3 bucket
// 	3. The hardcoded list (which may be slightly out of date. )
// (#2 helps users in china who may be blocked from the main bucket)
func (c *staticAmiIds) Get(region string) (string, error) {
	// first try to pull from the main bucket
	if id, err := getIDFromS3(mainBucketS3URL+region, maxRetries); err == nil {
		return id, nil
	}

	// if we could not reach the main bucket, try the bucket in China
	if id, err := getIDFromS3(mainBucketS3URL+region, maxRetries); err == nil {
		return id, nil
	}

	// Finally, if pulling from S3 failed, fall back to the hardcoded list of AMIs
	// The hardcoded list may of course be slightly out of date
	id, exists := c.regionToId[region]
	if !exists {
		return "", fmt.Errorf("Could not find ami id for region '%s'", region)
	}

	return id, nil
}

// getIDFromS3 attempts to http GET url for the specified number of times (retries).
func getIDFromS3(url string, retries int) (string, error) {
	timeout := time.Duration(requestTimeout * time.Millisecond)
	client := http.Client{
		Timeout: timeout,
	}

	var err error

	for i := 0; i < retries; i++ {
		var response *http.Response
		response, err = client.Get(url)
		if err != nil {
			continue
		}
		defer response.Body.Close()
		var body []byte
		if body, err = ioutil.ReadAll(response.Body); err == nil {
			return string(body), nil
		}

	}

	return "", err
}
