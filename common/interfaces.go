// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package common

import (
	"github.com/athanorlabs/atomic-swap/common/types"
)

// Message is implemented by all network messages
type Message interface {
	String() string
	Encode() ([]byte, error)
	Type() byte
}

// SwapState is the interface used by other packages in *xmrtaker.swapState or *xmrmaker.swapState.
type SwapState interface {
	SwapStateNet
	SwapStateRPC
}

// SwapStateNet handles incoming protocol messages for an initiated protocol.
// It is implemented by *xmrtaker.swapState and *xmrmaker.swapState
type SwapStateNet interface {
	HandleProtocolMessage(msg Message) error
	ID() types.Hash
	Exit() error
}

// SwapStateRPC contains the methods used by the RPC server into the SwapState.
type SwapStateRPC interface {
	SendKeysMessage() Message
	ID() types.Hash
	Exit() error
}
