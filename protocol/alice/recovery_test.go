package alice

import (
	"testing"

	"github.com/noot/atomic-swap/common"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

func newTestRecoveryState(t *testing.T) *recoveryState {
	inst, s := newTestInstance(t)
	akp, err := generateKeys()
	require.NoError(t, err)

	s.privkeys = akp.PrivateKeyPair
	s.pubkeys = akp.PublicKeyPair
	s.secp256k1Pub = akp.Secp256k1PublicKey
	s.dleqProof = akp.DLEqProof

	s.setBobKeys(s.pubkeys.SpendKey(), s.privkeys.ViewKey(), akp.Secp256k1PublicKey)
	s.bobAddress = inst.callOpts.From
	err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	rs, err := NewRecoveryState(inst, s.privkeys.SpendKey(), inst.contractAddr, s.contractSwapID)
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
	sc := rs.ss.getSecret()
	_, err = rs.ss.alice.contract.Claim(rs.ss.txOpts, rs.ss.contractSwapID, sc)
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
