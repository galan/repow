-include .env

VERSION ?= master
PROJECTNAME := $(shell basename "$(PWD)")

# Use linker flags to provide version/build settings
LDFLAGS=-ldflags "-X=main.VersionLdFlag=$(VERSION)"

GOCMD=go
GOBUILD=$(GOCMD) build $(LDFLAGS)
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

GO_MAIN=cmd/app/main.go
BINARY_PATH=bin
BINARY_NAME=repow
BINARY_LINUX_AMD=$(BINARY_NAME)-linux-amd64
BINARY_LINUX_ARM=$(BINARY_NAME)-linux-arm64
BINARY_DARWIN_AMD=$(BINARY_NAME)-darwin-amd64
BINARY_DARWIN_ARM=$(BINARY_NAME)-darwin-arm64
BINARY_WINDOWS_AMD=$(BINARY_NAME)-windows-amd64

DOCKER_IMAGE_NAME=repow

## install: Install missing dependencies. Runs `go get` internally. e.g; make install get=github.com/foo/bar
#install: go-get

all: test build

clean:
	@$(GOCLEAN)
	@rm -rf $(BINARY_PATH)/

test:
	$(GOTEST) -v ./...

build: build-linux-amd build-linux-arm build-osx-amd build-osx-arm build-windows-amd
build-linux-amd:
	@mkdir -p $(BINARY_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_PATH)/$(BINARY_LINUX_AMD) $(GO_MAIN)
build-linux-arm:
	@mkdir -p $(BINARY_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) -v -o $(BINARY_PATH)/$(BINARY_LINUX_ARM) $(GO_MAIN)
build-osx-amd:
	@mkdir -p $(BINARY_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_PATH)/$(BINARY_DARWIN_AMD) $(GO_MAIN)
build-osx-arm:
	@mkdir -p $(BINARY_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) -v -o $(BINARY_PATH)/$(BINARY_DARWIN_ARM) $(GO_MAIN)
build-windows-amd:
	@mkdir -p $(BINARY_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -v -o $(BINARY_PATH)/$(BINARY_WINDOWS_AMD) $(GO_MAIN)
docker-build:
	docker build . -t $(DOCKER_IMAGE_NAME):$(VERSION)
docker-run:
	docker run -it --rm -p 8080:8080 --name repow $(DOCKER_IMAGE_NAME):$(VERSION)






# potential make alternative: https://taskfile.dev
# cons: requires binary to be downloaded/installed

# potental dockerfile alternative: https://buildpacks.io/
