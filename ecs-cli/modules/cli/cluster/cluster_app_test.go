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
	"bytes"
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	command "github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	clusterName = "defaultCluster"
	stackName   = "defaultCluster"
)

type mockReadWriter struct {
	clusterName string
	stackName   string
}

func (rdwr *mockReadWriter) Get(cluster string, profile string) (*config.CLIConfig, error) {
	cliConfig := config.NewCLIConfig(rdwr.clusterName)
	cliConfig.CFNStackNamePrefix = ""
	return cliConfig, nil
}

func (rdwr *mockReadWriter) Save(*config.CLIConfig) error {
	return nil
}

func newMockReadWriter() *mockReadWriter {
	return &mockReadWriter{clusterName: clusterName}
}

func setupTest(t *testing.T) (*mock_ecs.MockECSClient, *mock_cloudformation.MockCloudformationClient) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	os.Setenv("AWS_REGION", "us-west-1")
	return mockECS, mockCloudformation
}

func TestClusterUp(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithForce(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(stackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(stackName).Return(nil),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithoutPublicIP(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, gomock.Any()).Do(func(x, y, z interface{}) {
			cfnParams := z.(*cloudformation.CfnStackParams)
			associateIPAddress, err := cfnParams.GetParameter(cloudformation.ParameterKeyAssociatePublicIPAddress)
			assert.NoError(t, err, "Unexpected error getting cfn parameter")
			assert.Equal(t, "false", aws.StringValue(associateIPAddress.ParameterValue), "Should not associate public IP address")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.NoAutoAssignPublicIPAddressFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	vpcID := "vpc-02dd3038"
	subnetIds := "subnet-04726b21,subnet-04346b21"

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.String(command.VpcIdFlag, vpcID, "")
	flagSet.String(command.SubnetIdsFlag, subnetIds, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithAvailabilityZones(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	vpcAZs := "us-west-2c,us-west-2a"

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.String(command.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithoutKeyPair(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.Bool(command.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error for key pair name")
}

func TestClusterUpWithSecurityGroupWithoutVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	securityGroupID := "sg-eeaabc8d"

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.ForceFlag, true, "")
	flagSet.String(command.SecurityGroupFlag, securityGroupID, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error for security group without VPC")
}

func TestClusterUpWith2SecurityGroups(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	securityGroupIds := "sg-eeaabc8d,sg-eaaebc8d"
	vpcId := "vpc-02dd3038"
	subnetIds := "subnet-04726b21,subnet-04346b21"

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.ForceFlag, true, "")
	flagSet.String(command.SecurityGroupFlag, securityGroupIds, "")
	flagSet.String(command.VpcIdFlag, vpcId, "")
	flagSet.String(command.SubnetIdsFlag, subnetIds, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithSubnetsWithoutVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	subnetID := "subnet-72f52e32"

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.ForceFlag, true, "")
	flagSet.String(command.SubnetIdsFlag, subnetID, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error for subnets without VPC")
}

func TestClusterUpWithVPCWithoutSubnets(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	vpcID := "vpc-02dd3038"

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.ForceFlag, true, "")
	flagSet.String(command.VpcIdFlag, vpcID, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error for VPC without subnets")
}

func TestClusterUpWithAvailabilityZonesWithVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	vpcID := "vpc-02dd3038"
	vpcAZs := "us-west-2c,us-west-2a"

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.ForceFlag, true, "")
	flagSet.String(command.VpcIdFlag, vpcID, "")
	flagSet.String(command.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error for VPC with AZs")
}

func TestClusterUpWithout2AvailabilityZones(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	vpcAZs := "us-west-2c"

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.Bool(command.ForceFlag, true, "")
	flagSet.String(command.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error for 2 AZs")
}

func TestCliFlagsToCfnStackParams(t *testing.T) {

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	params := cliFlagsToCfnStackParams(context)

	_, err := params.GetParameter(cloudformation.ParameterKeyAsgMaxSize)
	assert.Error(t, err, "Expected error for parameter ParameterKeyAsgMaxSize")
	assert.Equal(t, cloudformation.ParameterNotFoundError, err, "Expect error to be ParameterNotFoundError")

	flagSet.String(command.AsgMaxSizeFlag, "2", "")
	context = cli.NewContext(nil, flagSet, nil)
	params = cliFlagsToCfnStackParams(context)
	_, err = params.GetParameter(cloudformation.ParameterKeyAsgMaxSize)
	assert.NoError(t, err, "Unexpected error getting parameter ParameterKeyAsgMaxSize")
}

func TestClusterUpForImageIdInput(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	imageID := "ami-12345"

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, gomock.Any()).Do(func(x, y, z interface{}) {
			cfnStackParams := z.(*cloudformation.CfnStackParams)
			param, err := cfnStackParams.GetParameter(cloudformation.ParameterKeyAmiId)
			assert.NoError(t, err, "Expected image id params to be present")
			assert.Equal(t, imageID, aws.StringValue(param.ParameterValue), "Expected image id to match")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")
	flagSet.String(command.ImageIdFlag, imageID, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithClusterNameEmpty(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, &mockReadWriter{clusterName: ""}, mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error bringing up cluster")
}

func TestClusterUpWithoutRegion(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)
	os.Unsetenv("AWS_REGION")

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	context := cli.NewContext(nil, flagSet, nil)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	assert.Error(t, err, "Expected error bringing up cluster")
}

func TestClusterDown(t *testing.T) {
	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CLIParams, error) {
		return &config.CLIParams{
			Cluster: clusterName,
		}, nil
	}

	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil),
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(stackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(stackName).Return(nil),
		mockECS.EXPECT().DeleteCluster(clusterName).Return(clusterName, nil),
	)
	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(command.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := deleteCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	assert.NoError(t, err, "Unexpected error deleting cluster")
}

func TestClusterDownWithoutForce(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	context := cli.NewContext(nil, flagSet, nil)
	err := deleteCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	assert.Error(t, err, "Expected error when force deleting cluster")
}

func TestDeleteClusterPrompt(t *testing.T) {
	readBuffer := bytes.NewBuffer([]byte("yes\ny\nno\n"))
	reader := bufio.NewReader(readBuffer)
	err := deleteClusterPrompt(reader)
	assert.NoError(t, err, "Expected no error with prompt to delete cluster")
	err = deleteClusterPrompt(reader)
	assert.NoError(t, err, "Expected no error with prompt to delete cluster")
	err = deleteClusterPrompt(reader)
	assert.Error(t, err, "Expected error with prompt to delete cluster")
}

func TestClusterScale(t *testing.T) {
	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CLIParams, error) {
		return &config.CLIParams{
			Cluster: clusterName,
		}, nil
	}
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	mockECS.EXPECT().Initialize(gomock.Any())
	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)

	mockCloudformation.EXPECT().Initialize(gomock.Any())
	mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil)
	mockCloudformation.EXPECT().UpdateStack(stackName, gomock.Any()).Return("", nil)
	mockCloudformation.EXPECT().WaitUntilUpdateComplete(stackName).Return(nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")
	flagSet.String(command.AsgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, nil)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	assert.NoError(t, err, "Unexpected error scaling cluster")
}

func TestClusterScaleWithoutIamCapability(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(command.AsgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, nil)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	assert.Error(t, err, "Expected error scaling cluster when iam capability is not specified")
}

func TestClusterScaleWithoutSize(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(command.CapabilityIAMFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	assert.Error(t, err, "Expected error scaling cluster when size is not specified")
}

func TestClusterPSTaskGetInfoFail(t *testing.T) {
	testSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CLIParams, error) {
		return &config.CLIParams{
			Cluster: clusterName,
			Session: testSession,
		}, nil
	}
	defer os.Clearenv()
	mockECS, _ := setupTest(t)

	mockECS.EXPECT().Initialize(gomock.Any())
	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)
	mockECS.EXPECT().GetTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
	}).Return(errors.New("error"))

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)

	context := cli.NewContext(nil, flagSet, nil)
	_, err = clusterPS(context, newMockReadWriter(), mockECS)
	assert.Error(t, err, "Expected error in cluster ps")
}
