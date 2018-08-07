// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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
	"fmt"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	privateDNSNamespaceStackNameFormat     = "amazon-ecs-cli-setup-%s-%s-private-dns-namespace"
	serviceDiscoveryServiceStackNameFormat = "amazon-ecs-cli-setup-%s-%s-service-discovery-service"
)

// CloudFormation template parameters
const (
	parameterKeyNamespaceDescription = "NamespaceDescription"
	parameterKeyVPCID                = "VPCID"
	parameterKeyNamespaceName        = "NamespaceName"
)

const (
	parameterKeySDSDescription                          = "SDSDescription"
	parameterKeySDSName                                 = "SDSName"
	parameterKeyNamespaceID                             = "NamespaceID"
	parameterKeyDNSType                                 = "DNSType"
	parameterKeyDNSTTL                                  = "DNSTTL"
	parameterKeyHealthCheckCustomConfigFailureThreshold = "FailureThreshold"
)

const (
	CFNTemplateOutputPrivateNamespaceID = "PrivateDNSNamespaceID"
	CFNTemplateOutputSDSARN             = "ServiceDiscoveryServiceARN"
)

const (
	dnsRecordTypeA   = "A"
	dnsRecordTypeSRV = "SRV"
)

var requiredParamsSDS = []string{parameterKeyNamespaceID, parameterKeySDSName, parameterKeyDNSType}
var requiredParamsNamespace = []string{parameterKeyVPCID, parameterKeyNamespaceName}

// Create creates resources for service discovery and returns the ID of the Service Discovery Service
func Create(c *cli.Context, networkMode, serviceName, clusterName string) (*string, error) {
	rdwr, err := config.NewReadWriter()
	if err != nil {
		return nil, err
	}

	commandConfig, err := config.NewCommandConfig(c, rdwr)
	if err != nil {
		return nil, err
	}

	cfnClient := cloudformation.NewCloudformationClient(commandConfig)

	return create(c, networkMode, serviceName, clusterName, cfnClient)
}

func create(c *cli.Context, networkMode, serviceName, clusterName string, cfnClient cloudformation.CloudformationClient) (*string, error) {
	// create namespace
	namespaceParams := namespaceCFNParams(c)
	if err := namespaceParams.Validate(); err != nil {
		return nil, err
	}

	namespaceStackName := cfnStackName(privateDNSNamespaceStackNameFormat, clusterName, serviceName)
	if _, err := cfnClient.CreateStack(cloudformation.GetPrivateNamespaceTemplate(), namespaceStackName, false, namespaceParams); err != nil {
		return nil, err
	}

	logrus.Info("Waiting for the private DNS namespace to be created...")
	// Wait for stack creation
	cfnClient.WaitUntilCreateComplete(namespaceStackName)

	// Get the ID of the namespace we just created
	namespaceID, err := getOutputIDFromStack(cfnClient, namespaceStackName, CFNTemplateOutputPrivateNamespaceID)
	if err != nil {
		return nil, err
	}

	// create SDS
	sdsParams := sdsCFNParams(aws.StringValue(namespaceID), serviceName, networkMode)
	if err := sdsParams.Validate(); err != nil {
		return nil, err
	}

	sdsStackName := cfnStackName(serviceDiscoveryServiceStackNameFormat, clusterName, serviceName)
	if _, err := cfnClient.CreateStack(cloudformation.GetSDSTemplate(), sdsStackName, false, sdsParams); err != nil {
		return nil, err
	}

	logrus.Info("Waiting for the Service Discovery Service to be created...")
	// Wait for stack creation
	cfnClient.WaitUntilCreateComplete(sdsStackName)

	// Return the ID of the SDS we just created
	return getOutputIDFromStack(cfnClient, sdsStackName, CFNTemplateOutputSDSARN)
}

func getOutputIDFromStack(cfnClient cloudformation.CloudformationClient, stackName, outputKey string) (*string, error) {
	response, err := cfnClient.DescribeStacks(stackName)
	if err != nil {
		return nil, err
	}
	if len(response.Stacks) == 0 {
		return nil, fmt.Errorf("Could not find CloudFormation stack: %s", stackName)
	}

	for _, output := range response.Stacks[0].Outputs {
		if aws.StringValue(output.OutputKey) == outputKey {
			return output.OutputValue, nil
		}
	}
	return nil, fmt.Errorf("Failed to find output %s in stack %s", outputKey, stackName)

}

func cfnStackName(stackName, cluster, service string) string {
	return fmt.Sprintf(stackName, cluster, service)
}

func sdsCFNParams(namespaceID, sdsName, networkMode string) *cloudformation.CfnStackParams {
	cfnParams := cloudformation.NewCfnStackParams(requiredParamsSDS)

	cfnParams.Add(parameterKeyNamespaceID, namespaceID)
	cfnParams.Add(parameterKeySDSName, sdsName)

	dnsType := dnsRecordTypeSRV
	if networkMode == ecs.NetworkModeAwsvpc {
		dnsType = dnsRecordTypeA
	}
	cfnParams.Add(parameterKeyDNSType, dnsType)

	return cfnParams
}

func namespaceCFNParams(context *cli.Context) *cloudformation.CfnStackParams {
	cfnParams := cloudformation.NewCfnStackParams(requiredParamsNamespace)

	namespaceName := context.String(flags.PrivateDNSNamespaceNameFlag)
	cfnParams.Add(parameterKeyNamespaceName, namespaceName)

	vpcID := context.String(flags.VpcIdFlag)
	cfnParams.Add(parameterKeyVPCID, vpcID)

	return cfnParams
}
