// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/tests"
)

func createTestSwap(claimer ethcommon.Address) *contracts.SwapCreatorSwap {
	return &contracts.SwapCreatorSwap{
		Owner:        ethcommon.Address{0x1},
		Claimer:      claimer,
		PubKeyClaim:  [32]byte{0x1},
		PubKeyRefund: [32]byte{0x1},
		Timeout1:     big.NewInt(time.Now().Add(30 * time.Minute).Unix()),
		Timeout2:     big.NewInt(time.Now().Add(60 * time.Minute).Unix()),
		Asset:        ethcommon.Address(types.EthAssetETH),
		Value:        big.NewInt(1e18),
		Nonce:        big.NewInt(1),
	}
}

func TestCreateRelayClaimRequest(t *testing.T) {
	ethKey := tests.GetMakerTestKey(t)
	claimer := crypto.PubkeyToAddress(*ethKey.Public().(*ecdsa.PublicKey))
	ec, _ := tests.NewEthClient(t)
	secret := [32]byte{0x1}
	swapCreatorAddr, _ := contracts.DevDeploySwapCreator(t, ec, ethKey)

	// success path
	swap := createTestSwap(claimer)
	relaySwap := &contracts.SwapCreatorRelaySwap{
		Swap:        *swap,
		Fee:         big.NewInt(1),
		SwapCreator: swapCreatorAddr,
		RelayerHash: types.Hash{},
	}
	req, err := CreateRelayClaimRequest(ethKey, relaySwap, secret)
	require.NoError(t, err)
	require.NotNil(t, req)

	// change the ethkey to not match the claimer address to trigger the error path
	ethKey = tests.GetTakerTestKey(t)
	_, err = CreateRelayClaimRequest(ethKey, relaySwap, secret)
	require.ErrorContains(t, err, "does not match claimer")
}
