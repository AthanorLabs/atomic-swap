package main

import (
	"fmt"
)

//nolint:lll
var (
	errNoDuration                         = fmt.Errorf("must provide non-zero --duration")
	errCannotHaveGreaterThan100Commission = fmt.Errorf("%s must be less than 1", flagRelayerCommission)
	errMustSetRelayerEndpoint             = fmt.Errorf("%s must be set if %s is set", flagRelayerEndpoint, flagRelayerCommission)
)

func errInvalidFlagValue(flagName string, err error) error {
	return fmt.Errorf("invalid value passed to --%s: %w", flagName, err)
}
