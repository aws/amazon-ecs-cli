# Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You
# may not use this file except in compliance with the License. A copy of
# the License is located at
#
# 	http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is
# distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF
# ANY KIND, either express or implied. See the License for the specific
# language governing permissions and limitations under the License.

ROOT := $(shell pwd)

all: build

SOURCEDIR := ./ecs-cli
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')
LOCAL_BINARY := bin/local/ecs-cli
LINUX_BINARY := bin/linux-amd64/ecs-cli
DARWIN_BINARY := bin/darwin-amd64/ecs-cli
WINDOWS_BINARY := bin/windows-amd64/ecs-cli.exe
LOCAL_PATH := $(ROOT)/scripts:${PATH}

.PHONY: build
build: $(LOCAL_BINARY)

$(LOCAL_BINARY): $(SOURCES)
	./scripts/build_binary.sh ./bin/local
	@echo "Built ecs-cli"

.PHONY: test
test:
	env -i PATH=$$PATH GOPATH=$$GOPATH GOROOT=$$GOROOT go test -timeout=120s -v -cover ./ecs-cli/modules/...

.PHONY: generate
generate: $(SOURCES)
	PATH=$(LOCAL_PATH) go generate ./ecs-cli/modules/...

.PHONY: generate-deps
generate-deps:
	go get github.com/tools/godep
	go get github.com/golang/mock/mockgen
	go get golang.org/x/tools/cmd/goimports

.PHONY: windows-build
windows-build: $(WINDOWS_BINARY)

.PHONY: docker-build
docker-build:
	docker run -v $(shell pwd):/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--workdir=/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=$(ECS_RELEASE) \
		golang:1.8 make $(LINUX_BINARY)
	docker run -v $(shell pwd):/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--workdir=/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=$(ECS_RELEASE) \
		golang:1.8 make $(DARWIN_BINARY)
	docker run -v $(shell pwd):/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--workdir=/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=$(ECS_RELEASE) \
		golang:1.8 make $(WINDOWS_BINARY)

.PHONY: docker-test
docker-test:
	docker run -v $(shell pwd):/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--workdir=/usr/src/app/src/github.com/aws/amazon-ecs-cli \
		--env GOPATH=/usr/src/app \
		--env ECS_RELEASE=$(ECS_RELEASE) \
		golang:1.8 make test

.PHONY: supported-platforms
supported-platforms: $(LINUX_BINARY) $(DARWIN_BINARY) $(WINDOWS_BINARY)

$(WINDOWS_BINARY): $(SOURCES)
	@mkdir -p ./bin/windows-amd64
	TARGET_GOOS=windows GOARCH=amd64 ./scripts/build_binary.sh ./bin/windows-amd64
	mv ./bin/windows-amd64/ecs-cli ./bin/windows-amd64/ecs-cli.exe
	@echo "Built ecs-cli.exe for windows"

$(LINUX_BINARY): $(SOURCES)
	@mkdir -p ./bin/linux-amd64
	TARGET_GOOS=linux GOARCH=amd64 ./scripts/build_binary.sh ./bin/linux-amd64
	@echo "Built ecs-cli for linux"

$(DARWIN_BINARY): $(SOURCES)
	@mkdir -p ./bin/darwin-amd64
	TARGET_GOOS=darwin GOARCH=amd64 ./scripts/build_binary.sh ./bin/darwin-amd64
	@echo "Built ecs-cli for darwin"

.PHONY: clean
clean:
	rm -rf ./bin/ ||:
