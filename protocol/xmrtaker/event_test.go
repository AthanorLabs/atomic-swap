package xmrtaker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func TestSwapState_handleEvent_EventETHClaimed(t *testing.T) {
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
	err = s.HandleProtocolMessage(msg)
	require.Equal(t, errMissingKeys, err)

	err = s.generateAndSetKeys()
	require.NoError(t, err)

	// handle valid SendKeysMessage
	msg, err = s.SendKeysMessage()
	require.NoError(t, err)
	msg.PrivateViewKey = s.privkeys.ViewKey().Hex()
	msg.EthAddress = s.EthAddress().String()

	err = s.HandleProtocolMessage(msg)
	require.NoError(t, err)

	resp := s.Net().(*mockNet).LastSentMessage()
	require.NotNil(t, resp)
	require.Equal(t, message.NotifyETHLockedType, resp.Type())
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
		AccountIdx:       0,
	})
	require.NoError(t, err)
	t.Logf("Transfer mined at block=%d with %d confirmations", transfer.Height, transfer.Confirmations)

	// send notification that monero was locked
	lmsg := &message.NotifyXMRLock{
		Address: string(xmrAddr),
		TxID:    transfer.TxID,
	}

	// TODO assert ready was called
	err = s.HandleProtocolMessage(lmsg)
	require.NoError(t, err)
	require.Equal(t, s.nextExpectedEvent, &EventETHClaimed{})
	require.Equal(t, types.ContractReady, s.info.Status())

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
	require.Equal(t, types.CompletedSuccess, s.info.Status())
}
