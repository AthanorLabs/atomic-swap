// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package backend

import (
	"errors"
)

var (
	errNilSwapContractOrAddress = errors.New("must provide swap contract and address")
)
