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

package command

import (
	"fmt"
	"os"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/cloudformation"
	ec2client "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ec2"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs"
	ecscompose "github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/codegangsta/cli"
	"github.com/docker/libcompose/project"
)

// displayTitle flag is used to print the title for the fields
const displayTitle = true

var flagNamesToStackParameterKeys map[string]string

func init() {
	flagNamesToStackParameterKeys = map[string]string{
		asgMaxSizeFlag:    cloudformation.ParameterKeyAsgMaxSize,
		vpcAzFlag:         cloudformation.ParameterKeyVPCAzs,
		securityGroupFlag: cloudformation.ParameterKeySecurityGroup,
		sourceCidrFlag:    cloudformation.ParameterKeySourceCidr,
		ecsPortFlag:       cloudformation.ParameterKeyEcsPort,
		subnetIdsFlag:     cloudformation.ParameterKeySubnetIds,
		vpcIdFlag:         cloudformation.ParameterKeyVpcId,
		instanceTypeFlag:  cloudformation.ParameterKeyInstanceType,
		keypairNameFlag:   cloudformation.ParameterKeyKeyPairName,
		imageIdFlag:       cloudformation.ParameterKeyAmiId,
	}
}

func ClusterUp(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error executing 'up': ", err)
		return
	}

	ecsClient := ecsclient.NewECSClient()
	cfnClient := cloudformation.NewCloudformationClient()
	amiIds := ami.NewStaticAmiIds()
	if err := createCluster(c, rdwr, ecsClient, cfnClient, amiIds); err != nil {
		logrus.Error("Error executing 'up': ", err)
		return
	}
}

func ClusterDown(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error executing 'down': ", err)
		return
	}

	ecsClient := ecsclient.NewECSClient()
	cfnClient := cloudformation.NewCloudformationClient()
	if err := deleteCluster(c, rdwr, ecsClient, cfnClient); err != nil {
		logrus.Error("Error executing 'down': ", err)
		return
	}
}

func ClusterScale(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error executing 'scale': ", err)
		return
	}

	ecsClient := ecsclient.NewECSClient()
	cfnClient := cloudformation.NewCloudformationClient()
	if err := scaleCluster(c, rdwr, ecsClient, cfnClient); err != nil {
		logrus.Error("Error executing 'scale': ", err)
		return
	}
}

func ClusterPS(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Error("Error executing 'ps ", err)
		return
	}

	ecsClient := ecsclient.NewECSClient()
	infoSet, err := clusterPS(c, rdwr, ecsClient)
	if err != nil {
		logrus.Error("Error executing 'ps ", err)
		return
	}
	os.Stdout.WriteString(infoSet.String(displayTitle))
}

func createCluster(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient, amiIds ami.ECSAmiIds) error {
	// Validate cli flags
	if !isIAMAcknowledged(context) {
		return fmt.Errorf("Please acknowledge that this command may create IAM resources with the '--%s' flag", capabilityIAMFlag)
	}
	ecsParams, err := newCliParams(context, rdwr)
	if err != nil {
		return err
	}

	// Check if cfn stack already exists
	cfnClient.Initialize(ecsParams)
	stackName := ecsParams.GetCfnStackName()
	if err := cfnClient.ValidateStackExists(stackName); err == nil {
		return fmt.Errorf("A CloudFormation stack already exists for the cluster '%s'", ecsParams.Cluster)
	}

	// Populate cfn params
	cfnParams := cliFlagsToCfnStackParams(context)
	cfnParams.Add(cloudformation.ParameterKeyCluster, ecsParams.Cluster)

	// Check if key pair exists
	_, err = cfnParams.GetParameter(cloudformation.ParameterKeyKeyPairName)
	if err == cloudformation.ParameterNotFoundError {
		return fmt.Errorf("Please specify the keypair name with '--%s' flag", keypairNameFlag)
	} else if err != nil {
		return err
	}

	// Check if image id was supplied, else populate
	_, err = cfnParams.GetParameter(cloudformation.ParameterKeyAmiId)
	if err == cloudformation.ParameterNotFoundError {
		amiId, err := amiIds.Get(aws.StringValue(ecsParams.Config.Region))
		if err != nil {
			return err
		}
		cfnParams.Add(cloudformation.ParameterKeyAmiId, amiId)
	} else if err != nil {
		return err
	}
	if err := cfnParams.Validate(); err != nil {
		return err
	}

	// Create ECS cluster
	ecsClient.Initialize(ecsParams)
	if _, err := ecsClient.CreateCluster(ecsParams.Cluster); err != nil {
		return err
	}

	// Create cfn stack
	template := cloudformation.GetTemplate()
	if _, err := cfnClient.CreateStack(template, stackName, cfnParams); err != nil {
		return err
	}

	logrus.Info("Waiting for your cluster resources to be created")
	// Wait for stack creation
	return cfnClient.WaitUntilCreateComplete(stackName)
}

var newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CliParams, error) {
	return config.NewCliParams(context, rdwr)
}

func deleteCluster(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient) error {
	// Validate cli flags
	if !isForceSet(context) {
		return fmt.Errorf("Missing required flag '--%s'", forceFlag)
		// TODO prompt override for force
	}
	ecsParams, err := newCliParams(context, rdwr)
	if err != nil {
		return err
	}

	// Validate that cluster exists in ECS
	ecsClient.Initialize(ecsParams)
	if err := validateCluster(ecsParams.Cluster, ecsClient); err != nil {
		return err
	}

	// Validate that a cfn stack exists for the cluster
	cfnClient.Initialize(ecsParams)
	stackName := ecsParams.GetCfnStackName()
	if err := cfnClient.ValidateStackExists(stackName); err != nil {
		return fmt.Errorf("CloudFormation stack not found for cluster '%s'", ecsParams.Cluster)
	}

	// Delete cfn stack
	if err := cfnClient.DeleteStack(stackName); err != nil {
		return err
	}
	logrus.Info("Waiting for your cluster resources to be deleted")
	if err := cfnClient.WaitUntilDeleteComplete(stackName); err != nil {
		return err
	}

	// Delete cluster in ECS
	if _, err := ecsClient.DeleteCluster(ecsParams.Cluster); err != nil {
		return err
	}

	return nil
}

// scaleCluster executes the 'scale' command.
func scaleCluster(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient) error {
	// Validate cli flags
	if !isIAMAcknowledged(context) {
		return fmt.Errorf("Please acknowledge that this command may create IAM resources with the '--%s' flag", capabilityIAMFlag)
	}

	size, err := getClusterSize(context)
	if err != nil {
		return err
	}
	if size == "" {
		return fmt.Errorf("Missing required flag '--%s'", asgMaxSizeFlag)
	}

	ecsParams, err := newCliParams(context, rdwr)
	if err != nil {
		return err
	}

	// Validate that cluster exists in ECS
	ecsClient.Initialize(ecsParams)
	if err := validateCluster(ecsParams.Cluster, ecsClient); err != nil {
		return err
	}

	// Validate that we have a cfn stack for the cluster
	cfnClient.Initialize(ecsParams)
	stackName := ecsParams.GetCfnStackName()
	if err := cfnClient.ValidateStackExists(stackName); err != nil {
		return fmt.Errorf("CloudFormation stack not found for cluster '%s'", ecsParams.Cluster)
	}

	// Populate update params for the cfn stack
	cfnParams := cloudformation.NewCfnStackParamsForUpdate()
	cfnParams.Add(cloudformation.ParameterKeyAsgMaxSize, size)

	// Update the stack.
	if _, err := cfnClient.UpdateStack(stackName, cfnParams); err != nil {
		return err
	}

	logrus.Info("Waiting for your cluster resources to be updated")
	return cfnClient.WaitUntilUpdateComplete(stackName)
}

func clusterPS(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient) (project.InfoSet, error) {
	ecsParams, err := newCliParams(context, rdwr)
	if err != nil {
		return nil, err
	}

	// Validate that cluster exists in ECS
	ecsClient.Initialize(ecsParams)
	if err := validateCluster(ecsParams.Cluster, ecsClient); err != nil {
		return nil, err
	}
	ec2Client := ec2client.NewEC2Client(ecsParams)

	ecsContext := &ecscompose.Context{ECSClient: ecsClient, EC2Client: ec2Client}
	task := ecscompose.NewTask(ecsContext)
	return task.Info(false)
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

// cliFlagsToCfnStackParams converts values set for CLI flags to cloudformation stack parameters.
func cliFlagsToCfnStackParams(context *cli.Context) *cloudformation.CfnStackParams {
	cfnParams := cloudformation.NewCfnStackParams()
	for cliFlag, cfnParamKeyName := range flagNamesToStackParameterKeys {
		cfnParamKeyValue := context.String(cliFlag)
		if cfnParamKeyValue != "" {
			cfnParams.Add(cfnParamKeyName, cfnParamKeyValue)
		}
	}

	return cfnParams
}

// isIAMAcknowledged returrns true if the 'capability-iam' flag is set from CLI.
func isIAMAcknowledged(context *cli.Context) bool {
	return context.Bool(capabilityIAMFlag)
}

// isForceSet returns true if the 'force' flag is set from CLI.
func isForceSet(context *cli.Context) bool {
	return context.Bool(forceFlag)
}

// getClusterSize gets the value for the 'size' flag from CLI.
func getClusterSize(context *cli.Context) (string, error) {
	size := context.String(asgMaxSizeFlag)
	if size != "" {
		if _, err := strconv.Atoi(size); err != nil {
			return "", err
		}
	}

	return size, nil
}
