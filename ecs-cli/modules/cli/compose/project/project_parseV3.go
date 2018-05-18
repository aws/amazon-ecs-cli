package project

import (
	"io/ioutil"
	"path/filepath"
	"reflect"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/libcompose/yaml"
)

func (p *ecsProject) parseV3() (*[]containerconfig.ContainerConfig, error) {
	log.Debug("Parsing v3 project...")

	v3Config, err := getV3Config(p.ecsContext.ComposeFiles)
	if err != nil {
		return nil, err
	}

	// convert ServiceConfigs to ContainerConfigs
	conConfigs := []containerconfig.ContainerConfig{}
	for _, service := range v3Config.Services {
		cCon, err := convertToContainerConfig(service)
		if err != nil {
			return nil, err
		}
		conConfigs = append(conConfigs, *cCon)
	}

	// TODO: process v3Config.Volumes as well
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

	configDetails := types.ConfigDetails{
		WorkingDir:  wrkDir,
		ConfigFiles: configFiles,
		Environment: nil,
	}

	// load config from config details
	config, err := loader.Load(configDetails)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func convertToContainerConfig(serviceConfig types.ServiceConfig) (*containerconfig.ContainerConfig, error) {
	//TODO: Add Healthcheck, Devices to ContainerConfig
	c := &containerconfig.ContainerConfig{
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
		env.SetValue(*v)
		envVars = append(envVars, &env)
	}
	c.Environment = envVars

	extraHosts, err := utils.ConvertToExtraHosts(serviceConfig.ExtraHosts)
	if err != nil {
		return nil, err
	}
	c.ExtraHosts = extraHosts

	// TODO: refactor utils.ConvertToLogConfiguration to take in driver (string) and Options (map[string]string)
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
		ecsTmpfs, err := utils.ConvertToTmpfs(tmpfs)
		if err != nil {
			return nil, err
		}
		c.Tmpfs = ecsTmpfs
	}
	// TODO: reconcile with top-level Volumes key
	if len(serviceConfig.Volumes) > 0 {
		mountPoints := []*ecs.MountPoint{}

		for _, volConfig := range serviceConfig.Volumes {
			if volConfig.Type == "volume" {
				mp := &ecs.MountPoint{
					ContainerPath: &volConfig.Target,
					SourceVolume:  &volConfig.Source,
					ReadOnly:      &volConfig.ReadOnly,
				}
				mountPoints = append(mountPoints, mp)
			} else {
				log.Warnf("Unsupported mount type found: %s", volConfig.Type)
			}
		}
		c.MountPoints = mountPoints
	}

	logWarningForDeployFields(serviceConfig.Deploy, serviceConfig.Name)

	// TODO: log out unsupported fields
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
