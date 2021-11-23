package alice

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var _ = logging.SetLogLevel("alice", "debug")

type mockNet struct {
	msg net.Message
}

func (n *mockNet) SendSwapMessage(msg net.Message) error {
	n.msg = msg
	return nil
}

func newTestAlice(t *testing.T) (*alice, *swapState) {
	alice, err := NewAlice(context.Background(), common.DefaultAliceMoneroEndpoint, common.DefaultEthEndpoint, common.DefaultPrivKeyAlice)
	require.NoError(t, err)
	swapState := newSwapState(alice, 1, 1)
	return alice, swapState
}

func TestSwapState_HandleProtocolMessage_SendKeysMessage(t *testing.T) {
	_, s := newTestAlice(t)
	defer s.cancel()

	msg := &net.SendKeysMessage{}
	_, _, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingKeys, err)

	_, err = s.generateKeys()
	require.NoError(t, err)

	bobPrivKeys, err := monero.GenerateKeys()
	require.NoError(t, err)

	msg = &net.SendKeysMessage{
		PublicSpendKey: bobPrivKeys.SpendKey().Public().Hex(),
		PrivateViewKey: bobPrivKeys.ViewKey().Hex(),
		SpendKeyHash:   bobPrivKeys.SpendKey().HashString(),
		EthAddress:     "0x",
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, time.Second*time.Duration(defaultTimeoutDuration.Int64()), s.t1.Sub(s.t0))
	require.Equal(t, bobPrivKeys.SpendKey().Public().Hex(), s.bobPublicSpendKey.Hex())
	require.Equal(t, bobPrivKeys.ViewKey().Hex(), s.bobPrivateViewKey.Hex())
	require.Equal(t, bobPrivKeys.SpendKey().Hash(), s.bobClaimHash)
}

func TestSwapState_HandleProtocolMessage_SendKeysMessage_Refund(t *testing.T) {
	_, s := newTestAlice(t)
	defer s.cancel()
	s.net = new(mockNet)

	// set timeout to 2s
	// TODO: pass this as a param to newSwapState
	defaultTimeoutDuration = big.NewInt(2)
	defer func() {
		defaultTimeoutDuration = big.NewInt(60 * 60 * 24)
	}()

	_, err := s.generateKeys()
	require.NoError(t, err)

	bobPrivKeys, err := monero.GenerateKeys()
	require.NoError(t, err)

	msg := &net.SendKeysMessage{
		PublicSpendKey: bobPrivKeys.SpendKey().Public().Hex(),
		PrivateViewKey: bobPrivKeys.ViewKey().Hex(),
		SpendKeyHash:   bobPrivKeys.SpendKey().HashString(),
		EthAddress:     "0x",
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyContractDeployedType, resp.Type())
	require.Equal(t, time.Second*time.Duration(defaultTimeoutDuration.Int64()), s.t1.Sub(s.t0))
	require.Equal(t, bobPrivKeys.SpendKey().Public().Hex(), s.bobPublicSpendKey.Hex())
	require.Equal(t, bobPrivKeys.ViewKey().Hex(), s.bobPrivateViewKey.Hex())

	// ensure we refund before t0
	time.Sleep(time.Second * 2)
	require.NotNil(t, s.net.(*mockNet).msg)
	require.Equal(t, net.NotifyRefundType, s.net.(*mockNet).msg.Type())
	// TODO: check balance
}

func TestSwapState_NotifyXMRLock(t *testing.T) {
	_, s := newTestAlice(t)
	defer s.cancel()
	s.nextExpectedMessage = &net.NotifyXMRLock{}

	_, err := s.generateKeys()
	require.NoError(t, err)

	bobPrivKeys, err := monero.GenerateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobPrivKeys.SpendKey().Public(), bobPrivKeys.ViewKey())

	_, err = s.deployAndLockETH(1)
	require.NoError(t, err)

	s.desiredAmount = 0
	kp := monero.SumSpendAndViewKeys(bobPrivKeys.PublicKeyPair(), s.pubkeys)
	xmrAddr := kp.Address()

	msg := &net.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyReadyType, resp.Type())

	// TODO: test refund case
}

func TestSwapState_NotifyClaimed(t *testing.T) {
	// TODO
}
