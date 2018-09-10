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
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	utils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	sdk "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	testClusterName        = "cluster"
	testServiceName        = "service"
	testDescription        = "clyde loves pudding"
	otherTestDescription   = "pudding plays hard to get"
	testNamespaceStackName = "amazon-ecs-cli-setup-cluster-service-private-dns-namespace"
	testSDSStackName       = "amazon-ecs-cli-setup-cluster-service-service-discovery-service"
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

	testCreateServiceDiscovery(t, "awsvpc", &utils.ServiceDiscovery{}, simpleWorkflowContext(), validateNamespace, validateSDS, true)
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

	testCreateServiceDiscovery(t, "bridge", input, simpleWorkflowContext(), validateNamespace, validateSDS, true)
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

	testCreateServiceDiscovery(t, "host", input, simpleWorkflowContext(), validateNamespace, validateSDS, true)
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

	testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, true)
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

	testCreateServiceDiscovery(t, "awsvpc", input, overrides, validateNamespace, validateSDS, true)
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

	testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
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

	testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
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

	testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
}

func TestCreateServiceDiscoveryWithECSParamsExistingPublicNamespaceByName(t *testing.T) {
	input := &utils.ServiceDiscovery{
		ContainerName: testContainerName,
		ContainerPort: aws.Int64(80),
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

	oldFindPublicNamespace := findPublicNamespace
	defer func() { findPublicNamespace = oldFindPublicNamespace }()
	findPublicNamespace = func(name string, config *config.CommandConfig) (*string, error) {
		// In this test the namespace does pre-exist and we search for it
		return aws.String(otherTestNamespaceID), nil
	}

	testCreateServiceDiscovery(t, "awsvpc", input, emptyContext(), validateNamespace, validateSDS, false)
}

func testCreateServiceDiscovery(t *testing.T, networkMode string, ecsParamsSD *utils.ServiceDiscovery, c *cli.Context, validateNamespace validateNamespaceParamsFunc, validateSDS validateSDSParamsFunc, createNamespace bool) {
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

	sdsARN, err := create(c, networkMode, testServiceName, mockCloudformation, ecsParamsSD, config)

	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(sdsARN), "Expected SDS ARN to match")

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
