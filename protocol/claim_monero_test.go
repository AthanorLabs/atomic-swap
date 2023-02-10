package protocol

import (
	"context"
	"path"
	"testing"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/monero"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var (
	_ = logging.SetLogLevel("monero", "debug")
	_ = logging.SetLogLevel("protocol", "debug")
)

func TestClaimMonero_NoTransferBack(t *testing.T) {
	env := common.Development

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	conf := &monero.WalletClientConf{
		Env:            env,
		WalletFilePath: path.Join(t.TempDir(), "test-wallet-tcm"),
	}
	err = conf.Fill()
	require.NoError(t, err)

	moneroCli, err := monero.CreateSpendWalletFromKeys(conf, kp, 0)
	require.NoError(t, err)
	height, err := moneroCli.GetHeight()
	require.NoError(t, err)
	amt := coins.StrToDecimal("1")
	pnamt := coins.PiconeroAmount(*amt)
	monero.MineMinXMRBalance(t, moneroCli, &pnamt)

	_, err = ClaimMonero(
		context.Background(),
		common.Development,
		[32]byte{},
		moneroCli,
		height,
		kp,
		"",
		false,
	)
	require.NoError(t, err)
}

func TestClaimMonero_WithTransferBack(t *testing.T) {
	monero.TestBackgroundMineBlocks(t)
	env := common.Development

	kp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	conf := &monero.WalletClientConf{
		Env:            env,
		WalletFilePath: path.Join(t.TempDir(), "test-wallet-tcm"),
	}
	err = conf.Fill()
	require.NoError(t, err)

	moneroCli, err := monero.CreateSpendWalletFromKeys(conf, kp, 0)
	require.NoError(t, err)
	height, err := moneroCli.GetHeight()
	require.NoError(t, err)
	amt := coins.StrToDecimal("1")
	pnamt := coins.PiconeroAmount(*amt)
	monero.MineMinXMRBalance(t, moneroCli, &pnamt)

	kp2, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	depositAddr := kp2.PublicKeyPair().Address(env)

	_, err = ClaimMonero(
		context.Background(),
		common.Development,
		[32]byte{},
		moneroCli,
		height,
		kp,
		depositAddr,
		true,
	)
	require.NoError(t, err)
}
