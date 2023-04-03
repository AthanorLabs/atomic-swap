package rpc

import (
	"errors"
)

var (
	// net_ errors
	errNoOfferWithID = errors.New("peer does not have offer with given ID")

	// ws errors
	errUnimplemented = errors.New("unimplemented")
	errInvalidMethod = errors.New("invalid method")
)
