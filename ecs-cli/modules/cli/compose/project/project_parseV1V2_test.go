package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/adapter"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/stretchr/testify/assert"
)

func TestParseV1V2_Version1_HappyPath(t *testing.T) {
	// expected ContainerConfig values
	redisImage := "redis"
	webImage := "web"
	cpuShares := int64(73)
	command := []string{"bundle exec thin -p 3000"}
	devices := []*ecs.Device{
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
	dnsServers := []string{"1.2.3.4"}
	dnsSearchDomains := []string{"search.example.com"}
	entryPoint := []string{"/code/entrypoint.sh"}
	env := []*ecs.KeyValuePair{
		{
			Name:  aws.String("RACK_ENV"),
			Value: aws.String("development"),
		},
	}
	extraHosts := []*ecs.HostEntry{
		{
			Hostname:  aws.String("test.local"),
			IpAddress: aws.String("127.10.10.10"),
		},
	}
	hostname := "foobarbaz"
	labels := map[string]*string{
		"label1":         aws.String(""),
		"com.foo.label2": aws.String("value"),
	}
	links := []string{"redis:redis"}
	logDriver := aws.String("json-file")
	logOpts := map[string]*string{
		"max-file": aws.String("50"),
		"max-size": aws.String("50k"),
	}
	logging := &ecs.LogConfiguration{
		LogDriver: logDriver,
		Options:   logOpts,
	}
	memory := int64(953) // 1000000000 in Mib
	mountPoints := []*ecs.MountPoint{
		{
			ContainerPath: aws.String("./code"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-0"),
		},
	}
	ports := []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(5000),
			HostPort:      aws.Int64(5000),
			Protocol:      aws.String("tcp"),
		},
		{
			ContainerPort: aws.Int64(8001),
			HostPort:      aws.Int64(8001),
			Protocol:      aws.String("tcp"),
		},
	}
	privileged := true
	readonly := true
	securityOpts := []string{"label:type:test_virt"}
	shmSize := int64(128)
	user := "user"
	ulimits := []*ecs.Ulimit{
		{
			Name:      aws.String("nproc"),
			HardLimit: aws.Int64(65535),
			SoftLimit: aws.Int64(65535),
		},
		{
			Name:      aws.String("nofile"),
			HardLimit: aws.Int64(40000),
			SoftLimit: aws.Int64(20000),
		},
	}
	volumesFrom := []*ecs.VolumeFrom{
		{
			ReadOnly:        aws.Bool(true),
			SourceContainer: aws.String("web"),
		},
	}
	workingDir := "/var"

	composeV1FileString := `web:
  cpu_shares: 73
  command:
   - bundle exec thin -p 3000
  devices:
   - "/dev/sda:/dev/sdd:r"
   - "/dev/sdd:/dev/xdr"
   - "/dev/sda"
  dns:
   - 1.2.3.4
  dns_search: search.example.com
  entrypoint: /code/entrypoint.sh
  environment:
    RACK_ENV: development
  extra_hosts:
   - test.local:127.10.10.10
  hostname: "foobarbaz"
  image: web
  labels:
   - label1
   - com.foo.label2=value
  links:
   - "redis:redis"
  log_driver: json-file
  log_opt:
    max-file: 50
    max-size: 50k
  mem_limit: 1000000000
  ports:
   - '5000:5000'
   - "127.0.0.1:8001:8001"
  privileged: true
  read_only: true
  security_opt:
   - label:type:test_virt
  shm_size: 128M
  ulimits:
    nproc: 65535
    nofile:
      soft: 20000
      hard: 40000
  user: user
  volumes:
   - ./code
  working_dir: /var
redis:
  image: redis
  volumes_from:
    - web:ro`

	// Setup docker-compose file
	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeV1FileString))
	assert.NoError(t, err, "Unexpected error in writing to test file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// Set up project
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// Parse docker compose file
	actualConfigs, err := project.parseV1V2()
	assert.NoError(t, err, "Unexpected error parsing file")

	// verify redis ContainerConfig
	redis, err := getContainerConfigByName("redis", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving redis config")

	assert.Equal(t, redisImage, redis.Image, "Expected image to match")
	assert.Equal(t, volumesFrom, redis.VolumesFrom, "Expected VolumesFrom to match")

	// verify web ContainerConfig
	web, err := getContainerConfigByName("web", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving web config")

	assert.Equal(t, command, web.Command, "Expected Command to match")
	assert.Equal(t, cpuShares, web.CPU, "Expected CPU to match")
	assert.ElementsMatch(t, devices, web.Devices, "Expected Devices to match")
	assert.Equal(t, dnsSearchDomains, web.DNSSearchDomains, "Expected DNSSearchDomains to match")
	assert.Equal(t, dnsServers, web.DNSServers, "Expected DNSServers to match")
	assert.Equal(t, labels, web.DockerLabels, "Expected DockerLabels to match")
	assert.Equal(t, securityOpts, web.DockerSecurityOptions, "Expected DockerSecurityOptions to match")
	assert.Equal(t, entryPoint, web.Entrypoint, "Expected EntryPoint to be match")
	assert.Equal(t, env, web.Environment, "Expected Environment to match")
	assert.Equal(t, extraHosts, web.ExtraHosts, "Expected ExtraHosts to match")
	assert.Equal(t, hostname, web.Hostname, "Expected Hostname to match")
	assert.Equal(t, webImage, web.Image, "Expected Image to match")
	assert.Equal(t, links, web.Links, "Expected Links to match")
	assert.Equal(t, logging, web.LogConfiguration, "Expected LogConfiguration to match")
	assert.Equal(t, memory, web.Memory, "Expected Memory to match")
	assert.Equal(t, mountPoints, web.MountPoints, "Expected MountPoints to match")
	assert.ElementsMatch(t, ports, web.PortMappings, "Expected PortMappings to match")
	assert.Equal(t, privileged, web.Privileged, "Expected Privileged to match")
	assert.Equal(t, readonly, web.ReadOnly, "Expected ReadOnly to match")
	assert.Equal(t, shmSize, web.ShmSize, "Expected ShmSize to match")
	assert.Equal(t, user, web.User, "Expected User to match")
	assert.ElementsMatch(t, ulimits, web.Ulimits, "Expected Ulimits to match")
	assert.Equal(t, workingDir, web.WorkingDirectory, "Expected WorkingDirectory to match")
}

func TestParseV1V2_Version2Files(t *testing.T) {
	// expected ContainerConfig values
	wordpressImage := "wordpress"
	mysqlImage := "mysql"

	capAdd := []string{"ALL"}
	capDrop := []string{"NET_ADMIN", "SYS_ADMIN"}
	devices := []*ecs.Device{
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
	logOpts := map[string]*string{
		"syslog-address": aws.String("tcp://192.168.0.42:123"),
	}
	logging := &ecs.LogConfiguration{
		LogDriver: aws.String("syslog"),
		Options:   logOpts,
	}
	memoryReservation := int64(476) // 500000000 / miB
	memory := int64(512)
	mountPoints := []*ecs.MountPoint{
		{
			ContainerPath: aws.String("/var/lib/mysql"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-1"),
		},
		{
			ContainerPath: aws.String("/var/lib/mysql"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-2"),
		},
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-3"),
		},
		{
			ContainerPath: aws.String("/etc/configs/"),
			ReadOnly:      aws.Bool(true),
			SourceVolume:  aws.String("volume-4"),
		},
		{
			ContainerPath: aws.String("/var/lib/mysql"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("datavolume"),
		},
	}
	ports := []*ecs.PortMapping{
		{
			ContainerPort: aws.Int64(80),
			HostPort:      aws.Int64(80),
			Protocol:      aws.String("tcp"),
		},
	}
	shmSize := int64(1024) // 1 gb = 1024 miB
	tmpfs := []*ecs.Tmpfs{
		{
			ContainerPath: aws.String("/run"),
			MountOptions:  aws.StringSlice([]string{}),
			Size:          aws.Int64(1024),
		},
		{
			ContainerPath: aws.String("/tmp"),
			MountOptions:  aws.StringSlice([]string{"ro", "rw"}),
			Size:          aws.Int64(64),
		},
	}
	volumesFrom := []*ecs.VolumeFrom{{
		ReadOnly:        aws.Bool(false),
		SourceContainer: aws.String("mysql"),
	},
	}

	composeV2FileString := `version: '2'
services:
  wordpress:
    cap_add:
      - ALL
    cap_drop:
      - NET_ADMIN
      - SYS_ADMIN
    devices:
      - "/dev/sda:/dev/sdd:r"
      - "/dev/sdd:/dev/xdr"
      - "/dev/sda"
    image: wordpress
    ports: ["80:80"]
    mem_reservation: 500000000
    mem_limit: 512M
    shm_size: 1gb
    tmpfs:
      - /run:size=1gb
      - /tmp:size=65536k,ro,rw
    logging:
      driver: syslog
      options:
        syslog-address: "tcp://192.168.0.42:123"
    volumes_from:
      - mysql:rw
  mysql:
    image: mysql
    volumes:
      - /var/lib/mysql
      - /opt/data:/var/lib/mysql
      - ./cache:/tmp/cache:rw
      - ~/configs:/etc/configs/:ro
      - datavolume:/var/lib/mysql
volumes:
  datavolume:`

	// Setup docker-compose file
	tmpfile, err := ioutil.TempFile("", "test")
	assert.NoError(t, err, "Unexpected error in creating test file")

	defer os.Remove(tmpfile.Name())

	_, err = tmpfile.Write([]byte(composeV2FileString))
	assert.NoError(t, err, "Unexpected error in writing to test file")

	err = tmpfile.Close()
	assert.NoError(t, err, "Unexpected error closing file")

	// Set up project
	project := setupTestProject(t)
	project.ecsContext.ComposeFiles = append(project.ecsContext.ComposeFiles, tmpfile.Name())

	// Parse docker compose file
	actualConfigs, err := project.parseV1V2()
	assert.NoError(t, err, "Unexpected error parsing file")

	// verify wordpress ContainerConfig
	wordpress, err := getContainerConfigByName("wordpress", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving wordpress config")

	assert.Equal(t, wordpressImage, wordpress.Image, "Expected wordpress Image to match")
	assert.Equal(t, capAdd, wordpress.CapAdd, "Expected CapAdd to match")
	assert.Equal(t, capDrop, wordpress.CapDrop, "Expected CapDrop to match")
	assert.ElementsMatch(t, devices, wordpress.Devices, "Expected Devices to match")
	assert.Equal(t, logging, wordpress.LogConfiguration, "Expected Log Configuration to match")

	assert.Equal(t, memoryReservation, wordpress.MemoryReservation, "Expected memoryReservation to match")
	assert.Equal(t, memory, wordpress.Memory, "Expected Memory to match")
	assert.Equal(t, ports, wordpress.PortMappings, "Expected ports to match")
	assert.Equal(t, shmSize, wordpress.ShmSize, "Expected shmSize to match")
	assert.ElementsMatch(t, tmpfs, wordpress.Tmpfs, "Expected tmpfs to match")
	assert.Equal(t, volumesFrom, wordpress.VolumesFrom, "Expected VolumesFrom to match")

	// verify mysql ContainerConfig
	mysql, err := getContainerConfigByName("mysql", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving wordpress config")

	assert.Equal(t, mysqlImage, mysql.Image, "Expected mysql Image to match")
	assert.ElementsMatch(t, mountPoints, mysql.MountPoints, "Expected MountPoints to match")
}

func TestParseV1V2_Version1_WithEnvFile(t *testing.T) {
	// Set up env file
	envKey := "rails_env"
	envValue := "development"
	envContents := []byte(envKey + "=" + envValue)

	envFile, err := ioutil.TempFile("", "example")
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

	// Setup docker-compose file
	webImage := "webapp"
	composeFileString := `web:
  image: webapp
  env_file:
  - ` + envFile.Name()

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

	actualConfigs, err := project.parseV1V2()
	assert.NoError(t, err, "Unexpected error parsing file")

	// verify wordpress ServiceConfig
	web, err := getContainerConfigByName("web", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving container config")
	assert.Equal(t, webImage, web.Image, "Expected Image to match")
	assert.Equal(t, expectedEnv, web.Environment, "Expected Environment to match")
}

func TestParseV1V2_WithEnvFromShell(t *testing.T) {
	// If the value for a key in the environment field is blank, it
	// resolves to the shell value (good for specifying secrets).
	// If a value is specified for a key in the environment field, it
	// overrides the shell value for that key.

	envKey1 := "RACK_ENV"
	envValue1 := "staging"

	envKey2 := "SHOW"
	envValue2 := "true"
	envKey3 := "SESSION_SECRET"
	envValue3 := "clydeIsTheGoodestDog"

	os.Setenv(envKey1, envValue1)
	os.Setenv(envKey2, envValue2)
	os.Setenv(envKey3, envValue3)
	defer func() {
		os.Unsetenv(envKey1)
		os.Unsetenv(envKey2)
		os.Unsetenv(envKey3)
	}()

	expectedEnv := []*ecs.KeyValuePair{
		{
			Name:  aws.String(envKey1),
			Value: aws.String("development"),
		},
		{
			Name:  aws.String(envKey2),
			Value: aws.String(envValue2),
		},
		{
			Name:  aws.String(envKey3),
			Value: aws.String(envValue3),
		},
	}
	// Setup docker-compose file
	webImage := "webapp"
	composeFileString := `version: '2'
services:
  web:
    image: webapp
    environment:
     - RACK_ENV=development
     - SHOW=
     - SESSION_SECRET`

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

	actualConfigs, err := project.parseV1V2()
	assert.NoError(t, err, "Unexpected error parsing file")

	// verify wordpress ServiceConfig
	web, err := getContainerConfigByName("web", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving container config")
	assert.Equal(t, webImage, web.Image, "Expected Image to match")
	assert.Equal(t, expectedEnv, web.Environment, "Expected Environment to match")
}

func TestParseV1V2_MemoryValidation(t *testing.T) {
	// Setup docker-compose file
	memory := int64(128)
	composeFileString := `version: '2'
services:
  web:
    image: webapp
    mem_reservation: 128m`

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

	actualConfigs, err := project.parseV1V2()
	assert.NoError(t, err, "Unexpected error parsing file")

	// verify wordpress ServiceConfig
	web, err := getContainerConfigByName("web", actualConfigs)
	if assert.NoError(t, err) {
		assert.Equal(t, memory, web.Memory, "Expected Memory to match")
		assert.Equal(t, memory, web.MemoryReservation, "Expected MemoryReservation to match")
	}
}

func getContainerConfigByName(name string, configs *[]adapter.ContainerConfig) (*adapter.ContainerConfig, error) {
	for _, config := range *configs {
		if config.Name == name {
			return &config, nil
		}
	}
	return nil, fmt.Errorf("Container with name %v could not be found", name)
}
