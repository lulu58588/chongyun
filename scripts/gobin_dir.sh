#!/bin/sh
# Where `go install` puts the binary: GOBIN when it is set, else GOPATH/bin.
set -e

gobin=$(go env GOBIN 2>/dev/null || true)
if [ -z "$gobin" ]; then
	gobin="$(go env GOPATH)/bin"
fi
echo "$gobin"
