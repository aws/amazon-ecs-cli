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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cloudformationsdk "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	clusterName = "defaultCluster"
	stackName   = "defaultCluster"
)

type mockReadWriter struct {
	clusterName       string
	stackName         string
	defaultLaunchType string
}

func (rdwr *mockReadWriter) Get(cluster string, profile string) (*config.CLIConfig, error) {
	cliConfig := config.NewCLIConfig(rdwr.clusterName)
	cliConfig.CFNStackName = rdwr.clusterName
	cliConfig.DefaultLaunchType = rdwr.defaultLaunchType
	return cliConfig, nil
}

func (rdwr *mockReadWriter) SaveProfile(configName string, profile *config.Profile) error {
	return nil
}

func (rdwr *mockReadWriter) SaveCluster(configName string, cluster *config.Cluster) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultProfile(configName string) error {
	return nil
}

func (rdwr *mockReadWriter) SetDefaultCluster(configName string) error {
	return nil
}

func newMockReadWriter() *mockReadWriter {
	return &mockReadWriter{
		clusterName: clusterName,
	}
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

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.NoAutoAssignPublicIPAddressFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	vpcID := "vpc-02dd3038"
	subnetIds := "subnet-04726b21,subnet-04346b21"

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.VpcIdFlag, vpcID, "")
	flagSet.String(flags.SubnetIdsFlag, subnetIds, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithAvailabilityZones(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	vpcAZs := "us-west-2c,us-west-2a"

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithCustomRole(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	instanceRole := "sparklepony"

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.InstanceRoleFlag, instanceRole, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithTwoCustomRoles(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	instanceRole := "sparklepony, sparkleunicorn"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.InstanceRoleFlag, instanceRole, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.Error(t, err, "Expected error for custom instance role")
}

func TestClusterUpWithDefaultAndCustomRoles(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	instanceRole := "sparklepony"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.InstanceRoleFlag, instanceRole, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.Error(t, err, "Expected error for custom instance role")
}

func TestClusterUpWithNoRoles(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, false, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.Error(t, err, "Expected error for custom instance role")
}

func TestClusterUpWithoutKeyPair(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.Bool(flags.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.SecurityGroupFlag, securityGroupID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.Error(t, err, "Expected error for security group without VPC")
}

func TestClusterUpWith2SecurityGroups(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation)

	securityGroupIds := "sg-eeaabc8d,sg-eaaebc8d"
	vpcId := "vpc-02dd3038"
	subnetIds := "subnet-04726b21,subnet-04346b21"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.SecurityGroupFlag, securityGroupIds, "")
	flagSet.String(flags.VpcIdFlag, vpcId, "")
	flagSet.String(flags.SubnetIdsFlag, subnetIds, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.SubnetIdsFlag, subnetID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.VpcIdFlag, vpcID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.VpcIdFlag, vpcID, "")
	flagSet.String(flags.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.Error(t, err, "Expected error for 2 AZs")
}

func TestCliFlagsToCfnStackParams(t *testing.T) {

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	params := cliFlagsToCfnStackParams(context)

	_, err := params.GetParameter(cloudformation.ParameterKeyAsgMaxSize)
	assert.Error(t, err, "Expected error for parameter ParameterKeyAsgMaxSize")
	assert.Equal(t, cloudformation.ParameterNotFoundError, err, "Expect error to be ParameterNotFoundError")

	flagSet.String(flags.AsgMaxSizeFlag, "2", "")
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
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.ImageIdFlag, imageID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithClusterNameEmpty(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := &mockReadWriter{clusterName: ""}
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)
	assert.Error(t, err, "Expected error bringing up cluster")
}

func TestClusterUpWithoutRegion(t *testing.T) {
	defer os.Clearenv()
	os.Unsetenv("AWS_REGION")

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	_, err := newCliParams(context, rdwr)
	assert.Error(t, err, "Expected error due to missing region in bringing up cluster")
}

func TestClusterUpWithFargateLaunchTypeFlag(t *testing.T) {
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
			isFargate, err := cfnParams.GetParameter(cloudformation.ParameterKeyIsFargate)
			assert.NoError(t, err, "Unexpected error getting cfn parameter")
			assert.Equal(t, "true", aws.StringValue(isFargate.ParameterValue), "Should have Fargate launch type.")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
		mockCloudformation.EXPECT().DescribeNetworkResources(stackName).Return(nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.LaunchTypeFlag, config.LaunchTypeFargate, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := newMockReadWriter()
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)

	assert.Equal(t, config.LaunchTypeFargate, cliParams.LaunchType, "Launch Type should be FARGATE")
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithFargateDefaultLaunchTypeConfig(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: config.LaunchTypeFargate,
	}

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
			isFargate, err := cfnParams.GetParameter(cloudformation.ParameterKeyIsFargate)
			assert.NoError(t, err, "Unexpected error getting cfn parameter")
			assert.Equal(t, "true", aws.StringValue(isFargate.ParameterValue), "Should have Fargate launch type.")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
		mockCloudformation.EXPECT().DescribeNetworkResources(stackName).Return(nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)

	assert.Equal(t, config.LaunchTypeFargate, cliParams.LaunchType, "Launch Type should be FARGATE")
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithFargateLaunchTypeFlagOverride(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: config.LaunchTypeEC2,
	}

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
			isFargate, err := cfnParams.GetParameter(cloudformation.ParameterKeyIsFargate)
			assert.NoError(t, err, "Unexpected error getting cfn parameter")
			assert.Equal(t, "true", aws.StringValue(isFargate.ParameterValue), "Should have Fargate launch type.")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
		mockCloudformation.EXPECT().DescribeNetworkResources(stackName).Return(nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.LaunchTypeFlag, config.LaunchTypeFargate, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)

	assert.Equal(t, config.LaunchTypeFargate, cliParams.LaunchType, "Launch Type should be FARGATE")
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithEC2LaunchTypeFlagOverride(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: config.LaunchTypeFargate,
	}

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
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.LaunchTypeFlag, config.LaunchTypeEC2, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)

	// This is kind of hack - this error will only get checked if launch type is EC2
	assert.Error(t, err, "Expected error for bringing up cluster with empty default launch type.")
}

func TestClusterUpWithBlankDefaultLaunchTypeConfig(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: "",
	}

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
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, false, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	cliParams, err := newCliParams(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CLIParams")

	err = createCluster(context, rdwr, mockECS, mockCloudformation, ami.NewStaticAmiIds(), cliParams)

	// This is kind of hack - this error will only get checked if launch type is EC2
	assert.Error(t, err, "Expected error for bringing up cluster with empty default launch type.")
}

func TestClusterDown(t *testing.T) {
	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CLIParams, error) {
		return &config.CLIParams{
			Cluster:      clusterName,
			CFNStackName: stackName,
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
	flagSet.Bool(flags.ForceFlag, true, "")

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
			Cluster:      clusterName,
			CFNStackName: stackName,
		}, nil
	}
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	mockECS.EXPECT().Initialize(gomock.Any())
	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)

	existingParameters := []*cloudformationsdk.Parameter{
		&cloudformationsdk.Parameter{
			ParameterKey: aws.String("SomeParam1"),
		},
		&cloudformationsdk.Parameter{
			ParameterKey: aws.String("SomeParam2"),
		},
	}

	mockCloudformation.EXPECT().Initialize(gomock.Any())
	mockCloudformation.EXPECT().GetStackParameters(stackName).Return(existingParameters, nil)
	mockCloudformation.EXPECT().UpdateStack(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		observedStackName := x.(string)
		cfnParams := y.(*cloudformation.CfnStackParams)
		assert.Equal(t, stackName, observedStackName)
		_, err := cfnParams.GetParameter("SomeParam1")
		assert.NoError(t, err, "Unexpected error on scale.")
		_, err = cfnParams.GetParameter("SomeParam2")
		assert.NoError(t, err, "Unexpected error on scale.")
		param, err := cfnParams.GetParameter(cloudformation.ParameterKeyAsgMaxSize)
		assert.NoError(t, err, "Unexpected error on scale.")
		assert.Equal(t, "1", aws.StringValue(param.ParameterValue))
	}).Return("", nil)
	mockCloudformation.EXPECT().WaitUntilUpdateComplete(stackName).Return(nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.AsgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, nil)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	assert.NoError(t, err, "Unexpected error scaling cluster")
}

func TestClusterScaleWithoutIamCapability(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.AsgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, nil)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	assert.Error(t, err, "Expected error scaling cluster when iam capability is not specified")
}

func TestClusterScaleWithoutSize(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation := setupTest(t)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")

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

func mocksForSuccessfulClusterUp(mockECS *mock_ecs.MockECSClient, mockCloudformation *mock_cloudformation.MockCloudformationClient) {
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
}
