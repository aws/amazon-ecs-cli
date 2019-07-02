package adapter

import "github.com/aws/aws-sdk-go/service/ecs"

// ContainerConfig all compose fields supported by the ecs-cli
type ContainerConfig struct {
	Name string

	CapAdd                []string
	CapDrop               []string
	Command               []string
	CPU                   int64
	Devices               []*ecs.Device
	DNSSearchDomains      []string
	DNSServers            []string
	DockerLabels          map[string]*string
	DockerSecurityOptions []string
	Entrypoint            []string
	Environment           []*ecs.KeyValuePair
	ExtraHosts            []*ecs.HostEntry
	HealthCheck           *ecs.HealthCheck
	Hostname              string
	Image                 string
	InitProcessEnabled    bool
	Links                 []string
	LogConfiguration      *ecs.LogConfiguration
	Memory                int64
	MemoryReservation     int64
	MountPoints           []*ecs.MountPoint
	PortMappings          []*ecs.PortMapping
	Privileged            bool
	PseudoTerminal        bool
	ReadOnly              bool
	ShmSize               int64

	// `ContainerConfig` contains the union of
	// all fields supported by Docker Compose v1~v3 and some of the
	// fields did not exist prior to certain version, so we need to
	// to use pointer type for these field in order to distinguish
	// between "not set" and cases like: "customer explicitly set it to 0".
	StopTimeout      *int64
	Tmpfs            []*ecs.Tmpfs
	Ulimits          []*ecs.Ulimit
	VolumesFrom      []*ecs.VolumeFrom
	User             string
	WorkingDirectory string
}
