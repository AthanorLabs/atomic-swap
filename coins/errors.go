package coins

import (
	"errors"
)

var (
	errNegativePiconeros = errors.New("negative piconero values are not supported")
	errNegativeWei       = errors.New("negative wei values are not supported")
	// ErrNegativeRate is generated when an exchange rate is negative
	ErrNegativeRate = errors.New(`"exchangeRate" can not be negative`)
	// ErrZeroRate is generated when an exchange rate has a zero value
	ErrZeroRate = errors.New(`"exchangeRate" must be non-zero`)
	// ErrInvalidCoin is generated when a ProvidesCoin type has an invalid string
	ErrInvalidCoin = errors.New("invalid ProvidesCoin")
)
