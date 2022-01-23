package bob

import (
	"math/big"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"

	"github.com/stretchr/testify/require"
)

func newTestRecoveryState(t *testing.T) *recoveryState {
	inst, s := newTestInstance(t)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	sr := s.secp256k1Pub.Keccak256()

	duration, err := time.ParseDuration("1440m")
	require.NoError(t, err)
	addr, _ := deploySwap(t, inst, s, sr, big.NewInt(1), duration)
	rs, err := NewRecoveryState(inst, s.privkeys.SpendKey(), addr)
	require.NoError(t, err)

	return rs
}

func TestClaimOrRecover_Claim(t *testing.T) {
	// test case where Bob is able to claim ether from the contract
	rs := newTestRecoveryState(t)

	// set contract to Ready
	_, err := rs.ss.contract.SetReady(rs.ss.txOpts)
	require.NoError(t, err)

	// assert we can claim ether
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestClaimOrRecover_Recover(t *testing.T) {
	// test case where Bob is able to reclaim his monero, after Alice refunds
	rs := newTestRecoveryState(t)

	daemonClient := monero.NewClient(common.DefaultMoneroDaemonEndpoint)
	addr, err := rs.ss.bob.client.GetAddress(0)
	require.NoError(t, err)
	_ = daemonClient.GenerateBlocks(addr.Address, 121)

	// lock XMR
	rs.ss.setAlicePublicKeys(rs.ss.pubkeys, nil)
	addrAB, err := rs.ss.lockFunds(333)
	require.NoError(t, err)

	// call refund w/ Alice's spend key
	sc := rs.ss.getSecret()
	_, err = rs.ss.contract.Refund(rs.ss.txOpts, sc)
	require.NoError(t, err)

	// assert Bob can reclaim his monero
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Recovered)
	require.Equal(t, addrAB, res.MoneroAddress)
}
