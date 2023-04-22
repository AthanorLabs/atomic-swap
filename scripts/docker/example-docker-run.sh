#!/usr/bin/env bash
set -ex

CONTAINER_NAME=atomic-stagenet
IMAGE_NAME=atomic-swap
TAG=master

# Setting NETWORK to "host" allows you to run swapcli commands on the local
# host. You can also use "bridge", which requires all swapcli commands to
# be run from inside the container.
NETWORK=host

# Note: We mount one directory above what swapd considers its "data-dir".
DATA_MOUNT_DIR="${HOME}/.atomicswap/docker"

# Pre-create the mounted directory, or docker will create it with root
# as the owner.
mkdir -p "${DATA_MOUNT_DIR}"

docker run --rm -v "${DATA_MOUNT_DIR}:/data" \
	--env SWAPD_ENV=stagenet \
	--env SWAPD_ETH_ENDPOINT="https://rpc.sepolia.org/" \
	--env SWAPD_LOG_LEVEL=debug \
	--network="${NETWORK}" \
	--name="${CONTAINER_NAME}" \
	"${IMAGE_NAME}:${TAG}"
