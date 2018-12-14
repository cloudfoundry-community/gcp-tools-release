.EXPORT_ALL_VARIABLES:
COMMIT_HASH := $(shell git show-ref HEAD | cut -d' ' -f 1)
LATEST_TAG := $(shell git describe --abbrev=0 --tags)
TIMESTAMP := $(shell date +%s)
VERSION ?= $(shell git describe --tags --exact-match `git rev-parse HEAD` 2>/dev/null || echo 0.0.$(TIMESTAMP)-custom.$$(git rev-parse --short HEAD))
TILE_NAME ?= $(shell if [ `echo $(VERSION) | grep -o custom` ]; then echo stackdriver-nozzle-custom; else echo stackdriver-nozzle; fi)
TILE_LABEL ?= $(shell if [ `echo $(VERSION) | grep -o custom` ]; then echo "Stackdriver Nozzle (custom build)"; else echo Stackdriver Nozzle; fi)
TILE_FILENAME := $(TILE_NAME)-$(VERSION).pivotal
TILE_SHA256 := $(TILE_FILENAME).sha256
RELEASE_TARBALL := stackdriver-tools-release-$(VERSION).tar.gz
RELEASE_PATH := $(PWD)/$(RELEASE_TARBALL)

build: test
	go build -v ./...

build-all:
	gox -output="out/stackdriver-nozzle_{{.OS}}_{{.Arch}}" -ldflags="-X github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/version.release=`cat release 2>/dev/null`" .

test:
	go test -v $(shell go list ./... | grep -v github.com/cloudfoundry-community/stackdriver-tools/ci | grep -v gopath)

integration-test:
	go test -v $(shell go list ./... | grep github.com/cloudfoundry-community/stackdriver-tools/ci | grep -v gopath)

lint:
	# Tests for output
	# Disabling gosec for https://github.com/securego/gosec/issues/267
	# Disabling vetshadow for https://github.com/golang/go/issues/19490
	# Disabling maligned because it also affect the config struct. TODO(mattysweeps) re-enable maligned
	# Ignoring missing comments for now TODO(mattysweeps) fix godoc
	[ -z "$$($(GOPATH)/bin/gometalinter --deadline=300s --disable gosec --disable vetshadow --disable maligned --vendor ./... | grep -v exported)" ]

get-deps:
	# For gometalinter linting
	cd $(GOPATH) && wget -q -O - https://git.io/vp6lP | sh

	# Simplify cross-compiling
	go get github.com/mitchellh/gox

	# Ginkgo and omega test tools
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega

bosh-release:
	echo $(VERSION) > src/stackdriver-nozzle/release
	bosh sync-blobs
	bosh create-release --name=stackdriver-tools --version=$(VERSION) --tarball=$(RELEASE_TARBALL) --force --sha2

tile: bosh-release
	erb tile.yml.erb > tile.yml
	tile build $(VERSION)
	mv product/$(TILE_FILENAME) $(TILE_FILENAME)
	echo -n $((sha256sum $(TILE_FILENAME) | cut -d' ' -f 1)) > $(TILE_SHA256)

clean:
	go clean ./...
	rm -f stackdriver-tools-release-*
	rm -f stackdriver-nozzle*.pivotal
	rm -f stackdriver-nozzle*.pivotal.sha256
	rm -f tile.yml tile-history.yml

.PHONY: build test integration-test lint get-deps clean tile bosh-release