// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

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
