package project

import (
	"fmt"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

func (p *ecsProject) checkComposeVersion() (string, error) {
	var composeVersion string
	if len(p.ecsContext.ComposeFiles) == 0 {
		logrus.Error("No Compose files found.")
	}
	for _, file := range p.ecsContext.ComposeFiles {
		fileVersion, err := getFileVersion(file)
		if err != nil {
			return "", err
		}
		if composeVersion != "" && composeVersion != fileVersion {
			return "", fmt.Errorf("Compose files must be of the same version. Found: %s and %s", composeVersion, fileVersion)
		}
		composeVersion = fileVersion
	}
	return composeVersion, nil
}

func getFileVersion(file string) (string, error) {
	type ComposeVersion struct {
		Version string `json:"version"`
	}
	version := &ComposeVersion{}

	loadedFile, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal([]byte(loadedFile), version)
	if err != nil {
		return "", errors.Wrapf(err, "Error unmarshalling yaml data from Compose file: %v", file)
	}
	logrus.Debugf("Docker Compose version found: %s", version.Version)

	return version.Version, nil
}
