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

package cluster

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/container"
	composecontext "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/task"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	ec2client "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ec2"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/libcompose/project"
	"github.com/urfave/cli"
)

// displayTitle flag is used to print the title for the fields
const displayTitle = true

var flagNamesToStackParameterKeys map[string]string

func init() {
	flagNamesToStackParameterKeys = map[string]string{
		command.AsgMaxSizeFlag:    cloudformation.ParameterKeyAsgMaxSize,
		command.VpcAzFlag:         cloudformation.ParameterKeyVPCAzs,
		command.SecurityGroupFlag: cloudformation.ParameterKeySecurityGroup,
		command.SourceCidrFlag:    cloudformation.ParameterKeySourceCidr,
		command.EcsPortFlag:       cloudformation.ParameterKeyEcsPort,
		command.SubnetIdsFlag:     cloudformation.ParameterKeySubnetIds,
		command.VpcIdFlag:         cloudformation.ParameterKeyVpcId,
		command.InstanceTypeFlag:  cloudformation.ParameterKeyInstanceType,
		command.KeypairNameFlag:   cloudformation.ParameterKeyKeyPairName,
		command.ImageIdFlag:       cloudformation.ParameterKeyAmiId,
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
	os.Stdout.WriteString(infoSet.String(container.ContainerInfoColumns, displayTitle))
}

// If param1 exists, param2 is not allowed.
func validateMutuallyExclusiveParams(cfnParams *cloudformation.CfnStackParams, param1, param2 string) bool {
	if _, err := cfnParams.GetParameter(param1); err != nil {
		return false
	}
	if _, err := cfnParams.GetParameter(param2); err != cloudformation.ParameterNotFoundError {
		return true
	}
	return false
}

// If param1 exists, param2 is required.
func validateDependentParams(cfnParams *cloudformation.CfnStackParams, param1, param2 string) bool {
	if _, err := cfnParams.GetParameter(param1); err != nil {
		return false
	}
	if _, err := cfnParams.GetParameter(param2); err == cloudformation.ParameterNotFoundError {
		return true
	}
	return false
}

func validateCommaSeparatedParam(cfnParams *cloudformation.CfnStackParams, param string, minLength, maxLength int) bool {
	values, err := cfnParams.GetParameter(param)
	if err != nil {
		return false
	}
	if splitValues := strings.Split(*values.ParameterValue, ","); len(splitValues) < minLength || len(splitValues) > maxLength {
		return true
	}
	return false
}

func createCluster(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient, amiIds ami.ECSAmiIds) error {
	// Validate cli flags
	if !isIAMAcknowledged(context) {
		return fmt.Errorf("Please acknowledge that this command may create IAM resources with the '--%s' flag", command.CapabilityIAMFlag)
	}
	ecsParams, err := newCliParams(context, rdwr)
	if err != nil {
		return err
	}

	// Check if cluster is specified
	if ecsParams.Cluster == "" {
		return fmt.Errorf("Please configure a cluster using the configure command or the '--%s' flag", command.ClusterFlag)
	}

	// Check if cfn stack already exists
	cfnClient.Initialize(ecsParams)
	stackName := ecsParams.GetCFNStackName()
	var deleteStack bool
	if err = cfnClient.ValidateStackExists(stackName); err == nil {
		if !isForceSet(context) {
			return fmt.Errorf("A CloudFormation stack already exists for the cluster '%s'. Please specify '--%s' to clean up your existing resources", ecsParams.Cluster, command.ForceFlag)
		}
		deleteStack = true
	}

	// Populate cfn params
	cfnParams := cliFlagsToCfnStackParams(context)
	cfnParams.Add(cloudformation.ParameterKeyCluster, ecsParams.Cluster)
	if context.Bool(command.NoAutoAssignPublicIPAddressFlag) {
		cfnParams.Add(cloudformation.ParameterKeyAssociatePublicIPAddress, "false")
	}

	// Check if key pair exists
	_, err = cfnParams.GetParameter(cloudformation.ParameterKeyKeyPairName)
	if err == cloudformation.ParameterNotFoundError {
		return fmt.Errorf("Please specify the keypair name with '--%s' flag", command.KeypairNameFlag)
	} else if err != nil {
		return err
	}

	// Check if vpc and AZs are not both specified.
	if validateMutuallyExclusiveParams(cfnParams, cloudformation.ParameterKeyVPCAzs, cloudformation.ParameterKeyVpcId) {
		return fmt.Errorf("You can only specify '--%s' or '--%s'", command.VpcIdFlag, command.VpcAzFlag)
	}

	// Check if 2 AZs are specified
	if validateCommaSeparatedParam(cfnParams, cloudformation.ParameterKeyVPCAzs, 2, 2) {
		return fmt.Errorf("You must specify 2 comma-separated availability zones with the '--%s' flag", command.VpcAzFlag)
	}

	// Check if vpc exists when security group is specified
	if validateDependentParams(cfnParams, cloudformation.ParameterKeySecurityGroup, cloudformation.ParameterKeyVpcId) {
		return fmt.Errorf("You have selected a security group. Please specify a VPC with the '--%s' flag", command.VpcIdFlag)
	}

	// Check if subnets exists when vpc is specified
	if validateDependentParams(cfnParams, cloudformation.ParameterKeyVpcId, cloudformation.ParameterKeySubnetIds) {
		return fmt.Errorf("You have selected a VPC. Please specify 2 comma-separated subnets with the '--%s' flag", command.SubnetIdsFlag)
	}

	// Check if vpc exists when subnets is specified
	if validateDependentParams(cfnParams, cloudformation.ParameterKeySubnetIds, cloudformation.ParameterKeyVpcId) {
		return fmt.Errorf("You have selected subnets. Please specify a VPC with the '--%s' flag", command.VpcIdFlag)
	}

	// Check if image id was supplied, else populate
	_, err = cfnParams.GetParameter(cloudformation.ParameterKeyAmiId)
	if err == cloudformation.ParameterNotFoundError {
		amiId, err := amiIds.Get(aws.StringValue(ecsParams.Session.Config.Region))
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

	// Delete cfn stack
	if deleteStack {
		if err := cfnClient.DeleteStack(stackName); err != nil {
			return err
		}
		logrus.Info("Waiting for your CloudFormation stack resources to be deleted...")
		if err := cfnClient.WaitUntilDeleteComplete(stackName); err != nil {
			return err
		}
	}

	// Create cfn stack
	template := cloudformation.GetTemplate()
	if _, err := cfnClient.CreateStack(template, stackName, cfnParams); err != nil {
		return err
	}

	logrus.Info("Waiting for your cluster resources to be created...")
	// Wait for stack creation
	return cfnClient.WaitUntilCreateComplete(stackName)
}

var newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CLIParams, error) {
	return config.NewCLIParams(context, rdwr)
}

func deleteCluster(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient) error {
	// Validate cli flags
	if !isForceSet(context) {
		reader := bufio.NewReader(os.Stdin)
		if err := deleteClusterPrompt(reader); err != nil {
			return err
		}
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
	stackName := ecsParams.GetCFNStackName()
	if err := cfnClient.ValidateStackExists(stackName); err != nil {
		return fmt.Errorf("CloudFormation stack not found for cluster '%s'", ecsParams.Cluster)
	}

	// Delete cfn stack
	if err := cfnClient.DeleteStack(stackName); err != nil {
		return err
	}
	logrus.Info("Waiting for your cluster resources to be deleted...")
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
		return fmt.Errorf("Please acknowledge that this command may create IAM resources with the '--%s' flag", command.CapabilityIAMFlag)
	}

	size, err := getClusterSize(context)
	if err != nil {
		return err
	}
	if size == "" {
		return fmt.Errorf("Missing required flag '--%s'", command.AsgMaxSizeFlag)
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
	stackName := ecsParams.GetCFNStackName()
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

	logrus.Info("Waiting for your cluster resources to be updated...")
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

	ecsContext := &composecontext.Context{ECSClient: ecsClient, EC2Client: ec2Client}
	task := task.NewTask(ecsContext)
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

// deleteClusterPrompt prompts and checks for confirmation to delete the cluster
func deleteClusterPrompt(reader *bufio.Reader) error {
	fmt.Println("Are you sure you want to delete your cluster? [y/N]")
	input, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("Error reading input: %s", err.Error())
	}
	formattedInput := strings.ToLower(strings.TrimSpace(input))
	if formattedInput != "yes" && formattedInput != "y" {
		return fmt.Errorf("Aborted cluster deletion. To delete your cluster, re-run this command and specify the '--%s' flag or confirm that you'd like to delete your cluster at the prompt.", command.ForceFlag)
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
	return context.Bool(command.CapabilityIAMFlag)
}

// isForceSet returns true if the 'force' flag is set from CLI.
func isForceSet(context *cli.Context) bool {
	return context.Bool(command.ForceFlag)
}

// getClusterSize gets the value for the 'size' flag from CLI.
func getClusterSize(context *cli.Context) (string, error) {
	size := context.String(command.AsgMaxSizeFlag)
	if size != "" {
		if _, err := strconv.Atoi(size); err != nil {
			return "", err
		}
	}

	return size, nil
}
