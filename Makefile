GOPATH := $(shell go env GOPATH)
GODEP_BIN := $(GOPATH)/bin/dep
GOLINT := $(GOPATH)/bin/golint
VERSION := $(shell cat VERSION)-$(shell git rev-parse --short HEAD)

packages = $$(go list ./... | egrep -v '/vendor/')
files = $$(find . -name '*.go' | egrep -v '/vendor/')

ifeq "$(HOST_BUILD)" "yes"
	# Use host system for building
	BUILD_SCRIPT =./build-deb-host.sh
else
	# Use docker for building
	BUILD_SCRIPT = ./build-deb-docker.sh
endif


.PHONY: all
all: lint vet test build 


version:
	@ echo $(VERSION)

build:          ## Build the binary
	test $(BINARY_NAME)
	go build -o $(BINARY_NAME) -ldflags "-X main.Version=$(VERSION)" 

build-deb:      ## Build DEB package (needs other tools)
	test $(BINARY_NAME)
	test $(DEB_PACKAGE_NAME)
	test "$(DEB_PACKAGE_DESCRIPTION)"
	exec ${BUILD_SCRIPT}
	
test: 
	go test -race $(packages)

vet:            ## Run go vet
	go tool vet -printfuncs=Debug,Debugf,Debugln,Info,Infof,Infoln,Error,Errorf,Errorln $(files)

lint:           ## Run go lint
	$(GOLINT) -set_exit_status $(packages)
$(GOLINT):
	go get -u github.com/golang/lint/golint

clean:
	test $(BINARY_NAME)
	rm -f $(BINARY_NAME) 

help:           ## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

