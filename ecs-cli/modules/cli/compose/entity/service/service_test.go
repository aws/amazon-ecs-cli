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
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Equal(t, int64(deploymentMaxPercent), aws.Int64Value(deploymentConfig.MaximumPercent), "DeploymentConfig.MaxPercent should match")
			assert.Equal(t, int64(deploymentMinPercent), aws.Int64Value(deploymentConfig.MinimumHealthyPercent), "DeploymentConfig.MinimumHealthyPercent should match")
		},
		func(loadBalancer *ecs.LoadBalancer, role string) {
			assert.Nil(t, loadBalancer, "LoadBalancer should be nil")
			assert.Empty(t, role, "Role should be empty")
		},
	)
}

func TestCreateWithoutDeploymentConfig(t *testing.T) {
	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			assert.Nil(t, deploymentConfig.MaximumPercent, "DeploymentConfig.MaximumPercent should be nil")
			assert.Nil(t, deploymentConfig.MinimumHealthyPercent, "DeploymentConfig.MinimumHealthyPercent should be nil")
		},
		func(loadBalancer *ecs.LoadBalancer, role string) {
			assert.Nil(t, loadBalancer, "LoadBalancer should be nil")
			assert.Empty(t, role, "Role should be empty")
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
	)
}

type validateDeploymentConfiguration func(*ecs.DeploymentConfiguration)
type validateLoadBalancer func(*ecs.LoadBalancer, string)

func createServiceTest(t *testing.T, cliContext *cli.Context,
	validateDeploymentConfig validateDeploymentConfiguration,
	validateLB validateLoadBalancer) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskDefID := "taskDefinitionId"
	taskDefArn := "arn/" + taskDefID

	taskDefinition := ecs.TaskDefinition{
		Family:               aws.String("family"),
		ContainerDefinitions: []*ecs.ContainerDefinition{},
		Volumes:              []*ecs.Volume{},
	}
	respTaskDef := taskDefinition
	respTaskDef.TaskDefinitionArn = aws.String(taskDefArn)

	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	gomock.InOrder(
		mockEcs.EXPECT().RegisterTaskDefinitionIfNeeded(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
			// verify input fields
			req := x.(*ecs.RegisterTaskDefinitionInput)
			assert.Equal(t, aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family), "Task Definition family should match")
		}).Return(&respTaskDef, nil),
		mockEcs.EXPECT().CreateService(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(v, w, x, y, z interface{}) {
			observedTaskDefID := w.(string)
			assert.Equal(t, taskDefID, observedTaskDefID, "Task Definition name should match")
			observedLB := x.(*ecs.LoadBalancer)
			observedRole := y.(string)
			validateLB(observedLB, observedRole)
			observedDeploymentConfig := z.(*ecs.DeploymentConfiguration)
			validateDeploymentConfig(observedDeploymentConfig)

		}).Return(nil),
	)

	context := &context.Context{
		ECSClient:  mockEcs,
		ECSParams:  &config.CLIParams{},
		CLIContext: cliContext,
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
		projectContext: &context.Context{CLIContext: cliContext},
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
		projectContext: &context.Context{CLIContext: cliContext},
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
		projectContext: &context.Context{CLIContext: cliContext},
	}

	err := service.LoadContext()
	assert.Error(t, err, "Expected error to load context when flag is a string but got done")
}

func TestServiceInfo(t *testing.T) {
	entity.TestInfo(func(context *context.Context) entity.ProjectEntity {
		return NewService(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		assert.Contains(t, aws.StringValue(req.ServiceName), projectName, "ServiceName should contain ProjectName")
		assert.Nil(t, req.StartedBy, "StartedBy should be nil")
	}, t, true)
}

func TestServiceRun(t *testing.T) {
	service := NewService(&context.Context{})
	err := service.Run(map[string][]string{})
	assert.Error(t, err, "Expected unsupported error")
}
