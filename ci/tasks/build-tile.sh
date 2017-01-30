#!/usr/bin/env bash

set -e

source stackdriver-tools/ci/tasks/utils.sh
source /etc/profile.d/chruby-with-ruby-2.1.2.sh

cpi_release_name="stackdriver-tools"
semver=`cat version-semver/number`

image_name=${cpi_release_name}-${semver}.tgz
image_path="https://storage.googleapis.com/bosh-gcp/beta/stackdriver-tools/${image_name}"
output_path=candidate/latest-nozzle-tile.pivotal

pushd "stackdriver-tools"
	echo "building tile"
	tile build ${semver}
popd

echo "${image_path}"

echo "exposing tile"
mv stackdriver-tools/product/*.pivotal ${output_path}
