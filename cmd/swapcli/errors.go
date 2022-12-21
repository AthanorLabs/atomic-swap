package main

import (
	"fmt"
)

//nolint:lll
var (
	errNoMinAmount                        = fmt.Errorf("must provide non-zero %s", flagMinAmount)
	errNoMaxAmount                        = fmt.Errorf("must provide non-zero %s", flagMaxAmount)
	errNoExchangeRate                     = fmt.Errorf("must provide non-zero %s", flagExchangeRate)
	errNoProvidesAmount                   = fmt.Errorf("must provide non-zero %x", flagProvidesAmount)
	errNoDuration                         = fmt.Errorf("must provide non-zero --duration")
	errCannotHaveNegativeCommission       = fmt.Errorf("%s must be greater than zero", flagRelayerCommission)
	errCannotHaveGreaterThan100Commission = fmt.Errorf("%s must be less than 1", flagRelayerCommission)
	errMustSetRelayerCommission           = fmt.Errorf("%s must be set if %s is set", flagRelayerCommission, flagRelayerEndpoint)
	errMustSetRelayerEndpoint             = fmt.Errorf("%s must be set if %s is set", flagRelayerEndpoint, flagRelayerCommission)
)

func errInvalidFlagValue(flagName string, err error) error {
	return fmt.Errorf("invalid value passed to --%s: %w", flagName, err)
}
