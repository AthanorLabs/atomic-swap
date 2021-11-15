#!/bin/bash

cd cmd/daemon && go build -o swapd 
mv swapd ../..
cd ../client && go build -o swapcli
mv swapcli ../..
cd ../..