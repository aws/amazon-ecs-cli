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
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	composeutils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils/cache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
)

// Service type is placeholder for a single task definition and its cache
// and it performs operations on ECS Service level
type Service struct {
	taskDef          *ecs.TaskDefinition
	cache            cache.Cache
	projectContext   *Context
	timeSleeper      *utils.TimeSleeper
	deploymentConfig *ecs.DeploymentConfiguration
}

const (
	ecsActiveResourceCode  = "ACTIVE"
	ecsMissingResourceCode = "MISSING"
)

// NewService creates an instance of a Service and also sets up a cache for task definition
func NewService(context *Context) ProjectEntity {
	return &Service{
		cache:          setupTaskDefinitionCache(),
		projectContext: context,
		timeSleeper:    &utils.TimeSleeper{},
	}
}

// LoadContext reads the context set in NewService and loads DeploymentConfiguration
func (s *Service) LoadContext() error {
	maxPercent, err := getInt64FromCLIContext(s.Context(), DeploymentMaxPercentFlag)
	if err != nil {
		return err
	}
	minHealthyPercent, err := getInt64FromCLIContext(s.Context(), DeploymentMinHealthyPercentFlag)
	if err != nil {
		return err
	}
	s.deploymentConfig = &ecs.DeploymentConfiguration{
		MaximumPercent:        maxPercent,
		MinimumHealthyPercent: minHealthyPercent,
	}
	return nil
}

// getInt64FromCLIContext reads the flag from the cli context and typecasts into *int64
func getInt64FromCLIContext(context *Context, flag string) (*int64, error) {
	value := context.CLIContext.String(flag)
	if value == "" {
		return nil, nil
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Please pass integer value for the flag %s", flag)
	}
	return aws.Int64(intValue), nil
}

// SetTaskDefinition sets the ecs task definition to the current instance of Service
func (s *Service) SetTaskDefinition(taskDefinition *ecs.TaskDefinition) {
	s.taskDef = taskDefinition
}

// Context returs the context of this project
func (s *Service) Context() *Context {
	return s.projectContext
}

// Sleeper returs an instance of TimeSleeper used to wait until Service has gone to a stable state
func (s *Service) Sleeper() *utils.TimeSleeper {
	return s.timeSleeper
}

// TaskDefinition returns the task definition object that was created by
// transforming the Service Configs to ECS acceptable format
func (s *Service) TaskDefinition() *ecs.TaskDefinition {
	return s.taskDef
}

// TaskDefinitionCache returns the cache that should be used when checking for
// previous task definition
func (s *Service) TaskDefinitionCache() cache.Cache {
	return s.cache
}

// DeploymentConfig returns the configuration that control how many tasks run during the
// deployment and the ordering of stopping and starting tasks
func (s *Service) DeploymentConfig() *ecs.DeploymentConfiguration {
	return s.deploymentConfig
}

// ----------- Commands' implementations --------

// Create creates a task definition in ECS for the containers in the compose file
// and persists it in a cache locally. It always checks the cache before creating
func (s *Service) Create() error {
	_, err := getOrCreateTaskDefinition(s)
	if err != nil {
		return err
	}
	return s.createService()
}

// Start starts the containers if they weren't already running. Internally, start calls
// ECS.DescribeService to find out if the service is Active and if the count is 0,
// it updates the service with desired count as 1 else its a no-op
// TODO: Instead of always setting count=1, if the containers were Stopped before,
//       Start should fetch the previously set desired-count from the cache and start x count of containers
func (s *Service) Start() error {
	ecsService, err := s.describeService()
	if err != nil {
		// Describe API returns the failures for resources in the response (instead of returning an error)
		// Read the custom error returned from describeService to see if the resource was missing
		if strings.Contains(err.Error(), ecsMissingResourceCode) {
			return fmt.Errorf("Please use '%s' command to create the service '%s' first",
				CreateServiceCommandName, s.getServiceName())
		}
		return err
	}
	return s.startService(ecsService)
}

// Up creates the task definition and service and starts the containers if necessary.
// It does so by calling DescribeService to see if its present, else's calls Create() and Start()
// If the compose file had changed, it would update the service with the new task definition
// by calling UpdateService with the new task definition
func (s *Service) Up() error {
	// describe service to get the task definition and count running
	ecsService, err := s.describeService()
	var missingServiceErr bool
	if err != nil {
		if strings.Contains(err.Error(), ecsMissingResourceCode) {
			missingServiceErr = true
		} else {
			return err
		}
	}

	// get the current snapshot of compose yml
	// and update this instance with the latest task definition
	newTaskDefinition, err := getOrCreateTaskDefinition(s)
	if err != nil {
		return err
	}

	// if ECS service was not created before, or is inactive, create and start the ECS Service
	if missingServiceErr || aws.StringValue(ecsService.Status) != ecsActiveResourceCode {
		// uses the latest task definition to create the service
		err = s.createService()
		if err != nil {
			return err
		}
		return s.Start()
	}

	oldTaskDefinitionId := getIdFromArn(ecsService.TaskDefinition)
	newTaskDefinitionId := getIdFromArn(newTaskDefinition.TaskDefinitionArn)

	oldCount := aws.Int64Value(ecsService.DesiredCount)
	newCount := int64(1)
	if oldCount != 0 {
		newCount = oldCount // get the current non-zero count
	}

	// if both the task definitions are the same, just start the service
	if oldTaskDefinitionId == newTaskDefinitionId {
		return s.startService(ecsService)
	}

	ecsServiceName := aws.StringValue(ecsService.ServiceName)
	deploymentConfig := s.DeploymentConfig()
	// if the task definitions were different, updateService with new task definition
	// this creates a deployment in ECS and slowly takes down the containers with old ones and starts new ones
	err = s.Context().ECSClient.UpdateService(ecsServiceName, newTaskDefinitionId, newCount, deploymentConfig)
	if err != nil {
		return err
	}
	fields := log.Fields{
		"serviceName":    ecsServiceName,
		"taskDefinition": newTaskDefinitionId,
		"desiredCount":   newCount,
	}
	if deploymentConfig != nil && deploymentConfig.MaximumPercent != nil {
		fields["deployment-max-percent"] = aws.Int64Value(deploymentConfig.MaximumPercent)
	}
	if deploymentConfig != nil && deploymentConfig.MinimumHealthyPercent != nil {
		fields["deployment-min-healthy-percent"] = aws.Int64Value(deploymentConfig.MinimumHealthyPercent)
	}

	log.WithFields(fields).Info("Updated the ECS service with a new task definition. " +
		"Old containers will be stopped automatically, and replaced with new ones")
	return waitForServiceTasks(s, ecsServiceName)
}

// Info returns a formatted list of containers (running and stopped) started by this service
func (s *Service) Info(filterProjectTasks bool) (project.InfoSet, error) {
	// filterProjectTasks is not honored for services, because ECS Services have their
	// own custom StartedBy field, overriding that with startedBy=project will result in no tasks
	// We should instead filter by ServiceName=service
	return info(s, false)
}

// Scale the service desired count to be the specified count
func (s *Service) Scale(count int) error {
	return s.updateService(int64(count))
}

// Stop stops all the containers in the service by calling ECS.UpdateService(count=0)
// TODO, Store the current desiredCount in a cache, so that number of tasks(group of containers) can be started again
func (s *Service) Stop() error {
	return s.updateService(int64(0))
}

// Down stops any running containers(tasks) by calling Stop() and deletes an active ECS Service
// NoOp if the service is inactive
func (s *Service) Down() error {
	// describe the service
	ecsService, err := s.describeService()
	if err != nil {
		return err
	}

	ecsServiceName := aws.StringValue(ecsService.ServiceName)
	// if already deleted, NoOp
	if aws.StringValue(ecsService.Status) != ecsActiveResourceCode {
		log.WithFields(log.Fields{
			"serviceName": ecsServiceName,
		}).Info("ECS Service is already deleted")
		return nil
	}

	// stop any running tasks
	if aws.Int64Value(ecsService.DesiredCount) != 0 {
		if err = s.Stop(); err != nil {
			return err
		}
	}

	// deleteService
	if err = s.Context().ECSClient.DeleteService(ecsServiceName); err != nil {
		return err
	}
	return waitForServiceTasks(s, ecsServiceName)
}

// Run expects to issue a command override and start containers. But that doesnt apply to the context of ECS Services
func (s *Service) Run(commandOverrides map[string]string) error {
	return composeutils.ErrUnsupported
}

// ----------- Commands' helper functions --------

// createService calls the underlying ECS.CreateService
func (s *Service) createService() error {
	serviceName := s.getServiceName()
	taskDefinitionId := getIdFromArn(s.TaskDefinition().TaskDefinitionArn)
	err := s.Context().ECSClient.CreateService(serviceName, taskDefinitionId, s.DeploymentConfig())
	if err != nil {
		return err
	}
	return nil
}

// describeService calls underlying ECS.DescribeService and expects the service to be present,
// returns error otherwise
func (s *Service) describeService() (*ecs.Service, error) {
	serviceName := s.getServiceName()
	output, err := s.Context().ECSClient.DescribeService(serviceName)
	if err != nil {
		return nil, err
	}
	if len(output.Failures) > 0 {
		reason := aws.StringValue(output.Failures[0].Reason)
		return nil, fmt.Errorf("Got an error describing service '%s' : '%s'", serviceName, reason)
	} else if len(output.Services) == 0 {
		return nil, fmt.Errorf("Got an empty list of services while describing the service '%s'", serviceName)
	}
	return output.Services[0], nil
}

// startService checks if the service has a zero desired count and updates the count to 1 (of each container)
func (s *Service) startService(ecsService *ecs.Service) error {
	desiredCount := aws.Int64Value(ecsService.DesiredCount)
	if desiredCount != 0 {
		serviceName := aws.StringValue(ecsService.ServiceName)
		//NoOp
		log.WithFields(log.Fields{
			"serviceName":  serviceName,
			"desiredCount": desiredCount,
		}).Info("ECS Service is already running")

		return waitForServiceTasks(s, serviceName)
	}
	return s.updateService(int64(1))
}

// updateService calls the underlying ECS.UpdateService with the specified count
func (s *Service) updateService(count int64) error {
	serviceName := s.getServiceName()
	deploymentConfig := s.DeploymentConfig()
	if err := s.Context().ECSClient.UpdateServiceCount(serviceName, count, deploymentConfig); err != nil {
		return err
	}
	fields := log.Fields{
		"serviceName":  serviceName,
		"desiredCount": count,
	}
	if deploymentConfig != nil && deploymentConfig.MaximumPercent != nil {
		fields["deployment-max-percent"] = aws.Int64Value(deploymentConfig.MaximumPercent)
	}
	if deploymentConfig != nil && deploymentConfig.MinimumHealthyPercent != nil {
		fields["deployment-min-healthy-percent"] = aws.Int64Value(deploymentConfig.MinimumHealthyPercent)
	}

	log.WithFields(fields).Info("Updated ECS service successfully")
	return waitForServiceTasks(s, serviceName)
}

// ---------- naming utils -----------

// getServiceName returns an autogenerated name for the service based on the current project context
func (s *Service) getServiceName() string {
	return composeutils.GetServiceName(s.Context().ECSParams.ComposeServiceNamePrefix, getProjectName(s))
}
