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
		description        string
		value, fee, minFee *big.Int
		expectErr          string
	}

	testCases := []testCase{
		{
			description: "relayer fee exactly equal to min fee",
			value:       big.NewInt(1e18),
			fee:         DefaultRelayerFeeWei,
			minFee:      DefaultRelayerFeeWei,
		},
		{
			description: "relayer fee greater than minFee",
			value:       big.NewInt(1e18),
			fee:         big.NewInt(2e15),
			minFee:      big.NewInt(1e15),
		},
		{
			description: "relayer fee less than min fee",
			value:       big.NewInt(1e18),
			fee:         big.NewInt(8e15),
			minFee:      big.NewInt(9e15),
			expectErr:   "fee too low: got 0.008 ETH, expected minimum 0.009 ETH",
		},
		{
			description: "swap value equal to relayer fee",
			value:       big.NewInt(1e17),
			fee:         big.NewInt(1e17),
			minFee:      DefaultRelayerFeeWei,
			expectErr:   "swap value of 0.1 ETH is too low to support 0.1 ETH relayer fee",
		},
		{
			description: "swap value less than relayer fee",
			value:       big.NewInt(8e15),
			fee:         big.NewInt(9e15),
			minFee:      big.NewInt(9e15),
			expectErr:   "swap value of 0.008 ETH is too low to support 0.009 ETH relayer fee",
		},
		{
			description: "relayer fee absurdly larger than min fee",
			value:       big.NewInt(1e18),
			fee:         big.NewInt(1e15),
			minFee:      big.NewInt(1),
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
			RelayerFeeWei:     tc.fee,
			Swap:              swap,
			Secret:            secret[:],
		}

		err = validateClaimValues(ctx, request, ec, forwarderAddr, tc.minFee)
		if tc.expectErr != "" {
			require.ErrorContains(t, err, tc.expectErr, tc.description)
		} else {
			require.NoError(t, err, tc.description)
		}
	}
}
