#!/bin/bash

cd cmd/daemon 
if ! go build -o swapd ; then
	exit 1
fi
mv swapd ../..

cd ../client 
if ! go build -o swapcli ; then 
	exit 1
fi
mv swapcli ../..

cd ../recover
if ! go build -o swaprecover ; then 
	exit 1
fi
mv swaprecover ../..

cd ../tester
if ! go build -o swaptester ; then 
	exit 1
fi
mv swaptester ../..
cd ../..