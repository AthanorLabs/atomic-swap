#!/bin/bash
set -e

cmd="$(basename "${1}")"

#
# If we are running swapd and SWAPD_ENV is set, so this script
# knows where swapd will be writing data, we ensure that the
# atomic user that runs swapd has access the directories where
# the data is written.
#
if [[ "${cmd}" == 'swapd' ]] && [[ -n "${SWAPD_ENV}" ]]; then

	if ! [[ "${SWAPD_ENV}" =~ ^dev|stagenet|mainnet$ ]]; then
		echo "invalid SWAPD_ENV value"
		exit 1
	fi

	if [[ "${*}:1}" =~ '--data-dir' ]]; then
		echo "Setting --data-dir is not recommended for dockerized swapd."
		echo "If required, unset SWAPD_ENV or override the entrypoint."
		exit 1
	fi

	data_dir="/data/${SWAPD_ENV}"

	# create the directory if it does not exist
	if [[ ! -d "${data_dir}" ]]; then
		mkdir --mode=700 "${data_dir}"
	fi

	# ensure the files are owned by the atomic user
	chown -R atomic.atomic "${data_dir}"
fi

# Run swapd and swapcli commands as the atomic user for reduced
# privileges.
if [[ "${cmd}" == 'swapd' || "${cmd}" == 'swapcli' ]]; then
	exec gosu atomic "$@"
fi

exec "$@"
