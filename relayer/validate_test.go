package relayer

import (
	"context"
	"math/big"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
	"github.com/athanorlabs/atomic-swap/tests"
)

func TestValidateRelayerFee(t *testing.T) {
	ctx := context.Background()

	ec, _ := tests.NewEthClient(t)
	key := tests.GetTakerTestKey(t)

	forwarderAddr, err := contracts.DeployGSNForwarderWithKey(ctx, ec, key)
	require.NoError(t, err)

	contractAddr, _, err := contracts.DeploySwapFactoryWithKey(ctx, ec, key, forwarderAddr)
	require.NoError(t, err)

	type testCase struct {
		description string
		value       *big.Int
		expectErr   string
	}

	testCases := []testCase{
		{
			description: "swap value equal to relayer fee",
			value:       RelayerFeeWei,
			expectErr:   "swap value of 0.009 ETH is too low to support 0.009 ETH relayer fee",
		},
		{
			description: "swap value less than relayer fee",
			value:       new(big.Int).Sub(RelayerFeeWei, big.NewInt(1e15)),
			expectErr:   "swap value of 0.008 ETH is too low to support 0.009 ETH relayer fee",
		},
		{
			description: "swap value larger than min fee",
			value:       new(big.Int).Add(RelayerFeeWei, big.NewInt(1e15)),
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
			SFContractAddress: contractAddr,
			Swap:              swap,
			Secret:            secret[:],
		}

		err = validateClaimValues(ctx, request, ec, forwarderAddr)
		if tc.expectErr != "" {
			require.ErrorContains(t, err, tc.expectErr, tc.description)
		} else {
			require.NoError(t, err, tc.description)
		}
	}
}
