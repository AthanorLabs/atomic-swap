// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package protocol

import (
	"errors"
)

var (
	// ErrLogNotForUs is returned when a log is found that doesn't have the given contract swap ID.
	ErrLogNotForUs = errors.New("found log that isn't for our swap")

	errLogMissingParams    = errors.New("log didn't have enough topics")
	errInvalidEventTopic   = errors.New("log did not have correct event as first topic")
	errInvalidSecp256k1Key = errors.New("secp256k1 public key resulting from proof verification does not match key sent")
	errInvalidEd25519Key   = errors.New("ed25519 public key resulting from proof verification does not match key sent")
)
