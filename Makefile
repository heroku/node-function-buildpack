.PHONY: clean build test all

PROJECT_ROOT := $(shell pwd)
SHELL=/bin/bash -o pipefail

VERSION := "v$$(cat buildpack.toml | grep -m 1 version | sed -e 's/version = //g' | xargs)"

all: test build

test:
	go test -v -mod vendor -race -coverprofile c.out $(PROJECT_ROOT)/...

build:
	@GOOS=linux go build -o "bin/detector" ./cmd/detector/...
	@GOOS=linux go build -o "bin/builder" ./cmd/builder/...

package: clean build
	@tar cvzf node-function-buildpack-$(VERSION).tgz bin/ buildpack.toml README.md LICENSE

clean:
	@rm -fR artifactory/
	@rm -fR dependency-cache/
	@rm -f node-function-buildpack-$(VERSION).tgz
