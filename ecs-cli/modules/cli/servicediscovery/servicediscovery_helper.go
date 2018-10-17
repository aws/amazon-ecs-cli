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
	"regexp"
	"strconv"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/route53"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	utils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	cfnsdk "github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
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
	cfnStackNameMaxLength = 128
)

var requiredParamsSDS = []string{parameterKeyNamespaceID, parameterKeySDSName, parameterKeyDNSType}
var requiredParamsNamespace = []string{parameterKeyVPCID, parameterKeyNamespaceName}

// Imported Route53 Utility functions that can be mocked in tests
// Adding the type signature allows code editor autocompletion and checking to work normally
var findPrivateNamespace route53.FindPrivateNamespaceFunc = route53.FindPrivateNamespace
var findPublicNamespace route53.FindPublicNamespaceFunc = route53.FindPublicNamespace

func resolveIntPointerFieldOverride(c *cli.Context, flagName string, ecsParamsVal *int64, field string) (*int64, error) {
	flagVal, err := getInt64FromCLIContext(c, flagName)
	if err != nil {
		return nil, err
	}

	if flagVal != nil && ecsParamsVal != nil {
		paramsVal := strconv.FormatInt(*ecsParamsVal, 10)
		override := strconv.FormatInt(*flagVal, 10)
		showFieldOverrideMsg(paramsVal, override, field)
	}
	if flagVal != nil {
		return flagVal, nil
	}
	return ecsParamsVal, nil
}

// getInt64FromCLIContext reads the flag from the cli context and typecasts into *int64
func getInt64FromCLIContext(c *cli.Context, flag string) (*int64, error) {
	value := c.String(flag)
	if value == "" {
		return nil, nil
	}
	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Please pass integer value for the flag %s", flag)
	}
	return aws.Int64(intValue), nil
}

func resolveStringFieldOverride(c *cli.Context, flagName, ecsParamsVal string, field string) string {
	flagVal := c.String(flagName)
	if flagVal != "" && ecsParamsVal != "" {
		showFieldOverrideMsg(ecsParamsVal, flagVal, field)
	}
	if flagVal != "" {
		return flagVal
	}
	return ecsParamsVal
}

func showFieldOverrideMsg(val string, override string, field string) {
	overrideMsg := "Using flag value as override (was %v but is now %v)"

	logrus.WithFields(logrus.Fields{
		"Service Discovery field": field,
	}).Infof(overrideMsg, val, override)
}

// Validates that SD input does not contain conflicting values
// validateMergedSDInputFields should be called after the inputs from flags and ECS Params have been merged
func validateMergedSDInputFields(input *utils.ServiceDiscovery, networkMode string) error {
	dnsType := getDNSType(input.ServiceDiscoveryService, networkMode, false)
	if dnsType == servicediscovery.RecordTypeSrv && input.ContainerName == "" {
		return fmt.Errorf("container_name is a required field when using SRV DNS records")
	}
	if dnsType == servicediscovery.RecordTypeSrv && input.ContainerPort == nil {
		return fmt.Errorf("container_port is a required field when using SRV DNS records")
	}

	hasPublic := hasNamespace(input.PublicDNSNamespace.Namespace)
	hasPrivate := hasNamespace(input.PrivateDNSNamespace.Namespace)

	if hasPublic && hasPrivate {
		return fmt.Errorf("Both a public and private namespace can not be used with Service Discovery; please use only 1 namespace")
	}

	if !hasPublic && !hasPrivate {
		return fmt.Errorf("To use Service Discovery, please specify a DNS namespace")
	}

	if input.PrivateDNSNamespace.Name != "" && input.PrivateDNSNamespace.ID == "" {
		if input.PrivateDNSNamespace.VPC == "" {
			return fmt.Errorf("VPC is required when specifying private namespace by name")
		}
	}

	return nil
}

func resolveNamespaceOverride(namespaceNameFromFlag, namespaceIDFromFlag, namespaceType string, ecsParamsNamespace utils.Namespace) utils.Namespace {
	flagNamespace := utils.Namespace{
		ID:   namespaceIDFromFlag,
		Name: namespaceNameFromFlag,
	}
	if hasNamespace(flagNamespace) && hasNamespace(ecsParamsNamespace) {
		showFieldOverrideMsg(getNamespace(ecsParamsNamespace), getNamespace(ecsParamsNamespace), fmt.Sprintf("dns %s namespace", namespaceType))
	}
	if hasNamespace(flagNamespace) {
		return flagNamespace
	}
	return ecsParamsNamespace
}

// Validates that namespace name and ID fields were not both specified
// This function runs before flags and ECS Params are merged
func validateNameAndIdExclusive(c *cli.Context, ecsParamsSD *utils.ServiceDiscovery) error {
	if c.String(flags.PrivateDNSNamespaceIDFlag) != "" && c.String(flags.PrivateDNSNamespaceNameFlag) != "" {
		return fmt.Errorf("Validation Error: %s and %s both specified", flags.PrivateDNSNamespaceIDFlag, flags.PrivateDNSNamespaceNameFlag)
	}

	if c.String(flags.PublicDNSNamespaceIDFlag) != "" && c.String(flags.PublicDNSNamespaceNameFlag) != "" {
		return fmt.Errorf("Validation Error: %s and %s both specified", flags.PublicDNSNamespaceIDFlag, flags.PublicDNSNamespaceNameFlag)
	}

	if ecsParamsSD.PrivateDNSNamespace.Name != "" && ecsParamsSD.PrivateDNSNamespace.ID != "" {
		return fmt.Errorf("Validation Error: private_dns_namespace.name and private_dns_namespace.id both specified")
	}

	if ecsParamsSD.PublicDNSNamespace.Name != "" && ecsParamsSD.PublicDNSNamespace.ID != "" {
		return fmt.Errorf("Validation Error: public_dns_namespace.name and public_dns_namespace.id both specified")
	}

	return nil
}

// Logs warnings for fields which are ignored when a namespace ID is specified
func namespaceWarningsWhenIDSpecified(input *utils.ServiceDiscovery) {
	msg := "Ignoring %s because the ID of an existing private namespace was specified"

	privNamespace := input.PrivateDNSNamespace
	if privNamespace.ID != "" {
		if privNamespace.Description != "" {
			logrus.Warnf(msg, "description")
		}
		if privNamespace.VPC != "" {
			logrus.Warnf(msg, "vpc")
		}
	}
}

func hasNamespace(n utils.Namespace) bool {
	return n.ID != "" || n.Name != ""
}

func getNamespace(n utils.Namespace) string {
	if n.ID != "" {
		return n.ID
	}
	return n.Name
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

func cfnStackName(stackNameFmt, cluster, service string) string {
	maxLength := (cfnStackNameMaxLength - len(stackNameFmt)) / 2
	name := fmt.Sprintf(stackNameFmt, truncate(cluster, maxLength), truncate(service, maxLength))
	return sanitize(name)
}

// Makes the given string a valid CFN stack name
// by replacing all characters that are not alphanumeric or hyphen with 0
// and truncating at 128 characters
func sanitize(s string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		// the regex compiles, the unit tests verify this, there's no need to return this error
		logrus.Fatal(err)
	}
	return reg.ReplaceAllString(s, "0")
}

func truncate(s string, length int) string {
	if len(s) > length {
		return s[:length]
	}
	return s
}

func getSDSCFNParams(namespaceID, ecsServiceName, networkMode string, input *utils.ServiceDiscovery) *cloudformation.CfnStackParams {
	cfnParams := cloudformation.NewCfnStackParams(requiredParamsSDS)

	sdsName := input.ServiceDiscoveryService.Name
	if sdsName == "" {
		sdsName = ecsServiceName
	}

	cfnParams.Add(parameterKeyNamespaceID, namespaceID)
	cfnParams.Add(parameterKeySDSName, sdsName)

	if description := input.ServiceDiscoveryService.Description; description != "" {
		cfnParams.Add(parameterKeySDSDescription, description)
	}

	dnsType := getDNSType(input.ServiceDiscoveryService, networkMode, true)
	cfnParams.Add(parameterKeyDNSType, dnsType)

	populateSDSCFNParamsForCreateOrUpdate(cfnParams, networkMode, input.ServiceDiscoveryService)

	return cfnParams
}

func populateSDSCFNParamsForCreateOrUpdate(cfnParams *cloudformation.CfnStackParams, networkMode string, sds utils.ServiceDiscoveryService) {
	if dnsTTL := sds.DNSConfig.TTL; dnsTTL != nil {
		cfnParams.Add(parameterKeyDNSTTL, strconv.Itoa(int(*dnsTTL)))
	}

	if threshold := sds.HealthCheckCustomConfig.FailureThreshold; threshold != nil {
		cfnParams.Add(parameterKeyHealthCheckCustomConfigFailureThreshold, strconv.Itoa(int(*threshold)))
	}
}

func getSDSCFNParamsForUpdate(networkMode string, sds utils.ServiceDiscoveryService, existingParams []*cfnsdk.Parameter) (*cloudformation.CfnStackParams, error) {
	cfnParams, err := cloudformation.NewCfnStackParamsForUpdate(requiredParamsSDS, existingParams)
	if err != nil {
		return nil, err
	}

	populateSDSCFNParamsForCreateOrUpdate(cfnParams, networkMode, sds)

	return cfnParams, nil
}

func getDNSType(sds utils.ServiceDiscoveryService, networkMode string, warn bool) string {
	dnsType := sds.DNSConfig.Type
	if dnsType == "" {
		// set default
		if networkMode == ecs.NetworkModeAwsvpc {
			dnsType = servicediscovery.RecordTypeA
			if warn {
				logrus.Warnf("Defaulting DNS Type to %s because network mode was %s", servicediscovery.RecordTypeA, ecs.NetworkModeAwsvpc)
			}
		} else {
			dnsType = servicediscovery.RecordTypeSrv
			if warn {
				logrus.Warnf("Defaulting DNS Type to %s because network mode was %s", servicediscovery.RecordTypeSrv, networkMode)
			}
		}
	}
	return dnsType
}

func getNamespaceCFNParams(input *utils.ServiceDiscovery) *cloudformation.CfnStackParams {
	cfnParams := cloudformation.NewCfnStackParams(requiredParamsNamespace)

	privNamespace := input.PrivateDNSNamespace
	cfnParams.Add(parameterKeyNamespaceName, privNamespace.Name)

	cfnParams.Add(parameterKeyVPCID, privNamespace.VPC)

	if privNamespace.Description != "" {
		cfnParams.Add(parameterKeyNamespaceDescription, privNamespace.Description)
	}

	return cfnParams
}

func warnOnFlagsNotValidForUpdate(context *cli.Context) {
	flagsOnlyValidOnCreate := []string{
		flags.PrivateDNSNamespaceNameFlag,
		flags.VpcIdFlag,
		flags.PrivateDNSNamespaceIDFlag,
		flags.PublicDNSNamespaceIDFlag,
		flags.PublicDNSNamespaceNameFlag,
		flags.ServiceDiscoveryContainerNameFlag,
		flags.ServiceDiscoveryContainerPortFlag,
		flags.DNSTypeFlag,
	}

	for _, flag := range flagsOnlyValidOnCreate {
		if context.String(flag) != "" {
			logrus.Warnf("--%s is not valid when updating Service Discovery, its value can only be set during Service creation", flag)
		}
	}

}
