package bob

import (
	"errors"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"

	"github.com/fatih/color"
)

func (b *bob) Provides() common.ProvidesCoin {
	return common.ProvidesXMR
}

// InitiateProtocol is called when an RPC call is made from the user to initiate a swap.
func (b *bob) InitiateProtocol(providesAmount, desiredAmount uint64) (net.SwapState, error) {
	if err := b.initiate(providesAmount, desiredAmount); err != nil {
		return nil, err
	}

	return b.swapState, nil
}

func (b *bob) initiate(providesAmount, desiredAmount uint64) error {
	b.swapMu.Lock()
	defer b.swapMu.Unlock()

	if b.swapState != nil {
		return errors.New("protocol already in progress")
	}

	balance, err := b.client.GetBalance(0)
	if err != nil {
		return err
	}

	// check user's balance and that they actualy have what they will provide
	if balance.UnlockedBalance <= float64(providesAmount) {
		return errors.New("balance lower than amount to be provided")
	}

	b.swapState = newSwapState(b, providesAmount, desiredAmount)
	str := color.New(color.Bold).Sprintf("**initiated swap with ID=%d**", b.swapState.id)
	log.Info(str)
	return nil
}

// HandleInitiateMessage is called when we receive a network message from a peer that they wish to initiate a swap.
func (b *bob) HandleInitiateMessage(msg *net.InitiateMessage) (net.SwapState, net.Message, error) {
	if msg.Provides != common.ProvidesETH {
		return nil, nil, errors.New("peer does not provide ETH")
	}

	// TODO: allow user to accept/reject this via RPC
	str := color.New(color.Bold).Sprintf("**incoming swap with want amount %d**", msg.DesiredAmount)
	log.Info(str)

	if err := b.initiate(msg.DesiredAmount, msg.ProvidesAmount); err != nil {
		return nil, nil, err
	}

	if err := b.swapState.handleSendKeysMessage(msg.SendKeysMessage); err != nil {
		return nil, nil, err
	}

	resp, err := b.swapState.SendKeysMessage()
	if err != nil {
		return nil, nil, err
	}

	return b.swapState, resp, nil
}
