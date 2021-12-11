package alice

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
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
	cfg := &Config{
		Ctx:                  context.Background(),
		Basepath:             "/tmp/alice",
		MoneroWalletEndpoint: common.DefaultAliceMoneroEndpoint,
		EthereumEndpoint:     common.DefaultEthEndpoint,
		EthereumPrivateKey:   common.DefaultPrivKeyAlice,
		Environment:          common.Development,
		ChainID:              common.MainnetConfig.EthereumChainID,
	}

	alice, err := NewAlice(cfg)
	require.NoError(t, err)
	swapState, err := newSwapState(alice, common.NewEtherAmount(1), common.MoneroAmount(1))
	require.NoError(t, err)
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

	bobPrivKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	msg = &net.SendKeysMessage{
		PublicSpendKey: bobPrivKeys.SpendKey().Public().Hex(),
		PrivateViewKey: bobPrivKeys.ViewKey().Hex(),
		EthAddress:     "0x",
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, time.Second*time.Duration(defaultTimeoutDuration.Int64()), s.t1.Sub(s.t0))
	require.Equal(t, bobPrivKeys.SpendKey().Public().Hex(), s.bobPublicSpendKey.Hex())
	require.Equal(t, bobPrivKeys.ViewKey().Hex(), s.bobPrivateViewKey.Hex())
}

// test the case where Alice deploys and locks her eth, but Bob never locks his monero.
// Alice should call refund before the timeout t0.
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

	bobPrivKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	msg := &net.SendKeysMessage{
		PublicSpendKey: bobPrivKeys.SpendKey().Public().Hex(),
		PrivateViewKey: bobPrivKeys.ViewKey().Hex(),
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

	cdMsg, ok := resp.(*net.NotifyContractDeployed)
	require.True(t, ok)

	// ensure we refund before t0
	time.Sleep(time.Second * 15)
	require.NotNil(t, s.net.(*mockNet).msg)
	require.Equal(t, net.NotifyRefundType, s.net.(*mockNet).msg.Type())

	// check balance of contract is 0
	balance, err := s.alice.ethClient.BalanceAt(s.ctx, ethcommon.HexToAddress(cdMsg.Address), nil)
	require.NoError(t, err)
	require.Equal(t, int64(0), balance.Int64())
}

func TestSwapState_NotifyXMRLock(t *testing.T) {
	_, s := newTestAlice(t)
	defer s.cancel()
	s.nextExpectedMessage = &net.NotifyXMRLock{}

	_, err := s.generateKeys()
	require.NoError(t, err)

	bobPrivKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobPrivKeys.SpendKey().Public(), bobPrivKeys.ViewKey())

	_, err = s.deployAndLockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	s.desiredAmount = 0
	kp := mcrypto.SumSpendAndViewKeys(bobPrivKeys.PublicKeyPair(), s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	msg := &net.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyReadyType, resp.Type())
}

// test the case where the monero is locked, but Bob never claims.
// Alice should call refund after the timeout t1.
func TestSwapState_NotifyXMRLock_Refund(t *testing.T) {
	_, s := newTestAlice(t)
	defer s.cancel()
	s.net = new(mockNet)
	s.nextExpectedMessage = &net.NotifyXMRLock{}

	// set timeout to 2s
	// TODO: pass this as a param to newSwapState
	defaultTimeoutDuration = big.NewInt(3)
	defer func() {
		defaultTimeoutDuration = big.NewInt(60 * 60 * 24)
	}()

	_, err := s.generateKeys()
	require.NoError(t, err)

	bobPrivKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobPrivKeys.SpendKey().Public(), bobPrivKeys.ViewKey())

	contractAddr, err := s.deployAndLockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	s.desiredAmount = 0
	kp := mcrypto.SumSpendAndViewKeys(bobPrivKeys.PublicKeyPair(), s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	msg := &net.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyReadyType, resp.Type())

	_, ok := resp.(*net.NotifyReady)
	require.True(t, ok)

	time.Sleep(time.Second * 25)
	require.NotNil(t, s.net.(*mockNet).msg)
	require.Equal(t, net.NotifyRefundType, s.net.(*mockNet).msg.Type())

	// check balance of contract is 0
	balance, err := s.alice.ethClient.BalanceAt(s.ctx, contractAddr, nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), balance.Uint64())
}

func TestSwapState_NotifyClaimed(t *testing.T) {
	t.Skip() // TODO: fix this, fails saying the wallet doesn't have balance

	_, s := newTestAlice(t)
	defer s.cancel()

	msg := &net.SendKeysMessage{}
	_, _, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingKeys, err)

	_, err = s.generateKeys()
	require.NoError(t, err)

	// simulate bob sending his keys
	bobPrivKeys, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	pub := s.alice.ethPrivKey.Public().(*ecdsa.PublicKey)
	address := ethcrypto.PubkeyToAddress(*pub)

	msg = &net.SendKeysMessage{
		PublicSpendKey: bobPrivKeys.SpendKey().Public().Hex(),
		PrivateViewKey: bobPrivKeys.ViewKey().Hex(),
		EthAddress:     address.String(),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, time.Second*time.Duration(defaultTimeoutDuration.Int64()), s.t1.Sub(s.t0))
	require.Equal(t, bobPrivKeys.SpendKey().Public().Hex(), s.bobPublicSpendKey.Hex())
	require.Equal(t, bobPrivKeys.ViewKey().Hex(), s.bobPrivateViewKey.Hex())

	viewKey := mcrypto.SumPrivateViewKeys(bobPrivKeys.ViewKey(), s.privkeys.ViewKey())
	t.Log(viewKey.Hex())

	// simulate bob locking xmr
	bobAddr, err := s.alice.client.GetAddress(0)
	require.NoError(t, err)
	t.Log(bobAddr)

	// mine some blocks to get xmr first
	daemonClient := monero.NewClient(common.DefaultMoneroDaemonEndpoint)
	_ = daemonClient.GenerateBlocks(bobAddr.Address, 257)

	s.desiredAmount = 33333
	kp := mcrypto.SumSpendAndViewKeys(bobPrivKeys.PublicKeyPair(), s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	// lock xmr
	_, err = s.alice.client.Transfer(xmrAddr, 0, uint(s.desiredAmount))
	require.NoError(t, err)
	t.Log("transferred to account", xmrAddr)

	_ = daemonClient.GenerateBlocks(bobAddr.Address, 16)

	err = s.alice.client.Refresh()
	require.NoError(t, err)

	_ = s.alice.client.CloseWallet()

	lmsg := &net.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err = s.HandleProtocolMessage(lmsg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, net.NotifyReadyType, resp.Type())

	err = daemonClient.GenerateBlocks(bobAddr.Address, 1)
	require.NoError(t, err)

	// simulate bob calling claim
	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing Bob's secret spend key
	secret := bobPrivKeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.contract.Claim(s.txOpts, sc)
	require.NoError(t, err)

	cmsg := &net.NotifyClaimed{
		TxHash: tx.Hash().String(),
	}

	resp, done, err = s.HandleProtocolMessage(cmsg)
	require.NoError(t, err)
	require.True(t, done)
	require.Nil(t, resp)

	// check that wallet was generated
}
