package main

import (
	"math/big"
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
)

func TestValidateRelayerFee(t *testing.T) {
	swapABI, err := abi.JSON(strings.NewReader(contracts.SwapFactoryMetaData.ABI))
	require.NoError(t, err)

	type testCase struct {
		value, fee       *big.Int
		minFeePercentage *apd.Decimal
		expectErr        bool
	}

	testCases := []testCase{
		{
			value:            big.NewInt(100),
			fee:              big.NewInt(1),
			minFeePercentage: apd.New(1, -2),
		},
		{
			value:            big.NewInt(100),
			fee:              big.NewInt(2),
			minFeePercentage: apd.New(1, -2),
		},
		{
			value:            big.NewInt(1000),
			fee:              big.NewInt(1),
			minFeePercentage: apd.New(1, -2),
			expectErr:        true,
		},
		{
			value:            big.NewInt(100),
			fee:              big.NewInt(100),
			minFeePercentage: apd.New(1, 0),
		},
		{
			value:            big.NewInt(100),
			fee:              big.NewInt(10),
			minFeePercentage: apd.New(1, -1),
		},
		{
			value:            big.NewInt(10000),
			fee:              big.NewInt(99),
			minFeePercentage: apd.New(1, -2),
			expectErr:        true,
		},
		{
			value:            big.NewInt(10000),
			fee:              big.NewInt(101),
			minFeePercentage: apd.New(1, -2),
		},
	}

	for _, tc := range testCases {
		args := []interface{}{
			&contracts.SwapFactorySwap{
				Owner:        ethcommon.Address{},
				Claimer:      ethcommon.Address{},
				PubKeyClaim:  [32]byte{},
				PubKeyRefund: [32]byte{},
				Timeout0:     new(big.Int),
				Timeout1:     new(big.Int),
				Asset:        ethcommon.Address{},
				Value:        tc.value,
				Nonce:        new(big.Int),
			},
			[32]byte{},
			tc.fee,
		}
		data, err := swapABI.Pack("claimRelayer", args...)
		require.NoError(t, err)

		unpacked, err := unpackData(data[4:])
		require.NoError(t, err)
		require.Equal(t, unpacked["value"], tc.value)
		require.Equal(t, unpacked["fee"], tc.fee)

		err = validateRelayerFee(unpacked, tc.minFeePercentage)
		if tc.expectErr {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
	}
}
