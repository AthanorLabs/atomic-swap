package alice

import (
	"errors"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"

	"github.com/fatih/color" //nolint:misspell
)

// Provides returns common.ProvidesETH
func (a *Instance) Provides() common.ProvidesCoin {
	return common.ProvidesETH
}

// InitiateProtocol is called when an RPC call is made from the user to initiate a swap.
// The input units are ether and monero.
func (a *Instance) InitiateProtocol(providesAmount float64) (net.SwapState, error) {
	if err := a.initiate(common.EtherToWei(providesAmount)); err != nil {
		return nil, err
	}

	return a.swapState, nil
}

func (a *Instance) initiate(providesAmount common.EtherAmount) error {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	if a.swapState != nil {
		return errors.New("protocol already in progress")
	}

	balance, err := a.ethClient.BalanceAt(a.ctx, a.callOpts.From, nil)
	if err != nil {
		return err
	}

	// check user's balance and that they actually have what they will provide
	if balance.Cmp(providesAmount.BigInt()) <= 0 {
		return errors.New("balance lower than amount to be provided")
	}

	a.swapState, err = newSwapState(a, providesAmount)
	if err != nil {
		return err
	}

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%d**", a.swapState.id))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	return nil
}
