#!/usr/bin/env bash

set -e

source stackdriver-tools/ci/tasks/utils.sh
source /etc/profile.d/chruby-with-ruby-2.1.2.sh

cpi_release_name="stackdriver-tools"
semver=`cat version-semver/number`

image_name=${cpi_release_name}-${semver}.tgz
image_path="https://storage.googleapis.com/bosh-gcp/beta/stackdriver-tools/${image_name}"

echo "${image_path}"
echo "${image_path}" > canidate/latest-nozzle-tile.pivotal
