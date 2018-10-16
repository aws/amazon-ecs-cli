// Copyright 2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package servicediscovery

import (
	"flag"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	utils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	sdk "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	testClusterName        = "cluster"
	testServiceName        = "service"
	testDescription        = "clyde loves pudding"
	otherTestDescription   = "pudding plays hard to get"
	testNamespaceStackName = "amazon-ecs-cli-setup-private-dns-namespace-cluster-service"
	testSDSStackName       = "amazon-ecs-cli-setup-service-discovery-service-cluster-service"
	testNamespaceName      = "corp"
	otherTestNamespaceName = "dumpling"
	testVPCID              = "vpc-8BAADF00D"
	otherTestVPCID         = "vpc-F33DFAC3"
	testNamespaceID        = "ns-CA15CA15CA15CA15"
	otherTestNamespaceID   = "ns-D0G5D0G5D0G5D0G5D0G5"
	testPublicNamespaceID  = "ns-PUBL1C"
	testSDSARN             = "arn:aws:servicediscovery:eu-west-1:11111111111:service/srv-clydelovespudding"
	testContainerName      = "my-container"
	otherTestContainerName = "my-other-container"
)

type validateSDSParamsFunc func(*testing.T, *cloudformation.CfnStackParams)
type validateNamespaceParamsFunc func(*testing.T, *cloudformation.CfnStackParams)

func TestCreateServiceDiscoveryAWSVPC(t *testing.T) {
	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceName, parameterKeyNamespaceName, cfnParams, t)
		validateCFNParam(testVPCID, parameterKeyVPCID, cfnParams, t)
	}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeA, parameterKeyDNSType, cfnParams, t)
	}

	oldFindPrivateNamespace := findPrivateNamespace
	defer func() { findPrivateNamespace = oldFindPrivateNamespace }()
	findPrivateNamespace = func(name, vpc string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does pre-exist
		return nil, nil
	}

	registry, err := testCreateServiceDiscovery(t, "awsvpc", &utils.ServiceDiscovery{}, simpleWorkflowContext(), validateNamespace, validateSDS, true)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
}

func TestCreateServiceDiscoveryBridge(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
	}
	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceName, parameterKeyNamespaceName, cfnParams, t)
		validateCFNParam(testVPCID, parameterKeyVPCID, cfnParams, t)
	}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeSrv, parameterKeyDNSType, cfnParams, t)
	}

	oldFindPrivateNamespace := findPrivateNamespace
	defer func() { findPrivateNamespace = oldFindPrivateNamespace }()
	findPrivateNamespace = func(name, vpc string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does not pre-exist
		return nil, nil
	}

	registry, err := testCreateServiceDiscovery(t, "bridge", input, simpleWorkflowContext(), validateNamespace, validateSDS, true)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	assert.Equal(t, testContainerName, aws.StringValue(registry.ContainerName), "Expected container name to match")
	assert.Equal(t, int64(80), aws.Int64Value(registry.ContainerPort), "Expected container port to match")
}

func TestCreateServiceDiscoveryHost(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
	}
	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceName, parameterKeyNamespaceName, cfnParams, t)
		validateCFNParam(testVPCID, parameterKeyVPCID, cfnParams, t)
	}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeSrv, parameterKeyDNSType, cfnParams, t)
	}

	oldFindPrivateNamespace := findPrivateNamespace
	defer func() { findPrivateNamespace = oldFindPrivateNamespace }()
	findPrivateNamespace = func(name, vpc string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does not pre-exist
		return nil, nil
	}

	registry, err := testCreateServiceDiscovery(t, "host", input, simpleWorkflowContext(), validateNamespace, validateSDS, true)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	assert.Equal(t, testContainerName, aws.StringValue(registry.ContainerName), "Expected container name to match")
	assert.Equal(t, int64(80), aws.Int64Value(registry.ContainerPort), "Expected container port to match")
}

func TestCreateServiceDiscoveryForceRecreate(t *testing.T) {
	oldFindPrivateNamespace := findPrivateNamespace
	defer func() { findPrivateNamespace = oldFindPrivateNamespace }()
	findPrivateNamespace = func(name, vpc string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does pre-exist
		return nil, nil
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	describeNamespaceStackResponse := &sdk.DescribeStacksOutput{
		Stacks: []*sdk.Stack{
			&sdk.Stack{
				Outputs: []*sdk.Output{
					&sdk.Output{
						OutputKey:   aws.String(cfnTemplateOutputPrivateNamespaceID),
						OutputValue: aws.String(testNamespaceID),
					},
				},
			},
		},
	}

	describeSDSStackResponse := &sdk.DescribeStacksOutput{
		Stacks: []*sdk.Stack{
			&sdk.Stack{
				Outputs: []*sdk.Output{
					&sdk.Output{
						OutputKey:   aws.String(cfnTemplateOutputSDSARN),
						OutputValue: aws.String(testSDSARN),
					},
				},
			},
		},
	}

	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(testNamespaceStackName).Return(nil),
		// validate that existing SDS stack is deleted
		mockCloudformation.EXPECT().DeleteStack(testNamespaceStackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(testNamespaceStackName).Return(nil),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), testNamespaceStackName, false, gomock.Any()).Do(func(x, y, w, z interface{}) {
			stackName := y.(string)
			capabilityIAM := w.(bool)
			cfnParams := z.(*cloudformation.CfnStackParams)
			validateCFNParam(testNamespaceName, parameterKeyNamespaceName, cfnParams, t)
			validateCFNParam(testVPCID, parameterKeyVPCID, cfnParams, t)
			assert.False(t, capabilityIAM, "Expected capability capabilityIAM to be false")
			assert.Equal(t, testNamespaceStackName, stackName, "Expected stack name to match")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(testNamespaceStackName).Return(nil),
		mockCloudformation.EXPECT().DescribeStacks(testNamespaceStackName).Return(describeNamespaceStackResponse, nil),
		mockCloudformation.EXPECT().ValidateStackExists(testSDSStackName).Return(nil),
		// Validate that existing Namespace stack is deleted
		mockCloudformation.EXPECT().DeleteStack(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), testSDSStackName, false, gomock.Any()).Do(func(x, y, w, z interface{}) {
			stackName := y.(string)
			capabilityIAM := w.(bool)
			cfnParams := z.(*cloudformation.CfnStackParams)
			validateCFNParam(testNamespaceID, parameterKeyNamespaceID, cfnParams, t)
			validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
			validateCFNParam(servicediscovery.RecordTypeA, parameterKeyDNSType, cfnParams, t)
			assert.False(t, capabilityIAM, "Expected capability capabilityIAM to be false")
			assert.Equal(t, testSDSStackName, stackName, "Expected stack name to match")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().DescribeStacks(testSDSStackName).Return(describeSDSStackResponse, nil),
	)

	config := &config.CommandConfig{
		Cluster: testClusterName,
	}

	registry, err := create(simpleWorkflowContext(), "awsvpc", testServiceName, mockCloudformation, &utils.ServiceDiscovery{}, config)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
}

func TestCreateServiceDiscoveryWithECSParams(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
		PrivateDNSNamespace: utils.PrivateDNSNamespace{
			Namespace: utils.Namespace{
				Name: testNamespaceName,
			},
			VPC:         testVPCID,
			Description: testDescription,
		},
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			Name:        testServiceName,
			Description: testDescription,
			DNSConfig: utils.DNSConfig{
				TTL:  aws.Int64(60),
				Type: servicediscovery.RecordTypeSrv,
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceName, parameterKeyNamespaceName, cfnParams, t)
		validateCFNParam(testVPCID, parameterKeyVPCID, cfnParams, t)
		validateCFNParam(testDescription, parameterKeyNamespaceDescription, cfnParams, t)
	}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeSrv, parameterKeyDNSType, cfnParams, t)
		validateCFNParam(testDescription, parameterKeySDSDescription, cfnParams, t)
		validateCFNParam("60", parameterKeyDNSTTL, cfnParams, t)
		validateCFNParam("2", parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
	}

	oldFindPrivateNamespace := findPrivateNamespace
	defer func() { findPrivateNamespace = oldFindPrivateNamespace }()
	findPrivateNamespace = func(name, vpc string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does not pre-exist
		return nil, nil
	}

	registry, err := testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, true)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	assert.Equal(t, testContainerName, aws.StringValue(registry.ContainerName), "Expected container name to match")
	assert.Equal(t, int64(80), aws.Int64Value(registry.ContainerPort), "Expected container port to match")
}

func TestCreateServiceDiscoveryWithECSParamsOverriddenByFlags(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
		PrivateDNSNamespace: utils.PrivateDNSNamespace{
			Namespace: utils.Namespace{
				Name: testNamespaceName,
			},
			VPC:         testVPCID,
			Description: testDescription,
		},
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			Name:        testServiceName,
			Description: testDescription,
			DNSConfig: utils.DNSConfig{
				TTL:  aws.Int64(60),
				Type: servicediscovery.RecordTypeSrv,
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	flagDNSTTL := "120"
	flagContPort := "22"
	flagHealthThreshold := "3"
	flagSet := flag.NewFlagSet("create-sd", 0)
	flagSet.String(flags.PrivateDNSNamespaceNameFlag, otherTestNamespaceName, "")
	flagSet.String(flags.VpcIdFlag, otherTestVPCID, "")
	flagSet.String(flags.DNSTTLFlag, flagDNSTTL, "")
	flagSet.String(flags.DNSTypeFlag, servicediscovery.RecordTypeA, "")
	flagSet.String(flags.ServiceDiscoveryContainerNameFlag, otherTestContainerName, "")
	flagSet.String(flags.ServiceDiscoveryContainerPortFlag, flagContPort, "")
	flagSet.String(flags.HealthcheckCustomConfigFailureThresholdFlag, flagHealthThreshold, "")
	flagSet.Bool(flags.ForceFlag, true, "")

	overrides := cli.NewContext(nil, flagSet, nil)

	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(otherTestNamespaceName, parameterKeyNamespaceName, cfnParams, t)
		validateCFNParam(otherTestVPCID, parameterKeyVPCID, cfnParams, t)
		validateCFNParam(testDescription, parameterKeyNamespaceDescription, cfnParams, t)
	}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeA, parameterKeyDNSType, cfnParams, t)
		validateCFNParam(testDescription, parameterKeySDSDescription, cfnParams, t)
		validateCFNParam(flagDNSTTL, parameterKeyDNSTTL, cfnParams, t)
		validateCFNParam(flagHealthThreshold, parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
	}

	oldFindPrivateNamespace := findPrivateNamespace
	defer func() { findPrivateNamespace = oldFindPrivateNamespace }()
	findPrivateNamespace = func(name, vpc string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does not pre-exist
		return nil, nil
	}

	registry, err := testCreateServiceDiscovery(t, "awsvpc", input, overrides, validateNamespace, validateSDS, true)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	assert.Equal(t, otherTestContainerName, aws.StringValue(registry.ContainerName), "Expected container name to match")
	assert.Equal(t, int64(22), aws.Int64Value(registry.ContainerPort), "Expected container port to match")
}

func TestCreateServiceDiscoveryWithECSParamsExistingPrivateNamespaceByID(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
		PrivateDNSNamespace: utils.PrivateDNSNamespace{
			Namespace: utils.Namespace{
				ID: otherTestNamespaceID,
			},
		},
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			Name:        testServiceName,
			Description: testDescription,
			DNSConfig: utils.DNSConfig{
				TTL:  aws.Int64(60),
				Type: servicediscovery.RecordTypeSrv,
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(otherTestNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeSrv, parameterKeyDNSType, cfnParams, t)
		validateCFNParam(testDescription, parameterKeySDSDescription, cfnParams, t)
		validateCFNParam("60", parameterKeyDNSTTL, cfnParams, t)
		validateCFNParam("2", parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
	}

	registry, err := testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	assert.Equal(t, testContainerName, aws.StringValue(registry.ContainerName), "Expected container name to match")
	assert.Equal(t, int64(80), aws.Int64Value(registry.ContainerPort), "Expected container port to match")
}

func TestCreateServiceDiscoveryWithECSParamsExistingPrivateNamespaceByName(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
		PrivateDNSNamespace: utils.PrivateDNSNamespace{
			Namespace: utils.Namespace{
				Name: otherTestNamespaceName,
			},
			VPC: otherTestVPCID,
		},
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			Name:        testServiceName,
			Description: testDescription,
			DNSConfig: utils.DNSConfig{
				TTL:  aws.Int64(60),
				Type: servicediscovery.RecordTypeSrv,
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(otherTestNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeSrv, parameterKeyDNSType, cfnParams, t)
		validateCFNParam(testDescription, parameterKeySDSDescription, cfnParams, t)
		validateCFNParam("60", parameterKeyDNSTTL, cfnParams, t)
		validateCFNParam("2", parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
	}

	oldFindPrivateNamespace := findPrivateNamespace
	defer func() { findPrivateNamespace = oldFindPrivateNamespace }()
	findPrivateNamespace = func(name, vpc string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does pre-exist and we search for it
		return aws.String(otherTestNamespaceID), nil
	}

	registry, err := testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	assert.Equal(t, testContainerName, aws.StringValue(registry.ContainerName), "Expected container name to match")
	assert.Equal(t, int64(80), aws.Int64Value(registry.ContainerPort), "Expected container port to match")
}

func TestCreateServiceDiscoveryWithECSParamsExistingPublicNamespaceByID(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
		PublicDNSNamespace: utils.PublicDNSNamespace{
			Namespace: utils.Namespace{
				ID: otherTestNamespaceID,
			},
		},
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			Name:        testServiceName,
			Description: testDescription,
			DNSConfig: utils.DNSConfig{
				TTL:  aws.Int64(60),
				Type: servicediscovery.RecordTypeSrv,
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(otherTestNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeSrv, parameterKeyDNSType, cfnParams, t)
		validateCFNParam(testDescription, parameterKeySDSDescription, cfnParams, t)
		validateCFNParam("60", parameterKeyDNSTTL, cfnParams, t)
		validateCFNParam("2", parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
	}

	registry, err := testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	assert.Equal(t, testContainerName, aws.StringValue(registry.ContainerName), "Expected container name to match")
	assert.Equal(t, int64(80), aws.Int64Value(registry.ContainerPort), "Expected container port to match")
}

func TestCreateServiceDiscoveryWithECSParamsExistingPublicNamespaceByName(t *testing.T) {
	input := &utils.ServiceDiscovery{
		PublicDNSNamespace: utils.PublicDNSNamespace{
			Namespace: utils.Namespace{
				Name: otherTestNamespaceName,
			},
		},
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			Name:        testServiceName,
			Description: testDescription,
			DNSConfig: utils.DNSConfig{
				TTL:  aws.Int64(60),
				Type: servicediscovery.RecordTypeA,
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	var validateNamespace validateNamespaceParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {}
	var validateSDS validateSDSParamsFunc = func(t *testing.T, cfnParams *cloudformation.CfnStackParams) {
		validateCFNParam(otherTestNamespaceID, parameterKeyNamespaceID, cfnParams, t)
		validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
		validateCFNParam(servicediscovery.RecordTypeA, parameterKeyDNSType, cfnParams, t)
		validateCFNParam(testDescription, parameterKeySDSDescription, cfnParams, t)
		validateCFNParam("60", parameterKeyDNSTTL, cfnParams, t)
		validateCFNParam("2", parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
	}

	oldFindPublicNamespace := findPublicNamespace
	defer func() { findPublicNamespace = oldFindPublicNamespace }()
	findPublicNamespace = func(name string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does pre-exist and we search for it
		return aws.String(otherTestNamespaceID), nil
	}

	registry, err := testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(registry.RegistryArn), "Expected SDS ARN to match")
	logrus.Info(aws.StringValue(registry.ContainerName))
	logrus.Info(aws.StringValue(registry.ContainerName) == "")
	assert.Nil(t, registry.ContainerName, "Expected container name to be nil")
	assert.Nil(t, registry.ContainerPort, "Expected container port to be nil")
}

func TestUpdateServiceDiscovery(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			DNSConfig: utils.DNSConfig{
				TTL: aws.Int64(120),
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	existingParameters := []*sdk.Parameter{
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeySDSDescription),
			ParameterValue: aws.String(testDescription),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeySDSName),
			ParameterValue: aws.String(testServiceName),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyNamespaceID),
			ParameterValue: aws.String(testNamespaceID),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyDNSType),
			ParameterValue: aws.String(servicediscovery.RecordTypeA),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyDNSTTL),
			ParameterValue: aws.String("60"),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyHealthCheckCustomConfigFailureThreshold),
			ParameterValue: aws.String("1"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().GetStackParameters(testSDSStackName).Return(existingParameters, nil),
		mockCloudformation.EXPECT().UpdateStack(testSDSStackName, gomock.Any()).Do(func(x, y interface{}) {
			cfnParams := y.(*cloudformation.CfnStackParams)
			validateCFNParam("120", parameterKeyDNSTTL, cfnParams, t)
			validateCFNParam("2", parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
			validateUsePreviousValueSet(parameterKeyDNSType, cfnParams, t)
			validateUsePreviousValueSet(parameterKeySDSDescription, cfnParams, t)
			validateUsePreviousValueSet(parameterKeySDSName, cfnParams, t)
			validateUsePreviousValueSet(parameterKeyNamespaceID, cfnParams, t)
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilUpdateComplete(testSDSStackName).Return(nil),
	)

	err := update(emptyContext(), "awsvpc", testServiceName, testClusterName, mockCloudformation, input)
	assert.NoError(t, err, "Unexpected error calling update")
}

func TestUpdateServiceDiscoveryGetStackParametersError(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			DNSConfig: utils.DNSConfig{
				TTL: aws.Int64(120),
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().GetStackParameters(testSDSStackName).Return(nil, fmt.Errorf("Stack not found")),
	)

	err := update(emptyContext(), "awsvpc", testServiceName, testClusterName, mockCloudformation, input)
	assert.Error(t, err, "Expected error calling update")
}

func TestUpdateServiceDiscoveryUpdateStackError(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ServiceDiscoveryService: utils.ServiceDiscoveryService{
			DNSConfig: utils.DNSConfig{
				TTL: aws.Int64(120),
			},
			HealthCheckCustomConfig: utils.HealthCheckCustomConfig{
				FailureThreshold: aws.Int64(2),
			},
		},
	}

	existingParameters := []*sdk.Parameter{
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeySDSDescription),
			ParameterValue: aws.String(testDescription),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeySDSName),
			ParameterValue: aws.String(testServiceName),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyNamespaceID),
			ParameterValue: aws.String(testNamespaceID),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyDNSType),
			ParameterValue: aws.String(servicediscovery.RecordTypeSrv),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyDNSTTL),
			ParameterValue: aws.String("60"),
		},
		&sdk.Parameter{
			ParameterKey:   aws.String(parameterKeyHealthCheckCustomConfigFailureThreshold),
			ParameterValue: aws.String("1"),
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().GetStackParameters(testSDSStackName).Return(existingParameters, nil),
		mockCloudformation.EXPECT().UpdateStack(testSDSStackName, gomock.Any()).Do(func(x, y interface{}) {
			cfnParams := y.(*cloudformation.CfnStackParams)
			validateCFNParam("120", parameterKeyDNSTTL, cfnParams, t)
			validateCFNParam("2", parameterKeyHealthCheckCustomConfigFailureThreshold, cfnParams, t)
			validateUsePreviousValueSet(parameterKeyDNSType, cfnParams, t)
			validateUsePreviousValueSet(parameterKeySDSDescription, cfnParams, t)
			validateUsePreviousValueSet(parameterKeySDSName, cfnParams, t)
			validateUsePreviousValueSet(parameterKeyNamespaceID, cfnParams, t)
		}).Return("", fmt.Errorf("Some error")),
	)

	err := update(emptyContext(), "host", testServiceName, testClusterName, mockCloudformation, input)
	assert.Error(t, err, "Expected error calling update")
}

func TestDeleteServiceDiscovery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(testSDSStackName).Return(nil),
	)

	err := delete(emptyContext(), mockCloudformation, testServiceName, testServiceName, testClusterName)
	assert.NoError(t, err, "Unexpected error calling delete")
}

func TestDeleteServiceDiscoveryDeleteNamespace(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().ValidateStackExists(testNamespaceStackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(testNamespaceStackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(testNamespaceStackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("create-sd", 0)
	flagSet.Bool(flags.DeletePrivateNamespaceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)

	err := delete(context, mockCloudformation, testServiceName, testServiceName, testClusterName)
	assert.NoError(t, err, "Unexpected error calling delete")
}

func TestDeleteServiceDiscoveryStackNotFoundErrorForSDS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(testSDSStackName).Return(fmt.Errorf("Stack not found")),
	)

	// If no stack is found, then there is nothing to delete, so no error is returned
	err := delete(emptyContext(), mockCloudformation, testServiceName, testServiceName, testClusterName)
	assert.NoError(t, err, "Expected error calling delete")
}

func TestDeleteServiceDiscoveryStackNotFoundErrorForNamespaceWithDeleteNamespaceFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().ValidateStackExists(testNamespaceStackName).Return(fmt.Errorf("Stack not found")),
	)

	flagSet := flag.NewFlagSet("create-sd", 0)
	flagSet.Bool(flags.DeletePrivateNamespaceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)

	err := delete(context, mockCloudformation, testServiceName, testServiceName, testClusterName)
	// Since the user requested us to delete their namespace, if we failed to delete it, then that's an error case
	assert.Error(t, err, "Expected error calling delete")
}

func TestCFNStackName(t *testing.T) {
	// underscore is allowed in cluster and service names, but not CFNStack names
	clusterName := "supercalifragilisticexpialidocious_________1234_"
	serviceName := "anotherreallylongstring_______________________________________hi______________________________________________________wassup__________________________________________123456789"

	sdsStackName := cfnStackName(serviceDiscoveryServiceStackNameFormat, clusterName, serviceName)
	namespaceStackName := cfnStackName(privateDNSNamespaceStackNameFormat, clusterName, serviceName)

	// underscore is allowed in cluster and service names, but not CFNStack names
	assert.False(t, strings.Contains(sdsStackName, "_"), "Underscores are not allowed in CFN Stack names")
	assert.False(t, strings.Contains(namespaceStackName, "_"), "Underscores are not allowed in CFN Stack names")
	// CFN Stacknames must be no longer than 128 characters
	assert.True(t, len(sdsStackName) <= 128, "CFN Stack names must be no longer than 128 characters")
	assert.True(t, len(namespaceStackName) <= 128, "CFN Stack names must be no longer than 128 characters")
}

// Tests the following weird/rare case which is technically allowed:
// SDS wasn't create by CLI, so no SDS stack exists, But
// Namespace was created by CLI, and user wants us to remove it.
func TestDeleteServiceDiscoveryStackNotFoundErrorForSDSWithDeleteNamespaceFlag(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
		mockCloudformation.EXPECT().ValidateStackExists(testSDSStackName).Return(fmt.Errorf("Stack not found")),
		mockCloudformation.EXPECT().ValidateStackExists(testNamespaceStackName).Return(nil),
		mockCloudformation.EXPECT().DeleteStack(testNamespaceStackName).Return(nil),
		mockCloudformation.EXPECT().WaitUntilDeleteComplete(testNamespaceStackName).Return(nil),
	)

	flagSet := flag.NewFlagSet("create-sd", 0)
	flagSet.Bool(flags.DeletePrivateNamespaceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)

	err := delete(context, mockCloudformation, testServiceName, testServiceName, testClusterName)
	assert.NoError(t, err, "Unexpected error calling delete")
}

func testCreateServiceDiscovery(t *testing.T, networkMode string, ecsParamsSD *utils.ServiceDiscovery, c *cli.Context, validateNamespace validateNamespaceParamsFunc, validateSDS validateSDSParamsFunc, createNamespace bool) (*ecs.ServiceRegistry, error) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	describeNamespaceStackResponse := &sdk.DescribeStacksOutput{
		Stacks: []*sdk.Stack{
			&sdk.Stack{
				Outputs: []*sdk.Output{
					&sdk.Output{
						OutputKey:   aws.String(cfnTemplateOutputPrivateNamespaceID),
						OutputValue: aws.String(testNamespaceID),
					},
				},
			},
		},
	}

	describeSDSStackResponse := &sdk.DescribeStacksOutput{
		Stacks: []*sdk.Stack{
			&sdk.Stack{
				Outputs: []*sdk.Output{
					&sdk.Output{
						OutputKey:   aws.String(cfnTemplateOutputSDSARN),
						OutputValue: aws.String(testSDSARN),
					},
				},
			},
		},
	}

	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)

	// Create a list of expected calls for gomock.InOrder
	// The expected calls differ depending upon whether or not we are going to create a namespace or not
	var expectedCFNCalls []*gomock.Call
	if createNamespace {
		expectedCFNCalls = append(expectedCFNCalls, []*gomock.Call{
			mockCloudformation.EXPECT().ValidateStackExists(testNamespaceStackName).Return(fmt.Errorf("Stack Not Found")),
			mockCloudformation.EXPECT().CreateStack(gomock.Any(), testNamespaceStackName, false, gomock.Any()).Do(func(x, y, w, z interface{}) {
				stackName := y.(string)
				capabilityIAM := w.(bool)
				cfnParams := z.(*cloudformation.CfnStackParams)
				validateNamespace(t, cfnParams)
				assert.False(t, capabilityIAM, "Expected capability capabilityIAM to be false")
				assert.Equal(t, testNamespaceStackName, stackName, "Expected stack name to match")
			}).Return("", nil),
			mockCloudformation.EXPECT().WaitUntilCreateComplete(testNamespaceStackName).Return(nil),
			mockCloudformation.EXPECT().DescribeStacks(testNamespaceStackName).Return(describeNamespaceStackResponse, nil),
		}...)
	}
	expectedCFNCalls = append(expectedCFNCalls, []*gomock.Call{
		mockCloudformation.EXPECT().ValidateStackExists(testSDSStackName).Return(fmt.Errorf("Stack Not Found")),
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), testSDSStackName, false, gomock.Any()).Do(func(x, y, w, z interface{}) {
			stackName := y.(string)
			capabilityIAM := w.(bool)
			cfnParams := z.(*cloudformation.CfnStackParams)
			validateSDS(t, cfnParams)
			assert.False(t, capabilityIAM, "Expected capability capabilityIAM to be false")
			assert.Equal(t, testSDSStackName, stackName, "Expected stack name to match")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().DescribeStacks(testSDSStackName).Return(describeSDSStackResponse, nil),
	}...)

	gomock.InOrder(expectedCFNCalls...)

	config := &config.CommandConfig{
		Cluster: testClusterName,
	}

	return create(c, networkMode, testServiceName, mockCloudformation, ecsParamsSD, config)

}

func emptyContext() *cli.Context {
	flagSet := flag.NewFlagSet("create-sd", 0)
	return cli.NewContext(nil, flagSet, nil)
}

func simpleWorkflowContext() *cli.Context {
	flagSet := flag.NewFlagSet("create-sd", 0)
	flagSet.String(flags.PrivateDNSNamespaceNameFlag, testNamespaceName, "")
	flagSet.String(flags.VpcIdFlag, testVPCID, "")

	return cli.NewContext(nil, flagSet, nil)
}

func validateCFNParam(expectedValue, paramKey string, cfnParams *cloudformation.CfnStackParams, t *testing.T) {
	observedValue, err := cfnParams.GetParameter(paramKey)
	assert.NoError(t, err, "Unexpected error getting cfn parameter")
	assert.Equal(t, expectedValue, aws.StringValue(observedValue.ParameterValue), fmt.Sprintf("Expected %s to be %s", paramKey, expectedValue))
}

func validateUsePreviousValueSet(paramKey string, cfnParams *cloudformation.CfnStackParams, t *testing.T) {
	observedValue, err := cfnParams.GetParameter(paramKey)
	assert.NoError(t, err, "Unexpected error getting cfn parameter")
	assert.True(t, aws.BoolValue(observedValue.UsePreviousValue), "Expected UsePreviousValue to be true")
}
