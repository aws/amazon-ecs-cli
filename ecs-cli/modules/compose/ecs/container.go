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
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
)

// Container is a wrapper around ecsContainer
type Container struct {
	task         *ecs.Task
	ec2IPAddress string
	ecsContainer *ecs.Container
}

// NewContainer creates a new instance of the container and sets the task id and ecs container to it
func NewContainer(task *ecs.Task, ec2IPAddress string, container *ecs.Container) Container {
	return Container{
		task:         task,
		ec2IPAddress: ec2IPAddress,
		ecsContainer: container,
	}
}

// Id returns the ECS container's UUID
func (c *Container) Id() string {
	return utils.GetIdFromArn(aws.StringValue(c.ecsContainer.ContainerArn))
}

// Name returns the task's UUID with the container
// ECS doesn't have a describe container API so providing the task's UUID in the name enables users
// to easily look up this container in the ecs world by invoking DescribeTask
func (c *Container) Name() string {
	taskId := utils.GetIdFromArn(aws.StringValue(c.task.TaskArn))
	return utils.GetFormattedContainerName(taskId, aws.StringValue(c.ecsContainer.Name))
}

// TaskDefinition returns the ECS task definition id which encompasses the container definition, with
// which this container was created
func (c *Container) TaskDefinition() string {
	return utils.GetIdFromArn(aws.StringValue(c.task.TaskDefinitionArn))
}

// State returns the status of the container with the exit code and reason of stopped containers if exists
func (c *Container) State() string {
	ecsContainer := *c.ecsContainer
	status := aws.StringValue(ecsContainer.LastStatus)
	if status != ecs.DesiredStatusStopped {
		return status
	}
	// add the exitCode and reason if present for the stopped containers
	if ecsContainer.ExitCode != nil {
		status = fmt.Sprintf("%s ExitCode: %d", status, aws.Int64Value(ecsContainer.ExitCode))
	}
	if ecsContainer.Reason != nil {
		status = fmt.Sprintf("%s Reason: %s", status, aws.StringValue(ecsContainer.Reason))
	}
	return status
}

// PortString returns a formatted string with container's network bindings
// in a comma seperated fashion
func (c *Container) PortString() string {
	result := []string{}
	for _, port := range c.ecsContainer.NetworkBindings {
		protocol := ecs.TransportProtocolTcp
		if port.Protocol != nil {
			protocol = aws.StringValue(port.Protocol)
		}
		ipAddr := aws.StringValue(port.BindIP)
		if c.ec2IPAddress != "" {
			ipAddr = c.ec2IPAddress
		}
		result = append(result, fmt.Sprintf("%s:%d->%d/%s",
			ipAddr,
			aws.Int64Value(port.HostPort),
			aws.Int64Value(port.ContainerPort),
			protocol))
	}
	return strings.Join(result, ", ")
}

// convertContainersToInfoSet transforms the list of containers into a formatted set of fields
func convertContainersToInfoSet(containers []Container) project.InfoSet {
	result := project.InfoSet{}
	for _, cont := range containers {
		info := project.Info{}
		// TODO: Add more fields
		info = append(info, project.InfoPart{"Name", cont.Name()})
		info = append(info, project.InfoPart{"State", cont.State()})
		info = append(info, project.InfoPart{"Ports", cont.PortString()})
		info = append(info, project.InfoPart{"TaskDefinition", cont.TaskDefinition()})
		result = append(result, info)
	}
	return result
}
