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
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	privateDNSNamespaceStackNameFormat     = "amazon-ecs-cli-setup-private-dns-namespace-%s-%s"
	serviceDiscoveryServiceStackNameFormat = "amazon-ecs-cli-setup-service-discovery-service-%s-%s"
)

const (
	cfnTemplateOutputPrivateNamespaceID = "PrivateDNSNamespaceID"
	cfnTemplateOutputSDSARN             = "ServiceDiscoveryServiceARN"
)

// CreateFunc is the interface/signature for Create
// This helps when writing code in other packages that need to mock Create (specifically it's a nicety that helps IDE features work)
type CreateFunc func(networkMode, serviceName string, c *context.ECSContext) (*ecs.ServiceRegistry, error)

// Create creates a DNS namespace (or uses an existing one) and creates a Service Discovery Service
// The Service Discovery ARN is returned so that it can be used to enable ECS Service Discovery
func Create(networkMode, serviceName string, c *context.ECSContext) (*ecs.ServiceRegistry, error) {
	cfnClient := cloudformation.NewCloudformationClient(c.CommandConfig)

	var ecsParamsSD *utils.ServiceDiscovery
	if c.ECSParams != nil {
		ecsParamsSD = &c.ECSParams.RunParams.ServiceDiscovery
	} else {
		ecsParamsSD = &utils.ServiceDiscovery{}
	}

	return create(c.CLIContext, networkMode, serviceName, cfnClient, ecsParamsSD, c.CommandConfig)
}

// UpdateFunc is the interface/signature for Create
// This helps when writing code in other packages that need to mock Update (specifically it's a nicety that helps IDE features work)
type UpdateFunc func(networkMode, serviceName string, c *context.ECSContext) error

// Update updates values for Service Discovery
// Only a few values on the SDS are available for update: DNS TTL and FailureThreshold
func Update(networkMode, serviceName string, c *context.ECSContext) error {
	cfnClient := cloudformation.NewCloudformationClient(c.CommandConfig)

	var ecsParamsSD *utils.ServiceDiscovery
	if c.ECSParams != nil {
		ecsParamsSD = &c.ECSParams.RunParams.ServiceDiscovery
	} else {
		ecsParamsSD = &utils.ServiceDiscovery{}
	}

	return update(c.CLIContext, networkMode, serviceName, c.CommandConfig.Cluster, cfnClient, ecsParamsSD)
}

// DeleteFunc is the interface/signature for Delete
// This helps when writing code in other packages that need to mock Create (specifically it's a nicety that helps IDE features work)
type DeleteFunc func(serviceName string, c *context.ECSContext) error

// Delete deletes resources for service discovery
func Delete(serviceName string, c *context.ECSContext) error {
	cfnClient := cloudformation.NewCloudformationClient(c.CommandConfig)

	return delete(c.CLIContext, cfnClient, serviceName, c.ProjectName, c.CommandConfig.Cluster)
}

func update(c *cli.Context, networkMode, serviceName, clusterName string, cfnClient cloudformation.CloudformationClient, ecsParamsSD *utils.ServiceDiscovery) error {
	warnOnFlagsNotValidForUpdate(c)

	sdsInput, err := mergeSDSFields(c, ecsParamsSD.ServiceDiscoveryService)
	if err != nil {
		return err
	}

	sdsStackName := cfnStackName(serviceDiscoveryServiceStackNameFormat, clusterName, serviceName)
	existingParameters, err := cfnClient.GetStackParameters(sdsStackName)
	if err != nil {
		return errors.Wrap(err, "CloudFormation stack not found for Service Discovery Service")
	}

	sdsParams, err := getSDSCFNParamsForUpdate(networkMode, sdsInput, existingParameters)
	if err != nil {
		return err
	}
	if err := sdsParams.Validate(); err != nil {
		return err
	}

	if _, err := cfnClient.UpdateStack(sdsStackName, sdsParams); err != nil {
		return err
	}

	logrus.Info("Waiting for your Service Discovery resources to be updated...")
	return cfnClient.WaitUntilUpdateComplete(sdsStackName)
}

func delete(c *cli.Context, cfnClient cloudformation.CloudformationClient, serviceName, projectName, clusterName string) error {
	sdsStackName := cfnStackName(serviceDiscoveryServiceStackNameFormat, clusterName, serviceName)
	err := deleteStack(sdsStackName, projectName, "Service Discovery Service", cfnClient, true, false)
	if err != nil {
		return err
	}

	if c.Bool(flags.DeletePrivateNamespaceFlag) {
		namspaceStackName := cfnStackName(privateDNSNamespaceStackNameFormat, clusterName, serviceName)
		err = deleteStack(namspaceStackName, projectName, "Private DNS Namespace", cfnClient, false, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteStack(stackName, projectName, resource string, cfnClient cloudformation.CloudformationClient, ignoreValidation, cleanUp bool) error {
	if err := cfnClient.ValidateStackExists(stackName); err != nil {
		if ignoreValidation {
			return nil
		}
		return errors.Wrapf(err, "no %s CloudFormation stack found for project '%s'", resource, projectName)
	}
	if cleanUp {
		logrus.Info("Cleaning up existing CloudFormation stack...")
	} else {
		logrus.Infof("Waiting for your %s resource to be deleted...", resource)
	}
	if err := cfnClient.DeleteStack(stackName); err != nil {
		return err
	}
	return cfnClient.WaitUntilDeleteComplete(stackName)
}

func create(c *cli.Context, networkMode, serviceName string, cfnClient cloudformation.CloudformationClient, ecsParamsSD *utils.ServiceDiscovery, config *config.CommandConfig) (*ecs.ServiceRegistry, error) {
	err := validateNameAndIdExclusive(c, ecsParamsSD)
	if err != nil {
		return nil, err
	}
	mergedInput, err := mergeSDFlagsAndInput(c, ecsParamsSD)
	if err != nil {
		return nil, err
	}
	err = validateMergedSDInputFields(mergedInput, networkMode)
	if err != nil {
		return nil, err
	}
	namespaceWarningsWhenIDSpecified(mergedInput)

	namespaceID, err := getOrCreateNamespace(c, networkMode, serviceName, cfnClient, mergedInput, config)
	if err != nil {
		return nil, err
	}

	// create SDS
	sdsParams := getSDSCFNParams(aws.StringValue(namespaceID), serviceName, networkMode, mergedInput)
	if err := sdsParams.Validate(); err != nil {
		return nil, err
	}

	sdsStackName := cfnStackName(serviceDiscoveryServiceStackNameFormat, config.Cluster, serviceName)

	// first try to delete the SDS Stack to clean up previous attempts that failed
	if err := deleteStack(sdsStackName, serviceName, "Service Discovery Service", cfnClient, true, true); err != nil {
		return nil, errors.Wrapf(err, "A Service Discovery Service CloudFormation stack for %s already exists, failed to delete existing stack", serviceName)
	}

	if _, err := cfnClient.CreateStack(cloudformation.GetSDSTemplate(), sdsStackName, false, sdsParams); err != nil {
		return nil, err
	}

	logrus.Info("Waiting for the Service Discovery Service to be created...")
	cfnClient.WaitUntilCreateComplete(sdsStackName)

	registryARN, err := getOutputIDFromStack(cfnClient, sdsStackName, cfnTemplateOutputSDSARN)
	var containerName *string
	if mergedInput.ContainerName != "" {
		containerName = aws.String(mergedInput.ContainerName)
	}
	serviceRegistry := &ecs.ServiceRegistry{
		RegistryArn:   registryARN,
		ContainerName: containerName,
		ContainerPort: mergedInput.ContainerPort,
	}
	return serviceRegistry, err
}

// createNamespace creates a private DNS namespace
// This function is used if getExistingNamespace() fails to find an existing namespace with the required settings
// If a CFN stack with the same name exists already, we therefore know that it doesn't contain a namespace (stack create failed),
// So we can safely delete the existing stack
func createNamespace(c *cli.Context, networkMode, serviceName, clusterName string, cfnClient cloudformation.CloudformationClient, input *utils.ServiceDiscovery) (*string, error) {
	namespaceParams := getNamespaceCFNParams(input)
	if err := namespaceParams.Validate(); err != nil {
		return nil, err
	}

	namespaceStackName := cfnStackName(privateDNSNamespaceStackNameFormat, clusterName, serviceName)

	if err := deleteStack(namespaceStackName, serviceName, "Private DNS Namespace", cfnClient, true, true); err != nil {
		return nil, errors.Wrapf(err, "A Private DNS Namespace CloudFormation stack for %s already exists, failed to delete existing stack: %s", serviceName, err)
	}

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
	var err error
	ecsParamsSD.ServiceDiscoveryService, err = mergeSDSFields(c, sds)
	if err != nil {
		return nil, err
	}

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

// Merges flags and ECS Params for the SDS
func mergeSDSFields(c *cli.Context, sds utils.ServiceDiscoveryService) (utils.ServiceDiscoveryService, error) {
	sds.DNSConfig.Type = resolveStringFieldOverride(c, flags.DNSTypeFlag, sds.DNSConfig.Type, "dns_config.type")
	ttl, err := resolveIntPointerFieldOverride(c, flags.DNSTTLFlag, sds.DNSConfig.TTL, "dns_config.ttl")
	if err != nil {
		return sds, err
	}
	sds.DNSConfig.TTL = ttl
	threshold, err := resolveIntPointerFieldOverride(c, flags.HealthcheckCustomConfigFailureThresholdFlag, sds.HealthCheckCustomConfig.FailureThreshold, "failure_threshold")
	if err != nil {
		return sds, err
	}
	sds.HealthCheckCustomConfig.FailureThreshold = threshold

	return sds, nil
}
