package xmrmaker

import (
	"math/big"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

func newTestRecoveryState(t *testing.T, ec *ethclient.Client) *recoveryState {
	inst, s := newTestInstance(t, ec)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	sr := s.secp256k1Pub.Keccak256()

	duration, err := time.ParseDuration("1440m")
	require.NoError(t, err)
	newSwap(t, s, [32]byte{}, sr, big.NewInt(1), duration)

	rs, err := NewRecoveryState(inst.backend, "/tmp/test-infofile", s.privkeys.SpendKey(), s.ContractAddr(),
		s.contractSwapID, s.contractSwap)
	require.NoError(t, err)

	return rs
}

func TestClaimOrRecover_Claim(t *testing.T) {
	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	defer ec.Close()

	// test case where XMRMaker is able to claim ether from the contract
	rs := newTestRecoveryState(t, ec)
	txOpts, err := rs.ss.TxOpts()
	require.NoError(t, err)

	// set contract to Ready
	_, err = rs.ss.Contract().SetReady(txOpts, rs.ss.contractSwap)
	require.NoError(t, err)

	// assert we can claim ether
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestClaimOrRecover_Recover(t *testing.T) {
	if testing.Short() {
		t.Skip() // TODO: fails on CI w/ "not enough money"
	}

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	defer ec.Close()

	// test case where XMRMaker is able to reclaim his monero, after XMRTaker refunds
	rs := newTestRecoveryState(t, ec)
	txOpts, err := rs.ss.TxOpts()
	require.NoError(t, err)

	daemonClient := monero.NewClient(common.DefaultMoneroDaemonEndpoint)
	addr, err := rs.ss.GetAddress(0)
	require.NoError(t, err)
	_ = daemonClient.GenerateBlocks(addr.Address, 121)

	// lock XMR
	rs.ss.setXMRTakerPublicKeys(rs.ss.pubkeys, nil)
	addrAB, err := rs.ss.lockFunds(1)
	require.NoError(t, err)

	// call refund w/ XMRTaker's spend key
	sc := rs.ss.getSecret()
	_, err = rs.ss.Contract().Refund(txOpts, rs.ss.contractSwap, sc)
	require.NoError(t, err)

	// assert XMRMaker can reclaim his monero
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Recovered)
	require.Equal(t, addrAB, res.MoneroAddress)
}
