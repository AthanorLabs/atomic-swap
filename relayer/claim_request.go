// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package relayer provides libraries for creating and validating relay requests and responses.
package relayer

import (
	"crypto/ecdsa"
	"fmt"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	logging "github.com/ipfs/go-log"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

var log = logging.Logger("relayer")

// CreateRelayClaimRequest fills and returns a RelayClaimRequest ready for
// submission to a relayer.
func CreateRelayClaimRequest(
	claimerEthKey *ecdsa.PrivateKey,
	relaySwap *contracts.SwapCreatorRelaySwap,
	secret [32]byte,
) (*message.RelayClaimRequest, error) {
	signature, err := createRelayClaimSignature(
		claimerEthKey,
		relaySwap,
	)
	if err != nil {
		return nil, err
	}

	return &message.RelayClaimRequest{
		OfferID:   nil, // set elsewhere if sending to counterparty
		RelaySwap: relaySwap,
		Secret:    secret[:],
		Signature: signature,
	}, nil
}

func createRelayClaimSignature(
	claimerEthKey *ecdsa.PrivateKey,
	relaySwap *contracts.SwapCreatorRelaySwap,
) ([]byte, error) {
	signerAddress := ethcrypto.PubkeyToAddress(claimerEthKey.PublicKey)
	if relaySwap.Swap.Claimer != signerAddress {
		return nil, fmt.Errorf("signing key %s does not match claimer %s", signerAddress, relaySwap.Swap.Claimer)
	}

	// signature format is (r || s || v), v = 27/28
	signature, err := Sign(claimerEthKey, relaySwap.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to sign relay request: %w", err)
	}

	return signature, nil
}

// Sign signs the given digest and returns a 65-byte signature in (r,s,v) format.
func Sign(key *ecdsa.PrivateKey, digest [32]byte) ([]byte, error) {
	sig, err := ethcrypto.Sign(digest[:], key)
	if err != nil {
		return nil, err
	}

	// Ethereum wants 27/28 for v
	sig[64] += 27
	return sig, nil
}
