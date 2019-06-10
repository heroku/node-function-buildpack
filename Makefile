.PHONY: clean build test all

SHELL=/bin/bash -o pipefail

VERSION := "v$$(cat buildpack.toml | grep -m 1 version | sed -e 's/version = //g' | xargs)"

all: test build

test:
	go test -v ./...

build:
	@GOOS=linux go build -o "bin/detect" ./detect/main.go
	@GOOS=linux go build -o "bin/build" ./build/main.go

package: clean build
	@tar cvzf node-function-buildpack-$(VERSION).tgz bin/ buildpack.toml README.md LICENSE

clean:
	@rm -fR artifactory/
	@rm -fR dependency-cache/
	@rm -fR bin/
	@rm -f node-function-buildpack-$(VERSION).tgz