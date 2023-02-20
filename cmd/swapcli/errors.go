package main

import (
	"fmt"
)

var (
	errNoDuration                  = fmt.Errorf("must provide non-zero --duration")
	errCannotHaveGreaterThan100Fee = fmt.Errorf("%s must be less than 1", flagRelayerFee)
	errMustSetRelayerEndpoint      = fmt.Errorf("%s must be set if %s is set", flagRelayerEndpoint, flagRelayerFee)
)

func errInvalidFlagValue(flagName string, err error) error {
	return fmt.Errorf("invalid value passed to --%s: %w", flagName, err)
}
