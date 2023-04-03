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

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/tests"
)

func TestValidateRelayerFee(t *testing.T) {
	ctx := context.Background()
	ec, _ := tests.NewEthClient(t)
	key := tests.GetTakerTestKey(t)
	swapFactoryAddr, _ := deployContracts(t, ec, key)

	type testCase struct {
		description string
		value       *big.Int
		expectErr   string
	}

	testCases := []testCase{
		{
			description: "swap value equal to relayer fee",
			value:       FeeWei,
			expectErr:   "swap value of 0.009 ETH is too low to support 0.009 ETH relayer fee",
		},
		{
			description: "swap value less than relayer fee",
			value:       new(big.Int).Sub(FeeWei, big.NewInt(1e15)),
			expectErr:   "swap value of 0.008 ETH is too low to support 0.009 ETH relayer fee",
		},
		{
			description: "swap value larger than min fee",
			value:       new(big.Int).Add(FeeWei, big.NewInt(1e15)),
		},
	}

	for _, tc := range testCases {
		swap := &contracts.SwapFactorySwap{
			Owner:        ethcommon.Address{},
			Claimer:      ethcommon.Address{},
			PubKeyClaim:  [32]byte{},
			PubKeyRefund: [32]byte{},
			Timeout0:     new(big.Int),
			Timeout1:     new(big.Int),
			Asset:        ethcommon.Address{},
			Value:        tc.value,
			Nonce:        new(big.Int),
		}

		request := &message.RelayClaimRequest{
			SwapFactoryAddress: swapFactoryAddr,
			Swap:               swap,
			Secret:             make([]byte, 32),
		}

		err := validateClaimValues(ctx, request, ec, swapFactoryAddr)
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
	swapFactoryAddrInClaim := ethcommon.Address{0x1} // address in claim
	swapFactoryAddrOurs := ethcommon.Address{0x2}    // passed to validateClaimValues

	request := &message.RelayClaimRequest{
		OfferID:            &offerID,
		SwapFactoryAddress: swapFactoryAddrInClaim,
		Secret:             make([]byte, 32),
		Swap:               new(contracts.SwapFactorySwap), // test fails before we validate this
	}

	err := validateClaimValues(context.Background(), request, nil, swapFactoryAddrOurs)
	require.ErrorContains(t, err, "taker claim swap factory mismatch")
}

// When validating a claim made to a DHT advertised relayer, the contacts can have
// different addresses, but the claim's contract must be byte-code compatible. This
// tests for failure when it is not byte-code compatible.
func Test_validateClaimValues_dhtClaim_contractAddressNotEqual(t *testing.T) {
	ec, _ := tests.NewEthClient(t)
	key := tests.GetTakerTestKey(t)
	swapFactoryAddr, forwarderAddr := deployContracts(t, ec, key)

	request := &message.RelayClaimRequest{
		OfferID:            nil,           // DHT relayer claim
		SwapFactoryAddress: forwarderAddr, // not a valid swap factory contract
		Secret:             make([]byte, 32),
		Swap:               new(contracts.SwapFactorySwap), // test fails before we validate this
	}

	err := validateClaimValues(context.Background(), request, ec, swapFactoryAddr)
	require.ErrorContains(t, err, "contract address does not contain correct SwapFactory code")
}

func Test_validateSignature(t *testing.T) {
	ctx := context.Background()
	ethKey := tests.GetMakerTestKey(t)
	claimer := crypto.PubkeyToAddress(*ethKey.Public().(*ecdsa.PublicKey))
	ec, _ := tests.NewEthClient(t)
	secret := [32]byte{0x1}
	swapFactoryAddr, forwarderAddr := deployContracts(t, ec, ethKey)

	swap := createTestSwap(claimer)
	req, err := CreateRelayClaimRequest(ctx, ethKey, ec, swapFactoryAddr, forwarderAddr, swap, &secret)
	require.NoError(t, err)

	// success path
	err = validateClaimSignature(ctx, ec, req)
	require.NoError(t, err)

	// failure path (tamper with an arbitrary byte of the signature)
	req.Signature[10]++
	err = validateClaimSignature(ctx, ec, req)
	require.ErrorContains(t, err, "failed to verify signature")
}

func Test_validateClaimRequest(t *testing.T) {
	ctx := context.Background()
	ethKey := tests.GetMakerTestKey(t)
	claimer := crypto.PubkeyToAddress(*ethKey.Public().(*ecdsa.PublicKey))
	ec, _ := tests.NewEthClient(t)
	secret := [32]byte{0x1}
	swapFactoryAddr, forwarderAddr := deployContracts(t, ec, ethKey)

	swap := createTestSwap(claimer)
	req, err := CreateRelayClaimRequest(ctx, ethKey, ec, swapFactoryAddr, forwarderAddr, swap, &secret)
	require.NoError(t, err)

	// success path
	err = validateClaimRequest(ctx, req, ec, swapFactoryAddr)
	require.NoError(t, err)

	// test failure path by passing a non-eth asset
	asset := ethcommon.Address{0x1}
	req.Swap.Asset = asset
	err = validateClaimRequest(ctx, req, ec, swapFactoryAddr)
	require.ErrorContains(t, err, fmt.Sprintf("relaying for ETH Asset %s is not supported", types.EthAsset(asset)))
}
