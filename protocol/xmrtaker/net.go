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
		offer.ExchangeRate, offer.GetID())
	if err != nil {
		return nil, err
	}

	return a.swapStates[offer.GetID()], nil
}

func (a *Instance) initiate(providesAmount common.EtherAmount, receivedAmount common.MoneroAmount,
	exchangeRate types.ExchangeRate, offerID types.Hash) error {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	if a.swapStates[offerID] != nil {
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

	s, err := newSwapState(a.backend, pcommon.GetSwapInfoFilepath(a.basepath), a.transferBack,
		providesAmount, receivedAmount, exchangeRate)
	if err != nil {
		return err
	}

	go func() {
		<-s.done
		delete(a.swapStates, offerID)
	}()

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%d**", s.info.ID()))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	a.swapStates[offerID] = s
	return nil
}
