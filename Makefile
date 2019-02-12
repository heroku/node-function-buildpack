.PHONY: clean build test all
GO_SOURCES = $(shell find . -type f -name '*.go')

all: test build

build: artifactory/io/projectriff/node/io.projectriff.node

test:
	go test -v ./...

artifactory/io/projectriff/node/io.projectriff.node: buildpack.toml $(GO_SOURCES)
	rm -fR $@ 							&& \
	./ci/package.sh						&& \
	mkdir $@/latest 					&& \
	tar -C $@/latest -xzf $@/*/*.tgz


clean:
	rm -fR artifactory/
	rm -fR dependency-cache/