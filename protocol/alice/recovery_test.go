package alice

import (
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/stretchr/testify/require"
)

func newTestRecoveryState(t *testing.T) *recoveryState {
	inst, s := newTestInstance(t)
	akp, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	s.privkeys = akp
	s.pubkeys = akp.PublicKeyPair()

	s.setBobKeys(akp.SpendKey().Public(), akp.ViewKey(), nil)
	s.bobAddress = inst.callOpts.From
	addr, err := s.deployAndLockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	rs, err := NewRecoveryState(inst, akp.SpendKey(), addr)
	require.NoError(t, err)
	return rs
}

func TestClaimOrRefund_Claim(t *testing.T) {
	// test case where Bob has claimed the ether, so Alice should be able to
	// claim the monero.
	rs := newTestRecoveryState(t)

	// call swap.Ready()
	err := rs.ss.ready()
	require.NoError(t, err)

	// call swap.Claim()
	secret := rs.ss.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))
	_, err = rs.ss.contract.Claim(rs.ss.txOpts, sc)
	require.NoError(t, err)

	// assert we can claim the monero
	res, err := rs.ClaimOrRefund()
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestClaimOrRefund_Refund_beforeT0(t *testing.T) {
	// test case where Bob hasn't claimed the ether, and it's before
	// t0/IsReady, so Alice should be able to refund.
	rs := newTestRecoveryState(t)

	// assert we can refund the ether
	res, err := rs.ClaimOrRefund()
	require.NoError(t, err)
	require.True(t, res.Refunded)
}

func TestClaimOrRefund_Refund_afterT1(t *testing.T) {
	// test case where Bob hasn't claimed the ether, and it's after
	// t1, so Alice should be able to refund.
	rs := newTestRecoveryState(t)

	rpcClient, err := rpc.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	var result string
	err = rpcClient.Call(&result, "evm_snapshot")
	require.NoError(t, err)

	err = rpcClient.Call(nil, "evm_increaseTime", defaultTimeoutDuration.Int64()*2+360)
	require.NoError(t, err)

	defer func() {
		var ok bool
		err = rpcClient.Call(&ok, "evm_revert", result)
		require.NoError(t, err)
	}()

	// assert we can refund the ether
	res, err := rs.ClaimOrRefund()
	require.NoError(t, err)
	require.True(t, res.Refunded)
}
