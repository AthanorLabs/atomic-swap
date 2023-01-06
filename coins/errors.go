package coins

import (
	"errors"
)

var (
	errNegativePiconeros = errors.New("negative piconero values are not supported")
	errNegativeWei       = errors.New("negative wei values are not supported")
	errNegativeRate      = errors.New("negative exchange rates are not supported")
)
