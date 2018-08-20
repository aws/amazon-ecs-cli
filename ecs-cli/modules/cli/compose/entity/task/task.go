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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/types"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/cache"
	composeutils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
	log "github.com/sirupsen/logrus"
)

// Task type is placeholder for a single task definition and its cache
// and it performs compose operations at a task definition level
type Task struct {
	taskDef     *ecs.TaskDefinition
	cache       cache.Cache
	ecsContext  *context.ECSContext
	timeSleeper *utils.TimeSleeper
}

// NewTask creates an instance of a Task and also sets up a cache for task definition
func NewTask(context *context.ECSContext) entity.ProjectEntity {
	return &Task{
		cache:       entity.SetupTaskDefinitionCache(),
		ecsContext:  context,
		timeSleeper: &utils.TimeSleeper{},
	}
}

// LoadContext is a placeholder function to read the context set by NewTask. Its a NoOp for Task entity.
func (t *Task) LoadContext() error {
	// NoOp
	return nil
}

// SetTaskDefinition sets the ecs task definition to the current instance of Task
func (t *Task) SetTaskDefinition(taskDefinition *ecs.TaskDefinition) {
	t.taskDef = taskDefinition
}

// Context returs the context of this project
func (t *Task) Context() *context.ECSContext {
	return t.ecsContext
}

// Sleeper returs an instance of TimeSleeper used to wait until Tasks has either started running or stopped
func (t *Task) Sleeper() *utils.TimeSleeper {
	return t.timeSleeper
}

// TaskDefinition returns the task definition object that was created by
// transforming the Service Configs to ECS acceptable format
func (t *Task) TaskDefinition() *ecs.TaskDefinition {
	return t.taskDef
}

// TaskDefinitionCache returns the cache that should be used when checking for
// previous task definition
func (t *Task) TaskDefinitionCache() cache.Cache {
	return t.cache
}

// --- commands ---

// Create creates a task definition in ECS for the containers in the compose file
// and persists it in a cache locally. It always checks the cache before creating
func (t *Task) Create() error {
	_, err := entity.GetOrCreateTaskDefinition(t)
	if err != nil {
		return err
	}
	err = entity.OptionallyCreateLogs(t)
	return err
}

// Start starts the containers if they weren't already running.
func (t *Task) Start() error {
	return t.up(false)
}

// Up gets a list of running tasks and it updates it with the latest task definition
// if count of running tasks = 0, starts 1
// if count != 0, and the task definitions differed, then its stops the old ones and starts the new ones
func (t *Task) Up() error {
	updateTasks := t.Context().CLIContext.Bool(flags.ForceUpdateFlag)
	return t.up(updateTasks)
}

// Info returns a formatted list of containers (running and stopped) in the current cluster
// filtered by this project if filterLocal is set to true
func (t *Task) Info(filterLocal bool) (project.InfoSet, error) {
	return entity.Info(t, filterLocal)
}

// Scale finds out the current count of running tasks for this project and scales to the desired count.
// Any run params specified will be taken into account.
// if desired = current, noop
// if desired > current, stops the extra ones
// if desired < current, start new ones (also if current was 0, create a new task definition)
func (t *Task) Scale(desiredCount int) error {
	ecsTasks, err := entity.CollectTasksWithStatus(t, ecs.DesiredStatusRunning, true)
	if err != nil {
		return err
	}

	observedCount := len(ecsTasks)

	if desiredCount == observedCount {
		// NoOp
		log.WithFields(log.Fields{
			"countOfRunningTasks": observedCount,
		}).Info("Tasks are already running")
		// TODO, should we wait for PENDING -> RUNNING in this case?
		return nil
	}

	// running more than desired, stop the extra tasks
	if desiredCount < observedCount {
		diff := observedCount - desiredCount
		ecsTasksToStop := []*ecs.Task{}
		for i := 0; i < diff; i++ {
			ecsTasksToStop = append(ecsTasksToStop, ecsTasks[i])
		}
		return t.stopTasks(ecsTasksToStop)
	}

	// if desired > observed, then run the difference
	diff := desiredCount - observedCount

	var taskDef string
	// if nothing was running, create new task definition
	if observedCount == 0 {
		taskDefinition, err := entity.GetOrCreateTaskDefinition(t)
		if err != nil {
			return err
		}
		taskDef = aws.StringValue(taskDefinition.TaskDefinitionArn)
	} else {
		// Note: Picking the first task definition as a standard and scaling for that task definition
		taskDef = aws.StringValue(ecsTasks[0].TaskDefinitionArn)
	}

	newTasks, err := t.runTasks(taskDef, diff)
	if err != nil {
		return err
	}
	return t.waitForRunTasks(newTasks)
}

// Run starts all containers defined in the task definition once regardless of if they were started before
// It also overrides the commands for the specified containers
// TODO Account for other ContainerOverrides
func (t *Task) Run(commandOverrides map[string][]string) error {
	taskDef, err := entity.GetOrCreateTaskDefinition(t)
	if err != nil {
		return err
	}
	taskDefinitionId := aws.StringValue(taskDef.TaskDefinitionArn)
	count := 1

	runTaskInput, err := t.buildRunTaskInput(taskDefinitionId, count, commandOverrides)
	if err != nil {
		return err
	}

	ecsTasks, err := t.Context().ECSClient.RunTask(runTaskInput)
	if err != nil {
		return err
	}

	for _, failure := range ecsTasks.Failures {
		log.WithFields(log.Fields{
			"reason": aws.StringValue(failure.Reason),
		}).Info("Couldn't run containers")
	}

	return t.waitForRunTasks(ecsTasks.Tasks)
}

// Stop gets all the running tasks and issues ECS StopTask command to them
// and waits until they stop
func (t *Task) Stop() error {
	ecsTasks, err := entity.CollectTasksWithStatus(t, ecs.DesiredStatusRunning, true)
	if err != nil {
		return err
	}
	return t.stopTasks(ecsTasks)
}

// Down ECS doesn't permit removing the tasks. One can call stop, but the task is still describe-able for a while
// and then ECS deletes them
func (t *Task) Down() error {
	return composeutils.ErrUnsupported
}

// EntityType returns the type of the entity
func (t *Task) EntityType() types.Type {
	return types.Task
}

// ----------- Commands' helper functions --------

// waitForRunTasks waits for the containers to go to running state
func (t *Task) waitForRunTasks(ecsTasks []*ecs.Task) error {
	ecsTaskArns := make(map[string]bool)
	for _, ecsTask := range ecsTasks {
		ecsTaskArns[aws.StringValue(ecsTask.TaskArn)] = true
		for _, container := range ecsTask.Containers {
			log.WithFields(log.Fields{
				"container": getFormattedContainerName(ecsTask, container),
			}).Info("Starting container...")
		}
	}
	return waitForTasks(t, ecsTaskArns)
}

// stopTasks issues stop task requests to ECS Service and waits for them to stop
func (t *Task) stopTasks(ecsTasks []*ecs.Task) error {
	ecsTaskArns := make(map[string]bool)
	// TODO, parallelize
	for _, ecsTask := range ecsTasks {
		arn := aws.StringValue(ecsTask.TaskArn)
		ecsTaskArns[arn] = true
		err := t.Context().ECSClient.StopTask(arn)
		if err != nil {
			return err
		}
		for _, container := range ecsTask.Containers {
			log.WithFields(log.Fields{
				"container": getFormattedContainerName(ecsTask, container),
			}).Info("Stopping container...")
		}
	}
	return waitForTasks(t, ecsTaskArns)
}

// runTasks issues run task request to ECS Service in chunks of count=10
// it always takes into account the latest ECS params
func (t *Task) runTasks(taskDefinition string, totalCount int) ([]*ecs.Task, error) {
	result := []*ecs.Task{}
	chunkSize := 10 // can issue only up to 10 tasks in a RunTask Call

	for i := 0; i < totalCount; i += chunkSize {
		count := chunkSize
		if i+chunkSize > totalCount {
			count = totalCount - i
		}

		runTaskInput, err := t.buildRunTaskInput(taskDefinition, count, nil)
		if err != nil {
			return nil, err
		}

		ecsTasks, err := t.Context().ECSClient.RunTask(runTaskInput)
		if err != nil {
			return nil, err
		}

		for _, failure := range ecsTasks.Failures {
			log.WithFields(log.Fields{
				"reason": aws.StringValue(failure.Reason),
			}).Info("Couldn't run containers")
		}
		result = append(result, ecsTasks.Tasks...)
	}

	return result, nil
}

func convertToECSTaskOverride(overrides map[string][]string) (*ecs.TaskOverride, error) {
	if overrides == nil {
		return nil, nil
	}

	commandOverrides := []*ecs.ContainerOverride{}
	for cont, command := range overrides {
		contOverride := &ecs.ContainerOverride{
			Name:    aws.String(cont),
			Command: aws.StringSlice(command),
		}
		commandOverrides = append(commandOverrides, contOverride)
	}

	ecsOverrides := &ecs.TaskOverride{
		ContainerOverrides: commandOverrides,
	}

	return ecsOverrides, nil
}

// buildRunTaskInput will account for what is currently specified in ECS Params
func (t *Task) buildRunTaskInput(taskDefinition string, count int, overrides map[string][]string) (*ecs.RunTaskInput, error) {
	cluster := t.Context().CommandConfig.Cluster
	launchType := t.Context().CommandConfig.LaunchType
	group := entity.GetTaskGroup(t)

	ecsParams := t.ecsContext.ECSParams
	networkConfig, err := composeutils.ConvertToECSNetworkConfiguration(ecsParams)

	if err != nil {
		return nil, err
	}

	placementConstraints, err := composeutils.ConvertToECSPlacementConstraints(ecsParams)
	if err != nil {
		return nil, err
	}

	placementStrategy, err := composeutils.ConvertToECSPlacementStrategy(ecsParams)
	if err != nil {
		return nil, err
	}

	// NOTE: this validation is not useful if called after RegisterTaskDefinition
	if err := entity.ValidateFargateParams(ecsParams, launchType); err != nil {
		return nil, err
	}

	taskOverride, err := convertToECSTaskOverride(overrides)
	if err != nil {
		return nil, err
	}

	runTaskInput := &ecs.RunTaskInput{
		Cluster:        aws.String(cluster),
		TaskDefinition: aws.String(taskDefinition),
		Group:          aws.String(group),
		Count:          aws.Int64(int64(count)),
	}

	if networkConfig != nil {
		runTaskInput.NetworkConfiguration = networkConfig
	}

	if taskOverride != nil {
		runTaskInput.Overrides = taskOverride
	}

	if placementConstraints != nil {
		runTaskInput.PlacementConstraints = placementConstraints
	}

	if placementStrategy != nil {
		runTaskInput.PlacementStrategy = placementStrategy
	}

	if launchType != "" {
		runTaskInput.LaunchType = aws.String(launchType)
	}

	return runTaskInput, nil
}

// createOne issues run task with count=1 and waits for it to get to running state
func (t *Task) createOne() error {
	ecsTask, err := t.runTasks(aws.StringValue(t.TaskDefinition().TaskDefinitionArn), 1)
	if err != nil {
		return err
	}
	return t.waitForRunTasks(ecsTask)
}

// up gets a list of running tasks. If there are no running tasks, it starts 1 task.
// If there are no running tasks, and either the task definition has changed or
// forceUpdate is specified, then the running tasks are stopped and relaunched
// with the task definition and run parameters in the current call.
func (t *Task) up(forceUpdate bool) error {
	ecsTasks, err := entity.CollectTasksWithStatus(t, ecs.DesiredStatusRunning, true)
	if err != nil {
		return err
	}
	_, err = entity.GetOrCreateTaskDefinition(t)
	if err != nil {
		return err
	}

	err = entity.OptionallyCreateLogs(t)
	if err != nil {
		return err
	}

	countTasks := len(ecsTasks)
	if countTasks == 0 {
		return t.createOne()
	}

	log.WithFields(log.Fields{
		"ProjectName":  entity.GetProjectName(t),
		"CountOfTasks": countTasks,
	}).Info("Found existing ECS tasks for project")

	// Note: Picking the first task definition as a standard and comparing against that
	oldTaskDef := aws.StringValue(ecsTasks[0].TaskDefinitionArn)
	newTaskDef := aws.StringValue(t.TaskDefinition().TaskDefinitionArn)

	ecsTaskArns := make(map[string]bool)

	if oldTaskDef != newTaskDef || forceUpdate {
		log.WithFields(log.Fields{"taskDefinition": newTaskDef}).Info("Updating to new task definition")

		chunkSize := 10
		for i := 0; i < len(ecsTasks); i += chunkSize {
			var chunk []*ecs.Task
			if i+chunkSize > len(ecsTasks) {
				chunk = ecsTasks[i:len(ecsTasks)]
			} else {
				chunk = ecsTasks[i : i+chunkSize]
			}

			// stop 10 and then run 10

			for _, task := range chunk {
				arn := aws.StringValue(task.TaskArn)
				ecsTaskArns[arn] = true
				err := t.Context().ECSClient.StopTask(arn)
				if err != nil {
					return err
				}
			}
			newTasks, err := t.runTasks(newTaskDef, len(chunk))
			if err != nil {
				return err
			}
			for _, task := range newTasks {
				ecsTaskArns[aws.StringValue(task.TaskArn)] = true
			}
		}
		return waitForTasks(t, ecsTaskArns)
	}
	return nil
}

// ---------- naming utils -----------

func getFormattedContainerName(task *ecs.Task, container *ecs.Container) string {
	taskId := entity.GetIdFromArn(task.TaskArn)
	return composeutils.GetFormattedContainerName(taskId, aws.StringValue(container.Name))
}
