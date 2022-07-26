#!/bin/bash

GOBIN="$(go env GOPATH)/bin"

if [[ ! -x "${GOBIN}/golangci-lint" ]]; then
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${GOBIN}" v1.47.2
fi
