.EXPORT_ALL_VARIABLES:
TIMESTAMP := $(shell date +%s)

# Git
# tags are expecetd to be semver versions (example v1.2.3)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_TAG ?= $(shell git describe --tags --exact-match `git rev-parse HEAD` 2>/dev/null || echo custom)
COMMIT_HASH := $(shell git rev-parse HEAD)

# create a pseudo version based on the timestamp for easy dev releases
VERSION ?= $(shell if [ "custom" = "$(GIT_TAG)" ]; then echo 0.0.$(TIMESTAMP)-custom.$(COMMIT_HASH); else echo $(GIT_TAG) | egrep -o '[0-9]+\.[0-9]+\.[]0-9]+'; fi)

# BOSH release
RELEASE_TARBALL := stackdriver-tools-release-$(VERSION).tar.gz
RELEASE_SHA256 := $(RELEASE_TARBALL).sha256
RELEASE_BUILD_DIR := build
RELEASE_PATH := $(RELEASE_BUILD_DIR)/$(RELEASE_TARBALL)

# Tile
CUSTOM_TILE_QUALIFIER ?= $(GIT_BRANCH)
TILE_VERSION := $(shell if [ "custom" = "$(GIT_TAG)" ]; then echo custom; fi)
TILE_NAME ?= $(shell if [ `echo $(VERSION) | grep -o custom` ]; then echo stackdriver-nozzle-$(CUSTOM_TILE_QUALIFIER); else echo stackdriver-nozzle; fi)
TILE_LABEL ?= $(shell if [ `echo $(VERSION) | grep -o custom` ]; then echo "Stackdriver Nozzle $(CUSTOM_TILE_QUALIFIER)"; else echo Stackdriver Nozzle; fi)
TILE_FILENAME := $(TILE_NAME)-$(VERSION).pivotal
TILE_SHA256 := $(TILE_FILENAME).sha256
TILE_BUILD_DIR := product

# linting
METAGOLINTER_COMMIT := 102ac984005d45456a7e3ae6dc94ebcd95c2bb19
METALINTER_TAG := v2.0.12

build:
	go build -v ./...

test:
	go test -v ./...

lint:
	# Disabling gosec for https://github.com/securego/gosec/issues/267
	# Disabling vetshadow for https://github.com/golang/go/issues/19490
	# Disabling maligned because it also affect the config struct. https://github.com/cloudfoundry-community/stackdriver-tools/issues/244
	# Excluding stackdriver-nozzle's mocks directory. https://github.com/cloudfoundry-community/stackdriver-tools/issues/245
	$(GOPATH)/bin/gometalinter \
	--deadline=300s \
	--disable gosec \
	--disable vetshadow \
	--disable maligned \
	--vendor \
	--exclude src/stackdriver-nozzle/mocks \
	./...

get-deps:
	# For gometalinter linting
	cd $(GOPATH) && wget -q -O - https://raw.githubusercontent.com/alecthomas/gometalinter/$(METAGOLINTER_COMMIT)/scripts/install.sh | sh -s -- $(METALINTER_TAG)

	# Simplify cross-compiling
	go get github.com/mitchellh/gox

	# Ginkgo and omega test tools
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega

bosh-release:
	echo $(VERSION) > src/stackdriver-nozzle/release
	bosh sync-blobs
	bosh create-release --name=stackdriver-tools --version=$(VERSION) --tarball=$(RELEASE_TARBALL) --force --sha2
	mkdir -p $(RELEASE_BUILD_DIR)
	mv $(RELEASE_TARBALL) $(RELEASE_BUILD_DIR)/$(RELEASE_TARBALL)
	sha256sum $(RELEASE_BUILD_DIR)/$(RELEASE_TARBALL) | cut -d' ' -f 1 > $(RELEASE_BUILD_DIR)/$(RELEASE_SHA256)

tile: bosh-release
	erb tile.yml.erb > tile.yml
	mkdir -p $(TILE_BUILD_DIR)
	echo $(RELEASE_PATH)
	tile build $(VERSION)
	sha256sum $(TILE_BUILD_DIR)/$(TILE_FILENAME) | cut -d' ' -f 1 > $(TILE_BUILD_DIR)/$(TILE_SHA256)

clean:
	go clean ./...
	rm -f tile.yml tile-history.yml
	rm -fr blobs dev_releases product build release .blobs .dev_builds

.PHONY: build test integration-test lint get-deps clean tile bosh-release
