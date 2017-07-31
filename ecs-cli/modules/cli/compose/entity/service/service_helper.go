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
	"sort"
	"time"

	log "github.com/Sirupsen/logrus"
	ecscli "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/waiters"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const TimeOutUpdateService = 1

// serviceEvents is a wrapper for []*ecs.ServiceEvent
// that allows us to reverse it
type reverser []*ecs.ServiceEvent

func (s reverser) Len() int {
	return len(s)
}
func (s reverser) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s reverser) Less(i, j int) bool {
	time1 := *s[i].CreatedAt
	time2 := *s[j].CreatedAt
	diff := time1.Sub(time2)
	return diff.Seconds() < 0
}

// logNewServiceEvents logs events that have not been logged yet
func logNewServiceEvents(loggedEvents map[string]bool, events []*ecs.ServiceEvent) {

	// the slice comes ordered so that newer events are first. Logically, we
	// want to print older events first- so we reverse it
	sort.Sort(reverser(events))
	for _, event := range events {
		if _, ok := loggedEvents[*event.Id]; !ok {
			// New event that has not been logged yet
			loggedEvents[*event.Id] = true
			log.Infof("Service Event %s", event.String())
		}
	}

}

// waitForServiceTasks continuously polls ECS (by calling describeService) and waits for service to get stable
// with desiredCount == runningCount
func waitForServiceTasks(service *Service, ecsServiceName string) error {
	timeoutMessage := fmt.Sprintf("Timeout waiting for service %s to get stable", ecsServiceName)

	eventsLogged := make(map[string]bool)
	var lastRunningCount int64
	lastRunningCountChangedAt := time.Now()
	timeOut := float64(TimeOutUpdateService)
	log.Warnf("Command in Waiter: %s", service.Context().CLIContext.Command.Name)

	if val := service.Context().CLIContext.Float64(ecscli.ComposeServiceTimeOutFlag); val > 0 {
		timeOut = val
	}

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

		// The deployment was successful
		if len(ecsService.Deployments) == 1 && desiredCount == runningCount {
			log.WithFields(logFields).Info("ECS Service has reached a stable state")
			return true, nil
		}

		// Log information only if things have changed
		// running count has changed
		if runningCount != lastRunningCount {
			lastRunningCount = runningCount
			lastRunningCountChangedAt = time.Now()
			log.WithFields(logFields).Info("Describe ECS Service status")
		}

		// log new service events
		if len(ecsService.Events) > 0 {
			logNewServiceEvents(eventsLogged, ecsService.Events)
		}

		if time.Since(lastRunningCountChangedAt).Minutes() > timeOut {
			return true, fmt.Errorf("Deployment has not completed: Running count has not changed for %.2f minutes", timeOut)
		}

		return false, nil
	}, service, timeoutMessage, true)
}
