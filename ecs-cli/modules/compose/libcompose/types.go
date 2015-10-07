// This file is derived from Docker's Libcompose project, Copyright 2015 Docker, Inc.
// The original code may be found here :
// https://github.com/docker/libcompose/blob/master/project/types.go
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Modifications are Copyright 2015 Amazon.com, Inc. or its affiliates. Licensed under the Apache License 2.0
package libcompose

import "fmt"

// ServiceConfig is the Struct that stores the compose yml options for each service/container
type ServiceConfig struct {
	Build         string            `yaml:"build,omitempty"`
	CapAdd        []string          `yaml:"cap_add,omitempty"`
	CapDrop       []string          `yaml:"cap_drop,omitempty"`
	CpuSet        string            `yaml:"cpu_set,omitempty"`
	CpuShares     int64             `yaml:"cpu_shares,omitempty"`
	Command       Command           `yaml:"command"` // omitempty breaks serialization!
	Detach        string            `yaml:"detach,omitempty"`
	Devices       []string          `yaml:"devices,omitempty"`
	Dns           Stringorslice     `yaml:"dns"`        // omitempty breaks serialization!
	DnsSearch     Stringorslice     `yaml:"dns_search"` // omitempty breaks serialization!
	Dockerfile    string            `yaml:"dockerfile,omitempty"`
	DomainName    string            `yaml:"domainname,omitempty"`
	Entrypoint    Command           `yaml:"entrypoint"`  // omitempty breaks serialization!
	EnvFile       Stringorslice     `yaml:"env_file"`    // omitempty breaks serialization!
	Environment   MaporEqualSlice   `yaml:"environment"` // omitempty breaks serialization!
	Hostname      string            `yaml:"hostname,omitempty"`
	Image         string            `yaml:"image,omitempty"`
	Labels        SliceorMap        `yaml:"labels"` // omitempty breaks serialization!
	Links         MaporColonSlice   `yaml:"links"`  // omitempty breaks serialization!
	LogDriver     string            `yaml:"log_driver,omitempty"`
	MemLimit      int64             `yaml:"mem_limit,omitempty"` // TODO, accept string value: "m" for megabytes
	MemSwapLimit  int64             `yaml:"mem_swap_limit,omitempty"`
	Name          string            `yaml:"name,omitempty"`
	Net           string            `yaml:"net,omitempty"`
	Pid           string            `yaml:"pid,omitempty"`
	Uts           string            `yaml:"uts,omitempty"`
	Ipc           string            `yaml:"ipc,omitempty"`
	Ports         []string          `yaml:"ports,omitempty"`
	Privileged    bool              `yaml:"privileged,omitempty"`
	Restart       string            `yaml:"restart,omitempty"`
	ReadOnly      bool              `yaml:"read_only,omitempty"`
	StdinOpen     bool              `yaml:"stdin_open,omitempty"`
	SecurityOpt   []string          `yaml:"security_opt,omitempty"`
	Tty           bool              `yaml:"tty,omitempty"`
	User          string            `yaml:"user,omitempty"`
	VolumeDriver  string            `yaml:"volume_driver,omitempty"`
	Volumes       []string          `yaml:"volumes,omitempty"`
	VolumesFrom   []string          `yaml:"volumes_from,omitempty"`
	WorkingDir    string            `yaml:"working_dir,omitempty"`
	Expose        []string          `yaml:"expose,omitempty"`
	ExternalLinks []string          `yaml:"external_links,omitempty"`
	LogOpt        map[string]string `yaml:"log_opt,omitempty"`
	ExtraHosts    []string          `yaml:"extra_hosts,omitempty"`
}

type EnvironmentLookup interface {
	Lookup(key, serviceName string, config *ServiceConfig) []string
}

type ConfigLookup interface {
	Lookup(file, relativeTo string) ([]byte, string, error)
}

type Event int

const (
	CONTAINER_ID = "container_id"

	NO_EVENT = Event(iota)

	CONTAINER_CREATED = Event(iota)
	CONTAINER_STARTED = Event(iota)

	PROJECT_CREATE_START = Event(iota)
	PROJECT_CREATE_DONE  = Event(iota)
	PROJECT_UP_START     = Event(iota)
	PROJECT_UP_DONE      = Event(iota)
	PROJECT_DELETE_START = Event(iota)
	PROJECT_DELETE_DONE  = Event(iota)
	PROJECT_START_START  = Event(iota)
	PROJECT_START_DONE   = Event(iota)
)

func (e Event) String() string {
	var m string
	switch e {
	case CONTAINER_CREATED:
		m = "Created container"
	case CONTAINER_STARTED:
		m = "Started container"

	case PROJECT_CREATE_START:
		m = "Creating project"
	case PROJECT_CREATE_DONE:
		m = "Project created"
	case PROJECT_UP_START:
		m = "Starting project"
	case PROJECT_UP_DONE:
		m = "Project started"
	case PROJECT_DELETE_START:
		m = "Deleting project"
	case PROJECT_DELETE_DONE:
		m = "Project deleted"
	case PROJECT_START_START:
		m = "Starting project"
	case PROJECT_START_DONE:
		m = "Project started"
	}

	if m == "" {
		m = fmt.Sprintf("Event: %d", int(e))
	}

	return m
}

type InfoPart struct {
	Key, Value string
}

type InfoSet []Info
type Info []InfoPart
