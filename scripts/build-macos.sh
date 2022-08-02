#!/bin/bash

export GOOS=darwin 
export GOARCH=amd64

echo "building swapd..."
cd cmd/daemon 
if ! go build -o swapd-amd64-darwin ; then
	exit 1
fi
mv swapd ../..
echo "done building swapd."

echo "building swapcli..."
cd ../client 
if ! go build -o swapcli-amd64-darwin ; then 
	exit 1
fi
mv swapcli ../..
echo "done building swapcli."

if [[ -z "${ALL}" ]]; then
  exit 0
fi

echo "build swaprecover..."
cd ../recover
if ! go build -o swaprecover-amd64-darwin ; then 
	exit 1
fi
mv swaprecover ../..
echo "done building swaprecover."

echo "building swaptester..."
cd ../tester
if ! go build -o swaptester-amd64-darwin ; then 
	exit 1
fi
mv swaptester ../..
cd ../..
echo "done building swaptester."
