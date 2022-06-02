package xmrtaker

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
	"github.com/noot/atomic-swap/protocol/backend"
	pswap "github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"
)

var infofile = os.TempDir() + "/test.keys"

var _ = logging.SetLogLevel("xmrtaker", "debug")

type mockNet struct {
	msg net.Message
}

func (n *mockNet) SendSwapMessage(msg net.Message) error {
	n.msg = msg
	return nil
}

func newTestXMRTaker(t *testing.T) *Instance {
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRTaker)
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(common.MainnetConfig.EthereumChainID))
	require.NoError(t, err)
	addr, _, contract, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	bcfg := &backend.Config{
		Ctx:                  context.Background(),
		MoneroWalletEndpoint: common.DefaultXMRTakerMoneroEndpoint,
		MoneroDaemonEndpoint: common.DefaultXMRMakerMoneroEndpoint,
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          common.Development,
		ChainID:              big.NewInt(common.MainnetConfig.EthereumChainID),
		SwapManager:          pswap.NewManager(),
		SwapContract:         contract,
		SwapContractAddress:  addr,
		Net:                  new(mockNet),
	}

	b, err := backend.NewBackend(bcfg)
	require.NoError(t, err)

	cfg := &Config{
		Backend:  b,
		Basepath: "/tmp/xmrtaker",
	}

	xmrtaker, err := NewInstance(cfg)
	require.NoError(t, err)
	return xmrtaker
}

func newTestInstance(t *testing.T) (*Instance, *swapState) {
	xmrtaker := newTestXMRTaker(t)
	swapState, err := newSwapState(xmrtaker.backend, infofile, false, "", common.NewEtherAmount(1), common.MoneroAmount(0), 1)
	require.NoError(t, err)
	return xmrtaker, swapState
}

func newTestXMRMakerSendKeysMessage(t *testing.T) (*net.SendKeysMessage, *pcommon.KeysAndProof) {
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

	msg, xmrmakerKeysAndProof := newTestXMRMakerSendKeysMessage(t)

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, s.SwapTimeout(), s.t1.Sub(s.t0))
	require.Equal(t, xmrmakerKeysAndProof.PublicKeyPair.SpendKey().Hex(), s.xmrmakerPublicSpendKey.Hex())
	require.Equal(t, xmrmakerKeysAndProof.PrivateKeyPair.ViewKey().Hex(), s.xmrmakerPrivateViewKey.Hex())
}

// test the case where XMRTaker deploys and locks her eth, but XMRMaker never locks his monero.
// XMRTaker should call refund before the timeout t0.
func TestSwapState_HandleProtocolMessage_SendKeysMessage_Refund(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()

	// set timeout to 2s
	s.SetSwapTimeout(time.Second * 2)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	msg, xmrmakerKeysAndProof := newTestXMRMakerSendKeysMessage(t)

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyETHLockedType, resp.Type())
	require.Equal(t, s.SwapTimeout(), s.t1.Sub(s.t0))
	require.Equal(t, xmrmakerKeysAndProof.PublicKeyPair.SpendKey().Hex(), s.xmrmakerPublicSpendKey.Hex())
	require.Equal(t, xmrmakerKeysAndProof.PrivateKeyPair.ViewKey().Hex(), s.xmrmakerPrivateViewKey.Hex())

	for status := range s.statusCh {
		if status == types.CompletedRefund {
			break
		} else if !status.IsOngoing() {
			t.Fatalf("got wrong exit status %s, expected CompletedRefund", status)
		}
	}

	// ensure we refund before t0
	require.NotNil(t, s.Net().(*mockNet).msg)
	require.Equal(t, message.NotifyRefundType, s.Net().(*mockNet).msg.Type())

	// check swap is marked completed
	stage, err := s.Contract().Swaps(s.CallOpts(), s.contractSwapID)
	require.NoError(t, err)
	require.Equal(t, swapfactory.StageCompleted, stage)
}

func TestSwapState_NotifyXMRLock(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	kp := mcrypto.SumSpendAndViewKeys(xmrmakerKeysAndProof.PublicKeyPair, s.pubkeys)
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

// test the case where the monero is locked, but XMRMaker never claims.
// XMRTaker should call refund after the timeout t1.
func TestSwapState_NotifyXMRLock_Refund(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}
	s.SetSwapTimeout(time.Second * 3)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	kp := mcrypto.SumSpendAndViewKeys(xmrmakerKeysAndProof.PublicKeyPair, s.pubkeys)
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

	require.NotNil(t, s.Net().(*mockNet).msg)
	require.Equal(t, message.NotifyRefundType, s.Net().(*mockNet).msg.Type())

	// check balance of contract is 0
	balance, err := s.EthClient().BalanceAt(context.Background(), s.ContractAddr(), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), balance.Uint64())
}

func TestSwapState_NotifyClaimed(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.SetSwapTimeout(time.Minute * 2)

	// // close swap-deposit-wallet
	// _ = s.CloseWallet()

	// s.xmrtaker.client = monero.NewClient(common.DefaultXMRMakerMoneroEndpoint)
	// err := s.OpenWallet("test-wallet", "")
	// require.NoError(t, err)

	// invalid SendKeysMessage should result in an error
	msg := &net.SendKeysMessage{}
	_, _, err := s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingKeys, err)

	err = s.generateAndSetKeys()
	require.NoError(t, err)

	// handle valid SendKeysMessage
	msg, err = s.SendKeysMessage()
	require.NoError(t, err)
	msg.PrivateViewKey = s.privkeys.ViewKey().Hex()
	msg.EthAddress = s.EthAddress().String()

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, time.Minute*2, s.t1.Sub(s.t0))
	require.Equal(t, msg.PublicSpendKey, s.xmrmakerPublicSpendKey.Hex())
	require.Equal(t, msg.PrivateViewKey, s.xmrmakerPrivateViewKey.Hex())

	// simulate xmrmaker locking xmr
	xmrmakerAddr, err := s.GetAddress(0)
	require.NoError(t, err)

	// mine some blocks to get xmr first
	daemonClient := monero.NewClient(common.DefaultMoneroDaemonEndpoint)
	_ = daemonClient.GenerateBlocks(xmrmakerAddr.Address, 60)

	amt := common.MoneroAmount(1000000000)
	kp := mcrypto.SumSpendAndViewKeys(s.pubkeys, s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	// lock xmr
	_, err = s.Transfer(xmrAddr, 0, uint(amt))
	require.NoError(t, err)
	t.Log("transferred to account", xmrAddr)

	_ = daemonClient.GenerateBlocks(xmrmakerAddr.Address, 100)

	// send notification that monero was locked
	lmsg := &message.NotifyXMRLock{
		Address: string(xmrAddr),
	}

	resp, done, err = s.HandleProtocolMessage(lmsg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyReadyType, resp.Type())

	err = daemonClient.GenerateBlocks(xmrmakerAddr.Address, 1)
	require.NoError(t, err)

	// simulate xmrmaker calling claim
	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	secret := s.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	tx, err := s.Contract().Claim(s.txOpts, s.contractSwap, sc)
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
	s.nextExpectedMessage = &message.SendKeysMessage{}
	err := s.Exit()
	require.NoError(t, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedAbort, info.Status())
}

func TestExit_afterNotifyXMRLock(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedRefund, info.Status())
}

func TestExit_afterNotifyClaimed(t *testing.T) {
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyClaimed{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedRefund, info.Status())
}

func TestExit_invalidNextMessageType(t *testing.T) {
	// this case shouldn't ever really happen
	_, s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyETHLocked{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockETH(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.Equal(t, errUnexpectedMessageType, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedAbort, info.Status())
}
