package rpc

import (
	"errors"
)

var (
	// net_ errors
	errNoOfferWithID           = errors.New("peer does not have offer with given ID")
	errFailedToGetSwapInfo     = errors.New("failed to get swap info after initiating")
	errEthAssetIncorrectFormat = errors.New("ethAsset must be formatted as an address")

	// swap_ errors
	errCannotRefund = errors.New("cannot refund if not the ETH provider")

	// ws errors
	errUnimplemented = errors.New("unimplemented")
	errInvalidMethod = errors.New("invalid method")
)
