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

	offer, offerExtra, err := inst.offerManager.GetOffer(msg.OfferID)
	if err != nil {
		return nil, err
	}

	var maxDecimals uint8 = coins.NumEtherDecimals
	var token *coins.ERC20TokenInfo
	if offer.EthAsset.IsToken() {
		token, err = inst.backend.ETHClient().ERC20Info(inst.backend.Ctx(), offer.EthAsset.Address())
		if err != nil {
			return nil, err
		}
		maxDecimals = token.NumDecimals
	}

	err = coins.ValidatePositive("providedAmount", maxDecimals, msg.ProvidedAmount)
	if err != nil {
		return nil, err
	}

	expectedAmount := coins.NewEthAssetAmount(msg.ProvidedAmount, token)

	// The calculation below will return an error if the provided amount, when
	// represented in XMR, would require fractional piconeros. This can happen
	// more easily than one might expect, as ToXMR is doing a division by the
	// exchange rate. The taker also verifies that their provided amount will
	// not result in fractional piconeros, so the issue will normally be caught
	// before the taker ever contacts us.
	providedAmtAsXMR, err := offer.ExchangeRate.ToXMR(expectedAmount)
	if err != nil {
		return nil, err
	}

	if providedAmtAsXMR.Cmp(offer.MinAmount) < 0 {
		return nil, errAmountProvidedTooLow{msg.ProvidedAmount, offer.MinAmount}
	}

	if providedAmtAsXMR.Cmp(offer.MaxAmount) > 0 {
		return nil, errAmountProvidedTooHigh{msg.ProvidedAmount, offer.MaxAmount}
	}

	providedPiconero := coins.MoneroToPiconero(providedAmtAsXMR)

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
