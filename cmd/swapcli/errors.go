package main

import (
	"fmt"
)

var (
	errNoDuration           = fmt.Errorf("must provide non-zero --duration")
	errRelayerFeeOutOfRange = fmt.Errorf("valid --%s range is from %s to %s ETH",
		flagRelayerFee, minRelayerFee.Text('f'), maxRelayerFee.Text('f'))
)

func errInvalidFlagValue(flagName string, err error) error {
	return fmt.Errorf("invalid value passed to --%s: %w", flagName, err)
}
