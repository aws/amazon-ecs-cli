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
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/golang/mock/gomock"
	"github.com/urfave/cli"
)

type mockReadWriter struct {
	clusterName string
}

func (rdwr *mockReadWriter) GetConfig() (*config.CliConfig, error) {
	return config.NewCliConfig(rdwr.clusterName), nil
}

func (rdwr *mockReadWriter) ReadFrom(ecsConfig *config.CliConfig) error {
	return nil
}

func (rdwr *mockReadWriter) IsInitialized() (bool, error) {
	return true, nil
}

func (rdwr *mockReadWriter) Save(dest *config.Destination) error {
	return nil
}

func (rdwr *mockReadWriter) IsKeyPresent(section, key string) bool {
	return true
}

func newMockReadWriter() *mockReadWriter {
	return &mockReadWriter{clusterName: clusterName}
}

func TestClusterUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

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
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err != nil {
		t.Fatal("Error bringing up cluster: ", err)
	}
}

func TestClusterUpWithForce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

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

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err != nil {
		t.Fatal("Error bringing up cluster: ", err)
	}
}

func TestClusterUpWithVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	vpcId := "vpc-02dd3038"
	subnetIds := "subnet-04726b21,subnet-04346b21"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

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
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.String(vpcIdFlag, vpcId, "")
	flagSet.String(subnetIdsFlag, subnetIds, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err != nil {
		t.Fatal("Error bringing up cluster: ", err)
	}
}

func TestClusterUpWithAvailabilityZones(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	vpcAZs := "us-west-2c,us-west-2a"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

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
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.String(vpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err != nil {
		t.Fatal("Error bringing up cluster: ", err)
	}
}

func TestClusterUpWithoutKeyPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.Bool(forceFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for key pair name")
	}
}

func TestClusterUpWithSecurityGroupWithoutVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	securityGroupId := "sg-eeaabc8d"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")
	flagSet.String(securityGroupFlag, securityGroupId, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for security group without VPC")
	}
}

func TestClusterUpWith2SecurityGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	securityGroupIds := "sg-eeaabc8d,sg-eaaebc8d"
	vpcId := "vpc-02dd3038"
	subnetIds := "subnet-04726b21,subnet-04346b21"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")
	flagSet.String(securityGroupFlag, securityGroupIds, "")
	flagSet.String(vpcIdFlag, vpcId, "")
	flagSet.String(subnetIdsFlag, subnetIds, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for security group without VPC")
	}
}

func TestClusterUpWithSubnetsWithoutVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	subnetId := "subnet-72f52e32"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")
	flagSet.String(subnetIdsFlag, subnetId, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for subnets without VPC")
	}
}

func TestClusterUpWithVPCWithoutSubnets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	vpcId := "vpc-02dd3038"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")
	flagSet.String(vpcIdFlag, vpcId, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for VPC without subnets")
	}
}

func TestClusterUpWithout2Subnets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	vpcId := "vpc-02dd3038"
	subnetId := "subnet-04726b21"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")
	flagSet.String(vpcIdFlag, vpcId, "")
	flagSet.String(subnetIdsFlag, subnetId, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for 2 subnets")
	}
}

func TestClusterUpWithAvailabilityZonesWithVPC(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	vpcId := "vpc-02dd3038"
	vpcAZs := "us-west-2c,us-west-2a"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")
	flagSet.String(vpcIdFlag, vpcId, "")
	flagSet.String(vpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for VPC with AZs")
	}
}

func TestClusterUpWithout2AvailabilityZones(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	vpcAZs := "us-west-2c"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.Bool(forceFlag, true, "")
	flagSet.String(vpcAzFlag, vpcAZs, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for 2 AZs")
	}
}

func TestCliFlagsToCfnStackParams(t *testing.T) {
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	params := cliFlagsToCfnStackParams(context)

	_, err := params.GetParameter(cloudformation.ParameterKeyAsgMaxSize)
	if err == nil {
		t.Fatalf("Expected error for parameter '%s'", cloudformation.ParameterKeyAsgMaxSize)
	}
	if cloudformation.ParameterNotFoundError != err {
		t.Error("Enexpected error returned: ", err)
	}

	flagSet.String(asgMaxSizeFlag, "2", "")
	context = cli.NewContext(nil, flagSet, globalContext)
	params = cliFlagsToCfnStackParams(context)
	_, err = params.GetParameter(cloudformation.ParameterKeyAsgMaxSize)
	if err != nil {
		t.Error("Error getting parameter '%s'", cloudformation.ParameterKeyAsgMaxSize)
	}
}

func TestClusterUpForImageIdInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	imageId := "ami-12345"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

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
			if err != nil {
				t.Fatal("Expected image id params to be present")
			}
			if imageId != aws.StringValue(param.ParameterValue) {
				t.Fatalf("Expected image id to equal %s but got %s", imageId, aws.StringValue(param.ParameterValue))
			}
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(stackName).Return(nil),
	)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.String(imageIdFlag, imageId, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err != nil {
		t.Fatal("Error bringing up cluster: ", err)
	}
}

func TestClusterUpWithClusterNameEmpty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, &mockReadWriter{clusterName: ""}, mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error bringing up cluster")
	}
}

func TestClusterUpWithoutRegion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, newMockReadWriter(), mockECS, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error bringing up cluster")
	}
}

func TestClusterDown(t *testing.T) {
	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CliParams, error) {
		return &config.CliParams{
			Cluster: clusterName,
		}, nil
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	gomock.InOrder(
		mockECS.EXPECT().Initialize(gomock.Any()),
		mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil),
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(stackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(stackName).Return(nil),
		mockECS.EXPECT().DeleteCluster(clusterName).Return(clusterName, nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(forceFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := deleteCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	if err != nil {
		t.Fatal("Error deleting cluster: ", err)
	}
}

func TestClusterDownWithoutForce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)

	context := cli.NewContext(nil, flagSet, globalContext)
	err := deleteCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	if err == nil {
		t.Fatalf("Expected error deleting cluster when '--%s' is not specified", forceFlag)
	}
}

func TestDeleteClusterPrompt(t *testing.T) {
	readBuffer := bytes.NewBuffer([]byte("yes\ny\nno\n"))
	reader := bufio.NewReader(readBuffer)
	if err := deleteClusterPrompt(reader); err != nil {
		t.Error("Expected no error with prompt to delete cluster")
	}
	if err := deleteClusterPrompt(reader); err != nil {
		t.Error("Expected no error with prompt to delete cluster")
	}
	if err := deleteClusterPrompt(reader); err == nil {
		t.Error("Expected error with prompt to delete cluster")
	}
}

func TestClusterScale(t *testing.T) {
	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CliParams, error) {
		return &config.CliParams{
			Cluster: clusterName,
		}, nil
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	mockECS.EXPECT().Initialize(gomock.Any())
	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)

	mockCloudformation.EXPECT().Initialize(gomock.Any())
	mockCloudformation.EXPECT().ValidateStackExists(stackName).Return(nil)
	mockCloudformation.EXPECT().UpdateStack(stackName, gomock.Any()).Return("", nil)
	mockCloudformation.EXPECT().WaitUntilUpdateComplete(stackName).Return(nil)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(asgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	if err != nil {
		t.Fatal("Error scaling cluster: ", err)
	}
}

func TestClusterScaleWithoutIamCapability(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(asgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	if err == nil {
		t.Fatal("Expected error scaling cluster when iam capability is not specified")
	}
}

func TestClusterScaleWithoutSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := scaleCluster(context, newMockReadWriter(), mockECS, mockCloudformation)
	if err == nil {
		t.Fatal("Expected error scaling cluster when size is not specified")
	}
}

func TestClusterPSTaskGetInfoFail(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	testSession, err := session.NewSession()
	if err != nil {
		t.Fatal("Unexpected error in creating session")
	}

	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CliParams, error) {
		return &config.CliParams{
			Cluster: clusterName,
			Session: testSession,
		}, nil
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockECS := mock_ecs.NewMockECSClient(ctrl)
	mockECS.EXPECT().Initialize(gomock.Any())
	mockECS.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)
	mockECS.EXPECT().GetTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
	}).Return(errors.New("error"))

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)

	context := cli.NewContext(nil, flagSet, globalContext)
	_, err = clusterPS(context, newMockReadWriter(), mockECS)
	if err == nil {
		t.Fatal("Expected error in cluster ps")
	}
}
