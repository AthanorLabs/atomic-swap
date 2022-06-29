#!/bin/bash

arch=linux64

if [[ -d "monero-bin" ]]; then
    echo "$(dirname $(realpath monero-bin)) already installed"
    exit 0
fi

set -e

curl -L "https://downloads.getmonero.org/cli/${arch}" -o monero.tar.bz2
tar xjvf monero.tar.bz2

# Give the architecture and version specific release dir a fixed "monero-bin" symlink
versioned_dir="$(basename "$(tar tjf monero.tar.bz2 | head -1)")"
ln -sf "${versioned_dir}" monero-bin
