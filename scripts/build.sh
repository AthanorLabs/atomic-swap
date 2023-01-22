#!/usr/bin/env bash

PROJECT_ROOT="$(dirname "$(dirname "$(realpath "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

# Make sure no one is souring the script before we change GOBIN
if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then
	echo "Execute ${BASH_SOURCE[0]} instead of souring it"
	return
fi

# Fail script on any error
set -e

mkdir -p bin
export GOBIN="${PROJECT_ROOT}/bin"

if [[ -n "${ALL}" ]]; then
	go install -tags=prod ./cmd/...
else
	go install -tags=prod ./cmd/swapd ./cmd/swapcli
fi
