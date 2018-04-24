package containerconfig

import "github.com/aws/aws-sdk-go/service/ecs"

// ContainerConfig all compose fields supported by the ecs-cli
// TODO: finalize fields
type ContainerConfig struct {
	Name string

	CapAdd                []string
	CapDrop               []string
	Command               []string
	Devices               []string
	DNSSearchDomains      []string
	DNSServers            []string
	DockerLabels          map[string]*string
	DockerSecurityOptions []string
	Entrypoint            []string
	Environment           []*ecs.KeyValuePair
	ExtraHosts            []string
	Hostname              string
	HealthCheck           *ecs.HealthCheck
	Image                 string
	Links                 []string
	LogConfiguration      *ecs.LogConfiguration
	PortMappings          []*ecs.PortMapping
	Privileged            bool
	ReadOnly              bool
	Tmpfs                 []string
	User                  string
	WorkingDirectory      string
}
