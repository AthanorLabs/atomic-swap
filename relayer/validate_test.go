// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/tests"
)

func TestValidateRelayerFee(t *testing.T) {
	ctx := context.Background()
	ec, _ := tests.NewEthClient(t)
	key := tests.GetTakerTestKey(t)
	swapCreatorAddr, _ := contracts.DevDeploySwapCreator(t, ec, key)

	// 20-byte empty address, 4-byte zero salt
	empty := [24]byte{}
	relayerHash := crypto.Keccak256Hash(empty[:])

	type testCase struct {
		description string
		value       *big.Int
		expectErr   string
	}

	testCases := []testCase{
		{
			description: "swap value equal to relayer fee",
			value:       coins.RelayerFeeWei,
			expectErr:   "swap value of 0.009 ETH is too low to support 0.009 ETH relayer fee",
		},
		{
			description: "swap value less than relayer fee",
			value:       new(big.Int).Sub(coins.RelayerFeeWei, big.NewInt(1e15)),
			expectErr:   "swap value of 0.008 ETH is too low to support 0.009 ETH relayer fee",
		},
		{
			description: "swap value larger than min fee",
			value:       new(big.Int).Add(coins.RelayerFeeWei, big.NewInt(1e15)),
		},
	}

	for _, tc := range testCases {
		swap := contracts.SwapCreatorSwap{
			Owner:        ethcommon.Address{},
			Claimer:      ethcommon.Address{},
			PubKeyClaim:  [32]byte{},
			PubKeyRefund: [32]byte{},
			Timeout1:     new(big.Int),
			Timeout2:     new(big.Int),
			Asset:        ethcommon.Address{},
			Value:        tc.value,
			Nonce:        new(big.Int),
		}

		request := &message.RelayClaimRequest{
			RelaySwap: &contracts.SwapCreatorRelaySwap{
				Swap:        swap,
				Fee:         big.NewInt(1),
				SwapCreator: swapCreatorAddr,
				RelayerHash: relayerHash,
			},
			Secret: make([]byte, 32),
		}

		err := validateClaimValues(ctx, request, ec, ethcommon.Address{}, [4]byte{}, swapCreatorAddr)
		if tc.expectErr != "" {
			require.ErrorContains(t, err, tc.expectErr, tc.description)
		} else {
			require.NoError(t, err, tc.description)
		}
	}
}

// In the taker claim scenario, we need to fail if the contract address is not
// identical. If the claim has a different address, the swap was not created by
// the taker who is being asked to claim.
func Test_validateClaimValues_takerClaim_contractAddressNotEqualFail(t *testing.T) {
	offerID := types.Hash{0x1}                       // non-nil offer ID passed to indicate taker claim
	swapCreatorAddrInClaim := ethcommon.Address{0x1} // address in claim
	swapCreatorAddrOurs := ethcommon.Address{0x2}    // passed to validateClaimValues

	request := &message.RelayClaimRequest{
		OfferID: &offerID,
		Secret:  make([]byte, 32),
		RelaySwap: &contracts.SwapCreatorRelaySwap{
			SwapCreator: swapCreatorAddrInClaim,
		},
	}

	err := validateClaimValues(context.Background(), request, nil, ethcommon.Address{}, [4]byte{}, swapCreatorAddrOurs)
	require.ErrorContains(t, err, "taker claim swap creator mismatch")
}

// When validating a claim made to a DHT advertised relayer, the contacts can have
// different addresses, but the claim's contract must be byte-code compatible. This
// tests for failure when it is not byte-code compatible.
func Test_validateClaimValues_dhtClaim_contractAddressNotEqual(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	key := tests.GetTakerTestKey(t)
	swapCreatorAddr, _ := contracts.DevDeploySwapCreator(t, ec, key)

	request := &message.RelayClaimRequest{
		OfferID: nil, // DHT relayer claim
		Secret:  make([]byte, 32),
		RelaySwap: &contracts.SwapCreatorRelaySwap{
			SwapCreator: ethcommon.Address{1}, // not a valid swap creator contract
		},
	}

	err := validateClaimValues(context.Background(), request, ec, ethcommon.Address{}, [4]byte{}, swapCreatorAddr)
	require.ErrorContains(t, err, "contract address does not contain correct SwapCreator code")
}

func Test_validateSignature(t *testing.T) {
	ethKey := tests.GetMakerTestKey(t)
	claimer := crypto.PubkeyToAddress(*ethKey.Public().(*ecdsa.PublicKey))
	ec, _ := tests.NewEthClient(t)
	secret := [32]byte{0x1}
	swapCreatorAddr, _ := contracts.DevDeploySwapCreator(t, ec, ethKey)

	swap := createTestSwap(claimer)
	relaySwap := &contracts.SwapCreatorRelaySwap{
		SwapCreator: swapCreatorAddr,
		Swap:        *swap,
		RelayerHash: types.Hash{},
		Fee:         big.NewInt(1),
	}

	req, err := CreateRelayClaimRequest(ethKey, relaySwap, secret)
	require.NoError(t, err)

	// success path
	err = validateClaimSignature(req)
	require.NoError(t, err)

	// failure path (tamper with an arbitrary byte of the signature)
	req.Signature[10]++
	err = validateClaimSignature(req)
	// can be "recovery failed" or "signer of message is not swap claimer"
	require.Error(t, err)
}

func Test_validateClaimRequest(t *testing.T) {
	ctx := context.Background()
	ethKey := tests.GetMakerTestKey(t)
	claimer := crypto.PubkeyToAddress(*ethKey.Public().(*ecdsa.PublicKey))
	ec, _ := tests.NewEthClient(t)
	secret := [32]byte{0x1}
	swapCreatorAddr, _ := contracts.DevDeploySwapCreator(t, ec, ethKey)

	// 20-byte empty address, 4-byte zero salt
	empty := [24]byte{}
	relayerHash := crypto.Keccak256Hash(empty[:])

	swap := createTestSwap(claimer)
	relaySwap := &contracts.SwapCreatorRelaySwap{
		SwapCreator: swapCreatorAddr,
		Swap:        *swap,
		RelayerHash: relayerHash,
		Fee:         big.NewInt(1),
	}

	req, err := CreateRelayClaimRequest(ethKey, relaySwap, secret)
	require.NoError(t, err)

	// success path
	err = validateClaimRequest(ctx, req, ec, ethcommon.Address{}, [4]byte{}, swapCreatorAddr)
	require.NoError(t, err)

	// test failure path by passing a non-eth asset
	asset := ethcommon.Address{0x1}
	req.RelaySwap.Swap.Asset = asset
	err = validateClaimRequest(ctx, req, ec, ethcommon.Address{}, [4]byte{}, swapCreatorAddr)
	require.ErrorContains(t, err, fmt.Sprintf("relaying for ETH Asset %s is not supported", types.EthAsset(asset)))
}
