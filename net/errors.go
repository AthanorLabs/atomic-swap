// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package net

import (
	"errors"
)

var (
	errBootnodeCannotRelay   = errors.New("bootnode cannot be a relayer")
	errNilHandler            = errors.New("handler is nil")
	errNoOngoingSwap         = errors.New("no swap currently happening")
	errSwapAlreadyInProgress = errors.New("already have ongoing swap")
)
