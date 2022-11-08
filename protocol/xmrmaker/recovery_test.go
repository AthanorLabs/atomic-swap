package xmrmaker

import (
	"math/big"
	"path"
	"testing"
	"time"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker/offers"
	"github.com/athanorlabs/atomic-swap/tests"

	"github.com/stretchr/testify/require"
)

func newTestRecoveryState(t *testing.T, timeout time.Duration) (*recoveryState, *offers.MockDatabase) {
	inst, s, db := newTestInstanceAndDB(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	sr := s.secp256k1Pub.Keccak256()

	newSwap(t, s, [32]byte{}, sr, big.NewInt(1), timeout)

	dataDir := path.Join(t.TempDir(), "test-infofile")
	rs, err := NewRecoveryState(inst.backend, dataDir, s.privkeys.SpendKey(), s.ContractAddr(),
		s.contractSwapID, s.contractSwap)
	require.NoError(t, err)

	return rs, db
}

func TestClaimOrRecover_Claim(t *testing.T) {
	// test case where XMRMaker is able to claim ether from the contract
	rs, _ := newTestRecoveryState(t, 24*time.Hour)
	txOpts, err := rs.ss.TxOpts()
	require.NoError(t, err)

	// set contract to Ready
	tx, err := rs.ss.Contract().SetReady(txOpts, rs.ss.contractSwap)
	require.NoError(t, err)
	tests.MineTransaction(t, rs.ss, tx)

	// assert we can claim ether
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestClaimOrRecover_Recover(t *testing.T) {
	// test case where XMRMaker is able to reclaim their monero, after XMRTaker refunds
	rs, db := newTestRecoveryState(t, 24*time.Hour)
	txOpts, err := rs.ss.TxOpts()
	require.NoError(t, err)

	monero.MineMinXMRBalance(t, rs.ss, common.MoneroToPiconero(1))

	// lock XMR
	rs.ss.setXMRTakerPublicKeys(rs.ss.pubkeys, nil)
	lockedXMR, err := rs.ss.lockFunds(1)
	require.NoError(t, err)

	// call refund w/ XMRTaker's spend key
	sc := rs.ss.getSecret()
	tx, err := rs.ss.Contract().Refund(txOpts, rs.ss.contractSwap, sc)
	require.NoError(t, err)
	tests.MineTransaction(t, rs.ss, tx)

	db.EXPECT().PutOffer(rs.ss.offer)

	// assert XMRMaker can reclaim their monero
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Recovered)
	require.Equal(t, lockedXMR.Address, string(res.MoneroAddress))
	require.NotEmpty(t, lockedXMR.Address, "")
}
