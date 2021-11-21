package bob

import (
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"
)

var nextID uint64 = 0

type swapState struct {
	*bob

	id                            uint64
	providesAmount, desiredAmount uint64

	// our keys for this session
	privkeys *monero.PrivateKeyPair
	pubkeys  *monero.PublicKeyPair

	// swap contract and timeouts in it; set once contract is deployed
	contract     *swap.Swap
	contractAddr ethcommon.Address
	t0, t1       time.Time

	// Alice's keys for this session
	alicePublicKeys *monero.PublicKeyPair

	// next expected network message
	nextExpectedMessage net.Message

	// channels
	readyCh chan struct{}
}

func newSwapState(b *bob, providesAmount, desiredAmount uint64) *swapState {
	s := &swapState{
		bob:                 b,
		id:                  nextID,
		providesAmount:      providesAmount,
		desiredAmount:       desiredAmount,
		nextExpectedMessage: &net.SendKeysMessage{}, // should this be &net.InitiateMessage{}?
		readyCh:             make(chan struct{}),
	}

	nextID++
	return s
}
