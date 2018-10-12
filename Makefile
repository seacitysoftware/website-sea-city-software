# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
THIS_FILE := $(lastword $(MAKEFILE_LIST))
TARGET_DIR := "./target"

GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

default: dependencies build test

clean:
	rm -rf .clean/ && \
	mkdir -p .clean/ && \
	if [ -d target ]; then mv target/ .clean/; fi && \
	if [ -d bin ]; then mv bin/ .clean/; fi && \
	if [ -d vendor ]; then mv vendor/ .clean/; fi && \
	if [ -d build ]; then mv build/ .clean/; fi

## Build tools ensures that the tools used in the build toolchain are installed and configured
## this should only have to be run once
buildTools:
	go get -u github.com/golang/dep/cmd/dep

tools: buildTools
	go get -u github.com/alecthomas/gometalinter \
	&& gometalinter --install

dependenciesBackend:
	dep ensure

## Update/download dependencies
dependencies: dependenciesBackend

## Runs go fmt on the entire project, excluding the vendor directory
fmt:
	gofmt -w $(GOFMT_FILES)

## Generates mocks used for testing
mocks:
	mockery -dir services/ -all -case underscore

test:
	go test ./...

ensureTargetDirectory:
	if [ ! -d $(TARGET_DIR) ]; then mkdir -p $(TARGET_DIR); fi

buildBackend: ensureTargetDirectory
	go build -o $(TARGET_DIR)/website-sea-city-software

buildBackendLinux: ensureTargetDirectory
	GOOS=linux go build -o $(TARGET_DIR)/website-sea-city-software


## Builds the project
build: buildBackend

##
package:
	docker build . -t website-sea-city-software:latest

.PHONY: buildTools dependenciesBackend buildBackend buildBackendLinux clean test mocks serve build test package