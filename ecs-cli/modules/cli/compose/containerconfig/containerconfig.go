package containerconfig

import "github.com/aws/aws-sdk-go/service/ecs"

// ContainerConfig all compose fields supported by the ecs-cli
// TODO: finalize fields (Note: Devices and HealthCheck not currently supported)
type ContainerConfig struct {
	Name string

	CapAdd  []string
	CapDrop []string
	Command []string
	CPU     int64
	// Devices            []string
	DNSSearchDomains      []string
	DNSServers            []string
	DockerLabels          map[string]*string
	DockerSecurityOptions []string
	Entrypoint            []string
	Environment           []*ecs.KeyValuePair
	ExtraHosts            []string
	Hostname              string
	// HealthCheck        *ecs.HealthCheck
	Image             string
	Links             []string
	LogConfiguration  *ecs.LogConfiguration
	Memory            int64
	MemoryReservation int64
	MountPoints       []*ecs.MountPoint
	PortMappings      []*ecs.PortMapping
	Privileged        bool
	ReadOnly          bool
	ShmSize           int64
	Tmpfs             []string
	Ulimits           []*ecs.Ulimit
	VolumesFrom       []*ecs.VolumeFrom
	User              string
	WorkingDirectory  string
}
