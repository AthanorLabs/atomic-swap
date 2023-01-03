package swapnet

import (
	"errors"
)

var (
	errNilHandler            = errors.New("handler is nil")
	errNoOngoingSwap         = errors.New("no swap currently happening")
	errSwapAlreadyInProgress = errors.New("already have ongoing swap")
)
