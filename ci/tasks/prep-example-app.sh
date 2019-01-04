#!/usr/bin/env sh

set -e

apk add git
cp -R stackdriver-tools-source/* prepped_source/
echo "${GOOGLE_APPLICATION_CREDENTIALS}" > prepped_source/examples/cf-stackdriver-example/credentials.json
cd stackdriver-tools-source
go build
cat <<EOF > ../prepped_source/examples/cf-stackdriver-example/source-context.json

{
  "git": {
    "revisionId": "$(git rev-parse HEAD)",
    "url": "${STACKDRIVER_TOOLS_SOURCE_URI}"
  }
}
EOF
