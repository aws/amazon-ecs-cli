module github.com/aws/amazon-ecs-cli

go 1.13

require (
	bazil.org/fuse v0.0.0-20160811212531-371fbbdaa898 // indirect
	github.com/Microsoft/hcsshim v0.8.7-0.20190419161850-2226e083fc39 // indirect
	github.com/aws/aws-sdk-go v1.34.9
	github.com/awslabs/amazon-ecr-credential-helper v0.4.0
	github.com/containerd/containerd v1.4.9 // indirect
	github.com/containerd/ttrpc v1.0.0 // indirect
	github.com/containerd/typeurl v1.0.1 // indirect
	github.com/coreos/go-systemd/v22 v22.1.0 // indirect
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/docker/cli v0.0.0-20190814185437-1752eb3626e3
	github.com/docker/docker v1.4.2-0.20191101170500-ac7306503d23
	github.com/docker/go-units v0.4.0
	github.com/docker/libcompose v0.4.1-0.20171025083809-57bd716502dc
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/frankban/quicktest v1.11.3 // indirect
	github.com/fsouza/go-dockerclient v1.5.0
	github.com/go-ini/ini v1.60.1
	github.com/godbus/dbus v0.0.0-20190422162347-ade71ed3457e // indirect
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-cmp v0.5.5 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/hashicorp/golang-lru v0.5.1 // indirect
	github.com/imdario/mergo v0.3.4 // indirect
	github.com/kisielk/errcheck v1.5.0 // indirect
	github.com/mattn/go-shellwords v1.0.3 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/onsi/ginkgo v1.10.1 // indirect
	github.com/onsi/gomega v1.7.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/runc v1.0.0-rc2.0.20161227072456-f376b8033d2c // indirect
	github.com/opencontainers/runtime-spec v1.0.3-0.20200929063507-e6143ca7d51d // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/procfs v0.0.0-20190522114515-bc1a522cf7b1 // indirect
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/smartystreets/goconvey v1.7.2 // indirect
	github.com/spf13/cobra v1.0.0 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.2
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20170528113821-0c8571ac0ce1 // indirect
	github.com/yuin/goldmark v1.2.1 // indirect
	go.opencensus.io v0.22.3 // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/mod v0.3.0 // indirect
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b
	golang.org/x/sync v0.0.0-20201207232520-09787c993a3a // indirect
	golang.org/x/sys v0.0.0-20190812073006-9eafafc0a87e // indirect
	golang.org/x/text v0.3.3 // indirect
	golang.org/x/time v0.1.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/genproto v0.0.0-20200617032506-f1bdc9086088 // indirect
	google.golang.org/grpc v1.33.2 // indirect
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	gotest.tools/v3 v3.0.3 // indirect
)

replace github.com/docker/docker v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
