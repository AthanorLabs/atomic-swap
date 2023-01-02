package host

import (
	"errors"
)

var (
	errNilHandler            = errors.New("handler is nil")
	errNilOfferAPI           = errors.New("discovery.offerAPI is nil")
	errNilStream             = errors.New("stream is nil")
	errFailedToBootstrap     = errors.New("failed to bootstrap to any bootnode")
	errNoOngoingSwap         = errors.New("no swap currently happening")
	errSwapAlreadyInProgress = errors.New("already have ongoing swap")
)
