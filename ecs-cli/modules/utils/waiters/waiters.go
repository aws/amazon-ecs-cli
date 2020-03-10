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

package waiters

import (
	"fmt"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
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
	servicesWaitDelay = 5 * time.Second
)

// taskWaiterAction defines an action performed on a task and returns a bool to
// stop the wait or error if something unexpected happens
type taskWaiterAction func(retryCount int) (bool, error)

// serviceWaiterAction defines an action performed on a service and returns a bool to
// stop the wait or error if something unexpected happens
type serviceWaiterAction func() (bool, error)

// ServiceWaitUntilComplete runs the action in a for true loop, until the action returns true or an error
func ServiceWaitUntilComplete(action serviceWaiterAction, entity entity.ProjectEntity) error {
	// run the loop until done or error thrown
	for {
		done, err := action()
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		entity.Sleeper().Sleep(servicesWaitDelay)
	}

	return nil
}

// WaitUntilTimeout executes the waiterAction for maxRetries number of times, waiting for delayWait time between execution
func TaskWaitUntilTimeout(action taskWaiterAction, entity entity.ProjectEntity, timeoutMessage string) error {
	for retryCount := 0; retryCount < tasksMaxRetries; retryCount++ {
		done, err := action(retryCount)
		if err != nil {
			return err
		}
		if done {
			return nil
		}
		entity.Sleeper().Sleep(tasksWaitDelay)
	}

	return fmt.Errorf(timeoutMessage)
}
