#! /bin/bash
#
# This script builds the current git HEAD into a tile using the
# cfplatformeng/tile-generator image from Docker Hub.
#
# All you should need to do is run scripts/build-custom-tile.sh
# from the root directory of this repo.

export VERSION="0.0.1-custom.$(git rev-parse --short HEAD)"
docker build -t tile --build-arg "VERSION=${VERSION}" .
docker run --rm -v "${PWD}:/mnt" tile cp "/tmp/tile-build/stackdriver-tools/product/stackdriver-nozzle-custom-${VERSION}.pivotal" /mnt
docker rmi tile
