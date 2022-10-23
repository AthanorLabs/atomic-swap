package main

import (
	"errors"
)

var (
	errNoMinAmount                        = errors.New("must provide non-zero --min-amount")
	errNoMaxAmount                        = errors.New("must provide non-zero --max-amount")
	errNoExchangeRate                     = errors.New("must provide non-zero --exchange-rate")
	errNoProvidesAmount                   = errors.New("must provide non-zero --provides-amount")
	errNoDuration                         = errors.New("must provide non-zero --duration")
	errCannotHaveNegativeCommission       = errors.New("relayer-commission must be greater than zero")
	errCannotHaveGreaterThan100Commission = errors.New("relayer-commission must be less than 1")
	errMustSetRelayerCommission           = errors.New("relayer-commission must be set if relayer-endpoint is set")
	errMustSetRelayerEndpoint             = errors.New("relayer-endpoint must be set if relayer-commission is set")
)
