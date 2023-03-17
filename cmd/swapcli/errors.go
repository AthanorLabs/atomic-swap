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
