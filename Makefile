.PHONY: clean build test all
GO_SOURCES = $(shell find . -type f -name '*.go')

all: test build

build: artifactory/heroku/node-function

test:
	go test -v ./...

artifactory/heroku/node-function: buildpack.toml $(GO_SOURCES)
	rm -fR $@ 							&& \
	./ci/package.sh						&& \
	mkdir $@/latest 					&& \
	tar -C $@/latest -xzf $@/*/*.tgz

clean:
	rm -fR artifactory/
	rm -fR dependency-cache/
	rm -fR bin/