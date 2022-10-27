package xmrtaker

import (
	"context"
	"encoding/hex"
	"math/big"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	logging "github.com/ipfs/go-log"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	pswap "github.com/athanorlabs/atomic-swap/protocol/swap"
	"github.com/athanorlabs/atomic-swap/tests"
)

var infofile = os.TempDir() + "/test.keys"

var _ = logging.SetLogLevel("xmrtaker", "debug")

type mockNet struct {
	msgMu sync.Mutex  // lock needed, as SendSwapMessage is called async from timeout handlers
	msg   net.Message // last value passed to SendSwapMessage
}

func (n *mockNet) LastSentMessage() net.Message {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	return n.msg
}

func (n *mockNet) SendSwapMessage(msg net.Message, _ types.Hash) error {
	n.msgMu.Lock()
	defer n.msgMu.Unlock()
	n.msg = msg
	return nil
}

func newBackend(t *testing.T) backend.Backend {
	pk := tests.GetTakerTestKey(t)
	ec, chainID := tests.NewEthClient(t)
	ctx := context.Background()

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	var forwarderAddress ethcommon.Address
	_, tx, contract, err := contracts.DeploySwapFactory(txOpts, ec, forwarderAddress)
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(ctx, ec, tx)
	require.NoError(t, err)

	bcfg := &backend.Config{
		Ctx:                 context.Background(),
		MoneroClient:        monero.CreateWalletClient(t),
		EthereumClient:      ec,
		EthereumPrivateKey:  pk,
		Environment:         common.Development,
		SwapManager:         pswap.NewManager(),
		SwapContract:        contract,
		SwapContractAddress: addr,
		Net:                 new(mockNet),
	}

	b, err := backend.NewBackend(bcfg)
	require.NoError(t, err)
	return b
}

func newTestInstance(t *testing.T) *swapState {
	b := newBackend(t)
	swapState, err := newSwapState(b, types.Hash{}, infofile, false,
		common.NewEtherAmount(1), common.MoneroAmount(0), 1, types.EthAssetETH)
	require.NoError(t, err)
	return swapState
}

func newTestInstanceWithERC20(t *testing.T, initialBalance *big.Int) (*swapState, *contracts.ERC20Mock) {
	b := newBackend(t)

	txOpts, err := b.TxOpts()
	require.NoError(t, err)

	_, tx, contract, err := contracts.DeployERC20Mock(
		txOpts,
		b.EthClient(),
		"Mock",
		"MOCK",
		b.EthAddress(),
		initialBalance,
	)
	require.NoError(t, err)
	addr, err := bind.WaitDeployed(b.Ctx(), b.EthClient(), tx)
	require.NoError(t, err)

	swapState, err := newSwapState(b, types.Hash{}, infofile, false,
		common.NewEtherAmount(1), common.MoneroAmount(0), 1, types.EthAsset(addr))
	require.NoError(t, err)
	return swapState, contract
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
	s := newTestInstance(t)
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
	s := newTestInstance(t)
	defer s.cancel()

	s.SetSwapTimeout(time.Second * 15)

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
	lastMesg := s.Net().(*mockNet).LastSentMessage()
	require.NotNil(t, lastMesg)
	require.Equal(t, message.NotifyRefundType, lastMesg.Type())

	// check swap is marked completed
	stage, err := s.Contract().Swaps(nil, s.contractSwapID)
	require.NoError(t, err)
	require.Equal(t, contracts.StageCompleted, stage)
}

func TestSwapState_NotifyXMRLock(t *testing.T) {
	s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockAsset(common.NewEtherAmount(1))
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
	s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}
	s.SetSwapTimeout(time.Second * 3)

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockAsset(common.NewEtherAmount(1))
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

	sentMsg := s.Net().(*mockNet).LastSentMessage()
	require.NotNil(t, sentMsg)
	require.Equal(t, message.NotifyRefundType, sentMsg.Type())

	// check balance of contract is 0
	balance, err := s.BalanceAt(context.Background(), s.ContractAddr(), nil)
	require.NoError(t, err)
	require.Equal(t, uint64(0), balance.Uint64())
}

func TestSwapState_NotifyClaimed(t *testing.T) {
	s := newTestInstance(t)
	defer s.cancel()
	s.SetSwapTimeout(time.Minute * 2)

	// close swap-deposit-wallet
	backend := newBackend(t)
	err := backend.CreateWallet("test-wallet", "")
	require.NoError(t, err)

	monero.MineMinXMRBalance(t, backend, common.MoneroToPiconero(1))

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
	msg.EthAddress = s.EthAddress().String()

	resp, done, err := s.HandleProtocolMessage(msg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, time.Minute*2, s.t1.Sub(s.t0))
	require.Equal(t, msg.PublicSpendKey, s.xmrmakerPublicSpendKey.Hex())
	require.Equal(t, msg.PrivateViewKey, s.xmrmakerPrivateViewKey.Hex())

	// simulate xmrmaker locking xmr
	amt := common.MoneroAmount(1000000000)
	kp := mcrypto.SumSpendAndViewKeys(s.pubkeys, s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)

	// lock xmr
	tResp, err := backend.Transfer(xmrAddr, 0, uint64(amt))
	require.NoError(t, err)
	t.Logf("transferred %d pico XMR (fees %d) to account %s", tResp.Amount, tResp.Fee, xmrAddr)
	require.Equal(t, uint64(amt), tResp.Amount)

	transfer, err := backend.WaitForTransReceipt(&monero.WaitForReceiptRequest{
		Ctx:              s.ctx,
		TxID:             tResp.TxHash,
		DestAddr:         xmrAddr,
		NumConfirmations: monero.MinSpendConfirmations,
	})
	require.NoError(t, err)
	t.Logf("Transfer mined at block=%d with %d confirmations", transfer.Height, transfer.Confirmations)

	// send notification that monero was locked
	lmsg := &message.NotifyXMRLock{
		Address: string(xmrAddr),
		TxID:    transfer.TxID,
	}

	resp, done, err = s.HandleProtocolMessage(lmsg)
	require.NoError(t, err)
	require.False(t, done)
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyReadyType, resp.Type())

	// simulate xmrmaker calling claim
	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	secret := s.privkeys.SpendKeyBytes()
	var sc [32]byte
	copy(sc[:], common.Reverse(secret))

	txOpts, err := s.TxOpts()
	require.NoError(t, err)
	tx, err := s.Contract().Claim(txOpts, s.contractSwap, sc)
	require.NoError(t, err)
	tests.MineTransaction(t, s, tx)

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
	s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.SendKeysMessage{}
	err := s.Exit()
	require.NoError(t, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedAbort, info.Status())
}

func TestExit_afterNotifyXMRLock(t *testing.T) {
	s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyXMRLock{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockAsset(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedRefund, info.Status())
}

func TestExit_afterNotifyClaimed(t *testing.T) {
	s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyClaimed{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockAsset(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.NoError(t, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedRefund, info.Status())
}

func TestExit_invalidNextMessageType(t *testing.T) {
	// this case shouldn't ever really happen
	s := newTestInstance(t)
	defer s.cancel()
	s.nextExpectedMessage = &message.NotifyETHLocked{}

	err := s.generateAndSetKeys()
	require.NoError(t, err)

	xmrmakerKeysAndProof, err := generateKeys()
	require.NoError(t, err)

	s.setXMRMakerKeys(xmrmakerKeysAndProof.PublicKeyPair.SpendKey(), xmrmakerKeysAndProof.PrivateKeyPair.ViewKey(),
		xmrmakerKeysAndProof.Secp256k1PublicKey)

	_, err = s.lockAsset(common.NewEtherAmount(1))
	require.NoError(t, err)

	err = s.Exit()
	require.Equal(t, errUnexpectedMessageType, err)
	info := s.SwapManager().GetPastSwap(s.info.ID())
	require.Equal(t, types.CompletedAbort, info.Status())
}

func TestSwapState_ApproveToken(t *testing.T) {
	initialBalance := big.NewInt(999999)
	s, contract := newTestInstanceWithERC20(t, initialBalance)
	err := s.approveToken()
	require.NoError(t, err)
	allowance, err := contract.Allowance(&bind.CallOpts{}, s.EthAddress(), s.ContractAddr())
	require.NoError(t, err)
	require.Equal(t, initialBalance, allowance)
}
