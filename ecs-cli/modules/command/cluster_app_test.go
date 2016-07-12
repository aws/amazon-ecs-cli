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
	"errors"
	"flag"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/cloudformation/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config/ami"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/codegangsta/cli"
	"github.com/golang/mock/gomock"
)

type mockReadWriter struct{}

func (rdwr *mockReadWriter) GetConfig() (*config.CliConfig, error) {
	return config.NewCliConfig(clusterName), nil
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

func TestClusterUp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockEcs.EXPECT().Initialize(gomock.Any()),
		mockEcs.EXPECT().CreateCluster(gomock.Any()).Do(func(in interface{}) {
			if in.(string) != clusterName {
				t.Fatal("Expected to be called with " + clusterName + " not " + in.(string))
			}
		}).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(gomock.Any()).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), gomock.Any(), gomock.Any()).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(gomock.Any()).Return(nil),
	)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation, ami.NewStaticAmiIds())
	if err != nil {
		t.Fatal("Error bringing up cluster: ", err)
	}
}

func TestClusterUpWithoutKeyPair(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(gomock.Any()).Return(errors.New("error")),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation, ami.NewStaticAmiIds())
	if err == nil {
		t.Fatal("Expected error for key pair name")
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
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	imageId := "ami-12345"

	os.Setenv("AWS_ACCESS_KEY", "AKIDEXAMPLE")
	os.Setenv("AWS_SECRET_KEY", "secret")
	defer func() {
		os.Unsetenv("AWS_ACCESS_KEY")
		os.Unsetenv("AWS_SECRET_KEY")
	}()

	gomock.InOrder(
		mockEcs.EXPECT().Initialize(gomock.Any()),
		mockEcs.EXPECT().CreateCluster(gomock.Any()).Do(func(in interface{}) {
			if in.(string) != clusterName {
				t.Fatal("Expected to be called with " + clusterName + " not " + in.(string))
			}
		}).Return(clusterName, nil),
	)

	gomock.InOrder(
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(gomock.Any()).Return(errors.New("error")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), gomock.Any(), gomock.Any()).Do(func(x, y, z interface{}) {
			cfnStackParams := z.(*cloudformation.CfnStackParams)
			param, err := cfnStackParams.GetParameter(cloudformation.ParameterKeyAmiId)
			if err != nil {
				t.Fatal("Expected image id params to be present")
			}
			if imageId != aws.StringValue(param.ParameterValue) {
				t.Fatalf("Expected image id to equal %s but got %s", imageId, aws.StringValue(param.ParameterValue))
			}
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(gomock.Any()).Return(nil),
	)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(keypairNameFlag, "default", "")
	flagSet.String(imageIdFlag, imageId, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation, ami.NewStaticAmiIds())
	if err != nil {
		t.Fatal("Error bringing up cluster: ", err)
	}
}

func TestClusterUpWithoutRegion(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)

	context := cli.NewContext(nil, flagSet, globalContext)
	err := createCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation, ami.NewStaticAmiIds())
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
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockEcs.EXPECT().Initialize(gomock.Any()),
		mockEcs.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil),
		mockCloudformation.EXPECT().Initialize(gomock.Any()),
		mockCloudformation.EXPECT().ValidateStackExists(gomock.Any()).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(gomock.Any()).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(gomock.Any()).Return(nil),
		mockEcs.EXPECT().DeleteCluster(gomock.Any()).Do(func(in interface{}) {
			if in.(string) != clusterName {
				t.Fatal("Expected to be called with " + clusterName + " not " + in.(string))
			}
		}).Return(clusterName, nil),
	)
	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(forceFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := deleteCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation)
	if err != nil {
		t.Fatal("Error deleting cluster: ", err)
	}
}

func TestClusterDownWithoutForce(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)

	context := cli.NewContext(nil, flagSet, globalContext)
	err := deleteCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation)
	if err == nil {
		t.Fatalf("Expected error deleting cluster when '--%s' is not specified", forceFlag)
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
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	mockEcs.EXPECT().Initialize(gomock.Any())
	mockEcs.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)

	mockCloudformation.EXPECT().Initialize(gomock.Any())
	mockCloudformation.EXPECT().ValidateStackExists(gomock.Any()).Return(nil)
	mockCloudformation.EXPECT().UpdateStack(gomock.Any(), gomock.Any()).Return("", nil)
	mockCloudformation.EXPECT().WaitUntilUpdateComplete(gomock.Any()).Return(nil)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")
	flagSet.String(asgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := scaleCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation)
	if err != nil {
		t.Fatal("Error scaling cluster: ", err)
	}
}

func TestClusterScaleWithoutIamCapability(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.String(asgMaxSizeFlag, "1", "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := scaleCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation)
	if err == nil {
		t.Fatal("Expected error scaling cluster when iam capability is not specified")
	}
}

func TestClusterScaleWithoutSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-up", 0)
	flagSet.Bool(capabilityIAMFlag, true, "")

	context := cli.NewContext(nil, flagSet, globalContext)
	err := scaleCluster(context, &mockReadWriter{}, mockEcs, mockCloudformation)
	if err == nil {
		t.Fatal("Expected error scaling cluster when size is not specified")
	}
}

func TestClusterPSTaskGetInfoFail(t *testing.T) {
	newCliParams = func(context *cli.Context, rdwr config.ReadWriter) (*config.CliParams, error) {
		return &config.CliParams{
			Cluster: clusterName,
		}, nil
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockEcs := mock_ecs.NewMockECSClient(ctrl)

	mockEcs.EXPECT().Initialize(gomock.Any())
	mockEcs.EXPECT().IsActiveCluster(gomock.Any()).Return(true, nil)
	mockEcs.EXPECT().GetTasksPages(gomock.Any(), gomock.Any()).Do(func(x, y interface{}) {
	}).Return(errors.New("error"))

	globalSet := flag.NewFlagSet("ecs-cli", 0)
	globalSet.String("region", "us-west-1", "")
	globalContext := cli.NewContext(nil, globalSet, nil)

	flagSet := flag.NewFlagSet("ecs-cli-down", 0)

	context := cli.NewContext(nil, flagSet, globalContext)
	_, err := clusterPS(context, &mockReadWriter{}, mockEcs)
	if err == nil {
		t.Fatal("Expected error in cluster ps")
	}
}
