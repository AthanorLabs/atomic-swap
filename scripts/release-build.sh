#!/usr/bin/env bash

# Fail script on any error (do not change)
set -e

PROJECT_ROOT="$(dirname "$(dirname "$(realpath "$0")")")"
cd "${PROJECT_ROOT}"

# Make sure no one is sourcing the script, as we export variables
if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then
	echo "Execute ${BASH_SOURCE[0]} instead of souring it"
	exit 1
fi

version="HEAD" # use "latest" for most recent tagged release
install_targets=(
	"github.com/athanorlabs/atomic-swap/cmd/swapd@${version}"
	"github.com/athanorlabs/atomic-swap/cmd/swapcli@${version}"
	"github.com/athanorlabs/atomic-swap/cmd/bootnode@${version}"
)

# turn on echo
set -x

dest_dir=release-bin
rm -rf "${dest_dir}"
mkdir "${dest_dir}"

# Note: We don't bother with static builds (larger binaries) as swapd depends on
# a local monero-wallet-rpc binary and all releases of monero-wallet-rpc depend
# on glibc.
unset CGO_ENABLED

# Unfortunately, GOBIN can't be set when doing cross platform builds and
# go install doesn't take a -o flag:
# https://github.com/golang/go/issues/57485
# We are inside a go module project right now and we'll confuse tooling
# if we put the GOPATH inside of the project. We are using "go install",
# so nothing will go wrong even if a go.mod exists at the top of /tmp.
build_dir="$(mktemp -d /tmp/release-build-XXXXXXXXXX)"

for os in linux darwin; do
	for arch in amd64 arm64; do
		GOPATH="${build_dir}" GOOS="${os}" GOARCH="${arch}" \
			go install -tags=prod "${install_targets[@]}"
		from_dir="${build_dir}/bin/${os}_${arch}"
		to_dir="${dest_dir}/${os/darwin/macos}-${arch/amd64/x64}"
		if [[ -d "${from_dir}" ]]; then
			# non-native binaries
			mv "${from_dir}" "${to_dir}"
		else
			# native binaries
			mkdir "${to_dir}"
			mv "${build_dir}/bin/"* "${to_dir}"
		fi
	done
done

chmod -R u+w "${build_dir}"
rm -rf "${build_dir}"
