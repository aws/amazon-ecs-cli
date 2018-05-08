package project

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/compose"
	"github.com/aws/aws-sdk-go/aws"
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

func TestThrowErrorOnBadYaml(t *testing.T) {
	badPortsYaml := `version: '2'
services:
  wordpress:
    image: wordpress
    ports:
      - "80:80", "77:77"
  mysql:
	image: mysql`

	tmpfile, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatal("Unexpected error in creating test file: ", err)
	}
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.Write([]byte(badPortsYaml))
	assert.NoError(t, err, "Unexpected error parsing file")

	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	_, err = project.parseV3()
	assert.Error(t, err)
}

func TestThrowErrorIfFileDoesNotExist(t *testing.T) {
	var fakeFileName = "/missingFile"
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, fakeFileName)
	_, err := project.parseV3()
	assert.Error(t, err)
}

func verifyConvertToContainerConfigOutput(t *testing.T, expected types.ServiceConfig, actual containerconfig.ContainerConfig) {

	// verify equivalent fields
	assert.Equal(t, expected.CapAdd, actual.CapAdd, "Expected CapAdd to match")
	assert.Equal(t, expected.CapDrop, actual.CapDrop, "Expected CapDrop to match")
	assert.Equal(t, expected.SecurityOpt, actual.DockerSecurityOptions, "Expected SecurityOpt and DockerSecuirtyOptions to match")
	assert.Equal(t, []string(expected.Entrypoint), actual.Entrypoint, "Expected EntryPoint to match")
	assert.Equal(t, expected.Name, actual.Name, "Expected Name to match")
	assert.Equal(t, expected.Image, actual.Image, "Expected Image to match")
	assert.Equal(t, expected.Hostname, actual.Hostname, "Expected HostName to match")
	assert.Equal(t, expected.Links, actual.Links, "Expected Links to match")
	assert.Equal(t, expected.Privileged, actual.Privileged, "Expected Privileged to match")
	assert.Equal(t, expected.ReadOnly, actual.ReadOnly, "Expected ReadOnly to match")
	assert.Equal(t, []string(expected.Command), actual.Command, "Expected Command to match")
	assert.Equal(t, expected.User, actual.User, "Expected User to match")
	assert.Equal(t, expected.WorkingDir, actual.WorkingDirectory, "Expected WorkingDirectory to match")

	// verify nil-able lists
	if expected.DNSSearch != nil {
		assert.Equal(t, []string(expected.DNSSearch), actual.DNSSearchDomains, "Expected DNSSearch and DNSSearchDomains to match")
	} else {
		assert.Empty(t, actual.DNSSearchDomains)
	}
	if expected.DNS != nil {
		assert.Equal(t, []string(expected.DNS), actual.DNSServers, "Expected DNS and DNSServers to match")
	} else {
		assert.Empty(t, actual.DNSServers)
	}

	// verify custom conversions
	expectedHosts, err := utils.ConvertToExtraHosts(expected.ExtraHosts)
	assert.NoError(t, err, "Unexpected error converting extra hosts")
	assert.Equal(t, expectedHosts, actual.ExtraHosts, "Expected ExtraHosts to match")

	if expected.Tmpfs != nil {
		expectedTmpfs, err := utils.ConvertToTmpfs(yaml.Stringorslice(expected.Tmpfs))
		assert.NoError(t, err, "Unexpected error converting Tmpfs")
		assert.Equal(t, expectedTmpfs, actual.Tmpfs, "Expected Tmpfs to match")
	} else {
		assert.Empty(t, actual.Tmpfs)
	}

	if expected.Logging != nil {
		assert.Equal(t, expected.Logging.Driver, *actual.LogConfiguration.LogDriver, "Expected LogDriver to match")
		logOptsMap := aws.StringMap(expected.Logging.Options)
		assert.Equal(t, logOptsMap, actual.LogConfiguration.Options, "Expected Log Options to match")
	} else {
		assert.Empty(t, actual.LogConfiguration)
	}

	if expected.Labels != nil {
		labelsMap := aws.StringMap(expected.Labels)
		assert.Equal(t, labelsMap, actual.DockerLabels, "Expected Labels and DockerLabels to match")
	} else {
		assert.Empty(t, actual.DockerLabels)
	}

	if len(expected.Ports) > 0 {
		var exPorts []*ecs.PortMapping
		for _, portConfig := range expected.Ports {
			mapping := convertPortConfigToECSMapping(portConfig)
			exPorts = append(exPorts, mapping)
		}
		assert.Equal(t, exPorts, actual.PortMappings, "Expected PortMappings to match")
	} else {
		assert.Empty(t, actual.PortMappings)
	}

	if len(expected.Volumes) > 0 {
		for i, vol := range expected.Volumes {
			if vol.Type == "volume" {
				verifyMountPoint(t, vol, *actual.MountPoints[i])
			}
		}
	} else {
		assert.Empty(t, actual.MountPoints)
	}
	// TODO: verify expected.Environment
}

func verifyMountPoint(t *testing.T, servVolume types.ServiceVolumeConfig, mountPoint ecs.MountPoint) {
	assert.Equal(t, servVolume.Target, mountPoint.ContainerPath, "Expected volume Target and mount point ContainerPath to match")
	assert.Equal(t, servVolume.Source, mountPoint.SourceVolume, "Expected volume Source and mount point SourceVolume to match")
	assert.Equal(t, servVolume.ReadOnly, mountPoint.ReadOnly, "Expected volume and mount point readOnly to match")
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
