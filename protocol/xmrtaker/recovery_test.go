package xmrtaker

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/tests"
)

func newTestRecoveryState(t *testing.T, timeout time.Duration) *recoveryState {
	s := newTestInstance(t)
	s.SetSwapTimeout(timeout)
	akp, err := generateKeys()
	require.NoError(t, err)

	s.privkeys = akp.PrivateKeyPair
	s.pubkeys = akp.PublicKeyPair
	s.secp256k1Pub = akp.Secp256k1PublicKey
	s.dleqProof = akp.DLEqProof

	s.setXMRMakerKeys(s.pubkeys.SpendKey(), s.privkeys.ViewKey(), akp.Secp256k1PublicKey)
	s.xmrmakerAddress = s.ETHClient().Address()

	_, err = s.lockAsset()
	require.NoError(t, err)

	ethSwapInfo := &db.EthereumSwapInfo{
		SwapID:      s.contractSwapID,
		Swap:        s.contractSwap,
		StartNumber: big.NewInt(1),
	}

	dataDir := t.TempDir()
	rs, err := NewRecoveryState(s, types.Hash{}, dataDir, s.privkeys.SpendKey(), ethSwapInfo)
	require.NoError(t, err)
	return rs
}

func TestClaimOrRefund_Claim(t *testing.T) {
	// test case where XMRMaker has claimed the ether, so XMRTaker should be able to
	// claim the monero.
	rs := newTestRecoveryState(t, 12*time.Second)

	// call swap.Ready()
	err := rs.ss.ready()
	require.NoError(t, err)

	// call swap.Claim()
	sc := rs.ss.getSecret()
	txOpts, err := rs.ss.ETHClient().TxOpts(rs.ss.ctx)
	require.NoError(t, err)

	tx, err := rs.ss.Contract().Claim(txOpts, rs.ss.contractSwap, sc)
	require.NoError(t, err)
	tests.MineTransaction(t, rs.ss.ETHClient().Raw(), tx)
	t.Log("XMRMaker claimed ETH...")

	// assert we can claim the monero
	res, err := rs.ClaimOrRefund()
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestClaimOrRefund_Refund_beforeT0(t *testing.T) {
	// test case where XMRMaker hasn't claimed the ether, and it's before
	// t0/IsReady, so XMRTaker should be able to refund.
	rs := newTestRecoveryState(t, 12*time.Second)

	// assert we can refund the ether
	res, err := rs.ClaimOrRefund()
	require.NoError(t, err)
	require.True(t, res.Refunded)
}

func TestClaimOrRefund_Refund_afterT1(t *testing.T) {
	// test case where XMRMaker hasn't claimed the ether, and it's after
	// t1, so XMRTaker should be able to refund.
	rs := newTestRecoveryState(t, 1) // T1 expires before the new swap TX is confirmed
	// assert we can refund the ether
	res, err := rs.ClaimOrRefund()
	require.NoError(t, err)
	require.True(t, res.Refunded)
}
