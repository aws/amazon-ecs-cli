package project

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/logger"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/libcompose/yaml"
)

func (p *ecsProject) parseV3() (*[]adapter.ContainerConfig, error) {
	log.Debug("Parsing v3 project...")

	v3Config, err := getV3Config(p.ecsContext.ComposeFiles)
	if err != nil {
		return nil, err
	}

	servVols, err := adapter.ConvertToV3Volumes(v3Config.Volumes)
	if err != nil {
		return nil, err
	}
	p.volumes = servVols

	// convert ServiceConfigs to ContainerConfigs
	conConfigs := []adapter.ContainerConfig{}
	for _, service := range v3Config.Services {
		cCon, err := convertToContainerConfig(service, p.volumes)
		if err != nil {
			return nil, err
		}
		conConfigs = append(conConfigs, *cCon)
	}

	return &conConfigs, nil
}

// parses compose files into a docker/cli Config, which contains v3 ServiceConfigs
func getV3Config(composeFiles []string) (*types.Config, error) {
	configFiles := []types.ConfigFile{}
	for _, file := range composeFiles {

		loadedFile, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}
		parsedFile, err := loader.ParseYAML(loadedFile)
		if err != nil {
			return nil, err
		}
		configFile := types.ConfigFile{
			Filename: file,
			Config:   parsedFile,
		}
		configFiles = append(configFiles, configFile)
	}

	wrkDir, err := getWorkingDir(composeFiles[0])
	if err != nil {
		return nil, err
	}

	localEnv := getEnvironment()

	configDetails := types.ConfigDetails{
		WorkingDir:  wrkDir,
		ConfigFiles: configFiles,
		Environment: localEnv,
	}

	// load config from config details
	config, err := loader.Load(configDetails)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func convertToContainerConfig(serviceConfig types.ServiceConfig, serviceVols *adapter.Volumes) (*adapter.ContainerConfig, error) {
	logger.LogUnsupportedV3ServiceConfigFields(serviceConfig)
	logWarningForDeployFields(serviceConfig.Deploy, serviceConfig.Name)

	c := &adapter.ContainerConfig{
		CapAdd:                serviceConfig.CapAdd,
		CapDrop:               serviceConfig.CapDrop,
		Command:               serviceConfig.Command,
		DockerSecurityOptions: serviceConfig.SecurityOpt,
		Entrypoint:            serviceConfig.Entrypoint,
		Hostname:              serviceConfig.Hostname,
		Image:                 serviceConfig.Image,
		Links:                 serviceConfig.Links,
		Name:                  serviceConfig.Name,
		Privileged:            serviceConfig.Privileged,
		ReadOnly:              serviceConfig.ReadOnly,
		User:                  serviceConfig.User,
		WorkingDirectory:      serviceConfig.WorkingDir,
	}

	devices, err := adapter.ConvertToDevices(serviceConfig.Devices)
	if err != nil {
		return nil, err
	}
	c.Devices = devices

	if serviceConfig.HealthCheck != nil && !serviceConfig.HealthCheck.Disable {
		c.HealthCheck = adapter.ConvertToHealthCheck(serviceConfig.HealthCheck)
	}

	if serviceConfig.DNS != nil {
		c.DNSServers = serviceConfig.DNS
	}
	if serviceConfig.DNSSearch != nil {
		c.DNSSearchDomains = serviceConfig.DNSSearch
	}
	if serviceConfig.Labels != nil {
		labelsMap := aws.StringMap(serviceConfig.Labels)
		c.DockerLabels = labelsMap
	}

	envVars := []*ecs.KeyValuePair{}
	for k, v := range serviceConfig.Environment {
		env := ecs.KeyValuePair{}
		env.SetName(k)
		if v != nil {
			env.SetValue(*v)
		} else {
			env.SetValue("")
		}
		envVars = append(envVars, &env)
	}
	c.Environment = envVars

	extraHosts, err := adapter.ConvertToExtraHosts(serviceConfig.ExtraHosts)
	if err != nil {
		return nil, err
	}
	c.ExtraHosts = extraHosts

	// TODO: refactor adapter.ConvertToLogConfiguration to take in driver (string) and Options (map[string]string)
	if serviceConfig.Logging != nil {
		logConfig := ecs.LogConfiguration{}
		logConfig.SetLogDriver(serviceConfig.Logging.Driver)

		optionsMap := aws.StringMap(serviceConfig.Logging.Options)
		logConfig.SetOptions(optionsMap)
		c.LogConfiguration = &logConfig
	}

	if len(serviceConfig.Ports) > 0 {
		var portMappings []*ecs.PortMapping
		for _, portConfig := range serviceConfig.Ports {
			mapping := convertPortConfigToECSMapping(portConfig)
			portMappings = append(portMappings, mapping)
		}
		c.PortMappings = portMappings
	}
	// TODO: change ConvertToTmpfs to take in []string
	if serviceConfig.Tmpfs != nil {
		tmpfs := yaml.Stringorslice(serviceConfig.Tmpfs)
		ecsTmpfs, err := adapter.ConvertToTmpfs(tmpfs)
		if err != nil {
			return nil, err
		}
		c.Tmpfs = ecsTmpfs
	}

	if len(serviceConfig.Ulimits) > 0 {
		c.Ulimits = convertToECSUlimits(serviceConfig.Ulimits)
	}

	if len(serviceConfig.Volumes) > 0 {
		mountPoints := []*ecs.MountPoint{}

		for _, volConfig := range serviceConfig.Volumes {
			if volConfig.Type == "volume" || volConfig.Type == "bind" {

				sourceVolName, err := adapter.GetSourcePathAndUpdateVolumes(volConfig.Source, serviceVols)
				if err != nil {
					return nil, err
				}
				containerPath := volConfig.Target
				readOnly := volConfig.ReadOnly

				mp := &ecs.MountPoint{
					ContainerPath: &containerPath,
					SourceVolume:  &sourceVolName,
					ReadOnly:      &readOnly,
				}
				mountPoints = append(mountPoints, mp)
			} else {
				log.Warnf("Unsupported mount type found: %s", volConfig.Type)
			}
		}
		c.MountPoints = mountPoints
	}

	return c, nil
}

func getWorkingDir(fileName string) (string, error) {
	pwd, err := filepath.Abs(fileName)
	if err != nil {
		return "", errors.Wrap(err, "Unable to retrieve compose file directory")
	}
	return filepath.Dir(pwd), nil
}

func convertPortConfigToECSMapping(portConfig types.ServicePortConfig) *ecs.PortMapping {
	containerPort := int64(portConfig.Target)
	hostPort := int64(portConfig.Published)

	var ecsMapping = ecs.PortMapping{
		ContainerPort: &containerPort,
		HostPort:      &hostPort,
		Protocol:      &portConfig.Protocol,
	}
	return &ecsMapping
}

func convertToECSUlimits(ulimits map[string]*types.UlimitsConfig) []*ecs.Ulimit {
	ecsUlimits := []*ecs.Ulimit{}

	for name, ulimit := range ulimits {
		ecsULimit := ecs.Ulimit{}
		ecsULimit.SetName(name)

		if ulimit.Single > 0 {
			ecsULimit.SetSoftLimit(int64(ulimit.Single))
			ecsULimit.SetHardLimit(int64(ulimit.Single))
		} else {
			ecsULimit.SetSoftLimit(int64(ulimit.Soft))
			ecsULimit.SetHardLimit(int64(ulimit.Hard))
		}
		ecsUlimits = append(ecsUlimits, &ecsULimit)
	}
	return ecsUlimits
}

func logWarningForDeployFields(d types.DeployConfig, serviceName string) {
	if d.Resources.Limits != nil || d.Resources.Reservations != nil {
		log.WithFields(log.Fields{
			"option name":  "deploy",
			"service name": serviceName,
		}).Warn("Skipping unsupported YAML option for service... service-level resources should be configured in the ecs-param.yml file.")
	} else if !reflect.DeepEqual(d, types.DeployConfig{}) {
		log.WithFields(log.Fields{
			"option name":  "deploy",
			"service name": serviceName,
		}).Warn("Skipping unsupported YAML option for service...")
	}
}

func getEnvironment() map[string]string {
	env := os.Environ()
	envMap := make(map[string]string, len(env))
	for _, s := range env {
		varParts := strings.SplitN(s, "=", 2)
		envMap[varParts[0]] = varParts[1]
	}
	return envMap
}
