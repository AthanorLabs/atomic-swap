package main

import (
	"bytes"
	"math/big"
	"testing"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestValidateRelayerFee(t *testing.T) {
	buf := new(bytes.Buffer)
	_, err := buf.Write([]byte(contracts.SwapFactoryABI))
	require.NoError(t, err)

	swapABI, err := abi.JSON(buf)
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

		err = validateRelayerFee(data[4:], tc.minFeePercentage)
		if tc.expectErr {
			require.Error(t, err)
			continue
		}
		require.NoError(t, err)
	}
}
