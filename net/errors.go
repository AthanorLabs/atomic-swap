package net

import (
	"errors"
)

var (
	errNilStream             = errors.New("stream is nil")
	errFailedToBootstrap     = errors.New("failed to bootstrap to any bootnode")
	errNoOngoingSwap         = errors.New("no swap currently happening")
	errSwapAlreadyInProgress = errors.New("already have ongoing swap")
	errInvalidBufferLength   = errors.New("buffer has length 0")
)
