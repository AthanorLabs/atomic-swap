#!/usr/bin/env bash

LICENSE_HEADER="// Copyright 2023 Athanor Labs (ON)\n// SPDX-License-Identifier: LGPL-3.0-only\n"

find "." -type f -name "*.go" -print0 | while read -r -d $'\0' file; do
	echo -e "$LICENSE_HEADER" | cat - "$file" >temp && mv temp "$file"
done
