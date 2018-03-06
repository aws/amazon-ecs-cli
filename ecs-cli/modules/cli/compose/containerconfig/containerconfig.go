package containerconfig

// ContainerConfig all compose fields supported by the ecs-cli
// TODO: finalize fields
type ContainerConfig struct {
	CapAdd      []string
	CapDrop     []string
	Command     string // type TBD
	Devices     []string
	DNS         []string // StringList in docker/cli
	Entrypoint  string   // ShellCommand in docker/cli
	Environment []string // MappingWithEquals in docker/cli
	EnvFile     []string // StringList in docker/cli
	ExtraHosts  []string
	Hostname    string
	HealthCheck interface{} // *HealthCheckConfig in docker/cli
	Image       string
}
