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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/container"
	ecscontext "github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/task"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	ec2client "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ec2"
	ecsclient "github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/docker/libcompose/project"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// displayTitle flag is used to print the title for the fields
const displayTitle = true

var flagNamesToStackParameterKeys map[string]string

func init() {
	flagNamesToStackParameterKeys = map[string]string{
		flags.AsgMaxSizeFlag:    cloudformation.ParameterKeyAsgMaxSize,
		flags.VpcAzFlag:         cloudformation.ParameterKeyVPCAzs,
		flags.SecurityGroupFlag: cloudformation.ParameterKeySecurityGroup,
		flags.SourceCidrFlag:    cloudformation.ParameterKeySourceCidr,
		flags.EcsPortFlag:       cloudformation.ParameterKeyEcsPort,
		flags.SubnetIdsFlag:     cloudformation.ParameterKeySubnetIds,
		flags.VpcIdFlag:         cloudformation.ParameterKeyVpcId,
		flags.InstanceTypeFlag:  cloudformation.ParameterKeyInstanceType,
		flags.KeypairNameFlag:   cloudformation.ParameterKeyKeyPairName,
		flags.ImageIdFlag:       cloudformation.ParameterKeyAmiId,
		flags.InstanceRoleFlag:  cloudformation.ParameterKeyInstanceRole,
	}
}

func ClusterUp(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'up': ", err)
	}

	ecsClient := ecsclient.NewECSClient()
	cfnClient := cloudformation.NewCloudformationClient()

	cliParams, err := newCliParams(c, rdwr)
	if err != nil {
		logrus.Fatal("Error executing 'up': ", err)
	}

	err = createCluster(c, ecsClient, cfnClient, cliParams)
	if err != nil {
		logrus.Fatal("Error executing 'up': ", err)
	}

	if !c.Bool(flags.EmptyFlag) {
		// Displays resources create by CloudFormation, as a convenience for tasks launched
		// with Task Networking or in Fargate mode.
		if err := cfnClient.DescribeNetworkResources(cliParams.CFNStackName); err != nil {
			logrus.Error("Error describing Cloudformation resources: ", err)
		}
	}

	fmt.Println("Cluster creation succeeded.")
}

func ClusterDown(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'down': ", err)
	}

	ecsClient := ecsclient.NewECSClient()
	cfnClient := cloudformation.NewCloudformationClient()
	if err := deleteCluster(c, rdwr, ecsClient, cfnClient); err != nil {
		logrus.Fatal("Error executing 'down': ", err)
	}
}

func ClusterScale(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'scale': ", err)
	}

	ecsClient := ecsclient.NewECSClient()
	cfnClient := cloudformation.NewCloudformationClient()
	if err := scaleCluster(c, rdwr, ecsClient, cfnClient); err != nil {
		logrus.Fatal("Error executing 'scale': ", err)
	}
}

func ClusterPS(c *cli.Context) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		logrus.Fatal("Error executing 'ps ", err)
	}

	ecsClient := ecsclient.NewECSClient()
	infoSet, err := clusterPS(c, rdwr, ecsClient)
	if err != nil {
		logrus.Fatal("Error executing 'ps ", err)
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

func createCluster(context *cli.Context, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient, cliParams *config.CLIParams) error {
	var err error

	if context.Bool(flags.EmptyFlag) {
		err = createEmptyCluster(context, ecsClient, cfnClient, cliParams)
		if err != nil {
			return err
		}
		return nil
	}

	launchType := cliParams.LaunchType
	if launchType == "" {
		launchType = config.LaunchTypeDefault
	}

	// InstanceRole not needed when creating empty cluster for Fargate tasks
	if launchType == config.LaunchTypeEC2 {
		if err := validateInstanceRole(context); err != nil {
			return err
		}
		// Display warning if keypair not specified
		if context.String(flags.KeypairNameFlag) == "" {
			logrus.Warn("You will not be able to SSH into your EC2 instances without a key pair.")
		}

	}

	// Check if cluster is specified
	if cliParams.Cluster == "" {
		return fmt.Errorf("Please configure a cluster using the configure command or the '--%s' flag", flags.ClusterFlag)
	}

	// Check if cfn stack already exists
	cfnClient.Initialize(cliParams)
	stackName := cliParams.CFNStackName
	var deleteStack bool
	if err = cfnClient.ValidateStackExists(stackName); err == nil {
		if !isForceSet(context) {
			return fmt.Errorf("A CloudFormation stack already exists for the cluster '%s'. Please specify '--%s' to clean up your existing resources", cliParams.Cluster, flags.ForceFlag)
		}
		deleteStack = true
	}

	// Populate cfn params
	cfnParams := cliFlagsToCfnStackParams(context)
	cfnParams.Add(cloudformation.ParameterKeyCluster, cliParams.Cluster)
	if context.Bool(flags.NoAutoAssignPublicIPAddressFlag) {
		cfnParams.Add(cloudformation.ParameterKeyAssociatePublicIPAddress, "false")
	}

	if launchType == config.LaunchTypeFargate {
		cfnParams.Add(cloudformation.ParameterKeyIsFargate, "true")
	}

	// Check if vpc and AZs are not both specified.
	if validateMutuallyExclusiveParams(cfnParams, cloudformation.ParameterKeyVPCAzs, cloudformation.ParameterKeyVpcId) {
		return fmt.Errorf("You can only specify '--%s' or '--%s'", flags.VpcIdFlag, flags.VpcAzFlag)
	}

	// Check if 2 AZs are specified
	if validateCommaSeparatedParam(cfnParams, cloudformation.ParameterKeyVPCAzs, 2, 2) {
		return fmt.Errorf("You must specify 2 comma-separated availability zones with the '--%s' flag", flags.VpcAzFlag)
	}

	// Check if more than one custom instance role is specified
	if validateCommaSeparatedParam(cfnParams, cloudformation.ParameterKeyInstanceRole, 1, 1) {
		return fmt.Errorf("You can only specify one instance role name with the '--%s' flag", flags.InstanceRoleFlag)
	}

	// Check if vpc exists when security group is specified
	if validateDependentParams(cfnParams, cloudformation.ParameterKeySecurityGroup, cloudformation.ParameterKeyVpcId) {
		return fmt.Errorf("You have selected a security group. Please specify a VPC with the '--%s' flag", flags.VpcIdFlag)
	}

	// Check if subnets exists when vpc is specified
	if validateDependentParams(cfnParams, cloudformation.ParameterKeyVpcId, cloudformation.ParameterKeySubnetIds) {
		return fmt.Errorf("You have selected a VPC. Please specify 2 comma-separated subnets with the '--%s' flag", flags.SubnetIdsFlag)
	}

	// Check if vpc exists when subnets is specified
	if validateDependentParams(cfnParams, cloudformation.ParameterKeySubnetIds, cloudformation.ParameterKeyVpcId) {
		return fmt.Errorf("You have selected subnets. Please specify a VPC with the '--%s' flag", flags.VpcIdFlag)
	}

	// Check if image id was supplied, else populate
	amiIds := ami.NewStaticAmiIds()
	_, err = cfnParams.GetParameter(cloudformation.ParameterKeyAmiId)
	if err == cloudformation.ParameterNotFoundError {
		amiId, err := amiIds.Get(aws.StringValue(cliParams.Session.Config.Region))
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
	ecsClient.Initialize(cliParams)
	if _, err := ecsClient.CreateCluster(cliParams.Cluster); err != nil {
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

func createEmptyCluster(context *cli.Context, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient, cliParams *config.CLIParams) error {
	for _, flag := range flags.CFNResourceFlags() {
		if context.String(flag) != "" {
			logrus.Warnf("Value for flag '%v' will be ignored when creating an empty cluster", flag)
		}
	}
	if isIAMAcknowledged(context) {
		logrus.Warnf("The '--%v' flag will be ignored when creating an empty cluster", flags.CapabilityIAMFlag)
	}

	if isForceSet(context) {
		logrus.Warn("Force flag is unsupported when creating an empty cluster.")
	}

	if cliParams.Cluster == "" {
		return fmt.Errorf("Please configure a cluster using the configure command or the '--%s' flag", flags.ClusterFlag)
	}

	// Check if non-empty cluster with same name already exists
	cfnClient.Initialize(cliParams)
	stackName := cliParams.CFNStackName
	if err := cfnClient.ValidateStackExists(stackName); err == nil {
		return fmt.Errorf("A CloudFormation stack already exists for the cluster '%s'.", cliParams.Cluster)
	}

	ecsClient.Initialize(cliParams)
	if _, err := ecsClient.CreateCluster(cliParams.Cluster); err != nil {
		return err
	}

	return nil
}

var deleteCFNStack = func(cfnClient cloudformation.CloudformationClient, cliParams *config.CLIParams) error {
	stackName := cliParams.CFNStackName
	if err := cfnClient.DeleteStack(stackName); err != nil {
		return err
	}

	logrus.Info("Waiting for your cluster resources to be deleted...")
	if err := cfnClient.WaitUntilDeleteComplete(stackName); err != nil {
		return err
	}

	return nil
}

func deleteCluster(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient) error {
	// Validate cli flags
	if !isForceSet(context) {
		reader := bufio.NewReader(os.Stdin)
		if err := deleteClusterPrompt(reader); err != nil {
			return err
		}
	}

	cliParams, err := newCliParams(context, rdwr)
	if err != nil {
		return err
	}

	// Validate that cluster exists in ECS
	ecsClient.Initialize(cliParams)
	if err := validateCluster(cliParams.Cluster, ecsClient); err != nil {
		return err
	}

	// Validate that a cfn stack exists for the cluster
	cfnClient.Initialize(cliParams)
	stackName := cliParams.CFNStackName

	if err := cfnClient.ValidateStackExists(stackName); err != nil {
		logrus.Infof("No CloudFormation stack found for cluster '%s'.", cliParams.Cluster)
	} else {
		if err := deleteCFNStack(cfnClient, cliParams); err != nil {
			return err
		}
	}

	// Delete cluster in ECS
	if _, err := ecsClient.DeleteCluster(cliParams.Cluster); err != nil {
		return err
	}

	return nil
}

// scaleCluster executes the 'scale' command.
func scaleCluster(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient, cfnClient cloudformation.CloudformationClient) error {
	// Validate cli flags
	if !isIAMAcknowledged(context) {
		return fmt.Errorf("Please acknowledge that this command may create IAM resources with the '--%s' flag", flags.CapabilityIAMFlag)
	}

	size, err := getClusterSize(context)
	if err != nil {
		return err
	}
	if size == "" {
		return fmt.Errorf("Missing required flag '--%s'", flags.AsgMaxSizeFlag)
	}

	cliParams, err := newCliParams(context, rdwr)
	if err != nil {
		return err
	}

	// Validate that cluster exists in ECS
	ecsClient.Initialize(cliParams)
	if err := validateCluster(cliParams.Cluster, ecsClient); err != nil {
		return err
	}

	// Validate that we have a cfn stack for the cluster
	cfnClient.Initialize(cliParams)
	stackName := cliParams.CFNStackName
	existingParameters, err := cfnClient.GetStackParameters(stackName)
	if err != nil {
		return fmt.Errorf("CloudFormation stack not found for cluster '%s'", cliParams.Cluster)
	}

	// Populate update params for the cfn stack
	cfnParams, err := cloudformation.NewCfnStackParamsForUpdate(existingParameters)
	if err != nil {
		return err
	}
	cfnParams.Add(cloudformation.ParameterKeyAsgMaxSize, size)

	// Update the stack.
	if _, err := cfnClient.UpdateStack(stackName, cfnParams); err != nil {
		return err
	}

	logrus.Info("Waiting for your cluster resources to be updated...")
	return cfnClient.WaitUntilUpdateComplete(stackName)
}

func clusterPS(context *cli.Context, rdwr config.ReadWriter, ecsClient ecsclient.ECSClient) (project.InfoSet, error) {
	cliParams, err := newCliParams(context, rdwr)
	if err != nil {
		return nil, err
	}

	// Validate that cluster exists in ECS
	ecsClient.Initialize(cliParams)
	if err := validateCluster(cliParams.Cluster, ecsClient); err != nil {
		return nil, err
	}
	ec2Client := ec2client.NewEC2Client(cliParams)

	ecsContext := &ecscontext.ECSContext{ECSClient: ecsClient, EC2Client: ec2Client}
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
		return fmt.Errorf("Aborted cluster deletion. To delete your cluster, re-run this command and specify the '--%s' flag or confirm that you'd like to delete your cluster at the prompt.", flags.ForceFlag)
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

// isIAMAcknowledged returns true if the 'capability-iam' flag is set from CLI.
func isIAMAcknowledged(context *cli.Context) bool {
	return context.Bool(flags.CapabilityIAMFlag)
}

// returns true if customer specifies a custom instance role via 'role' flag.
func hasCustomRole(context *cli.Context) bool {
	return context.String(flags.InstanceRoleFlag) != "" // validate arn?
}

func validateInstanceRole(context *cli.Context) error {
	defaultRole := isIAMAcknowledged(context)
	customRole := hasCustomRole(context)

	if !defaultRole && !customRole {
		return fmt.Errorf("You must either specify a custom role with the '--%s' flag or set the '--%s' flag", flags.InstanceRoleFlag, flags.CapabilityIAMFlag)
	}
	if defaultRole && customRole {
		return fmt.Errorf("Cannot specify custom role when '--%s' flag is set", flags.CapabilityIAMFlag)
	}
	return nil
}

// isForceSet returns true if the 'force' flag is set from CLI.
func isForceSet(context *cli.Context) bool {
	return context.Bool(flags.ForceFlag)
}

// getClusterSize gets the value for the 'size' flag from CLI.
func getClusterSize(context *cli.Context) (string, error) {
	size := context.String(flags.AsgMaxSizeFlag)
	if size != "" {
		if _, err := strconv.Atoi(size); err != nil {
			return "", err
		}
	}

	return size, nil
}
