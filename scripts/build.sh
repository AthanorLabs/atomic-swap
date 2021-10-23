#!/bin/bash

cd cmd && go build -o atomic-swap 
mv atomic-swap ..
cd ..