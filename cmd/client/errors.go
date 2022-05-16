package main

import (
	"errors"
)

var (
	errNoMultiaddr      = errors.New("must provide peer's multiaddress with --multiaddr")
	errNoMinAmount      = errors.New("must provide non-zero --min-amount")
	errNoMaxAmount      = errors.New("must provide non-zero --max-amount")
	errNoExchangeRate   = errors.New("must provide non-zero --exchange-rate")
	errNoOfferID        = errors.New("must provide --offer-id")
	errNoProvidesAmount = errors.New("must provide --provides-amount")
)
