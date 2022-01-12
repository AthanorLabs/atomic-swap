package bob

import (
	"math/big"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"

	"github.com/stretchr/testify/require"
)

func newTestRecoveryState(t *testing.T) *recoveryState {
	inst, s := newTestInstance(t)

	bkp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	s.privkeys = bkp
	s.pubkeys = bkp.PublicKeyPair()

	refundKey := s.pubkeys.SpendKey().Bytes()
	var sr [32]byte
	copy(sr[:], common.Reverse(refundKey))

	duration, err := time.ParseDuration("1440m")
	require.NoError(t, err)
	addr, _ := deploySwap(t, inst, s, sr, big.NewInt(1), duration)
	rs, err := NewRecoveryState(inst, bkp.SpendKey(), addr)
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

	// lock XMR
	rs.ss.setAlicePublicKeys(rs.ss.pubkeys)
	addrAB, err := rs.ss.lockFunds(33000)
	require.NoError(t, err)

	// call refund w/ Alice's spend key
	secret := rs.ss.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	_, err = rs.ss.contract.Refund(rs.ss.txOpts, sc)
	require.NoError(t, err)

	// assert Bob can reclaim his monero
	res, err := rs.ClaimOrRecover()
	require.NoError(t, err)
	require.True(t, res.Recovered)
	require.Equal(t, addrAB, res.MoneroAddress)
}
