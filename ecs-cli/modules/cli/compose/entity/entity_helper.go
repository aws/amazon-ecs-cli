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

package entity

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	composecontainer "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/container"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/types"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/logs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudwatchlogs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/cache"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
)

const (
	eniIDKey          = "networkInterfaceId"
	ENIStatusAttached = "ATTACHED"
	ENIAttachmentType = "ElasticNetworkInterface"
)

// TaskDefinitionStore is an in memory cache of Task definitions
// This is provided to reduce the number of calls to describe-task-definition
type TaskDefinitionStore struct {
	inMemoryTaskDefStore map[string]*ecs.TaskDefinition
}

// NewTaskDefinitionStore creates a new in memory task definition cache
func NewTaskDefinitionStore() *TaskDefinitionStore {
	return &TaskDefinitionStore{
		inMemoryTaskDefStore: make(map[string]*ecs.TaskDefinition),
	}
}

func (tdStore *TaskDefinitionStore) getTaskDefintion(entity ProjectEntity, taskDefArn string) (*ecs.TaskDefinition, error) {
	// TODO: optimize even further by asynchronously storing to disk so that its available in the next ecs-cli invocation
	td, ok := tdStore.inMemoryTaskDefStore[taskDefArn]
	if !ok {
		var err error
		td, err = entity.Context().ECSClient.DescribeTaskDefinition(taskDefArn)
		if err != nil {
			return nil, err
		}
		tdStore.inMemoryTaskDefStore[taskDefArn] = td
	}

	return td, nil
}

// SetupTaskDefinitionCache finds a file system cache to store the ecs task definitions
func SetupTaskDefinitionCache() cache.Cache {
	tdCache, err := cache.NewFSCache(utils.ProjectTDCache)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Unable to create cache for task definitions; extraneous ones may be registered")
		tdCache = cache.NewNoopCache()
	}
	return tdCache
}

// ------- command helper functions ------------
// getOrCreateTaskDefinition
// info
// collectContainers
// collectTasks

// GetOrCreateTaskDefinition gets the task definition from cache if present, else
// creates it in ECS and persists in a local cache. It also sets the latest
// taskDefinition to the current instance of task
func GetOrCreateTaskDefinition(entity ProjectEntity) (*ecs.TaskDefinition, error) {
	taskDefinition := entity.TaskDefinition()
	log.WithFields(log.Fields{
		"TaskDefinition": taskDefinition,
	}).Debug("Finding task definition in cache or creating if needed")

	request := createRegisterTaskDefinitionRequest(taskDefinition)

	resp, err := entity.Context().ECSClient.RegisterTaskDefinitionIfNeeded(request, entity.TaskDefinitionCache())

	if err != nil {
		utils.LogError(err, "Create task definition failed")
		return nil, err
	}

	log.WithFields(log.Fields{
		"TaskDefinition": GetIdFromArn(resp.TaskDefinitionArn),
	}).Info("Using ECS task definition")

	// update the taskdefinition of the entity with the newly received TaskDefinition
	entity.SetTaskDefinition(resp)
	return resp, nil
}

func createRegisterTaskDefinitionRequest(taskDefinition *ecs.TaskDefinition) *ecs.RegisterTaskDefinitionInput {
	// Valid values for network mode are none, host or bridge. If no value
	// is passed for network mode, ECS will set it to 'bridge' on most
	// platforms, but Windows has different network modes. Passing nil allows ECS
	// to do the right thing for each platform.
	request := &ecs.RegisterTaskDefinitionInput{
		Family:                  taskDefinition.Family,
		ContainerDefinitions:    taskDefinition.ContainerDefinitions,
		Volumes:                 taskDefinition.Volumes,
		TaskRoleArn:             taskDefinition.TaskRoleArn,
		RequiresCompatibilities: taskDefinition.RequiresCompatibilities,
		ExecutionRoleArn:        taskDefinition.ExecutionRoleArn,
	}

	if networkMode := taskDefinition.NetworkMode; aws.StringValue(networkMode) != "" {
		request.NetworkMode = networkMode
	}

	if cpu := taskDefinition.Cpu; aws.StringValue(cpu) != "" {
		request.Cpu = cpu
	}

	if memory := taskDefinition.Memory; aws.StringValue(memory) != "" {
		request.Memory = memory
	}

	return request
}

// Info returns a formatted list of containers (running and stopped) in the current cluster
// filtered by this project if filterLocal is set to true
func Info(entity ProjectEntity, filterLocal bool) (project.InfoSet, error) {
	containers, err := collectContainers(entity, filterLocal)
	if err != nil {
		return nil, err
	}
	return composecontainer.ConvertContainersToInfoSet(containers), nil
}

// collectContainers gets all the desiredStatus=RUNNING and STOPPED tasks with EC2 IP Addresses
// if filterLocal is set to true, it filters tasks created by this project
func collectContainers(entity ProjectEntity, filterLocal bool) ([]composecontainer.Container, error) {
	ecsTasks, err := collectTasks(entity, filterLocal)
	if err != nil {
		return nil, err
	}
	info, ecsTasks, err := getContainersForTasksWithTaskNetworking(entity, ecsTasks)
	if err != nil {
		return nil, err
	}
	return getContainersForTasks(entity, ecsTasks, info)
}

// collectTasks gets all the desiredStatus=RUNNING and STOPPED tasks
// if filterLocal is set to true, it filters tasks created by this project
func collectTasks(entity ProjectEntity, filterLocal bool) ([]*ecs.Task, error) {
	// TODO, parallelize, perhaps using channels
	result := []*ecs.Task{}
	ecsTasks, err := CollectTasksWithStatus(entity, ecs.DesiredStatusRunning, filterLocal)
	if err != nil {
		return nil, err
	}
	result = append(result, ecsTasks...)

	ecsTasks, err = CollectTasksWithStatus(entity, ecs.DesiredStatusStopped, filterLocal)
	if err != nil {
		return nil, err
	}
	result = append(result, ecsTasks...)
	return result, nil
}

// CollectTasksWithStatus gets all the tasks of specified desired status
// If filterLocal is true, it filters out with Group or StartedBy as this project
func CollectTasksWithStatus(entity ProjectEntity, status string, filterLocal bool) ([]*ecs.Task, error) {
	request := constructListPagesRequest(entity, status, filterLocal)
	result := []*ecs.Task{}

	err := entity.Context().ECSClient.GetTasksPages(request, func(respTasks []*ecs.Task) error {
		// Filter the results by task.Group
		if entity.EntityType() == types.Task && filterLocal {
			for _, task := range respTasks {
				if aws.StringValue(task.Group) == GetTaskGroup(entity) {
					result = append(result, task)
				} else if aws.StringValue(task.StartedBy) == GetTaskDefinitionFamily(entity) { // Deprecated, filter by StartedBy
					result = append(result, task)
				}
			}
		} else {
			result = append(result, respTasks...)
		}
		return nil
	})

	return result, err
}

// constructListPagesRequest constructs the request based on the entity type and function parameters
func constructListPagesRequest(entity ProjectEntity, status string, filterLocal bool) *ecs.ListTasksInput {
	request := &ecs.ListTasksInput{
		DesiredStatus: aws.String(status),
	}

	// if service set ServiceName to the request, else set Task definition family to filter out (provided filterLocal is true)
	if entity.EntityType() == types.Service {
		request.SetServiceName(GetServiceName(entity))
	} else if filterLocal {
		// TODO: filter by Group when available in API
		request.SetFamily(GetTaskDefinitionFamily(entity))
	}
	return request
}

func convertToNetworkBindings(containerDef *ecs.ContainerDefinition) (bindings []*ecs.NetworkBinding) {
	for _, portMapping := range containerDef.PortMappings {
		bindings = append(bindings, &ecs.NetworkBinding{
			ContainerPort: portMapping.ContainerPort,
			HostPort:      portMapping.HostPort,
			Protocol:      portMapping.Protocol,
		})
	}

	return bindings
}

func getContainerDef(taskDef *ecs.TaskDefinition, name string) (*ecs.ContainerDefinition, error) {
	for _, containerDef := range taskDef.ContainerDefinitions {
		if aws.StringValue(containerDef.Name) == name {
			return containerDef, nil
		}
	}
	return nil, fmt.Errorf("Unexpected Error: Could not find container %s in task definition", name)
}

// processAttachment takes the attachment and associates the ID of an attached ENI with the TaskArn
// Mutates: eniIDs, taskENIs
func processAttachment(taskENIs map[string]string, eniIDs *[]*string, ecsTask *ecs.Task, attachment *ecs.Attachment) {
	if aws.StringValue(attachment.Status) == ENIStatusAttached && aws.StringValue(attachment.Type) == ENIAttachmentType {
		for _, detail := range attachment.Details {
			if aws.StringValue(detail.Name) == eniIDKey {
				eniID := detail.Value
				*eniIDs = append(*eniIDs, eniID)
				taskENIs[aws.StringValue(eniID)] = aws.StringValue(ecsTask.TaskArn)
			}
		}
	}
}

func getPublicIPsFromENIs(entity ProjectEntity, ecsTasks []*ecs.Task) (map[string]string, error) {
	taskPublicIPs := make(map[string]string)
	var eniIDs []*string
	taskENIs := make(map[string]string)
	for _, ecsTask := range ecsTasks {
		if aws.StringValue(ecsTask.LaunchType) == config.LaunchTypeFargate && aws.StringValue(ecsTask.LastStatus) == ecs.DesiredStatusRunning {
			for _, attachment := range ecsTask.Attachments {
				processAttachment(taskENIs, &eniIDs, ecsTask, attachment)
			}
		}
	}

	if len(eniIDs) == 0 {
		return taskPublicIPs, nil
	}

	netInterfaces, err := entity.Context().EC2Client.DescribeNetworkInterfaces(eniIDs)
	if err != nil {
		log.Warnf("Failed to describe Elastic Network Interfaces; falling back to private IP obtained from DescribeTask. Reason: %s", err)
		return taskPublicIPs, nil
	}

	for _, eni := range netInterfaces {
		if eni.Association != nil {
			taskArn := taskENIs[aws.StringValue(eni.NetworkInterfaceId)]
			taskPublicIPs[taskArn] = aws.StringValue(eni.Association.PublicIp)
		}
	}

	return taskPublicIPs, nil
}

func getContainersForTasksWithTaskNetworking(entity ProjectEntity, ecsTasks []*ecs.Task) ([]composecontainer.Container, []*ecs.Task, error) {
	var tasksWithInstanceIPs []*ecs.Task
	info := []composecontainer.Container{}
	tdStore := NewTaskDefinitionStore()

	if len(ecsTasks) == 0 {
		return info, ecsTasks, nil
	}

	// For Fargate tasks
	taskENIPublicIPs, err := getPublicIPsFromENIs(entity, ecsTasks)
	if err != nil {
		return nil, nil, err
	}

	for _, ecsTask := range ecsTasks {
		var hasTaskNetworking bool
		for _, container := range ecsTask.Containers {
			if len(container.NetworkInterfaces) > 0 {
				taskDef, err := tdStore.getTaskDefintion(entity, aws.StringValue(ecsTask.TaskDefinitionArn))
				if err != nil {
					return nil, nil, err
				}
				containerDef, err := getContainerDef(taskDef, aws.StringValue(container.Name))
				if err != nil {
					return nil, nil, err
				}
				bindings := convertToNetworkBindings(containerDef)
				ipAddress := aws.StringValue(container.NetworkInterfaces[0].PrivateIpv4Address)
				if aws.StringValue(ecsTask.LaunchType) == config.LaunchTypeFargate {
					if ip := taskENIPublicIPs[aws.StringValue(ecsTask.TaskArn)]; ip != "" {
						ipAddress = ip
					}
				}
				info = append(info, composecontainer.NewContainer(ecsTask, ipAddress, container, bindings))
				hasTaskNetworking = true
			}
		}
		if !hasTaskNetworking {
			tasksWithInstanceIPs = append(tasksWithInstanceIPs, ecsTask)
		}
	}
	return info, tasksWithInstanceIPs, nil
}

// getContainersForTasks returns the list of containers from the list of tasks.
// It also fetches the ip addresses of instances where the containers are running
func getContainersForTasks(entity ProjectEntity, ecsTasks []*ecs.Task, info []composecontainer.Container) ([]composecontainer.Container, error) {
	if len(ecsTasks) == 0 {
		return info, nil
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
		if ec2ID != "" && ec2Instances[ec2ID] != nil {
			ec2IPAddress = aws.StringValue(ec2Instances[ec2ID].PublicIpAddress)
			if ec2IPAddress == "" {
				ec2IPAddress = aws.StringValue(ec2Instances[ec2ID].PrivateIpAddress)
			}
		}
		for _, container := range ecsTask.Containers {
			info = append(info, composecontainer.NewContainer(ecsTask, ec2IPAddress, container, container.NetworkBindings))
		}
	}
	return info, nil
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
	return ConvertMapToSlice(out)
}

// ConvertMapToSlice converts the map [String -> bool] to a AWS String Slice that is used by our APIs as input
func ConvertMapToSlice(mapItems map[string]bool) []*string {
	sliceItems := make([]string, 0, len(mapItems))
	for key := range mapItems {
		sliceItems = append(sliceItems, key)
	}
	return aws.StringSlice(sliceItems)
}

// ---------- naming utils -----------

// GetTaskGroup returns an auto-generated formatted string
// that can be supplied while starting an ECS task and is used to identify the owner of ECS Task
func GetTaskGroup(entity ProjectEntity) string {
	return utils.GetTaskGroup(getProjectPrefix(entity), GetProjectName(entity))
}

// GetTaskDefinitionFamily returns the family name
func GetTaskDefinitionFamily(entity ProjectEntity) string {
	// ComposeProjectNamePrefix is deprecated, but its use must remain for backwards compatibility
	return entity.Context().CLIParams.ComposeProjectNamePrefix + GetProjectName(entity)
}

// GetProjectName returns the name of the project that was set in the context we are working with
func GetProjectName(entity ProjectEntity) string {
	return entity.Context().Context.ProjectName
}

// getProjectPrefix returns the prefix for the project name
func getProjectPrefix(entity ProjectEntity) string {
	return entity.Context().CLIParams.ComposeProjectNamePrefix
}

// GetServiceName using project entity
func GetServiceName(entity ProjectEntity) string {
	return utils.GetServiceName(getServicePrefix(entity), GetProjectName(entity))
}

func getServicePrefix(entity ProjectEntity) string {
	return entity.Context().CLIParams.ComposeServiceNamePrefix
}

// GetIdFromArn gets the aws String value of the input arn and returns the id part of the arn
func GetIdFromArn(arn *string) string {
	return utils.GetIdFromArn(aws.StringValue(arn))
}

// ValidateFargateParams ensures that the correct config has been given to run a Fargate task
func ValidateFargateParams(ecsParams *utils.ECSParams, launchType string) error {
	if launchType == config.LaunchTypeFargate {
		// If ecs-params.yml not passed in
		if ecsParams == nil {
			return fmt.Errorf("Launch Type %s requires network configuration to be set. Set network configuration using an ECS Params file.", launchType)
		}
		if ecsParams.TaskDefinition.NetworkMode != "awsvpc" {
			return fmt.Errorf("Launch Type %s requires network mode to be 'awsvpc'. Set network mode using an ECS Params file.", launchType)
		}
	}

	return nil
}

// OptionallyCreateLogs creates CW log groups if the --create-log-group flag is present.
func OptionallyCreateLogs(entity ProjectEntity) error {
	if entity.Context().CLIContext.Bool(flags.CreateLogsFlag) {
		err := logs.CreateLogGroups(entity.TaskDefinition(), cloudwatchlogs.NewLogClientFactory(entity.Context().CLIParams))
		if err != nil {
			return err
		}
	}

	return nil
}
