// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"fmt"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
	"github.com/athanorlabs/atomic-swap/net/message"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/fatih/color"
)

// Provides returns types.ProvidesXMR
func (inst *Instance) Provides() coins.ProvidesCoin {
	return coins.ProvidesXMR
}

func (inst *Instance) initiate(
	takerPeerID peer.ID,
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	providesAmount *coins.PiconeroAmount,
	desiredAmount coins.EthAssetAmount,
) (*swapState, error) {
	if inst.swapStates[offer.ID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	balance, err := inst.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	// check that the user's monero balance is sufficient for their max swap amount (strictly
	// greater check, since they need to cover chain fees).
	unlockedBal := coins.NewPiconeroAmount(balance.UnlockedBalance)
	if unlockedBal.Decimal().Cmp(providesAmount.Decimal()) <= 0 {
		return nil, errBalanceTooLow{
			unlockedBalance: unlockedBal.AsMonero(),
			providedAmount:  providesAmount.AsMonero(),
		}
	}

	// checks passed, delete the offer from memory for now
	_, _, err = inst.offerManager.TakeOffer(offer.ID)
	if err != nil {
		return nil, err
	}

	s, err := newSwapStateFromStart(
		inst.backend,
		takerPeerID,
		offer,
		offerExtra,
		inst.offerManager,
		providesAmount,
		desiredAmount,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		<-s.done
		inst.swapMu.Lock()
		defer inst.swapMu.Unlock()
		delete(inst.swapStates, offer.ID)
	}()

	symbol, err := pcommon.AssetSymbol(inst.backend, offer.EthAsset)
	if err != nil {
		_ = s.Exit()
		return nil, err
	}

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with offer ID=%s**", s.info.OfferID))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR THE SWAP MAY BE CANCELLED!"))
	log.Infof(color.New(color.Bold).Sprintf("receiving %v %s for %v XMR",
		s.info.ExpectedAmount,
		symbol,
		s.info.ProvidedAmount),
	)
	inst.swapStates[offer.ID] = s
	return s, nil
}

// HandleInitiateMessage is called when we receive a network message from a peer that they wish to initiate a swap.
func (inst *Instance) HandleInitiateMessage(
	takerPeerID peer.ID,
	msg *message.SendKeysMessage,
) (net.SwapState, error) {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	str := color.New(color.Bold).Sprintf("**incoming take of offer %s with provided amount %s**",
		msg.OfferID,
		msg.ProvidedAmount,
	)
	log.Info(str)

	// get offer and determine expected amount
	if types.IsHashZero(msg.OfferID) {
		return nil, errOfferIDNotSet
	}

	// TODO: If this is not ETH, we need quick/easy access to the number
	//       of token decimal places. Should it be in the OfferExtra struct?
	err := coins.ValidatePositive("providedAmount", coins.NumEtherDecimals, msg.ProvidedAmount)
	if err != nil {
		return nil, err
	}

	offer, offerExtra, err := inst.offerManager.GetOffer(msg.OfferID)
	if err != nil {
		return nil, err
	}

	providedAmount, err := offer.ExchangeRate.ToXMR(msg.ProvidedAmount)
	if err != nil {
		return nil, err
	}

	if providedAmount.Cmp(offer.MinAmount) < 0 {
		return nil, errAmountProvidedTooLow{msg.ProvidedAmount, offer.MinAmount}
	}

	if providedAmount.Cmp(offer.MaxAmount) > 0 {
		return nil, errAmountProvidedTooHigh{msg.ProvidedAmount, offer.MaxAmount}
	}

	providedPiconero := coins.MoneroToPiconero(providedAmount)

	// check decimals if ERC20
	// note: this is our counterparty's provided amount, ie. how much we're receiving
	expectedAmount, err := pcommon.GetEthAssetAmount(
		inst.backend.Ctx(),
		inst.backend.ETHClient(),
		msg.ProvidedAmount,
		offer.EthAsset,
	)
	if err != nil {
		return nil, err
	}

	state, err := inst.initiate(takerPeerID, offer, offerExtra, providedPiconero, expectedAmount)
	if err != nil {
		return nil, err
	}

	if err = state.handleSendKeysMessage(msg); err != nil {
		_ = state.Exit()
		return nil, err
	}

	resp := state.SendKeysMessage()
	err = inst.backend.SendSwapMessage(resp, offer.ID)
	if err != nil {
		_ = state.Exit()
		return nil, fmt.Errorf("failed to send SendKeysMessage to remote peer: %w", err)
	}

	return state, nil
}
