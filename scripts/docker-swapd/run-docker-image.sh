#!/usr/bin/env bash
set -e

# SWAPD_ENV/SWAPD_ETH_ENDPOINT are only set if not already set. See further down
# for all the SWAPD_* environment variables that can be set for swapd.
SWAPD_ENV="${SWAPD_ENV:-"stagenet"}"
SWAPD_ETH_ENDPOINT="${SWAPD_ETH_ENDPOINT:-"https://rpc.sepolia.org/"}"

# You can only run one container with the same name at the same time. Having
# docker run fail because a same-named container already exists is good, as both
# containers need to have a distinct mount dir.
CONTAINER_NAME="${CONTAINER_NAME:-"swapd-${SWAPD_ENV}"}"
IMAGE_NAME="atomic-swapd"
VERSION="latest" # image tag

# We mount one directory above what swapd considers its "data-dir". Data
# files will be created in ${DATA_MOUNT_DIR}/${SWAPD_ENV}.
DATA_MOUNT_DIR="${HOME}/.atomicswap/docker"

# Setting NETWORK to "host" allows you to run swapcli commands on the local
# host. You can also use "bridge", which requires all swapcli commands to
# be run from inside the container.
NETWORK="host"

env_args=()

add_env_arg() {
	local env_name=$1
	local env_value=${!env_name}

	# Add --env flag argument if the variable is defined and non-empty
	if [[ -n ${env_value} ]]; then
		env_args+=(--env "${env_name}=${env_value}")
	fi
}

add_env_arg SWAPD_ENV
add_env_arg SWAPD_ETH_ENDPOINT
add_env_arg SWAPD_RPC_PORT
add_env_arg SWAPD_LIBP2P_PORT
add_env_arg SWAPD_MONEROD_HOST
add_env_arg SWAPD_MONEROD_PORT
add_env_arg SWAPD_ETH_PRIVKEY
add_env_arg SWAPD_BOOTNODES
add_env_arg SWAPD_LOG_LEVEL

# Pre-create the mounted dir, or docker creates it with root as the owner.
mkdir -p "${DATA_MOUNT_DIR}"

# turn on command echo
set -x

exec docker run --rm -v "${DATA_MOUNT_DIR}:/data" "${env_args[@]}" \
	--network="${NETWORK}" \
	--name="${CONTAINER_NAME}" \
	"${IMAGE_NAME}:${VERSION}"
