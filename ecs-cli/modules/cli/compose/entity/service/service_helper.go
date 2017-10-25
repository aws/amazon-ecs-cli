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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/waiters"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	// DefaultUpdateServiceTimeout is the time that the CLI will wait to check if the
	// count of running tasks is changing. If count has not changed then an error is thrown
	// after DefaultUpdateServiceTimeout minutes
	DefaultUpdateServiceTimeout = 5

	// latestEventWindow defines "now"- it ensures that we only print events
	// which were created since roughly when the user entered the command in their
	// terminal. Units = seconds.
	latestEventWindow = 2
)

// serviceEvents is a wrapper for []*ecs.ServiceEvent
// that allows us to sort it by the timestamp
type serviceEvents []*ecs.ServiceEvent

func (s serviceEvents) Len() int {
	return len(s)
}
func (s serviceEvents) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s serviceEvents) Less(i, j int) bool {
	time1 := *s[i].CreatedAt
	time2 := *s[j].CreatedAt
	return time1.Before(time2)
}

// logNewServiceEvents logs events that have not been logged yet
func logNewServiceEvents(loggedEvents map[string]bool, events []*ecs.ServiceEvent, actionInvokedAt time.Time) {

	// sort the events so that newer ones are printed last
	sort.Sort(serviceEvents(events))
	for _, event := range events {
		if _, ok := loggedEvents[*event.Id]; !ok {
			// New event that has not been logged yet
			loggedEvents[*event.Id] = true
			if actionInvokedAt.Sub(*event.CreatedAt).Seconds() < latestEventWindow {
				log.WithFields(log.Fields{
					"timestamp": *event.CreatedAt},
				).Info(aws.StringValue(event.Message))
			}
		}
	}

}

// waitForServiceTasks continuously polls ECS (by calling describeService) and waits for service to get stable
// with desiredCount == runningCount
func waitForServiceTasks(service *Service, ecsServiceName string) error {
	eventsLogged := make(map[string]bool)
	var lastRunningCount int64
	lastRunningCountChangedAt := time.Now()
	timeOut := float64(DefaultUpdateServiceTimeout)
	actionInvokedAt := time.Now()

	if val := service.Context().CLIContext.Float64(flags.ComposeServiceTimeOutFlag); val > 0 {
		timeOut = val
	} else if val < 0 {
		return fmt.Errorf("Error with timeout flag: %f is not a valid timeout value", val)
	} else {
		log.Warn("Timeout was specified as zero. Your service deployment may not have completed yet.")
		return nil
	}

	return waiters.ServiceWaitUntilComplete(func(retryCount int) (bool, error) {
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

		// Log if running count has changed
		if runningCount != lastRunningCount {
			lastRunningCount = runningCount
			lastRunningCountChangedAt = time.Now()
			log.WithFields(logFields).Info("Service status")
		}

		// log new service events
		if len(ecsService.Events) > 0 {
			logNewServiceEvents(eventsLogged, ecsService.Events, actionInvokedAt)
		}

		// The deployment was successful
		if len(ecsService.Deployments) == 1 && desiredCount == runningCount {
			log.WithFields(logFields).Info("ECS Service has reached a stable state")
			return true, nil
		}

		if time.Since(lastRunningCountChangedAt).Minutes() > timeOut {
			return false, fmt.Errorf("Deployment has not completed: Running count has not changed for %.2f minutes", timeOut)
		}

		return false, nil

	}, service)
}
