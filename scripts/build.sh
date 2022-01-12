#!/bin/bash

cd cmd/daemon && go build -o swapd 
mv swapd ../..
cd ../client && go build -o swapcli
mv swapcli ../..
cd ../recover && go build -o swaprecover
mv swaprecover ../..
cd ../..