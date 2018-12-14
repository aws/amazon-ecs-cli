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
	"strconv"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/types"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/servicediscovery"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/route53"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/cache"
	composeutils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Service type is placeholder for a single task definition and its cache
// and it performs operations on ECS Service level
type Service struct {
	taskDef           *ecs.TaskDefinition
	cache             cache.Cache
	ecsContext        *context.ECSContext
	timeSleeper       *utils.TimeSleeper
	deploymentConfig  *ecs.DeploymentConfiguration
	loadBalancer      *ecs.LoadBalancer
	role              string
	healthCheckGP     *int64
	serviceRegistries []*ecs.ServiceRegistry
}

const (
	ecsActiveResourceCode  = "ACTIVE"
	ecsMissingResourceCode = "MISSING"
)

// make servicediscovery.Create easily mockable in tests
var servicediscoveryCreate servicediscovery.CreateFunc = servicediscovery.Create

// make servicediscovery.Update easily mockable in tests
var servicediscoveryUpdate servicediscovery.UpdateFunc = servicediscovery.Update

// make servicediscovery.Delete easily mockable in tests
var servicediscoveryDelete servicediscovery.DeleteFunc = servicediscovery.Delete

// make servicediscovery.Delete easily mockable in tests
var waitUntilSDSDeletable route53.WaitUntilSDSDeletableFunc = route53.WaitUntilSDSDeletable

// NewService creates an instance of a Service and also sets up a cache for task definition
func NewService(ecsContext *context.ECSContext) entity.ProjectEntity {
	return &Service{
		cache:       entity.SetupTaskDefinitionCache(),
		ecsContext:  ecsContext,
		timeSleeper: &utils.TimeSleeper{},
	}
}

// LoadContext reads the ECS context set in NewService and loads DeploymentConfiguration and LoadBalancer
// TODO: refactor to memoize s.Context().CLIContext, since that's the only
// thing that LoadContext seems to care about? (even in getInt64FromCLIContext)
func (s *Service) LoadContext() error {
	maxPercent, err := getInt64FromCLIContext(s.Context(), flags.DeploymentMaxPercentFlag)
	if err != nil {
		return err
	}
	minHealthyPercent, err := getInt64FromCLIContext(s.Context(), flags.DeploymentMinHealthyPercentFlag)
	if err != nil {
		return err
	}
	s.deploymentConfig = &ecs.DeploymentConfiguration{
		MaximumPercent:        maxPercent,
		MinimumHealthyPercent: minHealthyPercent,
	}

	// Load Balancer
	role := s.Context().CLIContext.String(flags.RoleFlag)
	targetGroupArn := s.Context().CLIContext.String(flags.TargetGroupArnFlag)
	loadBalancerName := s.Context().CLIContext.String(flags.LoadBalancerNameFlag)
	containerName := s.Context().CLIContext.String(flags.ContainerNameFlag)
	containerPort, err := getInt64FromCLIContext(s.Context(), flags.ContainerPortFlag)
	if err != nil {
		return err
	}

	// Health Check Grace Period
	healthCheckGP, err := getInt64FromCLIContext(s.Context(), flags.HealthCheckGracePeriodFlag)
	if err != nil {
		return err
	}
	s.healthCheckGP = healthCheckGP

	// Validates LoadBalancerName and TargetGroupArn cannot exist at the same time.
	// Other validation is taken care of by the API call. This currently
	// includes errors on absence of container name and port if target
	// group or ELB name is specified or if the load balancing resources
	// specified do not exist.
	// TODO: Add validation on targetGroupArn or loadBalancerName being
	// present if containerName or containerPort are specified
	if role != "" || targetGroupArn != "" || loadBalancerName != "" || containerName != "" || containerPort != nil {
		if targetGroupArn != "" && loadBalancerName != "" {
			return errors.Errorf("[--%s] and [--%s] flags cannot both be specified", flags.LoadBalancerNameFlag, flags.TargetGroupArnFlag)
		}

		s.loadBalancer = &ecs.LoadBalancer{
			ContainerName: aws.String(containerName),
			ContainerPort: containerPort,
		}
		if targetGroupArn != "" {
			s.loadBalancer.TargetGroupArn = aws.String(targetGroupArn)
		}
		if loadBalancerName != "" {
			s.loadBalancer.LoadBalancerName = aws.String(loadBalancerName)
		}
		s.role = role
	}
	return nil
}

// getInt64FromCLIContext reads the flag from the cli context and typecasts into *int64
func getInt64FromCLIContext(context *context.ECSContext, flag string) (*int64, error) {
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
func (s *Service) Context() *context.ECSContext {
	return s.ecsContext
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
	_, err := entity.GetOrCreateTaskDefinition(s)
	if err != nil {
		return err
	}
	err = entity.OptionallyCreateLogs(s)
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
	err := entity.OptionallyCreateLogs(s)
	if err != nil {
		return err
	}
	return s.startService()
}

// Up creates the task definition and service and starts the containers if necessary.
// It does so by calling DescribeService to see if it exists, then calls Create() and Start().
// Otherwise, if the compose or ecs-params files have changed, it will update
// the existing service with the new task definition by calling UpdateService
// with the new task definition and service parameters.
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
	newTaskDefinition, err := entity.GetOrCreateTaskDefinition(s)
	if err != nil {
		return err
	}

	err = entity.OptionallyCreateLogs(s)
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
		err = waitForServiceDescribable(s)
		if err != nil {
			return err
		}
		return s.startService()
	}

	// Update Existing Service
	if err = s.updateService(ecsService, newTaskDefinition); err != nil {
		return err
	}

	// Update Service Discovery
	if s.Context().CLIContext.Bool(flags.UpdateServiceDiscoveryFlag) {
		return servicediscoveryUpdate(aws.StringValue(newTaskDefinition.NetworkMode), entity.GetServiceName(s), s.Context())
	}

	return nil
}

func (s *Service) buildUpdateServiceInput(count *int64, serviceName, taskDefinition string) (*ecs.UpdateServiceInput, error) {
	cluster := s.Context().CommandConfig.Cluster
	deploymentConfig := s.DeploymentConfig()
	forceDeployment := s.Context().CLIContext.Bool(flags.ForceDeploymentFlag)
	networkConfig, err := composeutils.ConvertToECSNetworkConfiguration(s.ecsContext.ECSParams)
	if err != nil {
		return nil, err
	}

	input := &ecs.UpdateServiceInput{
		DesiredCount:            count,
		Service:                 aws.String(serviceName),
		Cluster:                 aws.String(cluster),
		DeploymentConfiguration: deploymentConfig,
		ForceNewDeployment:      &forceDeployment,
	}

	if s.healthCheckGP != nil {
		input.HealthCheckGracePeriodSeconds = aws.Int64(*s.healthCheckGP)
	}

	if networkConfig != nil {
		input.NetworkConfiguration = networkConfig
	}

	if taskDefinition != "" {
		input.TaskDefinition = aws.String(taskDefinition)
	}

	return input, nil
}

func (s *Service) updateService(ecsService *ecs.Service, newTaskDefinition *ecs.TaskDefinition) error {
	if s.Context().CLIContext.Bool(flags.EnableServiceDiscoveryFlag) {
		return fmt.Errorf("Service Discovery can not be enabled on an existing ECS Service")
	}

	schedulingStrategy := strings.ToUpper(s.Context().CLIContext.String(flags.SchedulingStrategyFlag))
	if schedulingStrategy != "" && schedulingStrategy != aws.StringValue(ecsService.SchedulingStrategy) {
		return fmt.Errorf("Scheduling Strategy cannot be updated on an existing ECS Service")
	}

	ecsServiceName := aws.StringValue(ecsService.ServiceName)
	if s.loadBalancer != nil {
		log.WithFields(log.Fields{
			"serviceName": ecsServiceName,
		}).Warn("You cannot update the load balancer configuration on an existing service.")
	}

	oldCount := aws.Int64Value(ecsService.DesiredCount)
	newCount := int64(1)
	count := &newCount
	if oldCount != 0 {
		count = &oldCount // get the current non-zero count
	}

	// if both the task definitions are the same, call update with the new count
	oldTaskDefinitionId := entity.GetIdFromArn(ecsService.TaskDefinition)
	newTaskDefinitionId := entity.GetIdFromArn(newTaskDefinition.TaskDefinitionArn)

	if aws.StringValue(ecsService.SchedulingStrategy) == ecs.SchedulingStrategyDaemon {
		count = nil
	}

	if oldTaskDefinitionId == newTaskDefinitionId {
		return s.updateServiceCount(count)
	}

	// if the task definitions were different, updateService with new task definition
	// this creates a deployment in ECS and slowly takes down the containers with old ones and starts new ones

	updateServiceInput, err := s.buildUpdateServiceInput(count, ecsServiceName, newTaskDefinitionId)
	if err != nil {
		return err
	}

	err = s.Context().ECSClient.UpdateService(updateServiceInput)
	if err != nil {
		return err
	}

	message := "Updated the ECS service with a new task definition. " +
		"Old containers will be stopped automatically, and replaced with new ones"
	s.logUpdateService(updateServiceInput, message)

	return waitForServiceTasks(s, ecsServiceName)
}

// Info returns a formatted list of containers (running and stopped) started by this service
func (s *Service) Info(filterProjectTasks bool) (project.InfoSet, error) {
	// filterProjectTasks is not honored for services, because ECS Services have their
	// own custom Group field, overriding that with startedBy=project will result in no tasks
	// We should instead filter by ServiceName=service
	return entity.Info(s, false)
}

// Scale the service desired count to be the specified count
func (s *Service) Scale(count int) error {
	return s.updateServiceCount(aws.Int64(int64(count)))
}

// Stop stops all the containers in the service by calling ECS.UpdateService(count=0)
// TODO, Store the current desiredCount in a cache, so that number of tasks(group of containers) can be started again
func (s *Service) Stop() error {
	return s.updateServiceCount(aws.Int64(0))
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
	if aws.Int64Value(ecsService.DesiredCount) != 0 && aws.StringValue(ecsService.SchedulingStrategy) != ecs.SchedulingStrategyDaemon {
		if err = s.Stop(); err != nil {
			return err
		}
	}

	// deleteService
	if err = s.Context().ECSClient.DeleteService(ecsServiceName); err != nil {
		return err
	}
	if err = waitForServiceTasks(s, ecsServiceName); err != nil {
		return err
	}

	// delete Service Discovery resources if they exist
	if len(ecsService.ServiceRegistries) > 0 {
		log.Info("Trying to delete any Service Discovery Resources that were created by the ECS CLI...")
		registryArn := aws.StringValue(ecsService.ServiceRegistries[0].RegistryArn)
		if err = s.deleteServiceDiscoveryResources(registryArn, ecsServiceName); err != nil {
			// SD deletion errors are logged but aren't fatal.
			log.Errorf("Problem deleting Service Discovery resources: %v", err)
		}
	}

	return nil
}

func (s *Service) deleteServiceDiscoveryResources(registryArn, ecsServiceName string) error {
	sdsID := getSDSIDFromArn(registryArn)
	if err := waitUntilSDSDeletable(sdsID, s.Context().CommandConfig); err != nil {
		return err
	}
	return servicediscoveryDelete(ecsServiceName, s.Context())

}

// Run expects to issue a command override and start containers. But that doesnt apply to the context of ECS Services
func (s *Service) Run(commandOverrides map[string][]string) error {
	return composeutils.ErrUnsupported
}

// EntityType returns service as the type
func (s *Service) EntityType() types.Type {
	return types.Service
}

// ----------- Commands' helper functions --------

func (s *Service) buildCreateServiceInput(serviceName, taskDefName string) (*ecs.CreateServiceInput, error) {
	launchType := s.Context().CommandConfig.LaunchType
	cluster := s.Context().CommandConfig.Cluster
	ecsParams := s.ecsContext.ECSParams
	schedulingStrategy := strings.ToUpper(s.Context().CLIContext.String(flags.SchedulingStrategyFlag))

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

	// NOTE: this validation is not useful if called after GetOrCreateTaskDefinition()
	if err = entity.ValidateFargateParams(s.Context().ECSParams, launchType); err != nil {
		return nil, err
	}

	if s.healthCheckGP != nil && s.loadBalancer == nil {
		return nil, fmt.Errorf("--%v is only valid for services configured to use load balancers", flags.HealthCheckGracePeriodFlag)
	}

	createServiceInput := &ecs.CreateServiceInput{
		DesiredCount:            aws.Int64(0),            // Required unless DAEMON schedulingStrategy
		ServiceName:             aws.String(serviceName), // Required
		TaskDefinition:          aws.String(taskDefName), // Required
		Cluster:                 aws.String(cluster),
		DeploymentConfiguration: s.deploymentConfig,
		LoadBalancers:           []*ecs.LoadBalancer{s.loadBalancer},
		Role:                    aws.String(s.role),
	}

	if schedulingStrategy != "" {
		createServiceInput.SchedulingStrategy = aws.String(schedulingStrategy)
		if schedulingStrategy == ecs.SchedulingStrategyDaemon {
			createServiceInput.DesiredCount = nil
		}
	}

	if s.healthCheckGP != nil {
		createServiceInput.HealthCheckGracePeriodSeconds = aws.Int64(*s.healthCheckGP)
	}

	if len(s.serviceRegistries) > 0 {
		createServiceInput.ServiceRegistries = s.serviceRegistries
	}

	if networkConfig != nil {
		createServiceInput.NetworkConfiguration = networkConfig
	}

	if placementConstraints != nil {
		createServiceInput.PlacementConstraints = placementConstraints
	}

	if placementStrategy != nil {
		createServiceInput.PlacementStrategy = placementStrategy
	}

	if launchType != "" {
		createServiceInput.LaunchType = aws.String(launchType)
	}

	if err = createServiceInput.Validate(); err != nil {
		return nil, err
	}

	return createServiceInput, nil
}

func (s *Service) logCreateService(serviceName, taskDefName string) {
	fields := log.Fields{
		"service":        serviceName,
		"taskDefinition": taskDefName,
	}
	if s.deploymentConfig != nil && s.deploymentConfig.MaximumPercent != nil {
		fields["deployment-max-percent"] = aws.Int64Value(s.deploymentConfig.MaximumPercent)
	}
	if s.deploymentConfig != nil && s.deploymentConfig.MinimumHealthyPercent != nil {
		fields["deployment-min-healthy-percent"] = aws.Int64Value(s.deploymentConfig.MinimumHealthyPercent)
	}
	if s.healthCheckGP != nil {
		fields["health-check-grace-period"] = *s.healthCheckGP
	}

	log.WithFields(fields).Info("Created an ECS service")
}

// createService calls the underlying ECS.CreateService
func (s *Service) createService() error {
	serviceName := entity.GetServiceName(s)
	taskDefName := entity.GetIdFromArn(s.TaskDefinition().TaskDefinitionArn)

	cliContext := s.Context().CLIContext

	if cliContext.Bool(flags.EnableServiceDiscoveryFlag) {
		networkMode := aws.StringValue(s.TaskDefinition().NetworkMode)

		serviceRegistry, err := servicediscoveryCreate(networkMode, serviceName, s.Context())
		if err != nil {
			return err
		}

		s.serviceRegistries = []*ecs.ServiceRegistry{
			serviceRegistry,
		}
	}

	// Create request input
	createServiceInput, err := s.buildCreateServiceInput(serviceName, taskDefName)
	if err != nil {
		return err
	}

	defer s.logCreateService(serviceName, taskDefName)

	// Call ECS Client
	err = s.Context().ECSClient.CreateService(createServiceInput)
	if err != nil {
		return err
	}

	return nil
}

// describeService calls underlying ECS.DescribeService and expects the service to be present,
// returns error otherwise
func (s *Service) describeService() (*ecs.Service, error) {
	serviceName := entity.GetServiceName(s)
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
func (s *Service) startService() error {
	ecsService, err := s.describeService()
	if err != nil {
		// Describe API returns the failures for resources in the response (instead of returning an error)
		// Read the custom error returned from describeService to see if the resource was missing
		if strings.Contains(err.Error(), ecsMissingResourceCode) {
			return fmt.Errorf("Please use '%s' command to create the service '%s' first",
				flags.CreateServiceCommandName, entity.GetServiceName(s))
		}
		return err
	}

	serviceName := aws.StringValue(ecsService.ServiceName)
	desiredCount := aws.Int64Value(ecsService.DesiredCount)
	forceDeployment := s.Context().CLIContext.Bool(flags.ForceDeploymentFlag)
	schedulingStrategy := aws.StringValue(ecsService.SchedulingStrategy)
	if desiredCount != 0 || schedulingStrategy == ecs.SchedulingStrategyDaemon {
		if forceDeployment {
			log.WithFields(log.Fields{
				"serviceName":        serviceName,
				"desiredCount":       desiredCount,
				"schedulingStrategy": schedulingStrategy,
				"force-deployment":   strconv.FormatBool(forceDeployment),
			}).Info("Forcing new deployment of running ECS Service")
			count := aws.Int64(desiredCount)
			if schedulingStrategy == ecs.SchedulingStrategyDaemon {
				count = nil
			}
			return s.updateServiceCount(count)
		}
		//NoOp
		log.WithFields(log.Fields{
			"serviceName":        serviceName,
			"desiredCount":       desiredCount,
			"schedulingStrategy": schedulingStrategy,
		}).Info("ECS Service is already running")

		return waitForServiceTasks(s, serviceName)
	}
	return s.updateServiceCount(aws.Int64(1))
}

// updateServiceCount calls the underlying ECS.UpdateService with the specified count
// NOTE: If network configuration has changed in ECS Params, this will also be updated
func (s *Service) updateServiceCount(count *int64) error {
	serviceName := entity.GetServiceName(s)

	updateServiceInput, err := s.buildUpdateServiceInput(count, serviceName, "")
	if err != nil {
		return err
	}

	if err = s.Context().ECSClient.UpdateService(updateServiceInput); err != nil {
		return err
	}

	s.logUpdateService(updateServiceInput, "Updated ECS service successfully")

	return waitForServiceTasks(s, serviceName)
}

func (s *Service) logUpdateService(input *ecs.UpdateServiceInput, message string) {
	fields := log.Fields{
		"service":      aws.StringValue(input.Service),
		"desiredCount": aws.Int64Value(input.DesiredCount),
	}
	if s.deploymentConfig != nil && s.deploymentConfig.MaximumPercent != nil {
		fields["deployment-max-percent"] = aws.Int64Value(s.deploymentConfig.MaximumPercent)
	}
	if s.deploymentConfig != nil && s.deploymentConfig.MinimumHealthyPercent != nil {
		fields["deployment-min-healthy-percent"] = aws.Int64Value(s.deploymentConfig.MinimumHealthyPercent)
	}
	if s.healthCheckGP != nil {
		fields["health-check-grace-period"] = *s.healthCheckGP
	}
	if input.ForceNewDeployment != nil {
		fields["force-deployment"] = aws.BoolValue(input.ForceNewDeployment)
	}

	log.WithFields(fields).Info(message)
}

func getSDSIDFromArn(sdsARN string) string {
	return strings.Split(sdsARN, "/")[1]
}
