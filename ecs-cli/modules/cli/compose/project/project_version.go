package project

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

func (p *ecsProject) checkComposeVersion() (string, error) {
	var composeVersion string
	if len(p.ecsContext.ComposeFiles) == 0 {
		return "", fmt.Errorf("No Compose files found")
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

	// if minor version of 1 or 2 found, log warning
	match, _ := regexp.MatchString("^.+\\..", composeVersion)
	if composeVersion != "" && match {
		versionNumber, err := strconv.ParseFloat(composeVersion, 64)
		if err != nil {
			return "", err
		}
		if 0 < versionNumber && versionNumber < 3 {
			logrus.Warnf("Minor version (%s) detected. Please format to include only major version (%d).", composeVersion, int(versionNumber))
		}
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
