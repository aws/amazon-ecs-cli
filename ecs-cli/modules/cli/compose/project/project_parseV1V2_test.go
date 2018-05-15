package project

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/containerconfig"
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
	workingDir := "/var"

	composeV1FileString := `web:
  cpu_shares: 73
  command:
   - bundle exec thin -p 3000
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
    nofile: 1024
  user: user
  volumes:
   - ./code
  working_dir: /var
redis:
  image: redis`

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

	// verify web ContainerConfig
	web, err := getContainerConfigByName("web", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving web config")

	assert.Equal(t, command, web.Command, "Expected Command to match")
	assert.Equal(t, cpuShares, web.CPU, "Expected CPU to match")
	assert.Equal(t, dnsSearchDomains, web.DNSSearchDomains, "Expected DNSSearchDomains to match")
	assert.Equal(t, dnsServers, web.DNSServers, "Expected DNSServers to match")
	assert.Equal(t, labels, web.DockerLabels, "Expected docker labels to match")
	assert.Equal(t, securityOpts, web.DockerSecurityOptions, "Expected DockerSecurityOptions to match")
	assert.Equal(t, entryPoint, web.Entrypoint, "Expected EntryPoint to be match")
	assert.Equal(t, env, web.Environment, "Expected Environment to match")
	assert.Equal(t, extraHosts, web.ExtraHosts, "Expected ExtraHosts to match")
	assert.Equal(t, hostname, web.Hostname, "Expected Hostname to match")
	assert.Equal(t, webImage, web.Image, "Expected Image to match")
	assert.Equal(t, links, web.Links, "Expected Links to match")
	assert.Equal(t, logging, web.LogConfiguration, "Expected LogConfiguration to match")
	assert.Equal(t, memory, web.Memory, "Expected memory to match")
	assert.Equal(t, mountPoints, web.MountPoints, "Expected MountPoints to match")
	assert.Equal(t, ports, web.PortMappings, "Expected PortMappings to match")
	assert.Equal(t, privileged, web.Privileged, "Expected Privileged to match")
	assert.Equal(t, readonly, web.ReadOnly, "Expected ReadOnly to match")
	assert.Equal(t, shmSize, web.ShmSize, "Expected ShmSize to match")
	assert.Equal(t, user, web.User, "Expected user to match")
	assert.Equal(t, workingDir, web.WorkingDirectory, "Expected WorkingDirectory to match")
}

func TestParseV1V2_Version2Files(t *testing.T) {
	// expected ContainerConfig values
	wordpressImage := "wordpress"
	mysqlImage := "mysql"

	capAdd := []string{"ALL"}
	capDrop := []string{"NET_ADMIN", "SYS_ADMIN"}
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
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("banana"),
		},
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(false),
			SourceVolume:  aws.String("volume-1"),
		},
		{
			ContainerPath: aws.String("/tmp/cache"),
			ReadOnly:      aws.Bool(true),
			SourceVolume:  aws.String("volume-2"),
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

	composeV2FileString := `version: '2'
services:
  wordpress:
    cap_add:
      - ALL
    cap_drop:
      - NET_ADMIN
      - SYS_ADMIN
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
  mysql:
    image: mysql
    volumes:
      - banana:/tmp/cache
      - :/tmp/cache
      - ./cache:/tmp/cache:ro
volumes:
  banana:`

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
	assert.Equal(t, logging, wordpress.LogConfiguration, "Expected Log Configuration to match")

	assert.Equal(t, memoryReservation, wordpress.MemoryReservation, "Expected memoryReservation to match")
	assert.Equal(t, memory, wordpress.Memory, "Expected Memory to match")
	assert.Equal(t, ports, wordpress.PortMappings, "Expected ports to match")
	assert.Equal(t, shmSize, wordpress.ShmSize, "Expected shmSize to match")
	assert.ElementsMatch(t, tmpfs, wordpress.Tmpfs, "Expected tmpfs to match")

	// verify mysql ContainerConfig
	mysql, err := getContainerConfigByName("mysql", actualConfigs)
	assert.NoError(t, err, "Unexpected error retrieving wordpress config")

	assert.Equal(t, mysqlImage, mysql.Image, "Expected mysql Image to match")
	assert.Equal(t, mountPoints, mysql.MountPoints, "Expected MountPoints to match")
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

func getContainerConfigByName(name string, configs *[]containerconfig.ContainerConfig) (*containerconfig.ContainerConfig, error) {
	for _, config := range *configs {
		if config.Name == name {
			return &config, nil
		}
	}
	return nil, fmt.Errorf("Container with name %v could not be found.", name)
}
