#!/usr/bin/env bash
# Installs abigen into the user's GOPATH bin directory if it is not there already
# or if the existing version does not match go-ethereum version in go.mod.

ABIGEN="$(go env GOPATH)/bin/abigen"
GO_MOD="$(dirname "$(readlink -f "$0")")/../go.mod"

VERSION="$(grep --max-count=1 "github.com/ethereum/go-ethereum" "${GO_MOD}" | cut '--delimiter= ' --fields=2)"
if [[ -z "${VERSION}" ]]; then
	echo "Failed to determine correct abigen version"
	exit 1
fi

if [[ ! -x "${ABIGEN}" ]] || [[ $("${ABIGEN}" --version) != *"${VERSION}-stable"* ]]; then
	go install "github.com/ethereum/go-ethereum/cmd/abigen@${VERSION}"
fi
