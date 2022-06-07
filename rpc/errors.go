package rpc

import (
	"errors"
)

var (
	// net_ errors
	errNoOfferWithID       = errors.New("peer does not have offer with given ID")
	errFailedToGetSwapInfo = errors.New("failed to get swap info after initiating")

	// swap_ errors
	errNoSwapWithID  = errors.New("unable to find swap with given ID")
	errNoOngoingSwap = errors.New("no current ongoing swap")
	errCannotRefund  = errors.New("cannot refund if not the ETH provider")

	// ws errors
	errUnimplemented     = errors.New("unimplemented")
	errInvalidMethod     = errors.New("invalid method")
	errSignerNotRequired = errors.New("signer not required")
)
