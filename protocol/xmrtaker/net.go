package xmrtaker

import (
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	pcommon "github.com/noot/atomic-swap/protocol"

	"github.com/fatih/color" //nolint:misspell
)

// Provides returns types.ProvidesETH
func (a *Instance) Provides() types.ProvidesCoin {
	return types.ProvidesETH
}

// InitiateProtocol is called when an RPC call is made from the user to initiate a swap.
// The input units are ether that we will provide.
func (a *Instance) InitiateProtocol(providesAmount float64, offer *types.Offer) (common.SwapState, error) {
	receivedAmount := offer.ExchangeRate.ToXMR(providesAmount)
	err := a.initiate(common.EtherToWei(providesAmount), common.MoneroToPiconero(receivedAmount),
		offer.ExchangeRate)
	if err != nil {
		return nil, err
	}

	return a.swapState, nil
}

func (a *Instance) initiate(providesAmount common.EtherAmount, receivedAmount common.MoneroAmount,
	exchangeRate types.ExchangeRate) error {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	if a.swapState != nil {
		return errProtocolAlreadyInProgress
	}

	balance, err := a.backend.BalanceAt(a.backend.Ctx(), a.backend.EthAddress(), nil)
	if err != nil {
		return err
	}

	// check user's balance and that they actually have what they will provide
	if balance.Cmp(providesAmount.BigInt()) <= 0 {
		return errBalanceTooLow
	}

	a.swapState, err = newSwapState(a.backend, pcommon.GetSwapInfoFilepath(a.basepath), a.transferBack,
		providesAmount, receivedAmount, exchangeRate)
	if err != nil {
		return err
	}

	go func() {
		<-a.swapState.done
		a.swapState = nil
	}()

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%d**", a.swapState.info.ID()))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	return nil
}
