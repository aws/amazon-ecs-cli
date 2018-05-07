package project

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
	"github.com/sirupsen/logrus"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"

	"github.com/docker/cli/cli/compose/types"
	"github.com/docker/libcompose/yaml"
)

func TestParseV3WithOneFile(t *testing.T) {
	// set up file
	composeFileString := `version: '3'
services:
  wordpress:
    cap_add:
      - ALL
    cap_drop:
      - NET_ADMIN
    command: echo "hello world"
    image: wordpress
    entrypoint: /wordpress/entry
    ports: ["80:80"]
    dns:
      - 2.2.2.2
    dns_search:
      - wrd.search.com
      - wrd.search2.com
    labels:
      com.example.wordpress: "wordpress label"
    links:
      - mysql
    logging:
      driver: awslogs
      options:
        awslogs-group: mywrdprs-logs
        awslogs-region: us-west-2
        awslogs-stream-prefix: wordpress
    tmpfs:
      - "/run:rw,size=1gb"
    read_only: true
  mysql:
    image: mysql
    labels:
      - "com.example.mysql=mysqllabel"
      - "com.example.mysql2=anothermysql label"
    user: mysqluser
    hostname: mysqlhost
    security_opt:
      - label:role:ROLE
      - label:user:USER
    working_dir: /mysqltestdir
    privileged: true
    extra_hosts:
      - "mysqlexhost:10.0.0.0"`
	logrus.Info("compose string: %s", composeFileString)

	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal("Unexpected error in creating test file: ", err)
	}
	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeFileString))
	assert.NoError(t, err, "Unexpected error parsing file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error parsing file")

	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// get map of docker-parsed ServiceConfigs
	expectedConfigs := getServiceConfigMap(t, project.ecsContext.ComposeFiles)

	// assert # and content of container configs matches expected services
	actualConfigs, err := project.parseV3()
	if err != nil {
		t.Fatal("Unexpected error parsing file: ", err)
	}
	assert.Equal(t, len(expectedConfigs), len(*actualConfigs))
	for _, containerConfig := range *actualConfigs {
		verifyConvertToContainerConfigOutput(t, expectedConfigs[containerConfig.Name], containerConfig)
	}
}

func TestParseV3WithMultipleFiles(t *testing.T) {
	// set up files
	fileString1 := `version: '3'
services:
  wordpress:
    image: wordpress
    ports: 
      - "80:80"
      - "777"
  mysql:
    image: mysql`

	fileString2 := `version: '3'
services:
  redis:
    image: redis
    ports: ["90:90"]`

	// initialize temp files
	tmpfile1, err1 := ioutil.TempFile("", "test")
	assert.NoError(t, err1, "Unexpected error parsing file")
	defer os.Remove(tmpfile1.Name())

	tmpfile2, err2 := ioutil.TempFile("", "test")
	assert.NoError(t, err2, "Unexpected error parsing file")
	defer os.Remove(tmpfile2.Name())

	// write compose contents to files
	_, err := tmpfile1.Write([]byte(fileString1))
	assert.NoError(t, err, "Unexpected error parsing file")

	err = tmpfile1.Close()
	assert.NoError(t, err, "Unexpected error parsing file")

	if _, err := tmpfile2.Write([]byte(fileString2)); err != nil {
		t.Fatal("Unexpected error writing to test file: ", err)
	}
	if err := tmpfile2.Close(); err != nil {
		t.Fatal("Unexpected error closing test file: ", err)
	}
	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile1.Name())
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile2.Name())

	// get map of docker-parsed ServiceConfigs
	expectedConfigs := getServiceConfigMap(t, project.ecsContext.ComposeFiles)

	// assert # and content of container configs matches expected services
	actualConfigs, err := project.parseV3()
	if err != nil {
		t.Fatal("Unexpected error parsing file: ", err)
	}
	assert.Equal(t, len(expectedConfigs), len(*actualConfigs))
	for _, containerConfig := range *actualConfigs {
		verifyConvertToContainerConfigOutput(t, expectedConfigs[containerConfig.Name], containerConfig)
	}
}

func verifyConvertToContainerConfigOutput(t *testing.T, observed types.ServiceConfig, expected containerconfig.ContainerConfig) {

	// verify equivalent fields
	assert.Equal(t, expected.CapAdd, observed.CapAdd)
	assert.Equal(t, expected.CapDrop, observed.CapDrop)
	assert.Equal(t, expected.DockerSecurityOptions, observed.SecurityOpt)
	assert.Equal(t, expected.Entrypoint, []string(observed.Entrypoint))
	assert.Equal(t, expected.Name, observed.Name)
	assert.Equal(t, expected.Image, observed.Image)
	assert.Equal(t, expected.Hostname, observed.Hostname)
	assert.Equal(t, expected.Links, observed.Links)
	assert.Equal(t, expected.Privileged, observed.Privileged)
	assert.Equal(t, expected.ReadOnly, observed.ReadOnly)
	assert.Equal(t, expected.Command, []string(observed.Command))
	assert.Equal(t, expected.User, observed.User)
	assert.Equal(t, expected.WorkingDirectory, observed.WorkingDir)

	// verify nil-able lists
	if observed.DNSSearch != nil {
		assert.Equal(t, expected.DNSSearchDomains, []string(observed.DNSSearch))
	} else {
		assert.Empty(t, expected.DNSSearchDomains)
	}
	if observed.DNS != nil {
		assert.Equal(t, expected.DNSServers, []string(observed.DNS))
	} else {
		assert.Empty(t, expected.DNSServers)
	}

	// verify custom conversions
	observedExHosts, err := utils.ConvertToExtraHosts(observed.ExtraHosts)
	assert.NoError(t, err, "Unexpected error converting extra hosts")
	assert.Equal(t, expected.ExtraHosts, observedExHosts)

	if observed.Tmpfs != nil {
		observedTmpfs, err := utils.ConvertToTmpfs(yaml.Stringorslice(observed.Tmpfs))
		assert.NoError(t, err, "Unexpected error converting Tmpfs")
		assert.Equal(t, expected.Tmpfs, observedTmpfs)
	} else {
		assert.Empty(t, expected.Tmpfs)
	}

	if observed.Logging != nil {
		assert.Equal(t, *expected.LogConfiguration.LogDriver, observed.Logging.Driver)
		logOptsMap := makePointerMapForStringMap(observed.Logging.Options)
		assert.Equal(t, expected.LogConfiguration.Options, logOptsMap)
	} else {
		assert.Empty(t, expected.LogConfiguration)
	}

	if observed.Labels != nil {
		labelsMap := makePointerMapForStringMap(observed.Labels)
		assert.Equal(t, expected.DockerLabels, labelsMap)
	} else {
		assert.Empty(t, expected.DockerLabels)
	}

	if len(observed.Ports) > 0 {
		var obsPorts []*ecs.PortMapping
		for _, portConfig := range observed.Ports {
			mapping := convertPortConfigToECSMapping(portConfig)
			obsPorts = append(obsPorts, mapping)
		}
		assert.Equal(t, expected.PortMappings, obsPorts)
	} else {
		assert.Empty(t, expected.PortMappings)
	}

	if len(observed.Volumes) > 0 {
		for i, vol := range observed.Volumes {
			if vol.Type == "volume" {

				verifyMountPoint(t, vol, *expected.MountPoints[i])
			}
		}
	} else {
		assert.Empty(t, expected.MountPoints)
	}
	// TODO: verify expected.Environment
}

func verifyMountPoint(t *testing.T, servVolume types.ServiceVolumeConfig, mountPoint ecs.MountPoint) {
	assert.Equal(t, servVolume.Target, mountPoint.ContainerPath)
	assert.Equal(t, servVolume.Source, mountPoint.SourceVolume)
	assert.Equal(t, servVolume.ReadOnly, mountPoint.ReadOnly)
}

func getServiceConfigMap(t *testing.T, composeFiles []string) map[string]types.ServiceConfig {
	// confirm files can be parsed by docker
	expectedDockerConfig, err := getV3Config(composeFiles)
	assert.NoError(t, err, "Unexpected error parsing v3 files")

	servMap := make(map[string]types.ServiceConfig)
	for _, service := range expectedDockerConfig.Services {
		servMap[service.Name] = service
	}
	return servMap
}
