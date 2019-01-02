#!/usr/bin/env bash
set -e

# Create a workspace for a GOPATH
gopath_prefix=/tmp/src/github.com/cloudfoundry-community
mkdir -p ${gopath_prefix}

# Link to the source repo
rm -f ${gopath_prefix}/stackdriver-tools
ln -s ${PWD} ${gopath_prefix}/stackdriver-tools

# Configure GOPATH
export GOPATH=/tmp
export PATH=${GOPATH}/bin:$PATH

# Run tests
cd ${gopath_prefix}/stackdriver-tools/

"$@"
