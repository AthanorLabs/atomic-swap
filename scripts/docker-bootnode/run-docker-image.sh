#!/usr/bin/env bash
set -e

# SWAPD_ENV is only set if not already set. See further down for all the SWAPD_*
# environment variables that can be set for the bootnode.
SWAPD_ENV="${SWAPD_ENV:-"mainnet"}"

# You can only run one container with the same name at the same time. Having
# docker run fail because a same-named container already exists is good, as both
# containers need to have a distinct mount dir.
CONTAINER_NAME="${CONTAINER_NAME:-"bootnode-${SWAPD_ENV}"}"
IMAGE_NAME="atomic-bootnode"
VERSION="$(git describe --abbrev=0 --tags)" # image tag

# Pre-create the mounted directory, or docker will create it with root as the
# owner. We mount one directory above what swapd considers its "data-dir". Data
# files will be created in ${DATA_MOUNT_DIR}/${SWAPD_ENV}.
DATA_MOUNT_DIR="${HOME}/.atomicswap/bootnode/docker"

env_args=()

# Add a --env flag to our array of passed environment flags for the container if
# the variable name has a non-empty value assigned to it.
add_env_arg() {
	local env_name=$1
	local env_value=${!env_name}

	if [[ -n ${env_value} ]]; then
		env_args+=(--env "${env_name}=${env_value}")
	fi
}

add_env_arg SWAPD_ENV
add_env_arg SWAPD_RPC_PORT
add_env_arg SWAPD_LIBP2P_PORT
add_env_arg SWAPD_BOOTNODES
add_env_arg SWAPD_LOG_LEVEL

# Pre-create the mounted dir, or docker creates it with root as the owner.
mkdir -p "${DATA_MOUNT_DIR}"

# turn on command echo
set -x

# Bootnodes should use host networking if they use docker. The container should
# also run on a host where the public IP is directly on an interface of the
# parent VM. This allows a bootnode to instantly know its libp2p IP and port on
# start, without depending on multiple other nodes for IP address discovery.
# With the flags below, the container will automatically restart on reboot.
exec docker run --restart=unless-stopped --detach \
	-v "${DATA_MOUNT_DIR}:/data" \
	"${env_args[@]}" \
	--network=host \
	--name="${CONTAINER_NAME}" \
	"${IMAGE_NAME}:${VERSION}"
