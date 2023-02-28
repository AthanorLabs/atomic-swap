package main

import (
	"fmt"
)

var (
	errNoDuration             = fmt.Errorf("must provide non-zero --duration")
	errRelayerFeeTooHigh      = fmt.Errorf("%s must be less than or equal to 1", flagRelayerFee)
	errMustSetRelayerEndpoint = fmt.Errorf("%s must be set if %s is set", flagRelayerEndpoint, flagRelayerFee)
)

func errInvalidFlagValue(flagName string, err error) error {
	return fmt.Errorf("invalid value passed to --%s: %w", flagName, err)
}
