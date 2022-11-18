package xmrmaker

import (
	"math/big"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/db"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func newTestRecoveryState(t *testing.T, timeout time.Duration) (*recoveryState, *offers.MockDatabase) {
	inst, s, tdb := newTestInstanceAndDB(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	sr := s.secp256k1Pub.Keccak256()

	newSwap(t, s, [32]byte{}, sr, big.NewInt(1), timeout)

	ethSwapInfo := &db.EthereumSwapInfo{
		ContractAddress: s.ContractAddr(),
		SwapID:          s.contractSwapID,
		Swap:            s.contractSwap,
		StartNumber:     big.NewInt(1),
	}

	dataDir := t.TempDir()
	rs, err := NewRecoveryState(
		inst.backend,
		types.Hash{},
		dataDir,
		s.privkeys.SpendKey(),
		ethSwapInfo,
	)
	require.NoError(t, err)

	return rs, tdb
}

func TestClaimOrRecover_Claim(t *testing.T) {
	// test case where XMRMaker is able to claim ether from the contract
	rs, _ := newTestRecoveryState(t, 24*time.Hour)
	txOpts, err := rs.ss.ETHClient().TxOpts(rs.ss.ctx)
	require.NoError(t, err)

	// set contract to Ready
	tx, err := rs.ss.Contract().SetReady(txOpts, rs.ss.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, rs.ss.ETHClient().Raw(), tx)

	// assert we can claim ether
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestClaimOrRecover_Recover(t *testing.T) {
	// test case where XMRMaker is able to reclaim their monero, after XMRTaker refunds
	rs, _ := newTestRecoveryState(t, 24*time.Hour)
	txOpts, err := rs.ss.ETHClient().TxOpts(rs.ss.ctx)
	require.NoError(t, err)

	// TODO: when recovery from disk is implemented, re-add this
	// db.EXPECT().PutOffer(rs.ss.offer)
	monero.MineMinXMRBalance(t, rs.ss.XMRClient(), common.MoneroToPiconero(1))

	// lock XMR
	rs.ss.setXMRTakerPublicKeys(rs.ss.pubkeys, nil)
	lockedXMR, err := rs.ss.lockFunds(1)
	require.NoError(t, err)

	// call refund w/ XMRTaker's spend key
	sc := rs.ss.getSecret()
	tx, err := rs.ss.Contract().Refund(txOpts, rs.ss.contractSwap, sc)
	require.NoError(t, err)
	tests.MineTransaction(t, rs.ss.ETHClient().Raw(), tx)

	// assert XMRMaker can reclaim their monero
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Recovered)
	require.Equal(t, lockedXMR.Address, string(res.MoneroAddress))
	require.NotEmpty(t, lockedXMR.Address, "")
}
