GOPATH := $(shell go env GOPATH)
GODEP := $(GOPATH)/bin/dep
GOLINT := $(GOPATH)/bin/golint
VERSION := $(shell cat VERSION)-$(shell git rev-parse --short HEAD)

BINARY_NAME := $(shell echo ${BINARY_NAME})
packages = $$(go list ./... | egrep -v '/vendor/')
files = $$(find . -name '*.go' | egrep -v '/vendor/')

.PHONY: all
all: lint vet test build

.PHONY: version
version:
	@ echo $(VERSION)

.PHONY: build
build:          ## Build the binary
build: vendor
# Cache packages if CI isn't defined
ifndef CI
		go build -i -o $(BINARY_NAME) -ldflags "-X main.Version=$(VERSION)" 
else
		go build -o $(BINARY_NAME) -ldflags "-X main.Version=$(VERSION)" 
endif

.PHONY: test
test: vendor
ifndef CI
		@ go test -i -race $(packages) # Install test packages; will speed up future `go test` runs.
endif
		go test -race $(packages)

.PHONY: vet
vet:            ## Run go vet
vet: vendor
	go tool vet -printfuncs=Debug,Debugf,Debugln,Info,Infof,Infoln,Error,Errorf,Errorln $(files)

.PHONY: lint
lint:           ## Run go lint
lint: vendor $(GOLINT)
	$(GOLINT) -set_exit_status $(packages)

vendor:
	@ echo "No vendor dir found. Fetching dependencies now..."
	go get -u github.com/golang/dep/cmd/dep
	GOPATH=$(GOPATH):. $(GODEP) ensure

clean:
	rm -f $(BINARY_NAME) 

help:           ## Show this help
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'
$(GOLINT):
	go get -u github.com/golang/lint/golint
