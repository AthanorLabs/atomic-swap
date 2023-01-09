package coins

import (
	"errors"
)

var (
	errNegativePiconeros = errors.New("negative piconero values are not supported")
	errNegativeWei       = errors.New("negative wei values are not supported")
	// ErrInvalidCoin is generated when a ProvidesCoin type has an invalid string
	ErrInvalidCoin = errors.New("invalid ProvidesCoin")
)
