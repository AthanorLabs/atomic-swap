// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/common/vjson"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

// RelayerQueryResponse is sent from a relayer to the opener of
// a /relayerquery/0 stream.
type RelayerQueryResponse struct {
	Address ethcommon.Address `json:"address" validate:"required"`
}

// String converts the RelayerQueryResponse to a string usable for debugging purposes
func (m *RelayerQueryResponse) String() string {
	return fmt.Sprintf("RelayerQueryResponse=%#v", m)
}

// Encode implements the Encode() method of the common.Message interface which
// prepends a message type byte before the message's JSON encoding.
func (m *RelayerQueryResponse) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{RelayerQueryResponseType}, b...), nil
}

// Type implements the Type() method of the common.Message interface
func (m *RelayerQueryResponse) Type() byte {
	return RelayerQueryResponseType
}

// RelayClaimRequest implements common.Message for our p2p relay claim requests.
type RelayClaimRequest struct {
	// OfferID is non-nil, if the request is from a maker to the taker of an
	// active swap. It is nil, if the request is being sent to a relay node,
	// because it advertised in the DHT.
	OfferID   *types.Hash                     `json:"offerID"`
	RelaySwap *contracts.SwapCreatorRelaySwap `json:"relaySwap" validate:"required"`
	Secret    []byte                          `json:"secret" validate:"required,len=32"`
	Signature []byte                          `json:"signature" validate:"required,len=65"`
}

// String converts the RelayClaimRequest to a string usable for debugging purposes
func (m *RelayClaimRequest) String() string {
	return fmt.Sprintf("RelayClaimResponse=%#v", m)
}

// Encode implements the Encode() method of the common.Message interface which
// prepends a message type byte before the message's JSON encoding.
func (m *RelayClaimRequest) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{RelayClaimRequestType}, b...), nil
}

// Type implements the Type() method of the common.Message interface
func (m *RelayClaimRequest) Type() byte {
	return RelayClaimRequestType
}

// RelayClaimResponse implements common.Message for our p2p relay claim responses
type RelayClaimResponse struct {
	TxHash ethcommon.Hash `json:"transactionHash" validate:"required"`
}

// String converts the RelayClaimRequest to a string usable for debugging purposes
func (m *RelayClaimResponse) String() string {
	return fmt.Sprintf("RelayClaimResponse=%#v", m)
}

// Encode implements the Encode() method of the common.Message interface which
// prepends a message type byte before the message's JSON encoding.
func (m *RelayClaimResponse) Encode() ([]byte, error) {
	b, err := vjson.MarshalStruct(m)
	if err != nil {
		return nil, err
	}

	return append([]byte{RelayClaimResponseType}, b...), nil
}

// Type implements the Type() method of the common.Message interface
func (m *RelayClaimResponse) Type() byte {
	return RelayClaimResponseType
}
