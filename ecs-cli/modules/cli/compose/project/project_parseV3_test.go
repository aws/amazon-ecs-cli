package project

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
)

func TestParseV3WithOneFile(t *testing.T) {
	// set up expected ContainerConfig values
	wordpressCon := adapter.ContainerConfig{}
	wordpressCon.Name = "wordpress"
	wordpressCon.CapAdd = []string{"ALL"}
	wordpressCon.CapDrop = []string{"NET_ADMIN"}
	wordpressCon.Command = []string{"echo \"hello world\""}
	wordpressCon.Image = "wordpress"
	wordpressCon.Entrypoint = []string{"/wordpress/entry"}
	wordpressCon.PortMappings = []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(80),
			HostPort:      aws.Int64(80),
			Protocol:      aws.String("tcp"),
		},
	}
	wordpressCon.Devices = []*ecs.Device{
		{
			HostPath:      aws.String("/dev/sda"),
			ContainerPath: aws.String("/dev/sdd"),
			Permissions:   aws.StringSlice([]string{"read"}),
		},
		{
			HostPath:      aws.String("/dev/sdd"),
			ContainerPath: aws.String("/dev/xdr"),
		},
		{
			HostPath: aws.String("/dev/sda"),
		},
	}
	wordpressCon.DNSServers = []string{"2.2.2.2"}
	wordpressCon.DNSSearchDomains = []string{"wrd.search.com", "wrd.search2.com"}
	wordpressCon.Environment = []*ecs.KeyValuePair{
		{
			Name:  aws.String("wordpress_env"),
			Value: aws.String("val1"),
		},
	}
	wordpressCon.DockerLabels = map[string]*string{
		"com.example.wordpress": aws.String("wordpress label"),
	}
	wordpressCon.Hostname = "wrdhost"
	wordpressCon.Links = []string{"mysql"}
	wordLogDriver := aws.String("awslogs")
	wordLogOpts := map[string]*string{
		"awslogs-group":         aws.String("mywrdprs-logs"),
		"awslogs-region":        aws.String("us-west-2"),
		"awslogs-stream-prefix": aws.String("wordpress"),
	}
	wordpressCon.LogConfiguration = &ecs.LogConfiguration{
		LogDriver: wordLogDriver,
		Options:   wordLogOpts,
	}
	wordTmpfsOpt := []string{"rw"}
	wordpressCon.Tmpfs = []*ecs.Tmpfs{
		{
			ContainerPath: aws.String("/run"),
			MountOptions:  aws.StringSlice(wordTmpfsOpt),
			Size:          aws.Int64(1024),
		},
	}
	wordpressCon.Privileged = true
	wordpressCon.ReadOnly = true
	wordpressCon.DockerSecurityOptions = []string{"label:role:ROLE", "label:user:USER"}
	wordpressCon.Ulimits = []*ecs.Ulimit{
		{
			Name:      aws.String("rss"),
			HardLimit: aws.Int64(65535),
			SoftLimit: aws.Int64(65535),
		},
		{
			Name:      aws.String("nofile"),
			HardLimit: aws.Int64(4000),
			SoftLimit: aws.Int64(2000),
		},
		{
			Name:      aws.String("nice"),
			HardLimit: aws.Int64(500),
			SoftLimit: aws.Int64(300),
		},
	}
	wordpressCon.WorkingDirectory = "/wrdprsdir"

	mysqlCon := adapter.ContainerConfig{}
	mysqlCon.Name = "mysql"
	mysqlCon.Image = "mysql"
	mysqlCon.DockerLabels = map[string]*string{
		"com.example.mysql":  aws.String("mysqllabel"),
		"com.example.mysql2": aws.String("anothermysql label"),
	}
	mysqlCon.User = "mysqluser"
	mysqlCon.ExtraHosts = []*ecs.HostEntry{
		{
			Hostname:  aws.String("mysqlexhost"),
			IpAddress: aws.String("10.0.0.0"),
		},
	}

	// set up file
	composeFileString := `version: '3'
services:
  wordpress:
    cap_add:
      - ALL
    cap_drop:
      - NET_ADMIN
    command: 
      - echo "hello world"
    image: wordpress
    entrypoint: /wordpress/entry
    ports: ["80:80"]
    devices:
      - "/dev/sda:/dev/sdd:r"
      - "/dev/sdd:/dev/xdr"
      - "/dev/sda"
    dns:
      - 2.2.2.2
    dns_search:
      - wrd.search.com
      - wrd.search2.com
    environment:
      wordpress_env: val1
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
    hostname: wrdhost
    security_opt:
      - label:role:ROLE
      - label:user:USER
    working_dir: /wrdprsdir
    privileged: true
    ulimits:
      rss: 65535
      nofile:
        soft: 2000
        hard: 4000
      nice:
        soft: 300
        hard: 500
  mysql:
    image: mysql
    labels:
      - "com.example.mysql=mysqllabel"
      - "com.example.mysql2=anothermysql label"
    user: mysqluser
    extra_hosts:
      - "mysqlexhost:10.0.0.0"`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeFileString))
	assert.NoError(t, err, "Unexpected error writing file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// assert # and content of container configs matches expected
	actualConfigs, err := project.parseV3()
	assert.NoError(t, err, "Unexpected error parsing file")

	assert.Equal(t, 2, len(*actualConfigs))

	wp, err := getContainerConfigByName(wordpressCon.Name, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving wordpress config")
	verifyContainerConfig(t, wordpressCon, *wp)

	mysql, err := getContainerConfigByName(mysqlCon.Name, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving mysql config")
	verifyContainerConfig(t, mysqlCon, *mysql)
}

func TestParseV3WithMultipleFiles(t *testing.T) {
	// set up expected ContainerConfig values
	wordpressCon := adapter.ContainerConfig{}
	wordpressCon.Name = "wordpress"
	wordpressCon.Image = "wordpress"
	wordpressCon.PortMappings = []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(80),
			HostPort:      aws.Int64(80),
			Protocol:      aws.String("tcp"),
		},
		{
			ContainerPort: aws.Int64(777),
			HostPort:      aws.Int64(0),
			Protocol:      aws.String("tcp"),
		},
	}
	wordpressCon.Environment = []*ecs.KeyValuePair{
		{
			Name:  aws.String("WRDPRS1"),
			Value: aws.String("val1"),
		},
		{
			Name:  aws.String("WRDPRS2"),
			Value: aws.String("val2"),
		},
	}

	mysqlCon := adapter.ContainerConfig{}
	mysqlCon.Name = "mysql"
	mysqlCon.Image = "mysql"

	redisCon := adapter.ContainerConfig{}
	redisCon.Name = "redis"
	redisCon.Image = "redis"
	redisCon.PortMappings = []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(90),
			HostPort:      aws.Int64(90),
			Protocol:      aws.String("tcp"),
		},
	}

	// set up files
	fileString1 := `version: '3'
services:
  wordpress:
    image: wordpress
    ports:
      - "80:80"
      - "777"
    environment:
      - WRDPRS1=val1
      - WRDPRS2=val2
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

	// assert # and content of container configs matches expected
	actualConfigs, err := project.parseV3()
	assert.NoError(t, err, "Unexpected error parsing file")

	assert.Equal(t, 3, len(*actualConfigs))

	wp, err := getContainerConfigByName(wordpressCon.Name, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving wordpress config")
	verifyContainerConfig(t, wordpressCon, *wp)

	mysql, err := getContainerConfigByName(mysqlCon.Name, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving mysql config")
	verifyContainerConfig(t, mysqlCon, *mysql)

	redis, err := getContainerConfigByName(redisCon.Name, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving redis config")
	verifyContainerConfig(t, redisCon, *redis)
}

func TestParseV3WithTopLevelVolume(t *testing.T) {
	// set up expected ContainerConfig values
	wordpressCon := adapter.ContainerConfig{}
	wordpressCon.Name = "wordpress"
	wordpressCon.Image = "wordpress"
	wordpressCon.CapAdd = []string{"ALL"}
	wordpressCon.PortMappings = []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(80),
			HostPort:      aws.Int64(80),
			Protocol:      aws.String("tcp"),
		},
	}
	wordpressCon.Links = []string{"mysql"}
	wordpressCon.ReadOnly = true

	mysqlCon := adapter.ContainerConfig{}
	mysqlCon.Name = "mysql"
	mysqlCon.Image = "mysql"
	mysqlCon.DockerLabels = map[string]*string{
		"com.example.mysql":  aws.String("mysqllabel"),
		"com.example.mysql2": aws.String("anothermysql label"),
	}
	mysqlCon.Privileged = true
	mysqlCon.MountPoints = []*ecs.MountPoint{
		{
			ContainerPath: aws.String("/var/lib/mysql"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-1"),
		},
		{
			ContainerPath: aws.String("/test/place"),
			ReadOnly:      aws.Bool(true),
			SourceVolume:  aws.String("logging"),
		},
		{
			ContainerPath: aws.String("/var/lib/mysql"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-2"),
		},
	}

	// set up file
	composeFileString := `version: '3'
services:
  wordpress:
    cap_add:
      - ALL
    image: wordpress
    ports: ["80:80"]
    links:
      - mysql
    read_only: true
  mysql:
    image: mysql
    labels:
      - "com.example.mysql=mysqllabel"
      - "com.example.mysql2=anothermysql label"
    privileged: true
    volumes:
     - /opt/data:/var/lib/mysql
     - logging:/test/place:ro
     - /var/lib/mysql
volumes:
  logging:`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeFileString))
	assert.NoError(t, err, "Unexpected error writing file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// assert # and content of container configs matches expected services
	actualConfigs, err := project.parseV3()
	assert.NoError(t, err, "Unexpected error parsing file")

	assert.Equal(t, 2, len(*actualConfigs))

	wp, err := getContainerConfigByName(wordpressCon.Name, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving wordpress config")
	verifyContainerConfig(t, wordpressCon, *wp)

	mysql, err := getContainerConfigByName(mysqlCon.Name, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving mysql config")
	verifyContainerConfig(t, mysqlCon, *mysql)
}

func TestParseV3_ErrorWithExternalVolume(t *testing.T) {
	// set up file with invalid Volume config ("external")
	composeFileString := `version: '3'
services:
  httpd:
    cap_add:
      - ALL
    cap_drop:
      - NET_ADMIN
    command: echo "hello world"
    image: httpd
    entrypoint: /web/entry
    ports: ["80:80"]
    volumes:
     - /opt/data:/var/lib/mysql
     - logging:/test/place:ro
     - /var/lib/test
volumes:
  logging:
    external:true`

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeFileString))
	assert.NoError(t, err, "Unexpected error writing file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// add files to projects
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// assert error when parsing v3 project
	_, err = project.parseV3()
	assert.Error(t, err, "Expected error when parsing project with invalid Volume configuration")
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

func TestParseV3WithEnvFile(t *testing.T) {
	// Set up env file
	envKey := "testEnv"
	envValue := "testValue"
	envContents := []byte(envKey + "=" + envValue)

	envFile, err := ioutil.TempFile("", "envTest")
	assert.NoError(t, err, "Unexpected error in creating test env file")

	defer os.Remove(envFile.Name())

	_, err = envFile.Write(envContents)
	assert.NoError(t, err, "Unexpected error in writing to test env file")

	expectedEnv := []*ecs.KeyValuePair{
		{
			Name:  aws.String(envKey),
			Value: aws.String(envValue),
		},
	}

	serviceName := "web"
	composeFileString := `version: '3'
services:
  web:
    image: httpd
    env_file: ` + envFile.Name()

	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeFileString))
	assert.NoError(t, err, "Unexpected error in writing to test file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// Set up project
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// assert # and content of container configs matches expected services
	actualConfigs, err := project.parseV3()
	assert.NoError(t, err, "Unexpected error parsing file")
	actualConfig, err := getContainerConfigByName(serviceName, actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving container config")

	assert.Equal(t, 1, len(*actualConfigs))
	assert.Equal(t, expectedEnv, actualConfig.Environment)
}

// TODO: add check for fields not used by V3, use to also check V1V2 ContainerConfigs?
func verifyContainerConfig(t *testing.T, expected, actual adapter.ContainerConfig) {
	assert.ElementsMatch(t, expected.CapAdd, actual.CapAdd, "Expected CapAdd to match")
	assert.ElementsMatch(t, expected.CapDrop, actual.CapDrop, "Expected CapDrop to match")
	assert.ElementsMatch(t, expected.Command, actual.Command, "Expected Command to match")
	assert.ElementsMatch(t, expected.Devices, actual.Devices, "Expected Devices to match")
	assert.ElementsMatch(t, expected.DNSSearchDomains, actual.DNSSearchDomains, "Expected DNSSearchDomains to match")
	assert.ElementsMatch(t, expected.DNSServers, actual.DNSServers, "Expected DNSServers to match")
	dockerLabelsEqual := reflect.DeepEqual(expected.DockerLabels, actual.DockerLabels)
	assert.True(t, dockerLabelsEqual, "Expected DockerLabels to match")
	assert.ElementsMatch(t, expected.DockerSecurityOptions, actual.DockerSecurityOptions, "Expected DockerSecurityOptions to match")
	assert.ElementsMatch(t, expected.Entrypoint, actual.Entrypoint, "Expected Entrypoint to match")
	assert.ElementsMatch(t, expected.Environment, actual.Environment, "Expected Environment to match")
	assert.ElementsMatch(t, expected.ExtraHosts, actual.ExtraHosts, "Expected ExtraHosts to match")
	assert.Equal(t, expected.Hostname, actual.Hostname, "Expected Hostname to match")
	assert.Equal(t, expected.Image, actual.Image, "Expected Image to match")
	assert.ElementsMatch(t, expected.Links, actual.Links, "Expected Links to match")
	assert.Equal(t, expected.LogConfiguration, actual.LogConfiguration, "Expected LogConfiguration to match")
	assert.ElementsMatch(t, expected.MountPoints, actual.MountPoints, "Expected MountPoints to match")
	assert.ElementsMatch(t, expected.PortMappings, actual.PortMappings, "Expected PortMappings to match")
	assert.Equal(t, expected.Privileged, actual.Privileged, "Expected Privileged to match")
	assert.Equal(t, expected.ReadOnly, actual.ReadOnly, "Expected ReadOnly to match")
	assert.ElementsMatch(t, expected.Tmpfs, actual.Tmpfs, "Expected Tmpfs to match")
	assert.ElementsMatch(t, expected.Ulimits, actual.Ulimits, "Expected Ulimits to match")
	assert.Equal(t, expected.User, actual.User, "Expected User to match")
	assert.Equal(t, expected.WorkingDirectory, actual.WorkingDirectory, "Expected WorkingDirectory to match")
}
