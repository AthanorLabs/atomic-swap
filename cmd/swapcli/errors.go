// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package main

import (
	"fmt"
)

var (
	errNoDuration = fmt.Errorf("must provide non-zero --duration")
)

func errInvalidFlagValue(flagName string, err error) error {
	return fmt.Errorf("invalid value passed to --%s: %w", flagName, err)
}
