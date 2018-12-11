build: test
	go build -v ./...

build-all:
	gox -output="out/stackdriver-nozzle_{{.OS}}_{{.Arch}}" -ldflags="-X github.com/cloudfoundry-community/stackdriver-tools/src/stackdriver-nozzle/version.release=`cat release 2>/dev/null`" .

test:
	go test `go list ./... | grep -v github.com/cloudfoundry-community/stackdriver-tools/ci`

integration-test:
	go test `go list ./... | grep github.com/cloudfoundry-community/stackdriver-tools/ci`

lint:
	# Tests for output
	# Disabling gosec for https://github.com/securego/gosec/issues/267
	# Disabling vetshadow for https://github.com/golang/go/issues/19490
	# Disabling maligned because it also affect the config struct. TODO(mattysweeps) re-enable maligned
	# Ignoring missing comments for now TODO(mattysweeps) fix godoc
	[ -z "$$(gometalinter --deadline=300s --disable gosec --disable vetshadow --disable maligned --vendor ./... | grep -v exported)" ]

get-deps:
	# For gometalinter linting
	pushd $(GOPATH) && curl -L https://git.io/vp6lP | sh && popd

	# Simplify cross-compiling
	go get github.com/mitchellh/gox

	# Ginkgo and omega test tools
	go get github.com/onsi/ginkgo/ginkgo
	go get github.com/onsi/gomega

clean:
	go clean ./...

.PHONY: build test integration-test lint get-deps clean
