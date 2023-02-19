package xmrtaker

import (
	"testing"
	"time"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/ethereum/watcher"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
)

func lockXMRAndCheckForReadyLog(t *testing.T, s *swapState, xmrAddr mcrypto.Address) {
	// backend simulates the xmrmaker's instance
	backend := newBackend(t)
	monero.MineMinXMRBalance(t, backend.XMRClient(), coins.MoneroToPiconero(coins.StrToDecimal("1")))

	amt := s.expectedPiconeroAmount()
	amtu64, err := amt.Uint64()
	require.NoError(t, err)
	// lock xmr
	transfer, err := backend.XMRClient().Transfer(s.ctx, xmrAddr, 0, amt, monero.MinSpendConfirmations)
	require.NoError(t, err)
	require.Equal(t, transfer.Amount, amtu64)
	t.Logf("Transferred %d pico XMR (fees %d) to account %s", transfer.Amount, transfer.Fee, xmrAddr)
	t.Logf("Transfer was mined at block=%d with %d confirmations", transfer.Height, transfer.Confirmations)

	txID, err := types.HexToHash(transfer.TxID)
	require.NoError(t, err)

	// send notification that monero was locked
	lmsg := &message.NotifyXMRLock{
		Address: xmrAddr,
		TxID:    txID,
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
		err = pcommon.CheckSwapID(&log, readyTopic, s.contractSwapID)
		require.NoError(t, err)
	case <-time.After(time.Second * 2):
		t.Fatalf("didn't get ready logs in time")
	}
}

func TestSwapState_handleEvent_EventETHClaimed(t *testing.T) {
	s, net := newTestSwapStateAndNet(t)
	defer s.cancel()
	s.SetSwapTimeout(time.Minute * 2)

	// test transferBack functionality
	s.transferBack = true
	s.Backend.SetBaseXMRDepositAddress(s.Backend.XMRClient().PrimaryAddress())

	// invalid SendKeysMessage should result in an error
	msg := &message.SendKeysMessage{}
	err := s.HandleProtocolMessage(msg)
	require.ErrorIs(t, err, errMissingProvidedAmount)

	// handle valid SendKeysMessage
	msg = s.SendKeysMessage().(*message.SendKeysMessage)
	msg.PrivateViewKey = s.privkeys.ViewKey()
	msg.EthAddress = s.ETHClient().Address()
	msg.ProvidedAmount = s.providedAmount.AsStandard()

	err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)

	resp := net.LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyETHLockedType, resp.Type())
	require.Equal(t, time.Minute*2, s.t1.Sub(s.t0))
	require.Equal(t, msg.PublicSpendKey.Hex(), s.xmrmakerPublicSpendKey.Hex())
	require.Equal(t, msg.PrivateViewKey.Hex(), s.xmrmakerPrivateViewKey.Hex())

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
