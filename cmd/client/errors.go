package main

import (
	"errors"
)

var (
	errNoMinAmount      = errors.New("must provide non-zero --min-amount")
	errNoMaxAmount      = errors.New("must provide non-zero --max-amount")
	errNoExchangeRate   = errors.New("must provide non-zero --exchange-rate")
	errNoProvidesAmount = errors.New("must provide non-zero --provides-amount")
	errNoDuration       = errors.New("must provide non-zero --duration")
)
