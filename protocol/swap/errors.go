package swap

import (
	"errors"
)

var (
	errHaveOngoingSwap = errors.New("already have ongoing swap")
)
