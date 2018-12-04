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
	"flag"
	"strconv"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

func TestCreateWithDeploymentConfig(t *testing.T) {
	deploymentMaxPercent := 200
	deploymentMinPercent := 100

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.DeploymentMaxPercentFlag, strconv.Itoa(deploymentMaxPercent), "")
	flagSet.String(flags.DeploymentMinHealthyPercentFlag, strconv.Itoa(deploymentMinPercent), "")

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			actual := input.DeploymentConfiguration
			assert.Equal(t, int64(deploymentMaxPercent), aws.Int64Value(actual.MaximumPercent), "DeploymentConfig.MaxPercent should match")
			assert.Equal(t, int64(deploymentMinPercent), aws.Int64Value(actual.MinimumHealthyPercent), "DeploymentConfig.MinimumHealthyPercent should match")
		},
	)
}

func TestCreateWithoutDeploymentConfig(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			actual := input.DeploymentConfiguration
			assert.Nil(t, actual.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, actual.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
	)
}

func ecsParamsWithNetworkConfig() *utils.ECSParams {
	return &utils.ECSParams{
		TaskDefinition: utils.EcsTaskDef{
			NetworkMode: "awsvpc",
		},
		RunParams: utils.RunParams{
			NetworkConfiguration: utils.NetworkConfiguration{
				AwsVpcConfiguration: utils.AwsVpcConfiguration{
					Subnets: []string{"sg-bafff1ed", "sg-c0ffeefe"},
				},
			},
		},
	}
}

func TestCreateWithNetworkConfig(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		ecsParamsWithNetworkConfig(),
		func(input *ecs.CreateServiceInput) {
			launchType := input.LaunchType
			assert.NotEqual(t, "FARGATE", launchType)

			networkConfig := input.NetworkConfiguration
			assert.NotNil(t, networkConfig, "NetworkConfiguration should not be nil")
			assert.Equal(t, 2, len(networkConfig.AwsvpcConfiguration.Subnets))
			assert.Nil(t, networkConfig.AwsvpcConfiguration.AssignPublicIp)
		},
	)
}

func ecsParamsWithFargateNetworkConfig() *utils.ECSParams {
	return &utils.ECSParams{
		TaskDefinition: utils.EcsTaskDef{
			ExecutionRole: "arn:aws:iam::123456789012:role/fargate_role",
			NetworkMode:   "awsvpc",
			TaskSize: utils.TaskSize{
				Cpu:    "512",
				Memory: "1GB",
			},
		},
		RunParams: utils.RunParams{
			NetworkConfiguration: utils.NetworkConfiguration{
				AwsVpcConfiguration: utils.AwsVpcConfiguration{
					Subnets:        []string{"sg-bafff1ed", "sg-c0ffeefe"},
					AssignPublicIp: utils.Enabled,
				},
			},
		},
	}
}

func TestCreateFargate(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{LaunchType: "FARGATE"},
		ecsParamsWithFargateNetworkConfig(),
		func(input *ecs.CreateServiceInput) {
			launchType := input.LaunchType
			assert.Equal(t, "FARGATE", aws.StringValue(launchType))

			networkConfig := input.NetworkConfiguration
			assert.NotNil(t, networkConfig, "NetworkConfiguration should not be nil")
			assert.Equal(t, 2, len(networkConfig.AwsvpcConfiguration.Subnets))
			assert.Equal(t, string(utils.Enabled), aws.StringValue(networkConfig.AwsvpcConfiguration.AssignPublicIp))
		},
	)
}

func TestCreateFargateNetworkModeNotAWSVPC(t *testing.T) {
	taskDefID := "taskDefinitionId"
	taskDefArn := "arn/" + taskDefID

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)

	taskDefinition := ecs.TaskDefinition{
		Family:               aws.String("family"),
		ContainerDefinitions: []*ecs.ContainerDefinition{},
		Volumes:              []*ecs.Volume{},
	}
	respTaskDef := taskDefinition
	respTaskDef.TaskDefinitionArn = aws.String(taskDefArn)

	gomock.InOrder(
		mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			// verify input fields
			req := x.(*ecs.RegisterTaskDefinitionInput)
			assert.Equal(t, aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family), "Task Definition family should match")
		}).Return(&respTaskDef, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)

	context := &context.ECSContext{
		ECSClient:     mockEcs,
		CommandConfig: &config.CommandConfig{LaunchType: "FARGATE"},
		CLIContext:    cliContext,
		ECSParams:     &utils.ECSParams{},
	}

	service := NewService(context)
	err := service.LoadContext()
	assert.NoError(t, err, "Unexpected error while loading context in create service test")

	service.SetTaskDefinition(&taskDefinition)
	err = service.Create()
	assert.Error(t, err, "Expected error creating service")
}

func TestCreateEC2Explicitly(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{LaunchType: "EC2"},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			launchType := input.LaunchType
			assert.Equal(t, "EC2", aws.StringValue(launchType))

			networkConfig := input.NetworkConfiguration
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
		},
	)
}

func TestCreateWithTaskPlacement(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		ecsParamsWithTaskPlacement(),
		func(input *ecs.CreateServiceInput) {
			placementConstraints := input.PlacementConstraints
			placementStrategy := input.PlacementStrategy
			expectedConstraints := []*ecs.PlacementConstraint{
				{
					Type: aws.String("distinctInstance"),
				}, {
					Expression: aws.String("attribute:ecs.instance-type =~ t2.*"),
					Type:       aws.String("memberOf"),
				},
			}
			expectedStrategy := []*ecs.PlacementStrategy{
				{
					Type: aws.String("random"),
				}, {
					Field: aws.String("instanceId"),
					Type:  aws.String("binpack"),
				},
			}

			assert.Len(t, placementConstraints, 2)
			assert.Equal(t, expectedConstraints, placementConstraints, "Expected Placement Constraints to match")
			assert.Len(t, placementStrategy, 2)
			assert.Equal(t, expectedStrategy, placementStrategy, "Expected Placement Strategy to match")
		},
	)
}

func ecsParamsWithTaskPlacement() *utils.ECSParams {
	return &utils.ECSParams{
		RunParams: utils.RunParams{
			TaskPlacement: utils.TaskPlacement{
				Constraints: []utils.Constraint{
					utils.Constraint{
						Type: ecs.PlacementConstraintTypeDistinctInstance,
					},
					utils.Constraint{
						Expression: "attribute:ecs.instance-type =~ t2.*",
						Type:       ecs.PlacementConstraintTypeMemberOf,
					},
				},
				Strategies: []utils.Strategy{
					utils.Strategy{
						Type: ecs.PlacementStrategyTypeRandom,
					},
					utils.Strategy{
						Field: "instanceId",
						Type:  ecs.PlacementStrategyTypeBinpack,
					},
				},
			},
		},
	}
}

// Specifies TargeGroupArn to test ALB
func TestCreateWithALB(t *testing.T) {
	targetGroupArn := "targetGroupArn"
	containerName := "containerName"
	containerPort := 80
	role := "role"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TargetGroupArnFlag, targetGroupArn, "")
	flagSet.String(flags.ContainerNameFlag, containerName, "")
	flagSet.String(flags.ContainerPortFlag, strconv.Itoa(containerPort), "")
	flagSet.String(flags.RoleFlag, role, "")

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			loadBalancer := input.LoadBalancers[0]
			observedRole := input.Role

			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.LoadBalancerName, "LoadBalancer.LoadBalancerName should be nil")
			assert.Equal(t, targetGroupArn, aws.StringValue(loadBalancer.TargetGroupArn), "LoadBalancer.TargetGroupArn should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, aws.StringValue(observedRole), "Role should match")
		},
	)
}

func TestCreateWithHealthCheckGracePeriodAndALB(t *testing.T) {
	targetGroupArn := "targetGroupArn"
	containerName := "containerName"
	containerPort := 80
	role := "role"
	healthCheckGP := 60

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TargetGroupArnFlag, targetGroupArn, "")
	flagSet.String(flags.ContainerNameFlag, containerName, "")
	flagSet.String(flags.ContainerPortFlag, strconv.Itoa(containerPort), "")
	flagSet.String(flags.RoleFlag, role, "")
	flagSet.String(flags.HealthCheckGracePeriodFlag, strconv.Itoa(healthCheckGP), "")

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			loadBalancer := input.LoadBalancers[0]
			observedRole := input.Role
			healthCheckGracePeriod := input.HealthCheckGracePeriodSeconds

			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.LoadBalancerName, "LoadBalancer.LoadBalancerName should be nil")
			assert.Equal(t, targetGroupArn, aws.StringValue(loadBalancer.TargetGroupArn), "LoadBalancer.TargetGroupArn should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, aws.StringValue(observedRole), "Role should match")
			assert.Equal(t, int64(healthCheckGP), *healthCheckGracePeriod, "HealthCheckGracePeriod should match")
		},
	)
}

// Specifies LoadBalancerName to test ELB
func TestCreateWithELB(t *testing.T) {
	loadbalancerName := "loadbalancerName"
	containerName := "containerName"
	containerPort := 80
	role := "role"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.LoadBalancerNameFlag, loadbalancerName, "")
	flagSet.String(flags.ContainerNameFlag, containerName, "")
	flagSet.String(flags.ContainerPortFlag, strconv.Itoa(containerPort), "")
	flagSet.String(flags.RoleFlag, role, "")

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			loadBalancer := input.LoadBalancers[0]
			observedRole := input.Role

			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.TargetGroupArn, "LoadBalancer.TargetGroupArn should be nil")
			assert.Equal(t, loadbalancerName, aws.StringValue(loadBalancer.LoadBalancerName), "LoadBalancer.LoadBalancerName should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, aws.StringValue(observedRole), "Role should match")
		},
	)
}

func TestCreateWithHealthCheckGracePeriodAndELB(t *testing.T) {
	loadbalancerName := "loadbalancerName"
	containerName := "containerName"
	containerPort := 80
	role := "role"
	healthCheckGP := 60

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.LoadBalancerNameFlag, loadbalancerName, "")
	flagSet.String(flags.ContainerNameFlag, containerName, "")
	flagSet.String(flags.ContainerPortFlag, strconv.Itoa(containerPort), "")
	flagSet.String(flags.RoleFlag, role, "")
	flagSet.String(flags.HealthCheckGracePeriodFlag, strconv.Itoa(healthCheckGP), "")

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			loadBalancer := input.LoadBalancers[0]
			observedRole := input.Role
			healthCheckGracePeriod := input.HealthCheckGracePeriodSeconds

			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.TargetGroupArn, "LoadBalancer.TargetGroupArn should be nil")
			assert.Equal(t, loadbalancerName, aws.StringValue(loadBalancer.LoadBalancerName), "LoadBalancer.LoadBalancerName should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, aws.StringValue(observedRole), "Role should match")
			assert.Equal(t, int64(healthCheckGP), *healthCheckGracePeriod, "HealthCheckGracePeriod should match")
		},
	)
}

type validateCreateServiceInputField func(*ecs.CreateServiceInput)

func createServiceTest(t *testing.T,
	flagSet *flag.FlagSet,
	commandConfig *config.CommandConfig,
	ecsParams *utils.ECSParams,
	validateInput validateCreateServiceInputField) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskDefID := "taskDefinitionId"
	taskDefArn, taskDefinition, registerTaskDefResponse := getTestTaskDef(taskDefID)

	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	gomock.InOrder(
		mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			// verify input fields
			req := x.(*ecs.RegisterTaskDefinitionInput)
			assert.Equal(t, aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family), "Task Definition family should match")
		}).Return(&registerTaskDefResponse, nil),

		mockEcs.EXPECT().CreateService(
			gomock.Any(),
			gomock.Any(),
		).Do(func(serviceName, input interface{}) {
			req := input.(*ecs.CreateServiceInput)
			validateInput(req)
		}).Return(nil),
	)

	cliContext := cli.NewContext(nil, flagSet, nil)

	context := &context.ECSContext{
		ECSClient:     mockEcs,
		CommandConfig: commandConfig,
		CLIContext:    cliContext,
		ECSParams:     ecsParams,
	}

	service := NewService(context)
	err := service.LoadContext()
	assert.NoError(t, err, "Unexpected error while loading context in create service test")

	service.SetTaskDefinition(&taskDefinition)
	err = service.Create()
	assert.NoError(t, err, "Unexpected error while create")

	// task definition should be set
	assert.Equal(t, taskDefArn, aws.StringValue(service.TaskDefinition().TaskDefinitionArn), "TaskDefArn should match")
}

func TestLoadContext(t *testing.T) {
	deploymentMaxPercent := 150

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.DeploymentMaxPercentFlag, strconv.Itoa(deploymentMaxPercent), "")
	cliContext := cli.NewContext(nil, flagSet, nil)
	service := &Service{
		ecsContext: &context.ECSContext{CLIContext: cliContext},
	}

	err := service.LoadContext()
	assert.NoError(t, err, "Unexpected error while loading context in load context test")

	observedDeploymentConfig := service.DeploymentConfig()

	assert.Equal(t, int64(deploymentMaxPercent),
		aws.Int64Value(observedDeploymentConfig.MaximumPercent),
		"DeploymentConfig.MaxPercent should match")
	assert.Nil(t, observedDeploymentConfig.MinimumHealthyPercent,
		"DeploymentConfig.MinimumHealthyPercent should be nil")

}

func TestLoadContextForIncorrectInput(t *testing.T) {
	deploymentMaxPercent := "string"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.DeploymentMaxPercentFlag, deploymentMaxPercent, "")
	cliContext := cli.NewContext(nil, flagSet, nil)
	service := &Service{
		ecsContext: &context.ECSContext{CLIContext: cliContext},
	}

	err := service.LoadContext()
	assert.Error(t, err, "Expected error to load context when flag is a string but got done")
}

func TestLoadContextForLoadBalancerInputError(t *testing.T) {
	targetGroupArn := "targetGroupArn"
	loadBalancerName := "loadBalancerName"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.TargetGroupArnFlag, targetGroupArn, "")
	flagSet.String(flags.LoadBalancerNameFlag, loadBalancerName, "")
	cliContext := cli.NewContext(nil, flagSet, nil)
	service := &Service{
		ecsContext: &context.ECSContext{CLIContext: cliContext},
	}

	err := service.LoadContext()
	assert.Error(t, err, "Expected error to load context when flag is a string but got done")
}

func TestServiceInfo(t *testing.T) {
	entity.TestInfo(func(context *context.ECSContext) entity.ProjectEntity {
		return NewService(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Contains(t, aws.StringValue(req.ServiceName), projectName, "ServiceName should contain ProjectName")
		assert.Nil(t, req.StartedBy, "StartedBy should be nil")
	}, t, true)
}

func TestServiceRun(t *testing.T) {
	service := NewService(&context.ECSContext{})
	err := service.Run(map[string][]string{})
	assert.Error(t, err, "Expected unsupported error")
}

// Up Service tests

type UpdateServiceParams struct {
	serviceName            string
	taskDefinition         string
	count                  int64
	deploymentConfig       *ecs.DeploymentConfiguration
	networkConfig          *ecs.NetworkConfiguration
	healthCheckGracePeriod *int64
	forceDeployment        bool
}

func TestDelayedServiceCreate(t *testing.T) {
	// define test flag set
	timeoutFlagValue := 1

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.ComposeServiceTimeOutFlag, strconv.Itoa(timeoutFlagValue), "")
	cliContext := cli.NewContext(nil, flagSet, nil)

	// call tests
	upNewServiceWithDelay(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{})
}

func TestUpdateExistingServiceWithForceFlag(t *testing.T) {
	// define test flag set
	forceFlagValue := true

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.ForceDeploymentFlag, forceFlagValue, "")
	cliContext := cli.NewContext(nil, flagSet, nil)

	// define existing service
	serviceName := "test-service"
	existingService := &ecs.Service{
		TaskDefinition: aws.String("arn/test-task-def"),
		Status:         aws.String("ACTIVE"),
		DesiredCount:   aws.Int64(0),
		ServiceName:    aws.String(serviceName),
	}

	// define expected client input given the above info
	expectedInput := getDefaultUpdateInput()
	expectedInput.serviceName = serviceName
	expectedInput.forceDeployment = forceFlagValue

	// call tests
	upServiceWithCurrentTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
	upServiceWithNewTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
}

func TestUpdateExistingServiceWithNewDeploymentConfig(t *testing.T) {
	// define test flag set
	deploymentMaxPercent := 200
	deploymentMinPercent := 100

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.DeploymentMaxPercentFlag, strconv.Itoa(deploymentMaxPercent), "")
	flagSet.String(flags.DeploymentMinHealthyPercentFlag, strconv.Itoa(deploymentMinPercent), "")
	cliContext := cli.NewContext(nil, flagSet, nil)

	// define existing service
	serviceName := "test-service"
	existingService := &ecs.Service{
		TaskDefinition: aws.String("arn/test-task-def"),
		Status:         aws.String("ACTIVE"),
		DesiredCount:   aws.Int64(0),
		ServiceName:    aws.String(serviceName),
	}

	// define expected client input given the above info
	expectedInput := getDefaultUpdateInput()
	expectedInput.serviceName = serviceName
	expectedInput.deploymentConfig = &ecs.DeploymentConfiguration{
		MaximumPercent:        aws.Int64(int64(deploymentMaxPercent)),
		MinimumHealthyPercent: aws.Int64(int64(deploymentMinPercent)),
	}

	// call tests
	upServiceWithCurrentTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
	upServiceWithNewTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
}

func TestUpdateExistingServiceWithNewHCGP(t *testing.T) {
	// define test flag set
	healthCheckGracePeriod := 200

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.HealthCheckGracePeriodFlag, strconv.Itoa(healthCheckGracePeriod), "")
	cliContext := cli.NewContext(nil, flagSet, nil)

	// define existing service
	serviceName := "test-service"
	existingService := &ecs.Service{
		TaskDefinition: aws.String("arn/test-task-def"),
		Status:         aws.String("ACTIVE"),
		DesiredCount:   aws.Int64(0),
		ServiceName:    aws.String(serviceName),
		LoadBalancers:  []*ecs.LoadBalancer{}, // LB required for HCGP, but not verified before calling client
	}

	// define expected client input given the above info
	expectedInput := getDefaultUpdateInput()
	expectedInput.serviceName = serviceName
	expectedInput.healthCheckGracePeriod = aws.Int64(int64(healthCheckGracePeriod))

	// call tests
	upServiceWithCurrentTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
	upServiceWithNewTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
}

func TestUpdateExistingServiceWithDesiredCountOverOne(t *testing.T) {
	// define test values
	existingDesiredCount := 2

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)

	// define existing service
	serviceName := "test-service"
	existingService := &ecs.Service{
		TaskDefinition: aws.String("arn/test-task-def"),
		Status:         aws.String("ACTIVE"),
		DesiredCount:   aws.Int64(int64(existingDesiredCount)), // existing count > 1
		ServiceName:    aws.String(serviceName),
	}

	// define expected client input given the above info
	expectedInput := getDefaultUpdateInput()
	expectedInput.serviceName = serviceName
	expectedInput.count = *aws.Int64(int64(existingDesiredCount))

	// call tests
	upServiceWithCurrentTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
	upServiceWithNewTaskDefTest(t, cliContext, &config.CommandConfig{}, &utils.ECSParams{}, expectedInput, existingService)
}

func TestCreateWithServiceDiscovery(t *testing.T) {
	sdsARN := "arn:aws:servicediscovery:eu-west-1:11111111111:service/srv-clydelovespudding"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.EnableServiceDiscoveryFlag, true, "")

	// Reset mockable function after test
	nonMockedServicediscoveryCreate := servicediscoveryCreate
	defer func() { servicediscoveryCreate = nonMockedServicediscoveryCreate }()

	servicediscoveryCreate = func(networkMode, serviceName string, c *context.ECSContext) (*ecs.ServiceRegistry, error) {
		return &ecs.ServiceRegistry{
			RegistryArn: aws.String(sdsARN),
		}, nil
	}

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			actualServiceRegistries := input.ServiceRegistries
			assert.Len(t, actualServiceRegistries, 1, "Expected a single Service Registry")
			assert.Equal(t, sdsARN, aws.StringValue(actualServiceRegistries[0].RegistryArn), "Service Registry should match")
		},
	)
}

func TestCreateWithServiceDiscoveryWithContainerNameAndPort(t *testing.T) {
	sdsARN := "arn:aws:servicediscovery:eu-west-1:11111111111:service/srv-clydelovespudding"
	containerName := "nginx"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.EnableServiceDiscoveryFlag, true, "")

	// Reset mockable function after test
	nonMockedServicediscoveryCreate := servicediscoveryCreate
	defer func() { servicediscoveryCreate = nonMockedServicediscoveryCreate }()

	servicediscoveryCreate = func(networkMode, serviceName string, c *context.ECSContext) (*ecs.ServiceRegistry, error) {
		return &ecs.ServiceRegistry{
			RegistryArn:   aws.String(sdsARN),
			ContainerName: aws.String(containerName),
			ContainerPort: aws.Int64(80),
		}, nil
	}

	createServiceTest(
		t,
		flagSet,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(input *ecs.CreateServiceInput) {
			actualServiceRegistries := input.ServiceRegistries
			assert.Len(t, actualServiceRegistries, 1, "Expected a single Service Registry")
			assert.Equal(t, int64(80), aws.Int64Value(actualServiceRegistries[0].ContainerPort), "Expected container port to be 80")
			assert.Equal(t, containerName, aws.StringValue(actualServiceRegistries[0].ContainerName), "Expected ContainerName to match")
			assert.Equal(t, sdsARN, aws.StringValue(actualServiceRegistries[0].RegistryArn), "Service Registry should match")
		},
	)
}

func getDefaultUpdateInput() UpdateServiceParams {
	return UpdateServiceParams{
		deploymentConfig: &ecs.DeploymentConfiguration{},
		count:            1,
	}
}

func getUpdateServiceMockClient(t *testing.T,
	ctrl *gomock.Controller,
	describeServiceResponse *ecs.DescribeServicesOutput,
	taskDefinition ecs.TaskDefinition,
	registerTaskDefResponse ecs.TaskDefinition,
	expectedInput UpdateServiceParams) *mock_ecs.MockECSClient {

	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	gomock.InOrder(
		mockEcs.EXPECT().DescribeService(gomock.Any()).Return(describeServiceResponse, nil),

		mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			verifyTaskDefinitionInput(t, taskDefinition, x.(*ecs.RegisterTaskDefinitionInput))
		}).Return(&registerTaskDefResponse, nil),

		mockEcs.EXPECT().UpdateService(
			gomock.Any(), // serviceName
			gomock.Any(), // taskDefinition
			gomock.Any(), // count
			gomock.Any(), // deploymentConfig
			gomock.Any(), // networkConfig
			gomock.Any(), // healthCheckGracePeriod
			gomock.Any(), // force
		).Do(func(a, b, c, d, e, f, g interface{}) {
			// validate the client is called with the expected inputs
			observedInput := UpdateServiceParams{
				serviceName:            a.(string),
				taskDefinition:         b.(string),
				count:                  c.(int64),
				deploymentConfig:       d.(*ecs.DeploymentConfiguration),
				networkConfig:          e.(*ecs.NetworkConfiguration),
				healthCheckGracePeriod: f.(*int64),
				forceDeployment:        g.(bool),
			}
			assert.Equal(t, expectedInput, observedInput)

		}).Return(nil),
	)
	return mockEcs
}

func getCreateServiceWithDelayMockClient(t *testing.T,
	ctrl *gomock.Controller,
	taskDefinition ecs.TaskDefinition,
	taskDefID string,
	registerTaskDefResponse ecs.TaskDefinition,
) *mock_ecs.MockECSClient {

	mockEcs := mock_ecs.NewMockECSClient(ctrl)

	createdService := &ecs.Service{
		TaskDefinition: aws.String("arn/" + taskDefID),
		Status:         aws.String("ACTIVE"),
		DesiredCount:   aws.Int64(0),
		RunningCount:   aws.Int64(0),
		ServiceName:    aws.String("test-created"),
	}
	updatedService := *createdService
	gomock.InOrder(
		mockEcs.EXPECT().DescribeService(gomock.Any()).Return(getDescribeServiceTestResponse(nil), nil),

		mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			verifyTaskDefinitionInput(t, taskDefinition, x.(*ecs.RegisterTaskDefinitionInput))
		}).Return(&registerTaskDefResponse, nil),

		mockEcs.EXPECT().CreateService(
			gomock.Any(), // serviceName
			gomock.Any(), // taskDefName
		).Do(func(serviceName, input interface{}) {
			req := input.(*ecs.CreateServiceInput)
			observedTaskDefID := req.TaskDefinition
			assert.Equal(t, taskDefID, aws.StringValue(observedTaskDefID), "Task Definition name should match")
		}).Return(nil),

		mockEcs.EXPECT().DescribeService(gomock.Any()).Return(getDescribeServiceTestResponse(nil), nil),
		mockEcs.EXPECT().DescribeService(gomock.Any()).Return(getDescribeServiceTestResponse(createdService), nil).MaxTimes(2),
		mockEcs.EXPECT().UpdateService(
			gomock.Any(), // serviceName
			gomock.Any(), // taskDefinition
			gomock.Any(), // count
			gomock.Any(), // deploymentConfig
			gomock.Any(), // networkConfig
			gomock.Any(), // healthCheckGracePeriod
			gomock.Any(), // force
		).Return(nil),
		mockEcs.EXPECT().DescribeService(gomock.Any()).Return(getDescribeServiceTestResponse(updatedService.SetDeployments([]*ecs.Deployment{&ecs.Deployment{}}).SetDesiredCount(1).SetRunningCount(1)), nil),
	)
	return mockEcs
}

func upServiceWithCurrentTaskDefTest(t *testing.T,
	cliContext *cli.Context,
	commandConfig *config.CommandConfig,
	ecsParams *utils.ECSParams,
	expectedInput UpdateServiceParams,
	existingService *ecs.Service) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// set up expected task def
	taskDefID := "taskDefinitionId"
	if existingService != nil {
		// set to existing if provided
		taskDefID = entity.GetIdFromArn(existingService.TaskDefinition)
	}
	taskDefArn, taskDefinition, registerTaskDefResponse := getTestTaskDef(taskDefID)

	// set up DescribeService() response
	describeServiceResponse := getDescribeServiceTestResponse(existingService)

	mockEcs := getUpdateServiceMockClient(t, ctrl, describeServiceResponse, taskDefinition, registerTaskDefResponse, expectedInput)

	ecsContext := &context.ECSContext{
		ECSClient:     mockEcs,
		CommandConfig: commandConfig,
		CLIContext:    cliContext,
		ECSParams:     ecsParams,
	}
	// if taskDef is unchanged, serviceName is taken from current context
	if existingService != nil {
		ecsContext.ProjectName = *existingService.ServiceName
	}
	service := NewService(ecsContext)
	err := service.LoadContext()
	assert.NoError(t, err, "Unexpected error while loading context in update service with current task def test")

	service.SetTaskDefinition(&taskDefinition)
	err = service.Up()
	assert.NoError(t, err, "Unexpected error on service up with current task def")

	// task definition should be set
	assert.Equal(t, taskDefArn, aws.StringValue(service.TaskDefinition().TaskDefinitionArn), "TaskDefArn should match")
}

func upNewServiceWithDelay(t *testing.T,
	cliContext *cli.Context,
	commandConfig *config.CommandConfig,
	ecsParams *utils.ECSParams) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskDefID := "newTaskDefinitionId"
	taskDefArn, taskDefinition, registerTaskDefResponse := getTestTaskDef(taskDefID)

	mockEcs := getCreateServiceWithDelayMockClient(t, ctrl, taskDefinition, taskDefID, registerTaskDefResponse)

	ecsContext := &context.ECSContext{
		ECSClient:     mockEcs,
		CommandConfig: commandConfig,
		CLIContext:    cliContext,
		ECSParams:     ecsParams,
	}
	ecsContext.SetProjectName()
	service := NewService(ecsContext)
	err := service.LoadContext()
	assert.NoError(t, err, "Unexpected error while loading context in update service with new task def test")

	service.SetTaskDefinition(&taskDefinition)
	err = service.Up()
	assert.NoError(t, err, "Unexpected error on service up with new task def")

	// task definition should be set
	assert.Equal(t, taskDefArn, aws.StringValue(service.TaskDefinition().TaskDefinitionArn), "TaskDefArn should match")
}

func upServiceWithNewTaskDefTest(t *testing.T,
	cliContext *cli.Context,
	commandConfig *config.CommandConfig,
	ecsParams *utils.ECSParams,
	expectedInput UpdateServiceParams,
	existingService *ecs.Service) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// set up expected (new) task def
	taskDefID := "newTaskDefinitionId"
	taskDefArn, taskDefinition, registerTaskDefResponse := getTestTaskDef(taskDefID)

	// expect input to include new task def
	expectedInput.taskDefinition = taskDefID

	// set up DescribeService() response
	describeServiceResponse := getDescribeServiceTestResponse(existingService)

	mockEcs := getUpdateServiceMockClient(t, ctrl, describeServiceResponse, taskDefinition, registerTaskDefResponse, expectedInput)

	ecsContext := &context.ECSContext{
		ECSClient:     mockEcs,
		CommandConfig: commandConfig,
		CLIContext:    cliContext,
		ECSParams:     ecsParams,
	}
	service := NewService(ecsContext)
	err := service.LoadContext()
	assert.NoError(t, err, "Unexpected error while loading context in update service with new task def test")

	service.SetTaskDefinition(&taskDefinition)
	err = service.Up()
	assert.NoError(t, err, "Unexpected error on service up with new task def")

	// task definition should be set
	assert.Equal(t, taskDefArn, aws.StringValue(service.TaskDefinition().TaskDefinitionArn), "TaskDefArn should match")
}

func getDescribeServiceTestResponse(existingService *ecs.Service) *ecs.DescribeServicesOutput {
	describeFailure := &ecs.Failure{
		Reason: aws.String("MISSING"),
		Arn:    aws.String("arn:missing-service"),
	}
	existingServiceResponse := &ecs.DescribeServicesOutput{
		Failures: []*ecs.Failure{},
		Services: []*ecs.Service{existingService},
	}
	emptyDescribeServiceResponse := &ecs.DescribeServicesOutput{
		Failures: []*ecs.Failure{describeFailure},
		Services: []*ecs.Service{},
	}
	if existingService != nil {
		return existingServiceResponse
	}
	return emptyDescribeServiceResponse
}

func getTestTaskDef(taskDefID string) (taskDefArn string, taskDefinition, registerTaskDefResponse ecs.TaskDefinition) {
	taskDefArn = "arn/" + taskDefID
	taskDefinition = ecs.TaskDefinition{
		Family:               aws.String("family"),
		ContainerDefinitions: []*ecs.ContainerDefinition{},
		Volumes:              []*ecs.Volume{},
	}
	registerTaskDefResponse = taskDefinition
	registerTaskDefResponse.TaskDefinitionArn = aws.String(taskDefArn)

	return taskDefArn, taskDefinition, registerTaskDefResponse
}

func verifyTaskDefinitionInput(t *testing.T,
	taskDef ecs.TaskDefinition,
	regInput *ecs.RegisterTaskDefinitionInput) {
	assert.Equal(t, aws.StringValue(taskDef.Family), aws.StringValue(regInput.Family), "Task Definition family should match")
}
