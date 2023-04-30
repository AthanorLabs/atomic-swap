// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/tests"
)

// Speed up tests a little by giving deployContracts(...) a package-level cache.
// These variables should not be accessed by other functions.
var _swapCreatorAddr *ethcommon.Address

// deployContracts deploys and returns the swapCreator and forwarder addresses.
func deployContracts(t *testing.T, ec *ethclient.Client, key *ecdsa.PrivateKey) ethcommon.Address {
	ctx := context.Background()

	if _swapCreatorAddr == nil {
		swapCreatorAddr, _, err := contracts.DeploySwapCreatorWithKey(ctx, ec, key)
		require.NoError(t, err)
		_swapCreatorAddr = &swapCreatorAddr
	}

	return *_swapCreatorAddr
}

func createTestSwap(claimer ethcommon.Address) *contracts.SwapCreatorSwap {
	return &contracts.SwapCreatorSwap{
		Owner:        ethcommon.Address{0x1},
		Claimer:      claimer,
		PubKeyClaim:  [32]byte{0x1},
		PubKeyRefund: [32]byte{0x1},
		Timeout0:     big.NewInt(time.Now().Add(30 * time.Minute).Unix()),
		Timeout1:     big.NewInt(time.Now().Add(60 * time.Minute).Unix()),
		Asset:        ethcommon.Address(types.EthAssetETH),
		Value:        big.NewInt(1e18),
		Nonce:        big.NewInt(1),
	}
}

func TestCreateRelayClaimRequest(t *testing.T) {
	ctx := context.Background()
	ethKey := tests.GetMakerTestKey(t)
	claimer := crypto.PubkeyToAddress(*ethKey.Public().(*ecdsa.PublicKey))
	ec, _ := tests.NewEthClient(t)
	secret := [32]byte{0x1}
	swapCreatorAddr := deployContracts(t, ec, ethKey)

	// success path
	swap := createTestSwap(claimer)
	relaySwap := &contracts.SwapCreatorRelaySwap{
		Swap:        *swap,
		Fee:         big.NewInt(1),
		SwapCreator: swapCreatorAddr,
		Relayer:     ethcommon.Address{},
	}
	req, err := CreateRelayClaimRequest(ctx, ethKey, ec, relaySwap, secret)
	require.NoError(t, err)
	require.NotNil(t, req)

	// change the ethkey to not match the claimer address to trigger the error path
	ethKey = tests.GetTakerTestKey(t)
	_, err = CreateRelayClaimRequest(ctx, ethKey, ec, relaySwap, secret)
	require.ErrorContains(t, err, "signing key does not match claimer")
}
