#!/bin/bash

VERSION="v1.47.2"
GOBIN="$(go env GOPATH)/bin"

if [[ ! -x "${GOBIN}/golangci-lint" ]]; then
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${GOBIN}" "${VERSION}"
fi
