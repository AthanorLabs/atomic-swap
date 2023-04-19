// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package backend

import (
	"errors"
)

var (
	errNilSwapContractOrAddress = errors.New("must provide swap contract and address")
)
