#!/usr/bin/env bash
set -e

IMAGE_NAME=atomic-swap

# VERSION can be "latest", a release tag, a hash or a branch name that does not contain slashes.
# It must exist on github, local changes are not visible inside the container.
VERSION=fde1cad9bf

# Run docker build from the directory of this script
cd "$(dirname "$0")"

docker build \
	--build-arg "VERSION=${VERSION}" \
	--build-arg "USER_UID=$(id -u)" \
	--build-arg "USER_GID=$(id -g)" \
	. -t "${IMAGE_NAME}:${VERSION}"

echo "built ${IMAGE_NAME}:${VERSION}"
