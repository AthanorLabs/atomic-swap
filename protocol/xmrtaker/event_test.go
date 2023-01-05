package xmrtaker

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/cockroachdb/apd/v3"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/ethereum/watcher"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
)

func lockXMRAndCheckForReadyLog(t *testing.T, s *swapState, xmrAddr mcrypto.Address) {
	// backend simulates the xmrmaker's instance
	backend := newBackend(t)
	err := backend.XMRClient().CreateWallet("test-wallet", "")
	require.NoError(t, err)
	one := apd.New(1, 0)
	monero.MineMinXMRBalance(t, backend.XMRClient(), coins.MoneroToPiconero(one))

	// lock xmr
	amt := coins.NewPiconeroAmount(1000000000)
	tResp, err := backend.XMRClient().Transfer(xmrAddr, 0, amt)
	require.NoError(t, err)
	t.Logf("transferred %d pico XMR (fees %d) to account %s", tResp.Amount, tResp.Fee, xmrAddr)
	require.Equal(t, amt.String(), fmt.Sprintf("%d", tResp.Amount))

	transfer, err := backend.XMRClient().WaitForReceipt(&monero.WaitForReceiptRequest{
		Ctx:              s.ctx,
		TxID:             tResp.TxHash,
		DestAddr:         xmrAddr,
		NumConfirmations: monero.MinSpendConfirmations,
		AccountIdx:       0,
	})
	require.NoError(t, err)
	t.Logf("Transfer mined at block=%d with %d confirmations", transfer.Height, transfer.Confirmations)

	// send notification that monero was locked
	lmsg := &message.NotifyXMRLock{
		Address: string(xmrAddr),
		TxID:    transfer.TxID,
	}

	// assert that ready() is called, setup contract watcher
	ethHeader, err := backend.ETHClient().Raw().HeaderByNumber(backend.Ctx(), nil)
	require.NoError(t, err)
	logReadyCh := make(chan ethtypes.Log)

	readyTopic := common.GetTopic(common.ReadyEventSignature)
	readyWatcher := watcher.NewEventFilter(
		s.Backend.Ctx(),
		s.Backend.ETHClient().Raw(),
		s.Backend.ContractAddr(),
		ethHeader.Number,
		readyTopic,
		logReadyCh,
	)
	err = readyWatcher.Start()
	require.NoError(t, err)

	// now handle the NotifyXMRLock message
	err = s.HandleProtocolMessage(lmsg)
	require.NoError(t, err)
	require.Equal(t, s.nextExpectedEvent, EventETHClaimedType)
	require.Equal(t, types.ContractReady, s.info.Status)

	select {
	case log := <-logReadyCh:
		err = pcommon.CheckSwapID(&log, "Ready", s.contractSwapID)
		require.NoError(t, err)
	case <-time.After(time.Second * 2):
		t.Fatalf("didn't get ready logs in time")
	}
}

func TestSwapState_handleEvent_EventETHClaimed(t *testing.T) {
	s := newTestSwapState(t)
	defer s.cancel()
	s.SetSwapTimeout(time.Minute * 2)

	// invalid SendKeysMessage should result in an error
	msg := &net.SendKeysMessage{ProvidedAmount: new(apd.Decimal)}
	err := s.HandleProtocolMessage(msg)
	require.True(t, errors.Is(err, errMissingKeys))

	// handle valid SendKeysMessage
	msg = s.SendKeysMessage()
	msg.ProvidedAmount = apd.New(1, -0)
	msg.PrivateViewKey = s.privkeys.ViewKey().Hex()
	msg.EthAddress = s.ETHClient().Address().String()

	err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)

	resp := s.Net().(*mockNet).LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyETHLockedType, resp.Type())
	require.Equal(t, time.Minute*2, s.t1.Sub(s.t0))
	require.Equal(t, msg.PublicSpendKey, s.xmrmakerPublicSpendKey.Hex())
	require.Equal(t, msg.PrivateViewKey, s.xmrmakerPrivateViewKey.Hex())

	// simulate xmrmaker locking xmr
	kp := mcrypto.SumSpendAndViewKeys(s.pubkeys, s.pubkeys)
	xmrAddr := kp.Address(common.Mainnet)
	lockXMRAndCheckForReadyLog(t, s, xmrAddr)

	// simulate xmrmaker calling claim
	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	secret := s.privkeys.SpendKeyBytes()
	sk, err := mcrypto.NewPrivateSpendKey(secret[:])
	require.NoError(t, err)

	// handled the claimed message should result in the monero wallet being created
	event := newEventETHClaimed(sk)
	s.eventCh <- event
	err = <-event.errCh
	require.NoError(t, err)
	require.Equal(t, types.CompletedSuccess, s.info.Status)
}
