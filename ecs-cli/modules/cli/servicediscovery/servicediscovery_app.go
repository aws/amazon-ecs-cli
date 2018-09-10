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

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients/aws/cloudformation"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/commands/flags"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	utils "github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	privateDNSNamespaceStackNameFormat     = "amazon-ecs-cli-setup-%s-%s-private-dns-namespace"
	serviceDiscoveryServiceStackNameFormat = "amazon-ecs-cli-setup-%s-%s-service-discovery-service"
)

const (
	cfnTemplateOutputPrivateNamespaceID = "PrivateDNSNamespaceID"
	cfnTemplateOutputSDSARN             = "ServiceDiscoveryServiceARN"
)

// Create creates a DNS namespace (or uses an existing one) and creates a Service Discovery Service
// The Service Discovery ARN is returned so that it can be used to enable ECS Service Discovery
func Create(networkMode, serviceName string, c *context.ECSContext) (*string, error) {
	cfnClient := cloudformation.NewCloudformationClient(c.CommandConfig)

	var ecsParamsSD *utils.ServiceDiscovery
	if c.ECSParams != nil {
		ecsParamsSD = &c.ECSParams.RunParams.ServiceDiscovery
	} else {
		ecsParamsSD = &utils.ServiceDiscovery{}
	}

	return create(c.CLIContext, networkMode, serviceName, cfnClient, ecsParamsSD, c.CommandConfig)
}

func create(c *cli.Context, networkMode, serviceName string, cfnClient cloudformation.CloudformationClient, ecsParamsSD *utils.ServiceDiscovery, config *config.CommandConfig) (*string, error) {
	err := validateNameAndIdExclusive(c, ecsParamsSD)
	if err != nil {
		return nil, err
	}
	input, err := mergeSDFlagsAndInput(c, ecsParamsSD)
	if err != nil {
		return nil, err
	}
	err = validateMergedSDInputFields(input, networkMode)
	if err != nil {
		return nil, err
	}
	namespaceWarningsWhenIDSpecified(input)

	namespaceID, err := getOrCreateNamespace(c, networkMode, serviceName, cfnClient, input, config)
	if err != nil {
		return nil, err
	}

	// create SDS
	sdsParams := getSDSCFNParams(aws.StringValue(namespaceID), serviceName, networkMode, input)
	if err := sdsParams.Validate(); err != nil {
		return nil, err
	}

	sdsStackName := cfnStackName(serviceDiscoveryServiceStackNameFormat, config.Cluster, serviceName)
	if _, err := cfnClient.CreateStack(cloudformation.GetSDSTemplate(), sdsStackName, false, sdsParams); err != nil {
		return nil, err
	}

	logrus.Info("Waiting for the Service Discovery Service to be created...")
	cfnClient.WaitUntilCreateComplete(sdsStackName)

	// Return the ID of the SDS we just created
	return getOutputIDFromStack(cfnClient, sdsStackName, cfnTemplateOutputSDSARN)
}

func createNamespace(c *cli.Context, networkMode, serviceName, clusterName string, cfnClient cloudformation.CloudformationClient, input *utils.ServiceDiscovery) (*string, error) {
	namespaceParams := getNamespaceCFNParams(input)
	if err := namespaceParams.Validate(); err != nil {
		return nil, err
	}

	namespaceStackName := cfnStackName(privateDNSNamespaceStackNameFormat, clusterName, serviceName)
	if _, err := cfnClient.CreateStack(cloudformation.GetPrivateNamespaceTemplate(), namespaceStackName, false, namespaceParams); err != nil {
		return nil, err
	}

	logrus.Info("Waiting for the private DNS namespace to be created...")
	cfnClient.WaitUntilCreateComplete(namespaceStackName)

	// Get the ID of the namespace we just created
	return getOutputIDFromStack(cfnClient, namespaceStackName, cfnTemplateOutputPrivateNamespaceID)
}

func getOrCreateNamespace(c *cli.Context, networkMode, serviceName string, cfnClient cloudformation.CloudformationClient, input *utils.ServiceDiscovery, config *config.CommandConfig) (*string, error) {
	namespace, err := getExistingNamespace(input, config)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		namespace, err = createNamespace(c, networkMode, serviceName, config.Cluster, cfnClient, input)
	} else {
		logrus.Infof("Using existing namespace %s", *namespace)
	}
	return namespace, err
}

func getExistingNamespace(input *utils.ServiceDiscovery, config *config.CommandConfig) (*string, error) {
	switch {
	case input.PrivateDNSNamespace.ID != "":
		return aws.String(input.PrivateDNSNamespace.ID), nil
	case input.PublicDNSNamespace.ID != "":
		return aws.String(input.PublicDNSNamespace.ID), nil
	case input.PrivateDNSNamespace.Name != "":
		return findPrivateNamespace(input.PrivateDNSNamespace.Name, input.PrivateDNSNamespace.VPC, config)
	case input.PublicDNSNamespace.Name != "":
		return getPublicNamespaceSpecifiedByName(input.PublicDNSNamespace.Name, config)
	default:
		return nil, nil
	}
}

func getPublicNamespaceSpecifiedByName(name string, config *config.CommandConfig) (*string, error) {
	namespace, err := findPublicNamespace(name, config)
	if err != nil {
		return nil, err
	}
	if namespace == nil {
		// we do not create public namespaces, so failing to find it is in an error case
		return nil, fmt.Errorf("Failed to find public namespace %s", name)
	}
	return namespace, err
}

// Flags override the values specified in ECS Params
// Merges fields for Namespace and SDS
// This function just merges fields; it doesn't validate them
func mergeSDFlagsAndInput(c *cli.Context, ecsParamsSD *utils.ServiceDiscovery) (*utils.ServiceDiscovery, error) {
	// Private DNS Namespace fields
	privNamespace := ecsParamsSD.PrivateDNSNamespace
	privNamespace.Namespace = resolveNamespaceOverride(c.String(flags.PrivateDNSNamespaceNameFlag), c.String(flags.PrivateDNSNamespaceIDFlag), "private", privNamespace.Namespace)
	privNamespace.VPC = resolveStringFieldOverride(c, flags.VpcIdFlag, privNamespace.VPC, "private_dns_namespace.vpc")
	ecsParamsSD.PrivateDNSNamespace = privNamespace

	// Public DNS Namespace fields
	pubNamespace := ecsParamsSD.PublicDNSNamespace
	pubNamespace.Namespace = resolveNamespaceOverride(c.String(flags.PublicDNSNamespaceNameFlag), c.String(flags.PublicDNSNamespaceIDFlag), "public", pubNamespace.Namespace)
	ecsParamsSD.PublicDNSNamespace = pubNamespace

	// SDS fields
	sds := ecsParamsSD.ServiceDiscoveryService
	sds.DNSConfig.Type = resolveStringFieldOverride(c, flags.DNSTypeFlag, sds.DNSConfig.Type, "dns_config.type")
	ttl, err := resolveIntPointerFieldOverride(c, flags.DNSTTLFlag, sds.DNSConfig.TTL, "dns_config.ttl")
	if err != nil {
		return nil, err
	}
	sds.DNSConfig.TTL = ttl
	threshold, err := resolveIntPointerFieldOverride(c, flags.HealthcheckCustomConfigFailureThresholdFlag, sds.HealthCheckCustomConfig.FailureThreshold, "failure_threshold")
	if err != nil {
		return nil, err
	}
	sds.HealthCheckCustomConfig.FailureThreshold = threshold
	ecsParamsSD.ServiceDiscoveryService = sds

	// top level fields
	// these container fields aren't used when creating Route53 resources in this package, but we aggregate them here so that we can do error checking- SRV records require these.
	ecsParamsSD.ContainerName = resolveStringFieldOverride(c, flags.ServiceDiscoveryContainerNameFlag, ecsParamsSD.ContainerName, "container_name")
	port, err := resolveIntPointerFieldOverride(c, flags.ServiceDiscoveryContainerPortFlag, ecsParamsSD.ContainerPort, "container_port")
	if err != nil {
		return nil, err
	}
	ecsParamsSD.ContainerPort = port

	return ecsParamsSD, nil
}
