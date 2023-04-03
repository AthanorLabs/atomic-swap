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
	swapFactoryAddr, forwarderAddr := deployContracts(t, ec, key)

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

		var secret [32]byte

		request := &message.RelayClaimRequest{
			SwapFactoryAddress: swapFactoryAddr,
			Swap:               swap,
			Secret:             secret[:],
		}

		err := validateClaimValues(ctx, request, ec, forwarderAddr)
		if tc.expectErr != "" {
			require.ErrorContains(t, err, tc.expectErr, tc.description)
		} else {
			require.NoError(t, err, tc.description)
		}
	}
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
	err = validateClaimRequest(ctx, req, ec, forwarderAddr)
	require.ErrorContains(t, err, fmt.Sprintf("relaying for ETH Asset %s is not supported", types.EthAsset(asset)))
}
