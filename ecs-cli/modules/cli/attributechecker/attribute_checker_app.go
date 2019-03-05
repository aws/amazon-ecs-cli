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
	"fmt"
	"os"
	"strings"

	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/libcompose/project"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	displayTitle            = true
	containerInstanceHeader = "Container Instance"
	missingAttributesHeader = "Missing Attributes"
)

var infoColumns = []string{containerInstanceHeader, missingAttributesHeader}

//AttributeChecker will compare task def and containers instances attributes and outputs missing attributes
func AttributeChecker(c *cli.Context) {
	err := validateAttributeCheckerFlags(c)
	if err != nil {
		logrus.Fatal("Error executing 'Attribute Checker': ", err)
	}
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'Attribute Checker': ", err)
	}
	commandConfig, err := config.NewCommandConfig(c, rdwr)
	if err != nil {
		logrus.Fatal("Error executing 'Attribute Checker': ", err)
	}

	ecsClient := ecsclient.NewECSClient(commandConfig)

	taskDefAttributeNames, err := taskdefattributesCheckRequest(c, ecsClient)
	if err != nil {
		logrus.Fatal("Error executing 'Attribute Checker': ", err)
	}
	if len(taskDefAttributeNames) == 0 {
		logrus.Info("The given task definition does not have any attributes")
		return
	}

	descrContainerInstancesResponse, err := describeContainerInstancesAttributeMap(c, ecsClient, commandConfig)
	if err != nil {
		logrus.Fatal("Error executing 'Attribute Checker': ", err)
	}

	compareOutput := compare(taskDefAttributeNames, descrContainerInstancesResponse)
	result := ConvertToInfoSet(compareOutput)
	os.Stdout.WriteString(result.String(infoColumns, displayTitle))
}

func contains(containerInstanceAttributeNames []*string, tdAttrNames *string) bool {
	for _, containerInstAttrNames := range containerInstanceAttributeNames {
		if *containerInstAttrNames == *tdAttrNames {
			return true
		}
	}
	return false
}

//compares between container instances and Task definition
func compare(taskDefAttributeNames []*string, descrContainerInstancesResponse map[string][]*string) map[string]string {
	attributeCheckerResult := make(map[string]string)
	for containerInstanceARN, containerInstanceAttributeNames := range descrContainerInstancesResponse {
		var missingAttributes []string
		for _, tdAttrNames := range taskDefAttributeNames {
			if !contains(containerInstanceAttributeNames, tdAttrNames) {
				missingAttributes = append(missingAttributes, *tdAttrNames)
			}
		}
		missingAttributesNames := strings.Join(missingAttributes, ", ")
		if len(missingAttributesNames) == 0 {
			missingAttributesNames = "None"
		}
		containerInstance := strings.Split(containerInstanceARN, "/")
		attributeCheckerResult[containerInstance[1]] = missingAttributesNames
	}
	return attributeCheckerResult
}

// DescribeContainerInstancesAttributeMap and get a map with Container instance ARN and Container instances attribute Names
func describeContainerInstancesAttributeMap(context *cli.Context, ecsClient ecsclient.ECSClient, commandConfig *config.CommandConfig) (map[string][]*string, error) {
	if err := validateCluster(commandConfig.Cluster, ecsClient); err != nil {
		return nil, err
	}
	var containerInstanceIdentifiers []*string
	containerInstanceIdentifier := context.String(flags.ContainerInstancesFlag)
	splitValues := strings.Split(containerInstanceIdentifier, ",")
	containerInstanceIdentifiers = aws.StringSlice(splitValues)

	descrContainerInstancesAttributes, err := ecsClient.GetAttributesFromDescribeContainerInstances(containerInstanceIdentifiers)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("Failed to Describe Container Instances, please check region/containerInstance/cluster values"))
	}
	return descrContainerInstancesAttributes, err
}

// validateCluster validates if the cluster exists in ECS and is in "ACTIVE" state.
func validateCluster(clusterName string, ecsClient ecsclient.ECSClient) error {
	isClusterActive, err := ecsClient.IsActiveCluster(clusterName)
	if err != nil {
		return err
	}

	if !isClusterActive {
		return fmt.Errorf("Cluster '%s' is not active. Ensure that it exists", clusterName)
	}
	return nil
}

//taskdefattributesCheckRequest describes task def and gets all attribute Names from the task definition
func taskdefattributesCheckRequest(context *cli.Context, ecsClient ecsclient.ECSClient) ([]*string, error) {

	taskDefIdentifier := context.String(flags.TaskDefinitionFlag)

	descrTaskDefinition, err := ecsClient.DescribeTaskDefinition(taskDefIdentifier)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("Failed to Describe TaskDefinition, please check the region/taskDefinition values"))
	}
	var taskattributeNames []*string
	for _, taskDefAttributesName := range descrTaskDefinition.RequiresAttributes {
		taskattributeNames = append(taskattributeNames, taskDefAttributesName.Name)
	}
	return taskattributeNames, err
}

//validates all required flags are passed to run the command
func validateAttributeCheckerFlags(context *cli.Context) error {
	if taskDefIdentifier := context.String(flags.TaskDefinitionFlag); taskDefIdentifier == "" {
		return fmt.Errorf("TaskDefinition must be specified with the --%s flag", flags.TaskDefinitionFlag)
	}
	if containerInstanceIdentifier := context.String(flags.ContainerInstancesFlag); containerInstanceIdentifier == "" {
		return fmt.Errorf("ContainerInstance(s) must be specified with the --%s flag", flags.ContainerInstancesFlag)
	}
	return nil
}

//ConvertToInfoSet transforms the Map of containerARN and MissingAttributes into a formatted set of fields
func ConvertToInfoSet(compareOutput map[string]string) project.InfoSet {
	result := project.InfoSet{}
	for key, element := range compareOutput {
		info := project.Info{
			containerInstanceHeader: key,
			missingAttributesHeader: element,
		}
		result = append(result, info)
	}
	return result
}
