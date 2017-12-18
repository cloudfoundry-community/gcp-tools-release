#! /bin/bash

# This script is run from inside the tile-generator container.
# It installs a few dependencies, then uses the tile generator to
# build the Stackdriver Nozzle tile. Most of the contents have
# been gratuitously stolen from the CI configs + Dockerfile.

## Env, dirs and PATHs
export GOPATH="${PWD}/gopath"
export PATH="${PATH}:${GOROOT}/bin:${GOPATH}/bin"
mkdir -p "${GOPATH}/bin"

## Deps for SD nozzle build / test.

apk update
# musl-dev needed for ld: https://bugs.alpinelinux.org/issues/6628
# coreutils needed for sha256sum and sha1sum
apk add go ruby musl-dev coreutils

go version # Go in Alpine v3.6 is 1.8.4
go get github.com/onsi/ginkgo
go install github.com/onsi/ginkgo/...
go get github.com/golang/lint/golint

## Install Bosh 2 CLI
BOSH2_VERSION="2.0.45"
BOSH2_SHA1="bf04be72daa7da0c9bbeda16fda7fc7b2b8af51e"

if ! test -f "bosh2_${BOSH2_VERSION}_SHA1SUM"; then
  wget -c "https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-${BOSH2_VERSION}-linux-amd64"
  echo "${BOSH2_SHA1}	bosh-cli-${BOSH2_VERSION}-linux-amd64" > "bosh2_${BOSH2_VERSION}_SHA1SUM"
  if ! sha1sum -cw --status "bosh2_${BOSH2_VERSION}_SHA1SUM"; then exit 1; fi
  mv "bosh-cli-${BOSH2_VERSION}-linux-amd64" "gopath/bin/bosh2"
  chmod a+x "gopath/bin/bosh2"
fi

bosh2 --version

## Let's build!

pushd stackdriver-tools
export RELEASE_PATH="dev_releases/stackdriver-tools-custom_${VERSION}.tgz"
echo "${VERSION}" > "src/stackdriver-nozzle/release"
rm -fr .dev_builds/* dev_releases/*
bosh2 sync-blobs
bosh2 create-release --force \
  --name="stackdriver-tools" \
  --version "${VERSION}" \
  --tarball="${RELEASE_PATH}"

export TILE_NAME="stackdriver-nozzle-custom"
export TILE_LABEL="Stackdriver Nozzle (custom build)"
erb tile.yml.erb > tile.yml
tile build "${VERSION}"
TILE="product/${TILE_NAME}-${VERSION}.pivotal"
sha256sum "${PWD}/${TILE}"
