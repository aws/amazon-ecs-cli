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

package utils

// ECS Params Reader is used to parse the ecs-params.yml file and marshal the data into the ECSParams struct

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	libYaml "github.com/docker/libcompose/yaml"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

///////////////////////////////////
///// ECS Params Schema types /////
///////////////////////////////////

// ECSParams contains the information parsed from the ecs-params.yml file
type ECSParams struct {
	Version        string
	TaskDefinition EcsTaskDef `yaml:"task_definition"`
	RunParams      RunParams  `yaml:"run_params"`
}

// EcsTaskDef corresponds to fields in an ECS TaskDefinition
type EcsTaskDef struct {
	NetworkMode          string         `yaml:"ecs_network_mode"`
	TaskRoleArn          string         `yaml:"task_role_arn"`
	PIDMode              string         `yaml:"pid_mode"`
	IPCMode              string         `yaml:"ipc_mode"`
	ContainerDefinitions ContainerDefs  `yaml:"services"`
	ExecutionRole        string         `yaml:"task_execution_role"`
	TaskSize             TaskSize       `yaml:"task_size"` // Needed to run FARGATE tasks
	DockerVolumes        []DockerVolume `yaml:"docker_volumes"`
	EFSVolumes           []EFSVolume    `yaml:"efs_volumes"`
	PlacementConstraints []Constraint   `yaml:"placement_constraints"`
}

// ContainerDefs is a map of ContainerDefs within a task definition
type ContainerDefs map[string]ContainerDef

// ContainerDef holds fields for an ECS Container Definition that are not supplied by docker-compose
type ContainerDef struct {
	Essential             bool                  `yaml:"essential"`
	InitProcessEnabled    bool                  `yaml:"init_process_enabled"`
	RepositoryCredentials RepositoryCredentials `yaml:"repository_credentials"`
	// resource field yaml names correspond to equivalent docker-compose field
	Cpu                   int64                  `yaml:"cpu_shares"`
	Memory                libYaml.MemStringorInt `yaml:"mem_limit"`
	MemoryReservation     libYaml.MemStringorInt `yaml:"mem_reservation"`
	HealthCheck           *HealthCheck           `yaml:"healthcheck"`
	Logging               Logging                `yaml:"logging"`
	FirelensConfiguration FirelensConfiguration  `yaml:"firelens_configuration"`
	Secrets               []Secret               `yaml:"secrets"`
	GPU                   string                 `yaml:"gpu"`
}

type Volume struct {
	DockerVolumeConfig DockerVolume
	EFSVolumeConfig    EFSVolume
}

type DockerVolume struct {
	Name          string            `yaml:"name"`
	Scope         *string           `yaml:"scope"`
	Autoprovision *bool             `yaml:"autoprovision"`
	Driver        *string           `yaml:"driver"`
	DriverOptions map[string]string `yaml:"driver_opts"`
	Labels        map[string]string `yaml:"labels"`
}

type EFSVolume struct {
	Name                  string  `yaml:"name"`
	FileSystemID          *string `yaml:"filesystem_id"` // Required
	RootDirectory         *string `yaml:"root_directory"`
	TransitEncryption     *string `yaml:"transit_encryption"` // Optional. default: DISABLED. options: ENABLED or DISABLED
	TransitEncryptionPort *int64  `yaml:"transit_encryption_port"`
	AccessPointID         *string `yaml:"access_point"`
	IAM                   *string `yaml:"iam"` // default: DISABLED. options: ENABLED or DISABLED
}

// Firelens holds all possible fields for logging via Firelens
// https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_FirelensConfiguration.html
type FirelensConfiguration struct {
	Type    string            `yaml:"type"`
	Options map[string]string `yaml:"options"`
}

// HealthCheck holds all possible fields for HealthCheck, including fields
// supported by docker compose vs ECS
type HealthCheck struct {
	Test        libYaml.Stringorslice
	Command     libYaml.Stringorslice
	Timeout     string `yaml:"timeout,omitempty"`
	Interval    string `yaml:"interval,omitempty"`
	Retries     int64  `yaml:"retries,omitempty"`
	StartPeriod string `yaml:"start_period,omitempty"`
}

// RepositoryCredentials holds CredentialParameters for a ContainerDef
type RepositoryCredentials struct {
	CredentialsParameter string `yaml:"credentials_parameter"`
}

// Logging holds a list of Secrets within SecretOptions. They are essentially
// identical to secrets injected into the container environment by ECS, but are
// instead injected into the Docker logging configuration.
type Logging struct {
	SecretOptions []Secret `yaml:"secret_options"`
}

// Secret supports the ECS Secrets integration with SSM Parameter Store
type Secret struct {
	ValueFrom string `yaml:"value_from"`
	Name      string `yaml:"name"`
}

// TaskSize holds Cpu and Memory values needed for Fargate tasks
// https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-cpu-memory-error.html
type TaskSize struct {
	Cpu    string `yaml:"cpu_limit"`
	Memory string `yaml:"mem_limit"`
}

// RunParams specifies non-TaskDefinition specific parameters
type RunParams struct {
	NetworkConfiguration NetworkConfiguration `yaml:"network_configuration"`
	TaskPlacement        TaskPlacement        `yaml:"task_placement"`
	ServiceDiscovery     ServiceDiscovery     `yaml:"service_discovery"`
}

// NetworkConfiguration specifies the network config for the task definition.
// Supports values 'awsvpc' (required for Fargate), 'bridge', 'host' or 'none'
type NetworkConfiguration struct {
	AwsVpcConfiguration AwsVpcConfiguration `yaml:"awsvpc_configuration"`
}

// AwsVpcConfiguration specifies the networking resources available to
// tasks running in 'awsvpc' networking mode
type AwsVpcConfiguration struct {
	Subnets        []string       `yaml:"subnets"`
	SecurityGroups []string       `yaml:"security_groups"`
	AssignPublicIp AssignPublicIp `yaml:"assign_public_ip"` // Needed to run FARGATE tasks
}

// TODO: Remove; use enum in aws-sdk-go instead (AssignPublicIpEnabled, AssignPublicIpDisabled)
type AssignPublicIp string

// ServiceDiscovery holds information related to ECS/Route53 Service Discovery
type ServiceDiscovery struct {
	ContainerName           string                  `yaml:"container_name"`
	ContainerPort           *int64                  `yaml:"container_port"`
	PrivateDNSNamespace     PrivateDNSNamespace     `yaml:"private_dns_namespace"`
	PublicDNSNamespace      PublicDNSNamespace      `yaml:"public_dns_namespace"`
	ServiceDiscoveryService ServiceDiscoveryService `yaml:"service_discovery_service"`
}

// Namespace holds the basic information for any type of namespace
type Namespace struct {
	ID   string `yaml:"id"`
	Name string `yaml:"name"`
}

// PrivateDNSNamespace holds information related to Route53 private DNS namespaces
type PrivateDNSNamespace struct {
	Namespace   `yaml:",inline"`
	VPC         string `yaml:"vpc"`
	Description string `yaml:"description"`
}

// PublicDNSNamespace holds information related to Route53 public DNS namespaces
type PublicDNSNamespace struct {
	Namespace `yaml:",inline"`
}

// ServiceDiscoveryService holds information related to Route53 Service Discovery Services
type ServiceDiscoveryService struct {
	Name                    string                  `yaml:"name"`
	Description             string                  `yaml:"description"`
	DNSConfig               DNSConfig               `yaml:"dns_config"`
	HealthCheckCustomConfig HealthCheckCustomConfig `yaml:"healthcheck_custom_config"`
}

// DNSConfig holds the dns configuration for Service Discovery Services
type DNSConfig struct {
	Type string `yaml:"type"`
	TTL  *int64 `yaml:"ttl"`
}

// HealthCheckCustomConfig
type HealthCheckCustomConfig struct {
	FailureThreshold *int64 `yaml:"failure_threshold"`
}

const (
	Enabled  AssignPublicIp = "ENABLED"
	Disabled AssignPublicIp = "DISABLED"
)

type TaskPlacement struct {
	Strategies  []Strategy   `yaml:"strategy"`
	Constraints []Constraint `yaml:"constraints"`
}

type Strategy struct {
	Field string `yaml:"field"`
	Type  string `yaml:"type"`
}

type Constraint struct {
	Expression string `yaml:"expression"`
	Type       string `yaml:"type"`
}

/////////////////////////////
///// Parsing Functions /////
/////////////////////////////

// Having a custom Unmarshaller on ContainerDef allows us to set custom defaults on the Container Definition
func (cd *ContainerDef) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type rawContainerDef ContainerDef
	raw := rawContainerDef{Essential: true} //  If essential is not specified, we want it to be true
	if err := unmarshal(&raw); err != nil {
		return err
	}

	*cd = ContainerDef(raw)
	return nil
}

// ReadECSParams parses the ecs-params.yml file and puts it into an ECSParams struct.
func ReadECSParams(filename string) (*ECSParams, error) {
	if filename == "" {
		defaultFilename := "ecs-params.yml"
		if _, err := os.Stat(defaultFilename); err == nil {
			filename = defaultFilename
		} else {
			return nil, nil
		}
	}

	// NOTE: Readfile reads all data into memory and closes file. Could
	// eventually refactor this to read different sections separately.
	ecsParamsData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading file '%v'", filename)
	}
	ecsParamsData = []byte(os.ExpandEnv(string(ecsParamsData)))
	ecsParams := &ECSParams{}

	if err = yaml.Unmarshal([]byte(ecsParamsData), &ecsParams); err != nil {
		return nil, errors.Wrapf(err, "Error unmarshalling yaml data from ECS params file: %v", filename)
	}

	return ecsParams, nil
}

/////////////////////
//// Converters ////
////////////////////

// ConvertToECSNetworkConfiguration extracts out the NetworkConfiguration from
// the ECSParams into a format that is compatible with ECSClient calls.
func ConvertToECSNetworkConfiguration(ecsParams *ECSParams) (*ecs.NetworkConfiguration, error) {
	if ecsParams == nil {
		return nil, nil
	}

	networkMode := ecsParams.TaskDefinition.NetworkMode

	if networkMode != "awsvpc" {
		return nil, nil
	}

	awsvpcConfig := ecsParams.RunParams.NetworkConfiguration.AwsVpcConfiguration

	subnets := awsvpcConfig.Subnets

	if len(subnets) < 1 {
		return nil, errors.New("at least one subnet is required in the network configuration")
	}

	securityGroups := awsvpcConfig.SecurityGroups
	assignPublicIp := string(awsvpcConfig.AssignPublicIp)

	ecsSubnets := make([]*string, len(subnets))
	for i, subnet := range subnets {
		ecsSubnets[i] = aws.String(subnet)
	}

	ecsSecurityGroups := make([]*string, len(securityGroups))
	for i, sg := range securityGroups {
		ecsSecurityGroups[i] = aws.String(sg)
	}

	ecsAwsVpcConfig := &ecs.AwsVpcConfiguration{
		Subnets:        ecsSubnets,
		SecurityGroups: ecsSecurityGroups,
	}

	// For tasks launched with network config in EC2 mode, assign_pubic_ip field is not accepted
	if assignPublicIp != "" {
		ecsAwsVpcConfig.AssignPublicIp = aws.String(assignPublicIp)
	}

	ecsNetworkConfig := &ecs.NetworkConfiguration{
		AwsvpcConfiguration: ecsAwsVpcConfig,
	}

	return ecsNetworkConfig, nil
}

// ConvertToECSHealthCheck extracts out the HealthCheck from the ECSParams into
// a format that is compatible with ECSClient calls.
func (h *HealthCheck) ConvertToECSHealthCheck() (*ecs.HealthCheck, error) {
	ecsHealthCheck := &ecs.HealthCheck{}
	if len(h.Command) > 0 && len(h.Test) > 0 {
		return nil, fmt.Errorf("healthcheck.test and healthcheck.command can not both be specified")
	}

	if len(h.Command) > 0 {
		ecsHealthCheck.Command = aws.StringSlice(getHealthCheckCommand(h.Command))
	}

	if len(h.Test) > 0 {
		ecsHealthCheck.Command = aws.StringSlice(getHealthCheckCommand(h.Test))
	}

	if h.Retries != 0 {
		ecsHealthCheck.Retries = &h.Retries
	}

	timeout, err := parseHealthCheckTime(h.Timeout)
	if err != nil {
		return ecsHealthCheck, err
	}
	ecsHealthCheck.Timeout = timeout

	startPeriod, err := parseHealthCheckTime(h.StartPeriod)
	if err != nil {
		return ecsHealthCheck, err
	}
	ecsHealthCheck.StartPeriod = startPeriod

	interval, err := parseHealthCheckTime(h.Interval)
	if err != nil {
		return ecsHealthCheck, err
	}
	ecsHealthCheck.Interval = interval

	return ecsHealthCheck, nil
}

// parses the command/test field for healthcheck
func getHealthCheckCommand(command []string) []string {
	if len(command) == 1 {
		// command/test was specified as a single string which wraps it in /bin/sh (CMD-SHELL)
		command = append([]string{"CMD-SHELL"}, command...)
	}
	return command
}

// parses a health check time string which could be a duration or an integer
func parseHealthCheckTime(field string) (*int64, error) {
	if field != "" {
		duration, err := time.ParseDuration(field)
		if err == nil {
			return adapter.ConvertToTimeInSeconds(&duration), nil
		} else if val, err := strconv.ParseInt(field, 10, 64); err == nil {
			return &val, nil
		} else {
			return nil, fmt.Errorf("Could not parse %s either as an integer or a duration (ex: 1m30s)", field)
		}
	}

	return nil, nil
}

// ConvertToECSPlacementConstraint converts a list of Constraints specified in the
// ecs-params into a format that is compatible with ECSClient calls.
func ConvertToECSPlacementConstraints(ecsParams *ECSParams) ([]*ecs.PlacementConstraint, error) {
	if ecsParams == nil {
		return nil, nil
	}

	constraints := ecsParams.RunParams.TaskPlacement.Constraints

	output := []*ecs.PlacementConstraint{}
	for _, constraint := range constraints {
		ecsConstraint := &ecs.PlacementConstraint{
			Type: aws.String(constraint.Type),
		}
		if constraint.Expression != "" {
			ecsConstraint.Expression = aws.String(constraint.Expression)
		}
		output = append(output, ecsConstraint)
	}

	return output, nil
}

// ConvertToECSPlacementStrategy converts a list of Strategies specified in the
// ecs-params into a format that is compatible with ECSClient calls.
func ConvertToECSPlacementStrategy(ecsParams *ECSParams) ([]*ecs.PlacementStrategy, error) {
	if ecsParams == nil {
		return nil, nil
	}
	strategies := ecsParams.RunParams.TaskPlacement.Strategies

	output := []*ecs.PlacementStrategy{}
	for _, strategy := range strategies {
		ecsStrategy := &ecs.PlacementStrategy{
			Type: aws.String(strategy.Type),
		}
		if strategy.Field != "" {
			ecsStrategy.Field = aws.String(strategy.Field)
		}
		output = append(output, ecsStrategy)
	}

	return output, nil
}
