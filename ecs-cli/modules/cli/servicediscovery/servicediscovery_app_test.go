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
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation/mock"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/aws-sdk-go/aws"
	sdk "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
)

const (
	testClusterName        = "cluster"
	testServiceName        = "service"
	testNamespaceStackName = "amazon-ecs-cli-setup-cluster-service-private-dns-namespace"
	testSDSStackName       = "amazon-ecs-cli-setup-cluster-service-service-discovery-service"
	testNamespaceName      = "corp"
	testVPCID              = "vpc-8BAADF00D"
	testNamespaceID        = "ns-CA15CA15CA15CA15"
	testSDSARN             = "arn:aws:servicediscovery:eu-west-1:11111111111:service/srv-clydelovespudding"
)

func TestCreateServiceDiscoveryAWSVPC(t *testing.T) {
	testCreateServiceDiscovery(t, "awsvpc", dnsRecordTypeA)
}

func TestCreateServiceDiscoveryBridge(t *testing.T) {
	testCreateServiceDiscovery(t, "bridge", dnsRecordTypeSRV)
}

func TestCreateServiceDiscoveryHost(t *testing.T) {
	testCreateServiceDiscovery(t, "host", dnsRecordTypeSRV)
}

func testCreateServiceDiscovery(t *testing.T, networkMode, expectedDNSType string) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	describeNamespaceStackResponse := &sdk.DescribeStacksOutput{
		Stacks: []*sdk.Stack{
			&sdk.Stack{
				Outputs: []*sdk.Output{
					&sdk.Output{
						OutputKey:   aws.String(CFNTemplateOutputPrivateNamespaceID),
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
						OutputKey:   aws.String(CFNTemplateOutputSDSARN),
						OutputValue: aws.String(testSDSARN),
					},
				},
			},
		},
	}

	mockCloudformation := mock_cloudformation.NewMockCloudformationClient(ctrl)
	gomock.InOrder(
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
		mockCloudformation.EXPECT().CreateStack(gomock.Any(), testSDSStackName, false, gomock.Any()).Do(func(x, y, w, z interface{}) {
			stackName := y.(string)
			capabilityIAM := w.(bool)
			cfnParams := z.(*cloudformation.CfnStackParams)
			validateCFNParam(testNamespaceID, parameterKeyNamespaceID, cfnParams, t)
			validateCFNParam(testServiceName, parameterKeySDSName, cfnParams, t)
			validateCFNParam(expectedDNSType, parameterKeyDNSType, cfnParams, t)
			assert.False(t, capabilityIAM, "Expected capability capabilityIAM to be false")
			assert.Equal(t, testSDSStackName, stackName, "Expected stack name to match")
		}).Return("", nil),
		mockCloudformation.EXPECT().WaitUntilCreateComplete(testSDSStackName).Return(nil),
		mockCloudformation.EXPECT().DescribeStacks(testSDSStackName).Return(describeSDSStackResponse, nil),
	)

	flagSet := flag.NewFlagSet("create-sd", 0)
	flagSet.String(flags.PrivateDNSNamespaceNameFlag, testNamespaceName, "")
	flagSet.String(flags.VpcIdFlag, testVPCID, "")
	flagSet.Bool(flags.ForceFlag, true, "")

	context := cli.NewContext(nil, flagSet, nil)

	sdsARN, err := create(context, networkMode, testServiceName, testClusterName, mockCloudformation)

	assert.NoError(t, err, "Unexpected Error calling create")
	assert.Equal(t, testSDSARN, aws.StringValue(sdsARN), "Expected SDS ARN to match")

}

func validateCFNParam(expectedValue, paramKey string, cfnParams *cloudformation.CfnStackParams, t *testing.T) {
	observedValue, err := cfnParams.GetParameter(paramKey)
	assert.NoError(t, err, "Unexpected error getting cfn parameter")
	assert.Equal(t, expectedValue, aws.StringValue(observedValue.ParameterValue))
}
