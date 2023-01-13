#!/bin/bash
#
# Installs the latest version of monero CLI tools. You can force a reinstall or upgrade by
# deleting the monero-bin symlink or the version specific folder that it links to. This
# script changes directories and should be executed, not sourced.
#

os_name="$(uname)"
arch_name="$(uname -m)"

if [[ "${os_name}" ==  Linux ]] && [[ "${arch_name}" == x86_64 ]]; then
	xmr_arch=linux64
elif [[ "${os_name}" ==  Linux ]] && [[ "${arch_name}" == aarch64 ]]; then
	xmr_arch=linuxarm8
elif [[ "${os_name}" ==  macOS ]] && [[ "${arch_name}" == x86_64 ]]; then
	xmr_arch=mac64
elif [[ "${os_name}" ==  macOS ]] && [[ "${arch_name}" == arm64 ]]; then
	xmr_arch=macarm8
else
	echo "OS=${os_name} on ARCH=${arch_name} is not supported"
	exit 1
fi

PROJECT_ROOT="$(dirname "$(dirname "$(readlink -f "$0")")")"
cd "${PROJECT_ROOT}"

if [[ -d "monero-bin" ]]; then
	echo "$(realpath monero-bin) already installed"
	exit 0
fi

set -e

curl -L "https://downloads.getmonero.org/cli/${xmr_arch}" -o monero.tar.bz2
tar xjvf monero.tar.bz2

# Give the architecture and version specific release dir a fixed "monero-bin" symlink
versioned_dir="$(basename "$(tar tjf monero.tar.bz2 | head -1)")"
ln -sf "${versioned_dir}" monero-bin
