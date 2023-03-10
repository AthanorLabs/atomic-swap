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
			value:       big.NewInt(100),
			fee:         big.NewInt(10),
			minFee:      big.NewInt(10),
		},
		{
			description: "relayer fee greater than minFee",
			value:       big.NewInt(100),
			fee:         big.NewInt(2),
			minFee:      big.NewInt(1),
		},
		{
			description: "relayer fee less than min fee",
			value:       big.NewInt(10000),
			fee:         big.NewInt(99),
			minFee:      big.NewInt(100),
			expectErr:   "fee too low: got 99, expected minimum 100",
		},
		{
			description: "swap value equal to relayer fee",
			value:       big.NewInt(100),
			fee:         big.NewInt(100),
			minFee:      big.NewInt(1),
			expectErr:   "relayer fee is not greater than swap value",
		},
		{
			description: "relayer fee absurdly larger than min fee",
			value:       big.NewInt(10000),
			fee:         big.NewInt(101),
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
