#!/bin/bash

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

if [[ -n "${ALL}" ]]; then
	GOBIN="${PROJECT_ROOT}" go install ./cmd/...
else
	GOBIN="${PROJECT_ROOT}" go install ./cmd/swapd ./cmd/swapcli
fi
