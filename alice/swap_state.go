package alice

import (
	"time"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"
)

var nextID uint64 = 0

// swapState is an instance of a swap. it holds the info needed for the swap,
// and its current state.
type swapState struct {
	*alice

	id uint64
	// amount of ETH we are providing this swap, and the amount of XMR we should receive.
	providesAmount, desiredAmount uint64

	// our keys for this session
	privkeys *monero.PrivateKeyPair
	pubkeys  *monero.PublicKeyPair

	// Bob's keys for this session
	bobSpendKey *monero.PublicKey
	bobViewKey  *monero.PrivateViewKey

	// swap contract and timeouts in it; set once contract is deployed
	contract *swap.Swap
	t0, t1   time.Time

	// next expected network message
	nextExpectedMessage net.Message // TODO: change to type?

	// channels
	xmrLockedCh chan struct{}
	claimedCh   chan struct{}
}

func newSwapState(a *alice, providesAmount, desiredAmount uint64) *swapState {
	s := &swapState{
		alice:               a,
		id:                  nextID,
		providesAmount:      providesAmount,
		desiredAmount:       desiredAmount,
		nextExpectedMessage: &net.SendKeysMessage{}, // should this be &net.InitiateMessage{}?
		xmrLockedCh:         make(chan struct{}),
		claimedCh:           make(chan struct{}),
	}

	nextID++
	return s
}
