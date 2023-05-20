// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package protocol

import (
	"context"
	"path"
	"testing"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/protocol/swap"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var (
	_ = logging.SetLogLevel("monero", "debug")
	_ = logging.SetLogLevel("protocol", "debug")
)

type mockSwapManager struct{}

func (*mockSwapManager) WriteSwapToDB(info *swap.Info) error {
	return nil
}

func (*mockSwapManager) PushNewStatus(_ types.Hash, _ types.Status) {
}

func TestClaimMonero_NoTransferBack(t *testing.T) {
	env := common.Development

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	conf := &monero.WalletClientConf{
		Env:                 env,
		WalletFilePath:      path.Join(t.TempDir(), "test-wallet-tcm"),
		MoneroWalletRPCPath: monero.GetWalletRPCDirectory(t),
	}
	err = conf.Fill()
	require.NoError(t, err)

	moneroCli, err := monero.CreateSpendWalletFromKeys(conf, kp, 0)
	require.NoError(t, err)
	height, err := moneroCli.GetHeight()
	require.NoError(t, err)
	xmrAmt := coins.StrToDecimal("1")
	pnAmt := coins.MoneroToPiconero(xmrAmt)
	monero.MineMinXMRBalance(t, moneroCli, pnAmt)

	info := &swap.Info{
		MoneroStartHeight: height,
	}

	err = ClaimMonero(
		context.Background(),
		common.Development,
		info,
		moneroCli,
		kp,
		nil, // deposit address can be nil, as noTransferBack is true
		true,
		new(mockSwapManager),
	)
	require.NoError(t, err)
}

func TestClaimMonero_WithTransferBack(t *testing.T) {
	monero.TestBackgroundMineBlocks(t)
	env := common.Development

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	conf := &monero.WalletClientConf{
		Env:                 env,
		WalletFilePath:      path.Join(t.TempDir(), "test-wallet-tcm"),
		MoneroWalletRPCPath: monero.GetWalletRPCDirectory(t),
	}
	err = conf.Fill()
	require.NoError(t, err)

	moneroCli, err := monero.CreateSpendWalletFromKeys(conf, kp, 0)
	require.NoError(t, err)
	height, err := moneroCli.GetHeight()
	require.NoError(t, err)
	xmrAmt := coins.StrToDecimal("1")
	pnAmt := coins.MoneroToPiconero(xmrAmt)
	monero.MineMinXMRBalance(t, moneroCli, pnAmt)

	kp2, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	depositAddr := kp2.PublicKeyPair().Address(env)

	info := &swap.Info{
		MoneroStartHeight: height,
	}

	err = ClaimMonero(
		context.Background(),
		common.Development,
		info,
		moneroCli,
		kp,
		depositAddr,
		false,
		new(mockSwapManager),
	)
	require.NoError(t, err)
}
