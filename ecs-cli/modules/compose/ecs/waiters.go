// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package ecs

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

const (
	// tasksWaitDelay is the delay between successive ECS DescribeTasks API calls
	// while determining if the task is running or stopped was created. This value
	// reflects the values set in the ecs waiters json file in the aws-go-sdk.
	tasksWaitDelay = 6 * time.Second

	// tasksMaxRetries is the maximum number of ECS DescribeTasks API will be invoked by the WaitUntilComplete method
	// to determine if the task is running or stopped before giving up. This value reflects the values set in the
	// ecs waiters json file in the aws-go-sdk.
	tasksMaxRetries = 100

	// servicesWaitDelay is the delay between successive ECS DescribeServices API calls
	// while determining if the service is stable or inactive. This value
	// reflects the values set in the ecs waiters json file in the aws-go-sdk.
	servicesWaitDelay = 15 * time.Second

	// servicesMaxRetries is the maximum number of ECS DescribeServices API will be invoked by the WaitUntilComplete method
	// to determine if the task is running or stopped before giving up. This value reflects the values set in the
	// ecs waiters json file in the aws-go-sdk.
	servicesMaxRetries = 40
)

// waiterAction defines an action performed on the project entity
// and returns a bool to stop the wait or error if something unexpected happens
type waiterAction func(retryCount int) (bool, error)

// waitForServiceTasks continuously polls ECS (by calling describeService) and waits for service to get stable
// with desiredCount == runningCount
func waitForServiceTasks(service *Service, ecsServiceName string) error {
	timeoutMessage := fmt.Sprintf("Timeout waiting for service %s to get stable", ecsServiceName)

	return waitUntilComplete(func(retryCount int) (bool, error) {

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

// waitForTasks continuously polls ECS (by calling descibeTasks) and waits for tasks status to match desired
func waitForTasks(task *Task, taskArns map[string]bool) error {
	timeoutMessage := "Timeout waiting for ECS tasks' status to match desired"

	return waitUntilComplete(func(retryCount int) (bool, error) {
		if len(taskArns) == 0 {
			return true, nil
		}

		// describe tasks
		taskArnsSlice := convertMapToSlice(taskArns)
		// TODO, limit to Describe 100 tasks at a time?
		ecsTasks, err := task.Context().ECSClient.DescribeTasks(taskArnsSlice)
		if err != nil {
			return false, err
		}

		// log tasks status
		checkECSTasksStatus(ecsTasks, taskArns, retryCount)

		if len(taskArns) == 0 {
			return true, nil
		}

		return false, nil
	}, task, timeoutMessage, false)
}

// checkECSTasksStatus iterates through the ecsTasks and checks if the desired status is same as last status
// and logs messages accordingly. If the statuses match, it removes from the poll-able list of task arns
func checkECSTasksStatus(ecsTasks []*ecs.Task, taskArns map[string]bool, retryCount int) map[string]bool {
	for _, ecsTask := range ecsTasks {
		desiredStatus := aws.StringValue(ecsTask.DesiredStatus)
		lastStatus := aws.StringValue(ecsTask.LastStatus)

		logMessage := "Describe ECS container status"
		logInfo := false

		// if task status is same as desired, then task has reached a stable state, no more describes needed
		// ELSE if task status is stopped, then task has reached a terminal state, no more describes needed
		if desiredStatus == lastStatus {
			// delete this current task from the next iteration describe
			delete(taskArns, aws.StringValue(ecsTask.TaskArn))
			logInfo = true
			if desiredStatus == ecs.DesiredStatusRunning {
				logMessage = "Started container..."
			} else if desiredStatus == ecs.DesiredStatusStopped {
				logMessage = "Stopped container..."
			}
		} else if lastStatus == ecs.DesiredStatusStopped {
			// delete this current task from the next iteration describe
			delete(taskArns, aws.StringValue(ecsTask.TaskArn))
			logInfo = true
			logMessage = "Container seem to have stopped..."
		}

		for _, container := range ecsTask.Containers {
			logFields := log.Fields{
				"taskDefinition": getIdFromArn(ecsTask.TaskDefinitionArn),
				"desiredStatus":  desiredStatus,
				"lastStatus":     lastStatus,
				"container":      getFormattedContainerName(ecsTask, container),
			}
			if logInfo || retryCount%2 == 0 {
				log.WithFields(logFields).Info(logMessage)
			} else {
				log.WithFields(logFields).Debug("Describe ECS container status")
			}

		}
	}
	return taskArns
}

// waitUntilComplete executes the waiterAction for maxRetries number of times, waiting for delayWait time between execution
func waitUntilComplete(action waiterAction, entity ProjectEntity, timeoutMessage string, isService bool) error {
	var delayWait time.Duration
	var maxRetries int
	if isService {
		delayWait = servicesWaitDelay
		maxRetries = servicesMaxRetries
	} else {
		delayWait = tasksWaitDelay
		maxRetries = tasksMaxRetries
	}

	for retryCount := 0; retryCount < maxRetries; retryCount++ {
		done, err := action(retryCount)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		entity.Sleeper().Sleep(delayWait)
	}

	return fmt.Errorf(timeoutMessage)
}
