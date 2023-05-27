// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/fatih/color"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
)

// Provides returns types.ProvidesETH
func (inst *Instance) Provides() coins.ProvidesCoin {
	return coins.ProvidesETH
}

// InitiateProtocol is called when an RPC call is made from the user to take a swap.
// The input units are ether that we will provide.
func (inst *Instance) InitiateProtocol(
	makerPeerID peer.ID,
	providesAmount *apd.Decimal,
	offer *types.Offer,
) (common.SwapState, error) {
	var maxDecimals uint8 = coins.NumEtherDecimals
	var token *coins.ERC20TokenInfo
	if offer.EthAsset.IsToken() {
		var err error
		token, err = inst.backend.ETHClient().ERC20Info(inst.backend.Ctx(), offer.EthAsset.Address())
		if err != nil {
			return nil, err
		}
		maxDecimals = token.NumDecimals
	}

	err := coins.ValidatePositive("providesAmount", maxDecimals, providesAmount)
	if err != nil {
		return nil, err
	}

	offerMinETH, err := offer.ExchangeRate.ToETH(offer.MinAmount)
	if err != nil {
		return nil, err
	}

	offerMaxETH, err := offer.ExchangeRate.ToETH(offer.MaxAmount)
	if err != nil {
		return nil, err
	}

	if offerMinETH.Cmp(providesAmount) > 0 {
		return nil, errAmountProvidedTooLow{providesAmount, offerMinETH}
	}

	if offerMaxETH.Cmp(providesAmount) < 0 {
		return nil, errAmountProvidedTooHigh{providesAmount, offerMaxETH}
	}

	err = validateMinBalance(
		inst.backend.Ctx(),
		inst.backend.ETHClient(),
		providesAmount,
		offer.EthAsset,
	)
	if err != nil {
		return nil, err
	}

	var providedAssetAmount coins.EthAssetAmount
	if token == nil {
		providedAssetAmount = coins.EtherToWei(providesAmount)
	} else {
		providedAssetAmount = coins.NewTokenAmountFromDecimals(providesAmount, token)
	}

	state, err := inst.initiate(makerPeerID, providedAssetAmount, offer.ExchangeRate, offer.EthAsset, offer.ID)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (inst *Instance) initiate(
	makerPeerID peer.ID,
	providesAmount coins.EthAssetAmount,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	offerID types.Hash,
) (*swapState, error) {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	if inst.swapStates[offerID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	s, err := newSwapStateFromStart(
		inst.backend,
		makerPeerID,
		offerID,
		inst.noTransferBack,
		providesAmount,
		exchangeRate,
		ethAsset,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		<-s.done
		inst.swapMu.Lock()
		defer inst.swapMu.Unlock()
		delete(inst.swapStates, offerID)
	}()

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with offer ID=%s**", s.info.OfferID))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR THE SWAP MAY BE CANCELLED!"))
	inst.swapStates[offerID] = s
	return s, nil
}
