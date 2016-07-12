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

	log "github.com/Sirupsen/logrus"
	composeutils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils/cache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
)

// ProjectEntity ties closely to how operations performed with the compose yaml are integrated with ECS
// It holds all the commands that are needed to operate the compose app
type ProjectEntity interface {
	Create() error
	Start() error
	Up() error
	Info(filterComposeTasks bool) (project.InfoSet, error)
	Run(commandOverrides map[string]string) error
	Scale(count int) error
	Stop() error
	Down() error

	LoadContext() error
	Context() *Context
	Sleeper() *utils.TimeSleeper
	TaskDefinition() *ecs.TaskDefinition
	TaskDefinitionCache() cache.Cache
	SetTaskDefinition(taskDefinition *ecs.TaskDefinition)
}

// setupTaskDefinitionCache finds a file system cache to store the ecs task definitions
func setupTaskDefinitionCache() cache.Cache {
	tdCache, err := cache.NewFSCache(composeutils.ProjectTDCache)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Unable to create cache for task definitions; extranious ones may be registered")
		tdCache = cache.NewNoopCache()
	}
	return tdCache
}

// ------- command helper functions ------------
// getOrCreateTaskDefinition
// info
// collectContainers
// collectTasks

// getOrCreateTaskDefinition gets the task definition from cache if present, else
// creates it in ECS and persists in a local cache. It also sets the latest
// taskDefinition to the current instance of task
func getOrCreateTaskDefinition(entity ProjectEntity) (*ecs.TaskDefinition, error) {
	taskDefinition := entity.TaskDefinition()
	log.WithFields(log.Fields{
		"TaskDefinition": taskDefinition,
	}).Debug("Finding task definition in cache or creating if needed")

	resp, err := entity.Context().ECSClient.RegisterTaskDefinitionIfNeeded(&ecs.RegisterTaskDefinitionInput{
		Family:               taskDefinition.Family,
		ContainerDefinitions: taskDefinition.ContainerDefinitions,
		Volumes:              taskDefinition.Volumes,
	}, entity.TaskDefinitionCache())

	if err != nil {
		composeutils.LogError(err, "Create task definition failed")
		return nil, err
	}

	log.WithFields(log.Fields{
		"TaskDefinition": getIdFromArn(resp.TaskDefinitionArn),
	}).Info("Using ECS task definition")

	// update the taskdefinition of the entity with the newly received TaskDefinition
	entity.SetTaskDefinition(resp)
	return resp, nil
}

// info returns a formatted list of containers (running and stopped) in the current cluster
// filtered by this project if filterLocal is set to true
func info(entity ProjectEntity, filterLocal bool) (project.InfoSet, error) {
	containers, err := collectContainers(entity, filterLocal)
	if err != nil {
		return nil, err
	}
	allInfo := convertContainersToInfoSet(containers)
	return allInfo, nil
}

// collectContainers gets all the desiredStatus=RUNNING and STOPPED tasks with EC2 IP Addresses
// if filterLocal is set to true, it filters tasks created by this project
func collectContainers(entity ProjectEntity, filterLocal bool) ([]Container, error) {
	ecsTasks, err := collectTasks(entity, filterLocal)
	if err != nil {
		return nil, err
	}
	return getContainersForTasks(entity, ecsTasks)
}

// collectTasks gets all the desiredStatus=RUNNING and STOPPED tasks
// if filterLocal is set to true, it filters tasks created by this project
func collectTasks(entity ProjectEntity, filterLocal bool) ([]*ecs.Task, error) {
	// TODO, parallelize, perhaps using channels
	result := []*ecs.Task{}
	ecsTasks, err := collectTasksWithStatus(entity, ecs.DesiredStatusRunning, filterLocal)
	if err != nil {
		return nil, err
	}
	result = append(result, ecsTasks...)

	ecsTasks, err = collectTasksWithStatus(entity, ecs.DesiredStatusStopped, filterLocal)
	if err != nil {
		return nil, err
	}
	result = append(result, ecsTasks...)
	return result, nil
}

// collectTasksWithStatus gets all the tasks of specified desired status
// If filterLocal is true, it filters out with startedBy as this project
func collectTasksWithStatus(entity ProjectEntity, status string, filterLocal bool) ([]*ecs.Task, error) {
	request := constructListPagesRequest(entity, status, filterLocal)
	result := []*ecs.Task{}

	err := entity.Context().ECSClient.GetTasksPages(request, func(respTasks []*ecs.Task) error {
		result = append(result, respTasks...)
		return nil
	})

	return result, err
}

// constructListPagesRequest constructs the request based on the entity type and function parameters
func constructListPagesRequest(entity ProjectEntity, status string, filterLocal bool) *ecs.ListTasksInput {
	request := &ecs.ListTasksInput{
		DesiredStatus: aws.String(status),
	}

	// if service set ServiceName to the request, else set StartedBy to filter out (provided filterLocal is true)
	service, ok := entity.(*Service)
	if ok {
		request.ServiceName = aws.String(service.getServiceName())
	} else if filterLocal {
		request.StartedBy = aws.String(getStartedBy(entity))
	}
	return request
}

// getContainersForTasks returns the list of containers from the list of tasks.
// It also fetches the ec2 public ip addresses of instances where the containers are running
func getContainersForTasks(entity ProjectEntity, ecsTasks []*ecs.Task) ([]Container, error) {
	result := []Container{}
	if len(ecsTasks) == 0 {
		return result, nil
	}

	// TODO, We are getting the container instances and then ec2 instances to fetch the IP Address of EC2 instance
	// Should we optimize by looking only for running tasks?
	containerInstanceArns := uniqueContainerInstanceArns(ecsTasks)
	if len(containerInstanceArns) == 0 {
		return nil, fmt.Errorf("No container instances for found tasks")
	}

	containerToEC2InstanceIDs, err := entity.Context().ECSClient.GetEC2InstanceIDs(containerInstanceArns)
	if err != nil {
		return nil, err
	}

	ec2InstanceIds := listEC2Ids(containerToEC2InstanceIDs)
	ec2Instances, err := entity.Context().EC2Client.DescribeInstances(ec2InstanceIds)
	if err != nil {
		return nil, err
	}

	for _, ecsTask := range ecsTasks {
		ec2ID := containerToEC2InstanceIDs[aws.StringValue(ecsTask.ContainerInstanceArn)]
		var ec2IPAddress string
		if ec2ID != "" {
			ec2IPAddress = aws.StringValue(ec2Instances[ec2ID].PublicIpAddress)
		}
		for _, container := range ecsTask.Containers {
			result = append(result, NewContainer(ecsTask, ec2IPAddress, container))
		}
	}
	return result, nil
}

// listEC2Ids converts a map of ContainerInstance:EC2Instance Ids to a
// list of ec2 instance Ids
func listEC2Ids(containerToEC2InstancesMap map[string]string) []*string {
	ec2InstanceIds := []*string{}
	for _, val := range containerToEC2InstancesMap {
		ec2InstanceIds = append(ec2InstanceIds, aws.String(val))
	}
	return ec2InstanceIds
}

// uniqueContainerInstanceArns returns the container instance arns
// present in the input array of tasks, after uniq'ing them
func uniqueContainerInstanceArns(tasks []*ecs.Task) []*string {
	out := make(map[string]bool, 0)
	for _, task := range tasks {
		if task.ContainerInstanceArn != nil {
			out[aws.StringValue(task.ContainerInstanceArn)] = true
		}
	}
	return convertMapToSlice(out)
}

// convertMapToSlice converts the map [String -> bool] to a AWS String Slice that is used by our APIs as input
func convertMapToSlice(mapItems map[string]bool) []*string {
	sliceItems := make([]string, 0, len(mapItems))
	for key := range mapItems {
		sliceItems = append(sliceItems, key)
	}
	return aws.StringSlice(sliceItems)
}

// ---------- naming utils -----------

// getStartedBy returns an auto-generated formatted string
// that can be supplied while starting an ECS task and is used to identify the owner of ECS Task
func getStartedBy(entity ProjectEntity) string {
	return composeutils.GetStartedBy(getProjectPrefix(entity), getProjectName(entity))
}

// getProjectName returns the name of the project that was set in the context we are working with
func getProjectName(entity ProjectEntity) string {
	return entity.Context().Context.ProjectName
}

// getProjectPrefix returns the prefix for the project name
func getProjectPrefix(entity ProjectEntity) string {
	return entity.Context().ECSParams.ComposeProjectNamePrefix
}

// getIdFromArn gets the aws String value of the input arn and returns the id part of the arn
func getIdFromArn(arn *string) string {
	return composeutils.GetIdFromArn(aws.StringValue(arn))
}
