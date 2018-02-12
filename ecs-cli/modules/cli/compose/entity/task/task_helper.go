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

package task

import (
	log "github.com/sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/waiters"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
)

// WaitForTasks continuously polls ECS (by calling descibeTasks) and waits for tasks status to match desired
func waitForTasks(task *Task, taskArns map[string]bool) error {
	timeoutMessage := "Timeout waiting for ECS running task count to match desired task count."

	return waiters.TaskWaitUntilTimeout(func(retryCount int) (bool, error) {
		if len(taskArns) == 0 {
			return true, nil
		}

		// describe tasks
		taskArnsSlice := entity.ConvertMapToSlice(taskArns)
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
	}, task, timeoutMessage)
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
				"taskDefinition": entity.GetIdFromArn(ecsTask.TaskDefinitionArn),
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
