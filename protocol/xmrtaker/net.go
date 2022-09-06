package xmrtaker

import (
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

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
	state, err := a.initiate(common.EtherToWei(providesAmount), common.MoneroToPiconero(receivedAmount),
		offer.ExchangeRate, offer.EthAsset, offer.GetID())
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (a *Instance) initiate(providesAmount common.EtherAmount, receivedAmount common.MoneroAmount,
	exchangeRate types.ExchangeRate, ethAsset types.EthAsset, offerID types.Hash) (*swapState, error) {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	if a.swapStates[offerID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	balance, err := a.backend.BalanceAt(a.backend.Ctx(), a.backend.EthAddress(), nil)
	if err != nil {
		return nil, err
	}

	// check user's balance and that they actually have what they will provide
	if balance.Cmp(providesAmount.BigInt()) <= 0 {
		return nil, errBalanceTooLow
	}

	s, err := newSwapState(a.backend, offerID, pcommon.GetSwapInfoFilepath(a.dataDir), a.transferBack, providesAmount, receivedAmount, exchangeRate, ethAsset)
	if err != nil {
		return nil, err
	}

	go func() {
		<-s.done
		a.swapMu.Lock()
		defer a.swapMu.Unlock()
		delete(a.swapStates, offerID)
	}()

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%s**", s.info.ID()))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	a.swapStates[offerID] = s
	return s, nil
}
