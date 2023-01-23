#!/usr/bin/env bash
# Installs golangci-lint (https://golangci-lint.run) into the user's personal GOPATH bin directory if it is
# not there already or if the version does not match the value defined below.

VERSION="v1.48.0"
GOBIN="$(go env GOPATH)/bin"
LINT="${GOBIN}/golangci-lint"

if [[ ! -x "${LINT}" ]] || [[ "v$("${LINT}" version --format=short)" != "${VERSION}" ]]; then
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${GOBIN}" "${VERSION}"
fi
