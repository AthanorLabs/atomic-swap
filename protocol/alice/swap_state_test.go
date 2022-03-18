package alice

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"
	pcommon "github.com/noot/atomic-swap/protocol"
	pswap "github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var infofile = os.TempDir() + "/test.keys"

var _ = logging.SetLogLevel("alice", "debug")

type mockNet struct {
	msg net.Message
}

func (n *mockNet) SendSwapMessage(msg net.Message) error {
	n.msg = msg
	return nil
}

func newTestAlice(t *testing.T) *Instance {
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(common.MainnetConfig.EthereumChainID))
	require.NoError(t, err)
	addr, _, contract, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	cfg := &Config{
		Ctx:                  context.Background(),
		Basepath:             "/tmp/alice",
		MoneroWalletEndpoint: common.DefaultAliceMoneroEndpoint,
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          common.Development,
		ChainID:              big.NewInt(common.MainnetConfig.EthereumChainID),
		SwapManager:          pswap.NewManager(),
		SwapContract:         contract,
		SwapContractAddress:  addr,
	}

	alice, err := NewInstance(cfg)
	require.NoError(t, err)
	return alice
}

func newTestInstance(t *testing.T) (*Instance, *swapState) {
	alice := newTestAlice(t)
	swapState, err := newSwapState(alice, infofile, common.NewEtherAmount(1))
	require.NoError(t, err)
	swapState.info.SetReceivedAmount(1)
	return alice, swapState
}

func newTestBobSendKeysMessage(t *testing.T) (*net.SendKeysMessage, *pcommon.KeysAndProof) {
	keysAndProof, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	msg := &net.SendKeysMessage{
		PublicSpendKey:     keysAndProof.PublicKeyPair.SpendKey().Hex(),
		PrivateViewKey:     keysAndProof.PrivateKeyPair.ViewKey().Hex(),
		DLEqProof:          hex.EncodeToString(keysAndProof.DLEqProof.Proof()),
		Secp256k1PublicKey: keysAndProof.Secp256k1PublicKey.String(),
		EthAddress:         "0x",
	}

	return msg, keysAndProof
}

func TestSwapState_HandleProtocolMessage_SendKeysMessage(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()

	msg := &net.SendKeysMessage{}
	_, _, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingKeys, err)

	err = s.generateAndSetKeys()
	require.NoError(t, err)

	msg, bobKeysAndProof := newTestBobSendKeysMessage(t)

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, time.Second*time.Duration(defaultTimeoutDuration.Int64()), s.t1.Sub(s.t0))
	require.Equal(t, bobKeysAndProof.PublicKeyPair.SpendKey().Hex(), s.bobPublicSpendKey.Hex())
	require.Equal(t, bobKeysAndProof.PrivateKeyPair.ViewKey().Hex(), s.bobPrivateViewKey.Hex())
}

// test the case where Alice deploys and locks her eth, but Bob never locks his monero.
// Alice should call refund before the timeout t0.
func TestSwapState_HandleProtocolMessage_SendKeysMessage_Refund(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.alice.net = new(mockNet)

	// set timeout to 2s
	// TODO: pass this as a param to newSwapState
	defaultTimeoutDuration = big.NewInt(2)
	defer func() {
		defaultTimeoutDuration = big.NewInt(60 * 60 * 24)
	}()

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	msg, bobKeysAndProof := newTestBobSendKeysMessage(t)

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyETHLockedType, resp.Type())
	require.Equal(t, time.Second*time.Duration(defaultTimeoutDuration.Int64()), s.t1.Sub(s.t0))
	require.Equal(t, bobKeysAndProof.PublicKeyPair.SpendKey().Hex(), s.bobPublicSpendKey.Hex())
	require.Equal(t, bobKeysAndProof.PrivateKeyPair.ViewKey().Hex(), s.bobPrivateViewKey.Hex())

	for status := range s.statusCh {
		if status == types.CompletedRefund {
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedRefund", status)
		}
	}

	// ensure we refund before t0
	require.NotNil(t, s.alice.net.(*mockNet).msg)
	require.Equal(t, message.NotifyRefundType, s.alice.net.(*mockNet).msg.Type())

	// check swap is marked completed
	info, err := s.alice.contract.Swaps(s.alice.callOpts, s.contractSwapID)
	require.NoError(t, err)
	require.True(t, info.Completed)
}

func TestSwapState_NotifyXMRLock(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	bobKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobKeysAndProof.PublicKeyPair.SpendKey(), bobKeysAndProof.PrivateKeyPair.ViewKey(),
		bobKeysAndProof.Secp256k1PublicKey)

	err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	s.info.SetReceivedAmount(0)
	kp := mcrypto.SumSpendAndViewKeys(bobKeysAndProof.PublicKeyPair, s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	msg := &message.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyReadyType, resp.Type())
}

// test the case where the monero is locked, but Bob never claims.
// Alice should call refund after the timeout t1.
func TestSwapState_NotifyXMRLock_Refund(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.alice.net = new(mockNet)
	s.nextExpectedMessage = &message.NotifyXMRLock{}

	// set timeout to 2s
	// TODO: pass this as a param to newSwapState
	defaultTimeoutDuration = big.NewInt(3)
	defer func() {
		defaultTimeoutDuration = big.NewInt(60 * 60 * 24)
	}()

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	bobKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobKeysAndProof.PublicKeyPair.SpendKey(), bobKeysAndProof.PrivateKeyPair.ViewKey(),
		bobKeysAndProof.Secp256k1PublicKey)

	err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	s.info.SetReceivedAmount(0)
	kp := mcrypto.SumSpendAndViewKeys(bobKeysAndProof.PublicKeyPair, s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	msg := &message.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyReadyType, resp.Type())

	_, ok := resp.(*message.NotifyReady)
	require.True(t, ok)

	for status := range s.statusCh {
		if status == types.CompletedRefund {
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedRefund", status)
		}
	}

	require.NotNil(t, s.alice.net.(*mockNet).msg)
	require.Equal(t, message.NotifyRefundType, s.alice.net.(*mockNet).msg.Type())

	// check balance of contract is 0
	balance, err := s.alice.ethClient.BalanceAt(context.Background(), s.alice.contractAddr, nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), balance.Uint64())
}

func TestSwapState_NotifyClaimed(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()

	s.alice.client = monero.NewClient(common.DefaultBobMoneroEndpoint)
	err := s.alice.client.OpenWallet("test-wallet", "")
	require.NoError(t, err)

	// invalid SendKeysMessage should result in an error
	msg := &net.SendKeysMessage{}
	_, _, err = s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingKeys, err)

	err = s.generateAndSetKeys()
	require.NoError(t, err)

	// handle valid SendKeysMessage
	msg, err = s.SendKeysMessage()
	require.NoError(t, err)
	msg.PrivateViewKey = s.privkeys.ViewKey().Hex()
	msg.EthAddress = common.EthereumPrivateKeyToAddress(s.alice.ethPrivKey).String()

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, time.Second*time.Duration(defaultTimeoutDuration.Int64()), s.t1.Sub(s.t0))
	require.Equal(t, msg.PublicSpendKey, s.bobPublicSpendKey.Hex())
	require.Equal(t, msg.PrivateViewKey, s.bobPrivateViewKey.Hex())

	// simulate bob locking xmr
	bobAddr, err := s.alice.client.GetAddress(0)
	require.NoError(t, err)

	// mine some blocks to get xmr first
	daemonClient := monero.NewClient(common.DefaultMoneroDaemonEndpoint)
	_ = daemonClient.GenerateBlocks(bobAddr.Address, 60)

	amt := common.MoneroAmount(1)
	s.info.SetReceivedAmount(amt.AsMonero())
	kp := mcrypto.SumSpendAndViewKeys(s.pubkeys, s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	// lock xmr
	_, err = s.alice.client.Transfer(xmrAddr, 0, uint(amt))
	require.NoError(t, err)
	t.Log("transferred to account", xmrAddr)

	_ = daemonClient.GenerateBlocks(bobAddr.Address, 100)

	// send notification that monero was locked
	lmsg := &message.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err = s.HandleProtocolMessage(lmsg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyReadyType, resp.Type())

	err = daemonClient.GenerateBlocks(bobAddr.Address, 1)
	require.NoError(t, err)

	// simulate bob calling claim
	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing Bob's secret spend key
	secret := s.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.alice.contract.Claim(s.txOpts, s.contractSwapID, sc)
	require.NoError(t, err)

	// handled the claimed message should result in the monero wallet being created
	cmsg := &message.NotifyClaimed{
		TxHash: tx.Hash().String(),
	}

	resp, done, err = s.HandleProtocolMessage(cmsg)
	require.NoError(t, err)
	require.True(t, done)
	require.Nil(t, resp)
}

func TestExit_afterSendKeysMessage(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.alice.net = new(mockNet)
	s.nextExpectedMessage = &message.SendKeysMessage{}
	err := s.Exit()
	require.Equal(t, errSwapAborted, err)
	info := s.alice.swapManager.GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedAbort, info.Status())
}

func TestExit_afterNotifyXMRLock(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	bobKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobKeysAndProof.PublicKeyPair.SpendKey(), bobKeysAndProof.PrivateKeyPair.ViewKey(),
		bobKeysAndProof.Secp256k1PublicKey)

	err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)
	info := s.alice.swapManager.GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedRefund, info.Status())
}

func TestExit_afterNotifyClaimed(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyClaimed{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	bobKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobKeysAndProof.PublicKeyPair.SpendKey(), bobKeysAndProof.PrivateKeyPair.ViewKey(),
		bobKeysAndProof.Secp256k1PublicKey)

	err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)
	info := s.alice.swapManager.GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedRefund, info.Status())
}

func TestExit_invalidNextMessageType(t *testing.T) {
	// this case shouldn't ever really happen
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyETHLocked{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	bobKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setBobKeys(bobKeysAndProof.PublicKeyPair.SpendKey(), bobKeysAndProof.PrivateKeyPair.ViewKey(),
		bobKeysAndProof.Secp256k1PublicKey)

	err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.Equal(t, errUnexpectedMessageType, err)
	info := s.alice.swapManager.GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedAbort, info.Status())
}
