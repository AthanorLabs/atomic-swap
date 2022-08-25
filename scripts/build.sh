#!/bin/bash

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}" || exit 1

echo "building swapd..."
cd cmd/daemon || exit 1
if ! go build -o swapd; then
	exit 1
fi
mv swapd ../.. || exit 1
echo "done building swapd."

echo "building swapcli..."
cd ../client || exit 1
if ! go build -o swapcli; then
	exit 1
fi
mv swapcli ../..
echo "done building swapcli."

if [[ -z "${ALL}" ]]; then
	exit 0
fi

echo "build swaprecover..."
cd ../recover || exit 1
if ! go build -o swaprecover; then
	exit 1
fi
mv swaprecover ../..
echo "done building swaprecover."

echo "building swaptester..."
cd ../tester || exit 1
if ! go build -o swaptester; then
	exit 1
fi
mv swaptester ../..
cd ../..
echo "done building swaptester."
