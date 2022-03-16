package bob

import (
	"errors"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net"

	"github.com/fatih/color" //nolint:misspell
)

// Provides returns types.ProvidesXMR
func (b *Instance) Provides() types.ProvidesCoin {
	return types.ProvidesXMR
}

func (b *Instance) initiate(offer *types.Offer, offerExtra *types.OfferExtra, providesAmount common.MoneroAmount,
	desiredAmount common.EtherAmount) error {
	b.swapMu.Lock()
	defer b.swapMu.Unlock()

	if b.swapState != nil {
		return errors.New("protocol already in progress")
	}

	balance, err := b.client.GetBalance(0)
	if err != nil {
		return err
	}

	// check user's balance and that they actually have what they will provide
	if balance.UnlockedBalance <= float64(providesAmount) {
		return errors.New("balance lower than amount to be provided")
	}

	b.swapState, err = newSwapState(b, offer, offerExtra.StatusCh, offerExtra.InfoFile, providesAmount, desiredAmount)
	if err != nil {
		return err
	}

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%d**", b.swapState.ID()))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	return nil
}

// HandleInitiateMessage is called when we receive a network message from a peer that they wish to initiate a swap.
func (b *Instance) HandleInitiateMessage(msg *net.SendKeysMessage) (net.SwapState, net.Message, error) {
	str := color.New(color.Bold).Sprintf("**incoming take of offer %s with provided amount %v**",
		msg.OfferID,
		msg.ProvidedAmount,
	)
	log.Info(str)

	// get offer and determine expected amount
	id, err := types.HexToHash(msg.OfferID)
	if err != nil {
		return nil, nil, err
	}

	offer, offerExtra := b.offerManager.getAndDeleteOffer(id)
	if offer == nil {
		return nil, nil, errors.New("failed to find offer with given ID")
	}

	providedAmount := offer.ExchangeRate.ToXMR(msg.ProvidedAmount)

	if providedAmount < offer.MinimumAmount {
		return nil, nil, errors.New("amount provided by taker is too low for offer")
	}

	if providedAmount > offer.MaximumAmount {
		return nil, nil, errors.New("amount provided by taker is too low for offer")
	}

	if err = b.initiate(offer, offerExtra, common.MoneroToPiconero(providedAmount), common.EtherToWei(msg.ProvidedAmount)); err != nil { //nolint:lll
		return nil, nil, err
	}

	offerExtra.IDCh <- b.swapState.info.ID()
	close(offerExtra.IDCh)

	if err = b.swapState.handleSendKeysMessage(msg); err != nil {
		return nil, nil, err
	}

	resp, err := b.swapState.SendKeysMessage()
	if err != nil {
		return nil, nil, err
	}

	return b.swapState, resp, nil
}
