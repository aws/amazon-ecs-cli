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
package attributechecker

import (
	"flag"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	taskDefArn           = "arn:aws:ecs:us-west-2:123412341234:task-definition/myTaskDef:1"
	taskDefName          = "testTaskDef:7"
	clusterName          = "defaultCluster"
	containerInstancearn = "arn:aws:ecs:eu-west-1:123456789012:container-instance/7493b0b5-6827-4828-88b3-47b71dd19428"
	containerInstanceID  = "7493b0b5-6827-4828-88b3-47b71dd19428"
	attribute1           = "com.amazonaws.ecs.capability.ecr-auth"
	attribute2           = "com.amazonaws.ecs.capability.task-iam-role"
	attribute3           = "com.amazonaws.ecs.capability.logging-driver.awslogs"
)

type mockReadWriter struct {
	clusterName string
}

func (rdwr *mockReadWriter) Get(cluster string, profile string) (*config.LocalConfig, error) {
	cliConfig := config.NewLocalConfig(rdwr.clusterName)
	return cliConfig, nil
}

func (rdwr *mockReadWriter) SaveCluster(configName string, cluster *config.Cluster) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultProfile(configName string, profile *config.Profile) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultCluster(configName string) error {
	return nil
}

func dummyTaskDef(Attributes []*ecs.Attribute) *ecs.TaskDefinition {
	taskDef := &ecs.TaskDefinition{}
	taskDef.SetTaskDefinitionArn(taskDefArn)
	taskDef.SetRequiresAttributes(Attributes)

	return taskDef
}

func dummyTaskdefAttributes() []*string {
	TaskDefAttrNamesOutput := []*string{}
	taskDefinition := &ecs.TaskDefinition{}
	taskDefinition.SetTaskDefinitionArn(taskDefArn)
	attributes := []*ecs.Attribute{
		{
			Name: aws.String(attribute1),
		},
		{
			Name: aws.String(attribute2),
		},
		{
			Name: aws.String(attribute3),
		},
	}
	taskDefinition.SetRequiresAttributes(attributes)
	for _, taskDefAttrNames := range taskDefinition.RequiresAttributes {
		TaskDefAttrNamesOutput = append(TaskDefAttrNamesOutput, taskDefAttrNames.Name)
	}
	return TaskDefAttrNamesOutput
}

func dummyContainerInstanceAndAttrMap() map[string][]*string {
	descrContainerInstancesoutputMap := map[string][]*string{}
	Attribute := []*string{}
	containerInstance := &ecs.ContainerInstance{}
	containerInstance.SetContainerInstanceArn(containerInstancearn)
	Attributes := []*ecs.Attribute{
		{
			Name: aws.String(attribute1),
		},
		{
			Name: aws.String(attribute2),
		},
	}
	containerInstance.SetAttributes(Attributes)

	for _, containerInstanceattributenames := range containerInstance.Attributes {
		Attribute = append(Attribute, containerInstanceattributenames.Name)
	}
	descrContainerInstancesoutputMap[containerInstancearn] = Attribute

	return descrContainerInstancesoutputMap
}

func setupTest(t *testing.T) *mock_ecs.MockECSClient {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	return mockECS
}

func TestCompare(t *testing.T) {
	defer os.Clearenv()
	taskDefAttributes := dummyTaskdefAttributes()
	containerInstanceAttributes := dummyContainerInstanceAndAttrMap()

	compareOutput := compare(taskDefAttributes, containerInstanceAttributes)
	assert.Equal(t, map[string]string{containerInstanceID: attribute3}, compareOutput)
}

func TestTaskdefattributesCheckRequest(t *testing.T) {
	defer os.Clearenv()
	mockECS := setupTest(t)
	Attribute := []*ecs.Attribute{
		{
			Name: aws.String(attribute1),
		},
		{
			Name: aws.String(attribute2),
		},
	}
	taskDefAttributes := dummyTaskDef(Attribute)

	mockECS.EXPECT().DescribeTaskDefinition(taskDefName).Return(taskDefAttributes, nil)

	flagSet := flag.NewFlagSet("task-definition", 0)
	flagSet.String(flags.TaskDefinitionFlag, taskDefName, "")
	context := cli.NewContext(nil, flagSet, nil)

	taskDefAttributeNames, err := taskdefattributesCheckRequest(context, mockECS)

	assert.NoError(t, err, "Unexpected error getting taskDef AttributeNames")
	assert.Equal(t, 2, len(taskDefAttributeNames))
}

func TestDescribeContainerInstancesRequestWithoneContainerInstanceArn(t *testing.T) {
	defer os.Clearenv()
	mockECS := setupTest(t)
	containerInstance := "containerInstance"
	containerInstanceAttributes := dummyContainerInstanceAndAttrMap()

	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)

	mockECS.EXPECT().GetAttributesFromDescribeContainerInstances([]*string{&containerInstance}).Return(containerInstanceAttributes, nil)

	flagSet := flag.NewFlagSet("container-instance", 0)
	flagSet.String(flags.ContainerInstancesFlag, containerInstance, "")
	context := cli.NewContext(nil, flagSet, nil)

	descrContainerInstancesResponse, err := describeContainerInstancesAttributeMap(context, mockECS, &config.CommandConfig{})
	assert.NoError(t, err, "Unexpected error invoking DescribeContainerInstanceAttribute Instance function")
	assert.Equal(t, 1, len(descrContainerInstancesResponse))
}
