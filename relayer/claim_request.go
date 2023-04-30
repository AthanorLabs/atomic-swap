// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

// Package relayer provides libraries for creating and validating relay requests and responses.
package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"

	"github.com/athanorlabs/atomic-swap/coins"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	rcommon "github.com/athanorlabs/go-relayer/common"
)

// FeeWei and FeeEth are the fixed 0.009 ETH fee for using a swap relayer to claim.
var (
	FeeWei = big.NewInt(9e15)
	FeeEth = coins.NewWeiAmount(FeeWei).AsEther()
)

var log = logging.Logger("relayer")

// CreateRelayClaimRequest fills and returns a RelayClaimRequest ready for
// submission to a relayer.
func CreateRelayClaimRequest(
	ctx context.Context,
	claimerEthKey *ecdsa.PrivateKey,
	ec *ethclient.Client,
	relaySwap *contracts.SwapCreatorRelaySwap,
	secret [32]byte,
) (*message.RelayClaimRequest, error) {
	signature, err := createRelayClaimSignature(
		ctx,
		claimerEthKey,
		ec,
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
	ctx context.Context,
	claimerEthKey *ecdsa.PrivateKey,
	ec *ethclient.Client,
	relaySwap *contracts.SwapCreatorRelaySwap,
) ([]byte, error) {
	signerAddress := ethcrypto.PubkeyToAddress(claimerEthKey.PublicKey)
	if relaySwap.Swap.Claimer != signerAddress {
		return nil, fmt.Errorf("signing key %s does not match claimer %s", signerAddress, relaySwap.Swap.Claimer)
	}

	msg := relaySwap.Hash()

	// signature format is (r || s || v), v = 27/28
	signature, err := rcommon.NewKeyFromPrivateKey(claimerEthKey).Sign(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to sign forward request digest: %w", err)
	}

	return signature, nil
}
