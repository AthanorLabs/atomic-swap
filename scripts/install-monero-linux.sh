#!/bin/bash

if [[ -d "./monero-x86_64-linux-gnu-v0.17.3.2" ]]; then 
    echo "monero-x86_64-linux-gnu-v0.17.3.2 already installed"
    exit 0
fi

curl -L https://downloads.getmonero.org/cli/linux64 > monero.tar.bz2
tar xjvf monero.tar.bz2
