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
	"flag"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/codegangsta/cli"
	"github.com/golang/mock/gomock"
)

func TestCreateWithDeploymentConfig(t *testing.T) {
	deploymentMaxPercent := 200
	deploymentMinPercent := 100

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(DeploymentMaxPercentFlag, strconv.Itoa(deploymentMaxPercent), "")
	flagSet.String(DeploymentMinHealthyPercentFlag, strconv.Itoa(deploymentMinPercent), "")
	cliContext := cli.NewContext(nil, flagSet, nil)

	createServiceTest(
		t,
		cliContext,
		func(deploymentConfig *ecs.DeploymentConfiguration) {
			if aws.Int64Value(deploymentConfig.MaximumPercent) != int64(deploymentMaxPercent) {
				t.Errorf("Expected DeploymentConfig.MaxPercent to be [%s] but got [%s]",
					deploymentMaxPercent, aws.Int64Value(deploymentConfig.MaximumPercent))
			}
			if aws.Int64Value(deploymentConfig.MinimumHealthyPercent) != int64(deploymentMinPercent) {
				t.Errorf("Expected DeploymentConfig.MinimumHealthyPercent to be [%s] but got [%s]",
					deploymentMinPercent, aws.Int64Value(deploymentConfig.MinimumHealthyPercent))
			}
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
			if deploymentConfig.MaximumPercent != nil {
				t.Errorf("Expected DeploymentConfig.MaximumPercent to be nil but got [%s]",
					aws.Int64Value(deploymentConfig.MaximumPercent))
			}
			if deploymentConfig.MinimumHealthyPercent != nil {
				t.Errorf("Expected DeploymentConfig.MinimumHealthyPercent to be nil but got [%s]",
					aws.Int64Value(deploymentConfig.MinimumHealthyPercent))
			}
		},
	)
}

type validateDeploymentConfiguration func(*ecs.DeploymentConfiguration)

func createServiceTest(t *testing.T, cliContext *cli.Context, validateDeploymentConfig validateDeploymentConfiguration) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	taskDefId := "taskDefinitionId"
	taskDefArn := "arn/" + taskDefId

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
			if aws.StringValue(taskDefinition.Family) != aws.StringValue(req.Family) {
				t.Errorf("Expected taskDefintion family to be [%s] but got [%s]",
					aws.StringValue(taskDefinition.Family), aws.StringValue(req.Family))
			}
		}).Return(&respTaskDef, nil),
		mockEcs.EXPECT().CreateService(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(x, y, z interface{}) {
			observedTaskDefId := y.(string)
			if taskDefId != observedTaskDefId {
				t.Errorf("Expected task definition name to be [%s] but got [%s]", taskDefId, observedTaskDefId)
			}
			observedDeploymentConfig := z.(*ecs.DeploymentConfiguration)
			validateDeploymentConfig(observedDeploymentConfig)
		}).Return(nil),
	)

	context := &Context{
		ECSClient:  mockEcs,
		ECSParams:  &config.CliParams{},
		CLIContext: cliContext,
	}

	service := NewService(context)
	if err := service.LoadContext(); err != nil {
		t.Fatal("Unexpected error while loading context in create service test")
	}
	service.SetTaskDefinition(&taskDefinition)
	if err := service.Create(); err != nil {
		t.Fatal("Unexpected error while create")
	}

	// task definition should be set
	if taskDefArn != aws.StringValue(service.TaskDefinition().TaskDefinitionArn) {
		t.Errorf("Expected service TaskDefArn to be [%s] but got [%s]",
			taskDefArn, aws.StringValue(service.TaskDefinition().TaskDefinitionArn))
	}
}

func TestLoadContext(t *testing.T) {
	deploymentMaxPercent := 150

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(DeploymentMaxPercentFlag, strconv.Itoa(deploymentMaxPercent), "")
	cliContext := cli.NewContext(nil, flagSet, nil)
	service := &Service{
		projectContext: &Context{CLIContext: cliContext},
	}

	if err := service.LoadContext(); err != nil {
		t.Fatal("Unexpected error while loading context in load context test")
	}

	observedDeploymentConfig := service.DeploymentConfig()
	if aws.Int64Value(observedDeploymentConfig.MaximumPercent) != int64(deploymentMaxPercent) {
		t.Errorf("Expected DeploymentConfig.MaxPercent to be [%s] but got [%s]",
			deploymentMaxPercent, aws.Int64Value(observedDeploymentConfig.MaximumPercent))
	}
	if observedDeploymentConfig.MinimumHealthyPercent != nil {
		t.Errorf("Expected DeploymentConfig.MinimumHealthyPercent to be nil but got [%s]",
			aws.Int64Value(observedDeploymentConfig.MinimumHealthyPercent))
	}

}

func TestLoadContextForIncorrectInput(t *testing.T) {
	deploymentMaxPercent := "string"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(DeploymentMaxPercentFlag, deploymentMaxPercent, "")
	cliContext := cli.NewContext(nil, flagSet, nil)
	service := &Service{
		projectContext: &Context{CLIContext: cliContext},
	}

	err := service.LoadContext()
	if err == nil {
		t.Error("Expected error to load context when flag is a string but got done")
	}
}

func TestServiceInfo(t *testing.T) {
	testInfo(func(context *Context) ProjectEntity {
		return NewService(context)
	}, func(req *ecs.ListTasksInput, projectName string, t *testing.T) {
		if !strings.Contains(aws.StringValue(req.ServiceName), projectName) {
			t.Errorf("Expected serviceName to contain projectName [%s] but got [%s]",
				projectName, aws.StringValue(req.ServiceName))
		}
		if req.StartedBy != nil {
			t.Error("Expected startedby to be not set")
		}
	}, t, true)
}

func TestServiceRun(t *testing.T) {
	service := NewService(&Context{})
	if err := service.Run(map[string]string{}); err == nil {
		t.Error("Expected unsupported error")
	}
}
