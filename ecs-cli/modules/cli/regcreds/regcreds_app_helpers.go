// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package regcreds

import (
	"fmt"
	"regexp"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
)

// returns the provided value with the ecs-cli resource prefix added
func generateECSResourceName(providedName string) *string {
	return aws.String(utils.ECSCLIResourcePrefix + providedName)
}

func generateSecretDescription(regName string) *string {
	return aws.String(fmt.Sprintf("Created with the ECS CLI for use with registry %s", regName))
}

func generateSecretString(username, password string) *string {
	return aws.String(`{"username":"` + username + `","password":"` + password + `"}`)
}

func getExecutionRolePolicyARN(region string) string {
	regionToPartition := map[string]string{
		"cn-north-1":     "aws-cn",
		"cn-northwest-1": "aws-cn",
		"us-gov-west-1":  "aws-us-gov",
	}

	expectedARN := arn.ARN{
		Service:   "iam",
		Resource:  "policy/service-role/AmazonECSTaskExecutionRolePolicy",
		AccountID: "aws",
	}

	if regionToPartition[region] != "" {
		expectedARN.Partition = regionToPartition[region]
	}

	expectedARN.Partition = "aws"

	return expectedARN.String()
}

func isARN(value string) bool {
	matches, _ := regexp.MatchString("arn:*:*:*:*:*:*", value)
	return matches
}
