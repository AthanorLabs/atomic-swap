package net

import (
	"errors"
)

var (
	errNilStream         = errors.New("stream is nil")
	errFailedToBootstrap = errors.New("failed to bootstrap to any bootnode")
)
