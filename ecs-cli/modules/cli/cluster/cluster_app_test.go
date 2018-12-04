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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/cluster/userdata"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ssm"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/ssm/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cloudformationsdk "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	clusterName    = "defaultCluster"
	stackName      = "defaultCluster"
	amiID          = "ami-deadb33f"
	mockedUserData = "some user data"
)

type mockReadWriter struct {
	clusterName       string
	stackName         string
	defaultLaunchType string
}

func (rdwr *mockReadWriter) Get(cluster string, profile string) (*config.LocalConfig, error) {
	cliConfig := config.NewLocalConfig(rdwr.clusterName)
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

type mockUserDataBuilder struct {
	userdata string
	files    []string
}

func (b *mockUserDataBuilder) AddFile(fileName string) error {
	b.files = append(b.files, fileName)
	return nil
}

func (b *mockUserDataBuilder) Build() (string, error) {
	return b.userdata, nil
}

func setupTest(t *testing.T) (*mock_ecs.MockECSClient, *mock_cloudformation.MockCloudformationClient, *mock_ssm.MockClient) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	mockSSM := mock_ssm.NewMockClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	os.Setenv("AWS_REGION", "us-west-1")

	return mockECS, mockCloudformation, mockSSM
}

/////////////////
// Cluster Up //
////////////////

func TestClusterUp(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	mocksForSuccessfulClusterUp(mockECS, mockCloudformation, mockSSM)

	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithForce(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(stackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(stackName).Return(nil),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithoutPublicIP(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Do(func(x, y, w, z interface{}) {
			capabilityIAM := w.(bool)
			cfnParams := z.(*cloudformation.CfnStackParams)
			associateIPAddress, err := cfnParams.GetParameter(ParameterKeyAssociatePublicIPAddress)
			assert.NoError(t, err, "Unexpected error getting cfn parameter")
			assert.Equal(t, "false", aws.StringValue(associateIPAddress.ParameterValue), "Should not associate public IP address")
			assert.True(t, capabilityIAM, "Expected capability capabilityIAM to be true")
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
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithUserData(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	oldNewUserDataBuilder := newUserDataBuilder
	defer func() { newUserDataBuilder = oldNewUserDataBuilder }()
	userdataMock := &mockUserDataBuilder{
		userdata: mockedUserData,
	}
	newUserDataBuilder = func(clusterName string) userdata.UserDataBuilder {
		return userdataMock
	}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Do(func(x, y, w, z interface{}) {
			cfnParams := z.(*cloudformation.CfnStackParams)
			param, err := cfnParams.GetParameter(ParameterKeyUserData)
			assert.NoError(t, err, "Expected User Data parameter to be set")
			assert.Equal(t, mockedUserData, aws.StringValue(param.ParameterValue), "Expected user data to match")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	userDataFiles := &cli.StringSlice{}
	userDataFiles.Set("some_file")
	userDataFiles.Set("some_file2")
	flagSet.Var(userDataFiles, flags.UserDataFlag, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")

	assert.ElementsMatch(t, []string{"some_file", "some_file2"}, userdataMock.files, "Expected userdata file list to match")
}

func TestClusterUpWithSpotPrice(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	spotPrice := "0.03"

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Do(func(x, y, w, z interface{}) {
			cfnParams := z.(*cloudformation.CfnStackParams)
			param, err := cfnParams.GetParameter(ParameterKeySpotPrice)
			assert.NoError(t, err, "Expected Spot Price parameter to be set")
			assert.Equal(t, spotPrice, aws.StringValue(param.ParameterValue), "Expected spot price to match")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.SpotPriceFlag, spotPrice, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	vpcID := "vpc-02dd3038"
	subnetIds := "subnet-04726b21,subnet-04346b21"

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation, mockSSM)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.VpcIdFlag, vpcID, "")
	flagSet.String(flags.SubnetIdsFlag, subnetIds, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithAvailabilityZones(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	vpcAZs := "us-west-2c,us-west-2a"

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation, mockSSM)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithCustomRole(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	instanceRole := "sparklepony"

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation, mockSSM)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.InstanceRoleFlag, instanceRole, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithTwoCustomRoles(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	instanceRole := "sparklepony, sparkleunicorn"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.InstanceRoleFlag, instanceRole, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for custom instance role")
}

func TestClusterUpWithDefaultAndCustomRoles(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	instanceRole := "sparklepony"

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.InstanceRoleFlag, instanceRole, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for custom instance role")
}

func TestClusterUpWithNoRoles(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, false, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for custom instance role")
}

func TestClusterUpWithoutKeyPair(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation, mockSSM)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.Bool(flags.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithSecurityGroupWithoutVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	securityGroupID := "sg-eeaabc8d"

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.SecurityGroupFlag, securityGroupID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for security group without VPC")
}

func TestClusterUpWith2SecurityGroups(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)

	mocksForSuccessfulClusterUp(mockECS, mockCloudformation, mockSSM)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

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
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithSubnetsWithoutVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	subnetID := "subnet-72f52e32"

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.SubnetIdsFlag, subnetID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for subnets without VPC")
}

func TestClusterUpWithVPCWithoutSubnets(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	vpcID := "vpc-02dd3038"

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.VpcIdFlag, vpcID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for VPC without subnets")
}

func TestClusterUpWithAvailabilityZonesWithVPC(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	vpcID := "vpc-02dd3038"
	vpcAZs := "us-west-2c,us-west-2a"

	gomock.InOrder(
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
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for VPC with AZs")
}

func TestClusterUpWithout2AvailabilityZones(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	vpcAZs := "us-west-2c"

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.Bool(flags.ForceFlag, true, "")
	flagSet.String(flags.VpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error for 2 AZs")
}

func TestCliFlagsToCfnStackParams(t *testing.T) {

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, nil)
	params, err := cliFlagsToCfnStackParams(context, clusterName, config.LaunchTypeEC2)
	assert.NoError(t, err, "Unexpected error from call to cliFlagsToCfnStackParams")

	_, err = params.GetParameter(ParameterKeyAsgMaxSize)
	assert.Error(t, err, "Expected error for parameter ParameterKeyAsgMaxSize")
	assert.Equal(t, cloudformation.ParameterNotFoundError, err, "Expect error to be ParameterNotFoundError")

	flagSet.String(flags.AsgMaxSizeFlag, "2", "")
	context = cli.NewContext(nil, flagSet, nil)
	params, err = cliFlagsToCfnStackParams(context, clusterName, config.LaunchTypeEC2)
	assert.NoError(t, err, "Unexpected error from call to cliFlagsToCfnStackParams")
	_, err = params.GetParameter(ParameterKeyAsgMaxSize)
	assert.NoError(t, err, "Unexpected error getting parameter ParameterKeyAsgMaxSize")
}

func TestClusterUpForImageIdInput(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	imageID := "ami-12345"

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(imageID), nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Do(func(x, y, w, z interface{}) {
			capabilityIAM := w.(bool)
			cfnStackParams := z.(*cloudformation.CfnStackParams)
			param, err := cfnStackParams.GetParameter(ParameterKeyAmiId)
			assert.NoError(t, err, "Expected image id params to be present")
			assert.Equal(t, imageID, aws.StringValue(param.ParameterValue), "Expected image id to match")
			assert.True(t, capabilityIAM, "Expected capability capabilityIAM to be true")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")
	flagSet.String(flags.ImageIdFlag, imageID, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithClusterNameEmpty(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.KeypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	rdwr := &mockReadWriter{clusterName: ""}
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error bringing up cluster")
}

func TestClusterUpWithoutRegion(t *testing.T) {
	defer os.Clearenv()
	os.Unsetenv("AWS_REGION")

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	_, err := newCommandConfig(context, rdwr)
	assert.Error(t, err, "Expected error due to missing region in bringing up cluster")
}

func TestClusterUpWithFargateLaunchTypeFlag(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Do(func(x, y, w, z interface{}) {
			cfnParams := z.(*cloudformation.CfnStackParams)
			isFargate, err := cfnParams.GetParameter(ParameterKeyIsFargate)
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
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)

	assert.Equal(t, config.LaunchTypeFargate, commandConfig.LaunchType, "Launch Type should be FARGATE")
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithFargateDefaultLaunchTypeConfig(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: config.LaunchTypeFargate,
	}

	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Do(func(x, y, w, z interface{}) {
			capabilityIAM := w.(bool)
			cfnParams := z.(*cloudformation.CfnStackParams)
			isFargate, err := cfnParams.GetParameter(ParameterKeyIsFargate)
			assert.NoError(t, err, "Unexpected error getting cfn parameter")
			assert.Equal(t, "true", aws.StringValue(isFargate.ParameterValue), "Should have Fargate launch type.")
			assert.True(t, capabilityIAM, "Expected capability capabilityIAM to be true")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
		mockCloudformation.EXPECT().DescribeNetworkResources(stackName).Return(nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)

	assert.Equal(t, config.LaunchTypeFargate, commandConfig.LaunchType, "Launch Type should be FARGATE")
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithFargateLaunchTypeFlagOverride(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: config.LaunchTypeEC2,
	}

	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Do(func(x, y, w, z interface{}) {
			capabilityIAM := w.(bool)
			cfnParams := z.(*cloudformation.CfnStackParams)
			isFargate, err := cfnParams.GetParameter(ParameterKeyIsFargate)
			assert.NoError(t, err, "Unexpected error getting cfn parameter")
			assert.Equal(t, "true", aws.StringValue(isFargate.ParameterValue), "Should have Fargate launch type.")
			assert.True(t, capabilityIAM, "Expected capability capabilityIAM to be true")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
		mockCloudformation.EXPECT().DescribeNetworkResources(stackName).Return(nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.LaunchTypeFlag, config.LaunchTypeFargate, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)

	assert.Equal(t, config.LaunchTypeFargate, commandConfig.LaunchType, "Launch Type should be FARGATE")
	assert.NoError(t, err, "Unexpected error bringing up cluster")
}

func TestClusterUpWithEC2LaunchTypeFlagOverride(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: config.LaunchTypeFargate,
	}

	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.LaunchTypeFlag, config.LaunchTypeEC2, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)

	// This is kind of hack - this error will only get checked if launch type is EC2
	assert.Error(t, err, "Expected error for bringing up cluster with empty default launch type.")
}

func TestClusterUpWithBlankDefaultLaunchTypeConfig(t *testing.T) {
	rdwr := &mockReadWriter{
		clusterName:       clusterName,
		defaultLaunchType: "",
	}

	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, false, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)

	// This is kind of hack - this error will only get checked if launch type is EC2
	assert.Error(t, err, "Expected error for bringing up cluster with empty default launch type.")
}

func TestClusterUpWithEmptyCluster(t *testing.T) {
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.EmptyFlag, true, "")
	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error bringing up empty cluster")
}

func TestClusterUpWithEmptyClusterWithExistingStack(t *testing.T) {
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.EmptyFlag, true, "")
	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = createCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Unexpected error bringing up empty cluster")
}

///////////////////
// Cluster Down //
//////////////////
func TestClusterDown(t *testing.T) {
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}
	defer os.Clearenv()

	gomock.InOrder(
		mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(stackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(stackName).Return(nil),
		mockECS.EXPECT().DeleteCluster(clusterName).Return(clusterName, nil),
	)
	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(flags.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = deleteCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error deleting cluster")
}

func TestClusterDownWithoutForce(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = deleteCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error when force deleting cluster")
}

func TestClusterDownForEmptyCluster(t *testing.T) {
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}
	defer os.Clearenv()

	gomock.InOrder(
		mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockECS.EXPECT().DeleteCluster(clusterName).Return(clusterName, nil),
	)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(flags.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = deleteCluster(context, awsClients, commandConfig)

	assert.NoError(t, err, "Unexpected error deleting cluster")
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

///////////////////
// Cluster Scale //
//////////////////

func TestClusterScale(t *testing.T) {
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}
	defer os.Clearenv()

	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)

	existingParameters := []*cloudformationsdk.Parameter{
		&cloudformationsdk.Parameter{
			ParameterKey: aws.String("SomeParam1"),
		},
		&cloudformationsdk.Parameter{
			ParameterKey: aws.String("SomeParam2"),
		},
	}

	mockCloudformation.EXPECT().GetStackParameters(stackName).Return(existingParameters, nil)
	mockCloudformation.EXPECT().UpdateStack(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
		observedStackName := x.(string)
		cfnParams := y.(*cloudformation.CfnStackParams)
		assert.Equal(t, stackName, observedStackName)
		_, err := cfnParams.GetParameter("SomeParam1")
		assert.NoError(t, err, "Unexpected error on scale.")
		_, err = cfnParams.GetParameter("SomeParam2")
		assert.NoError(t, err, "Unexpected error on scale.")
		param, err := cfnParams.GetParameter(ParameterKeyAsgMaxSize)
		assert.NoError(t, err, "Unexpected error on scale.")
		assert.Equal(t, "1", aws.StringValue(param.ParameterValue))
	}).Return("", nil)
	mockCloudformation.EXPECT().WaitUntilUpdateComplete(stackName).Return(nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")
	flagSet.String(flags.AsgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = scaleCluster(context, awsClients, commandConfig)
	assert.NoError(t, err, "Unexpected error scaling cluster")
}

func TestClusterScaleWithoutIamCapability(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(flags.AsgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = scaleCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error scaling cluster when iam capability is not specified")
}

func TestClusterScaleWithoutSize(t *testing.T) {
	defer os.Clearenv()
	mockECS, mockCloudformation, mockSSM := setupTest(t)
	awsClients := &AWSClients{mockECS, mockCloudformation, mockSSM}

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(flags.CapabilityIAMFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)
	rdwr := newMockReadWriter()
	commandConfig, err := newCommandConfig(context, rdwr)
	assert.NoError(t, err, "Unexpected error creating CommandConfig")

	err = scaleCluster(context, awsClients, commandConfig)
	assert.Error(t, err, "Expected error scaling cluster when size is not specified")
}

/////////////////
// Cluster PS //
////////////////

func TestClusterPSTaskGetInfoFail(t *testing.T) {
	testSession, err := session.NewSession()
	assert.NoError(t, err, "Unexpected error in creating session")

	newCommandConfig = func(context *cli.Context, rdwr config.ReadWriter) (*config.CommandConfig, error) {
		return &config.CommandConfig{
			Cluster: clusterName,
			Session: testSession,
		}, nil
	}
	defer os.Clearenv()
	mockECS, _, _ := setupTest(t)

	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)
	mockECS.EXPECT().GetTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
	}).Return(errors.New("error"))

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)

	context := cli.NewContext(nil, flagSet, nil)
	_, err = clusterPS(context, newMockReadWriter())
	assert.Error(t, err, "Expected error in cluster ps")
}

/////////////////////
// private methods //
/////////////////////

func amiMetadata(imageID string) *ssm.AMIMetadata {
	return &ssm.AMIMetadata{
		ImageID:        imageID,
		OsName:         "Amazon Linux",
		AgentVersion:   "1.7.2",
		RuntimeVersion: "Docker version 17.12.1-ce",
	}
}

func mocksForSuccessfulClusterUp(mockECS *mock_ecs.MockECSClient, mockCloudformation *mock_cloudformation.MockCloudformationClient, mockSSM *mock_ssm.MockClient) {
	gomock.InOrder(
		mockECS.EXPECT().CreateCluster(clusterName).Return(clusterName, nil),
	)
	gomock.InOrder(
		mockSSM.EXPECT().GetRecommendedECSLinuxAMI().Return(amiMetadata(amiID), nil),
	)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), stackName, true, gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)
}
