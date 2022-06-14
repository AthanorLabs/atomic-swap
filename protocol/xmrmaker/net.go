package xmrmaker

import (
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/common/types"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/net/message"

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

	if b.swapStates[offer.GetID()] != nil {
		return errProtocolAlreadyInProgress
	}

	balance, err := b.backend.GetBalance(0)
	if err != nil {
		return err
	}

	// check user's balance and that they actually have what they will provide
	if balance.UnlockedBalance <= float64(providesAmount) {
		return errBalanceTooLow
	}

	s, err := newSwapState(b.backend, offer, b.offerManager, offerExtra.StatusCh,
		offerExtra.InfoFile, providesAmount, desiredAmount)
	if err != nil {
		return err
	}

	go func() {
		<-s.done
		delete(b.swapStates, offer.GetID())
	}()

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%s**", s.ID()))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	log.Infof(color.New(color.Bold).Sprintf("receiving %v ETH for %v XMR",
		s.info.ReceivedAmount(),
		s.info.ProvidedAmount()),
	)
	b.swapStates[offer.GetID()] = s
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
		return nil, nil, errNoOfferWithID
	}

	providedAmount := offer.ExchangeRate.ToXMR(msg.ProvidedAmount)

	if providedAmount < offer.MinimumAmount {
		return nil, nil, errAmountProvidedTooLow
	}

	if providedAmount > offer.MaximumAmount {
		return nil, nil, errAmountProvidedTooHigh
	}

	if err = b.initiate(offer, offerExtra, common.MoneroToPiconero(providedAmount), common.EtherToWei(msg.ProvidedAmount)); err != nil { //nolint:lll
		return nil, nil, err
	}

	offerID, err := types.HexToHash(msg.OfferID)
	if err != nil {
		return nil, nil, err
	}

	s, has := b.swapStates[offerID]
	if !has {
		panic("did not store swap state in Instance map")
	}

	if err = s.handleSendKeysMessage(msg); err != nil {
		return nil, nil, err
	}

	resp, err := s.SendKeysMessage()
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		s.setNextExpectedMessage(&message.NotifyETHLocked{})
	}()
	return s, resp, nil
}
