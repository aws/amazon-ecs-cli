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
	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Equal(t, int64(deploymentMaxPercent), aws.Int64Value(deploymentConfig.MaximumPercent), "DeploymentConfig.MaxPercent should match")
			assert.Equal(t, int64(deploymentMinPercent), aws.Int64Value(deploymentConfig.MinimumHealthyPercent), "DeploymentConfig.MinimumHealthyPercent should match")
		},
		func(loadBalancer *ecs.LoadBalancer, role string) {
			assert.Nil(t, loadBalancer, "LoadBalancer should be nil")
			assert.Empty(t, role, "Role should be empty")
		},
		func(launchType string) {
			assert.Empty(t, launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
		},
	)
}

func TestCreateWithoutDeploymentConfig(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, role string) {
			assert.Nil(t, loadBalancer, "LoadBalancer should be nil")
			assert.Empty(t, role, "Role should be empty")
		},
		func(launchType string) {
			assert.Empty(t, launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
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
	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		&config.CommandConfig{},
		ecsParamsWithNetworkConfig(),
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, role string) {
			assert.Nil(t, loadBalancer, "LoadBalancer should be nil")
			assert.Empty(t, role, "Role should be empty")
		},
		func(launchType string) {
			assert.NotEqual(t, "FARGATE", launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
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
	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		&config.CommandConfig{LaunchType: "FARGATE"},
		ecsParamsWithFargateNetworkConfig(),
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, role string) {
			assert.Nil(t, loadBalancer, "LoadBalancer should be nil")
			assert.Empty(t, role, "Role should be empty")
		},
		func(launchType string) {
			assert.Equal(t, "FARGATE", launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
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
		ECSClient:  mockEcs,
		CommandConfig:  &config.CommandConfig{LaunchType: "FARGATE"},
		CLIContext: cliContext,
		ECSParams:  &utils.ECSParams{},
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
	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		&config.CommandConfig{LaunchType: "EC2"},
		&utils.ECSParams{},
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, role string) {
			assert.Nil(t, loadBalancer, "LoadBalancer should be nil")
			assert.Empty(t, role, "Role should be empty")
		},
		func(launchType string) {
			assert.Equal(t, "EC2", launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
		},
	)
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

	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, observedRole string) {
			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.LoadBalancerName, "LoadBalancer.LoadBalancerName should be nil")
			assert.Equal(t, targetGroupArn, aws.StringValue(loadBalancer.TargetGroupArn), "LoadBalancer.TargetGroupArn should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, observedRole, "Role should match")
		},
		func(launchType string) {
			assert.Empty(t, launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
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

	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceWithHealthCheckGPTest(
		t,
		cliContext,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, observedRole string) {
			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.LoadBalancerName, "LoadBalancer.LoadBalancerName should be nil")
			assert.Equal(t, targetGroupArn, aws.StringValue(loadBalancer.TargetGroupArn), "LoadBalancer.TargetGroupArn should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, observedRole, "Role should match")
		},
		func(launchType string) {
			assert.Empty(t, launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
		},
		func(healthCheckGracePeriod *int64) {
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

	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, observedRole string) {
			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.TargetGroupArn, "LoadBalancer.TargetGroupArn should be nil")
			assert.Equal(t, loadbalancerName, aws.StringValue(loadBalancer.LoadBalancerName), "LoadBalancer.LoadBalancerName should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, observedRole, "Role should match")
		},
		func(launchType string) {
			assert.Empty(t, launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
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

	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceWithHealthCheckGPTest(
		t,
		cliContext,
		&config.CommandConfig{},
		&utils.ECSParams{},
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, observedRole string) {
			assert.NotNil(t, loadBalancer, "LoadBalancer should not be nil")
			assert.Nil(t, loadBalancer.TargetGroupArn, "LoadBalancer.TargetGroupArn should be nil")
			assert.Equal(t, loadbalancerName, aws.StringValue(loadBalancer.LoadBalancerName), "LoadBalancer.LoadBalancerName should match")
			assert.Equal(t, containerName, aws.StringValue(loadBalancer.ContainerName), "LoadBalancer.ContainerName should match")
			assert.Equal(t, int64(containerPort), aws.Int64Value(loadBalancer.ContainerPort), "LoadBalancer.ContainerPort should match")
			assert.Equal(t, role, observedRole, "Role should match")
		},
		func(launchType string) {
			assert.Empty(t, launchType)
		},
		func(networkConfig *ecs.NetworkConfiguration) {
			assert.Nil(t, networkConfig, "NetworkConfiguration should be nil")
		},
		func(healthCheckGracePeriod *int64) {
			assert.Equal(t, int64(healthCheckGP), *healthCheckGracePeriod, "HealthCheckGracePeriod should match")
		},
	)
}

type validateDeploymentConfiguration func(*ecs.DeploymentConfiguration)
type validateLoadBalancer func(*ecs.LoadBalancer, string)
type validateLaunchType func(string)
type validateNetworkConfig func(*ecs.NetworkConfiguration)
type validateHealthCheckGracePeriod func(*int64)

func createServiceTest(t *testing.T,
	cliContext *cli.Context,
	commandConfig *config.CommandConfig,
	ecsParams *utils.ECSParams,
	validateDeploymentConfig validateDeploymentConfiguration,
	validateLB validateLoadBalancer,
	validateLT validateLaunchType,
	validateNC validateNetworkConfig) {

	createServiceWithHealthCheckGPTest(
		t,
		cliContext,
		commandConfig,
		ecsParams,
		validateDeploymentConfig,
		validateLB,
		validateLT,
		validateNC,
		func(healthCheckGP *int64) {
			assert.Nil(t, healthCheckGP, "HealthCheckGracePeriod should be nil")
		},
	)
}

func createServiceWithHealthCheckGPTest(t *testing.T,
	cliContext *cli.Context,
	commandConfig *config.CommandConfig,
	ecsParams *utils.ECSParams,
	validateDeploymentConfig validateDeploymentConfiguration,
	validateLB validateLoadBalancer,
	validateLT validateLaunchType,
	validateNC validateNetworkConfig,
	validateHCGP validateHealthCheckGracePeriod) {

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
			gomock.Any(), // serviceName
			gomock.Any(), // taskDefName
			gomock.Any(), // loadBalancer
			gomock.Any(), // role
			gomock.Any(), // deploymentConfig
			gomock.Any(), // networkConfig
			gomock.Any(), // launchType
			gomock.Any(), // healthCheckGracePeriod
		).Do(func(a, b, c, d, e, f, g, h interface{}) {
			observedTaskDefID := b.(string)
			assert.Equal(t, taskDefID, observedTaskDefID, "Task Definition name should match")

			observedLB := c.(*ecs.LoadBalancer)
			observedRole := d.(string)
			validateLB(observedLB, observedRole)

			observedDeploymentConfig := e.(*ecs.DeploymentConfiguration)
			validateDeploymentConfig(observedDeploymentConfig)

			observedLaunchType := g.(string)
			validateLT(observedLaunchType)

			observedNetworkConfig := f.(*ecs.NetworkConfiguration)
			validateNC(observedNetworkConfig)

			observedHealthCheckGracePeriod := h.(*int64)
			validateHCGP(observedHealthCheckGracePeriod)

		}).Return(nil),
	)

	context := &context.ECSContext{
		ECSClient:  mockEcs,
		CommandConfig:  commandConfig,
		CLIContext: cliContext,
		ECSParams:  ecsParams,
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
		ECSClient:  mockEcs,
		CommandConfig:  commandConfig,
		CLIContext: cliContext,
		ECSParams:  ecsParams,
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
		ECSClient:  mockEcs,
		CommandConfig:  commandConfig,
		CLIContext: cliContext,
		ECSParams:  ecsParams,
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
