# Use scripts/build-custom-tile.sh to build this Dockerfile!

FROM cfplatformeng/tile-generator:latest
MAINTAINER Alex Bramley <abramley@google.com>

# Hoop jumping so that VERSION is only set in one place.
ARG VERSION
ENV VERSION=$VERSION

WORKDIR /tmp/tile-build
COPY . stackdriver-tools/
RUN stackdriver-tools/scripts/build-custom-tile-docker.sh
