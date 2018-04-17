ROOT := $(shell pwd)
MOCKGEN := vendor/github.com/golang/mock/mockgen
MOCKFQ := $(ROOT)/$(MOCKGEN)
all: generate-deps

.PHONY: generate-deps
generate-deps: $(GOPATH)/bin/goimports $(GOPATH)/bin/dep
	dep ensure -v
	dep prune 2> /dev/null
	-rm $(ROOT)/../scripts/mockgen 2> /dev/null
	set -e ; cd $(MOCKFQ) ; go build ; cd $(ROOT)/../scripts ; ln -s ../$(shell basename $(ROOT))/$(MOCKGEN)/mockgen .

$(GOPATH)/bin/goimports:
	go get golang.org/x/tools/cmd/goimports

$(GOPATH)/bin/dep:
	curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
