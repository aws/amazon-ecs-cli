package project

import (
	"io/ioutil"
	"path/filepath"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/cli/cli/compose/types"
)

func (p *ecsProject) parseV3() (*[]containerconfig.ContainerConfig, error) {
	logrus.Debug("Parsing v3 project...")

	// load and parse each compose file into a configFile
	configFiles := []types.ConfigFile{}
	for _, file := range p.ecsContext.ComposeFiles {

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

	wrkDir, err := getWorkingDir(p.ecsContext.ComposeFiles[0])
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

	// convert ServiceConfigs to ContainerConfigs
	conConfigs := []containerconfig.ContainerConfig{}
	for _, service := range config.Services {
		cCon, err := convertToContainerConfig(service)
		if err != nil {
			return nil, err
		}
		conConfigs = append(conConfigs, *cCon)
	}

	return &conConfigs, nil
}

func convertToContainerConfig(serviceConfig types.ServiceConfig) (*containerconfig.ContainerConfig, error) {
	//TODO: fully convert ServiceConfig to ContainerConfig
	c := &containerconfig.ContainerConfig{
		Name: serviceConfig.Name,
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
