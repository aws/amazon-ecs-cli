module github.com/aws/amazon-ecs-cli

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.15-0.20200113171025-3fe6c5262873 // indirect
	github.com/Microsoft/hcsshim v0.8.9 // indirect
	github.com/aws/aws-sdk-go v1.34.9
	github.com/awslabs/amazon-ecr-credential-helper v0.4.0
	github.com/containerd/containerd v1.3.4 // indirect
	github.com/containerd/continuity v0.0.0-20200413184840-d3ef23f19fbb // indirect
	github.com/docker/cli v0.0.0-20190814185437-1752eb3626e3
	github.com/docker/docker v1.4.2-0.20191101170500-ac7306503d23
	github.com/docker/go-units v0.4.0
	github.com/docker/libcompose v0.4.1-0.20171025083809-57bd716502dc
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568
	github.com/fsouza/go-dockerclient v1.5.0
	github.com/go-ini/ini v1.60.1
	github.com/golang/mock v1.4.4
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/imdario/mergo v0.3.4 // indirect
	github.com/mattn/go-shellwords v1.0.3 // indirect
	github.com/mitchellh/go-homedir v0.0.0-20160301183130-981ab348d865 // indirect
	github.com/mitchellh/mapstructure v0.0.0-20180220230111-00c29f56e238 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/runc v1.0.0-rc2.0.20161227072456-f376b8033d2c // indirect
	github.com/pkg/errors v0.9.1
	github.com/russross/blackfriday v1.6.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/smartystreets/goconvey v1.7.2 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.0
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v0.0.0-20170528113821-0c8571ac0ce1 // indirect
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	golang.org/x/time v0.1.0 // indirect
	google.golang.org/genproto v0.0.0-20200617032506-f1bdc9086088 // indirect
	google.golang.org/grpc v1.29.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776 // indirect
	gotest.tools/v3 v3.4.0 // indirect
)

replace github.com/docker/docker v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
