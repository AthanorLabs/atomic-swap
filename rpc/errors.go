// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package rpc

import (
	"errors"
)

var (
	// net_ errors
	errNoOfferWithID          = errors.New("peer does not have offer with given ID")
	errUnsupportedForBootnode = errors.New("unsupported for bootnode")

	// ws errors
	errUnimplemented       = errors.New("unimplemented")
	errInvalidMethod       = errors.New("invalid method")
	errNamespaceNotEnabled = errors.New("namespace not enabled")
)
