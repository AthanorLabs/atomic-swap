#!/bin/bash
set -e

cmd="$(basename "${1}")"

#
# If we are running the bootnode, we ensure that the atomic user that runs swapd
# has access the directories where the data is written.
#
if [[ "${cmd}" == 'bootnode' ]]; then

	if [[ "${*}:1}" =~ '--data-dir' ]]; then
		echo "Dockerized bootnodes should not set the --data-dir flag."
		echo "Adjust where your container mounts /data instead."
		exit 1
	fi

	# ensure the volume files are owned by the atomic user
	chown -R atomic.atomic "/data"
fi

# Run bootnode and swapcli commands as the atomic user for reduced
# privileges.
if [[ "${cmd}" == 'bootnode' || "${cmd}" == 'swapcli' ]]; then
	exec gosu atomic "$@"
fi

exec "$@"
