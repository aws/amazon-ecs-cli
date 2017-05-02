// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package service

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/waiters"
	"github.com/aws/aws-sdk-go/aws"
)

// waitForServiceTasks continuously polls ECS (by calling describeService) and waits for service to get stable
// with desiredCount == runningCount
func waitForServiceTasks(service *Service, ecsServiceName string) error {
	timeoutMessage := fmt.Sprintf("Timeout waiting for service %s to get stable", ecsServiceName)

	return waiters.WaitUntilComplete(func(retryCount int) (bool, error) {

		ecsService, err := service.describeService()
		if err != nil {
			return false, err
		}

		desiredCount := aws.Int64Value(ecsService.DesiredCount)
		runningCount := aws.Int64Value(ecsService.RunningCount)

		logFields := log.Fields{
			"serviceName":  ecsServiceName,
			"desiredCount": desiredCount,
			"runningCount": runningCount,
		}
		if len(ecsService.Deployments) == 1 && desiredCount == runningCount {
			log.WithFields(logFields).Info("ECS Service has reached a stable state")
			return true, nil
		}

		if retryCount%2 == 0 {
			log.WithFields(logFields).Info("Describe ECS Service status")
		} else {
			log.WithFields(logFields).Debug("Describe ECS Service status")
		}
		return false, nil
	}, service, timeoutMessage, true)
}
