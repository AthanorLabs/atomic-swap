package alice

import (
	"errors"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"
)

func (a *alice) Provides() common.ProvidesCoin {
	return common.ProvidesETH
}

// InitiateProtocol is called when an RPC call is made from the user to initiate a swap.
func (a *alice) InitiateProtocol(providesAmount, desiredAmount uint64) (net.SwapState, error) {
	if err := a.initiate(providesAmount, desiredAmount); err != nil {
		return nil, err
	}

	return a.swapState, nil
}

// HandleInitiateMessage is called when we receive a network message from a peer that they wish to initiate a swap.
func (a *alice) HandleInitiateMessage(msg *net.InitiateMessage) (net.SwapState, net.Message, error) {
	if msg.Provides != common.ProvidesXMR {
		return nil, nil, errors.New("peer does not provide XMR")
	}

	// TODO: notify the user via the CLI/websockets that someone wishes to initiate a swap with them.

	// the other party initiated, saying what they will provide and what they desire.
	// we initiate our protocol, saying we will provide what they desire and vice versa.
	if err := a.initiate(msg.DesiredAmount, msg.ProvidesAmount); err != nil {
		return nil, nil, err
	}

	resp, err := a.swapState.handleSendKeysMessage(msg.SendKeysMessage)
	if err != nil {
		return nil, nil, err
	}

	return a.swapState, resp, nil
}

func (a *alice) initiate(providesAmount, desiredAmount uint64) error {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	if a.swapState != nil {
		return errors.New("protocol already in progress")
	}

	balance, err := a.ethClient.BalanceAt(a.ctx, a.auth.From, nil)
	if err != nil {
		return err
	}

	// check user's balance and that they actualy have what they will provide
	if balance.Uint64() <= providesAmount {
		return errors.New("balance lower than amount to be provided")
	}

	a.swapState = newSwapState(a, providesAmount, desiredAmount)
	return nil
}
