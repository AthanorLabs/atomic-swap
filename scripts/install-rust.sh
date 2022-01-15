#!/bin/bash
if ! command -v rustup &> /dev/null
then
	curl https://sh.rustup.rs -sSf | sh
fi
rustup default stable