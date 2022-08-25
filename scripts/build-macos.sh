#!/bin/bash

export GOOS=darwin
export GOARCH=amd64

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

echo "building swapd..."
cd cmd/daemon || exit 1
if ! go build -o swapd-amd64-darwin; then
	exit 1
fi
mv swapd ../..
echo "done building swapd."

echo "building swapcli..."
cd ../client || exit 1
if ! go build -o swapcli-amd64-darwin; then
	exit 1
fi
mv swapcli ../..
echo "done building swapcli."

if [[ -z "${ALL}" ]]; then
	exit 0
fi

echo "build swaprecover..."
cd ../recover || exit 1
if ! go build -o swaprecover-amd64-darwin; then
	exit 1
fi
mv swaprecover ../.. || exit 1
echo "done building swaprecover."

echo "building swaptester..."
cd ../tester || exit 1
if ! go build -o swaptester-amd64-darwin; then
	exit 1
fi
mv swaptester ../..
echo "done building swaptester."
