// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrtaker

import (
	"github.com/cockroachdb/apd/v3"
	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/fatih/color"
)

// Provides returns types.ProvidesETH
func (inst *Instance) Provides() coins.ProvidesCoin {
	return coins.ProvidesETH
}

// InitiateProtocol is called when an RPC call is made from the user to initiate a swap.
// The input units are ether that we will provide.
func (inst *Instance) InitiateProtocol(
	makerPeerID peer.ID,
	providesAmount *apd.Decimal,
	offer *types.Offer,
) (common.SwapState, error) {
	err := coins.ValidatePositive("providesAmount", coins.NumEtherDecimals, providesAmount)
	if err != nil {
		return nil, err
	}

	expectedAmount, err := offer.ExchangeRate.ToXMR(providesAmount)
	if err != nil {
		return nil, err
	}

	if expectedAmount.Cmp(offer.MinAmount) < 0 {
		return nil, errAmountProvidedTooLow{providesAmount, offer.MinAmount}
	}

	if expectedAmount.Cmp(offer.MaxAmount) > 0 {
		return nil, errAmountProvidedTooHigh{providesAmount, offer.MaxAmount}
	}

	providedAmount, err := pcommon.GetEthAssetAmount(
		inst.backend.Ctx(),
		inst.backend.ETHClient(),
		providesAmount,
		offer.EthAsset,
	)
	if err != nil {
		return nil, err
	}

	state, err := inst.initiate(makerPeerID, providedAmount, coins.MoneroToPiconero(expectedAmount),
		offer.ExchangeRate, offer.EthAsset, offer.ID)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (inst *Instance) initiate(
	makerPeerID peer.ID,
	providesAmount coins.EthAssetAmount,
	expectedAmount *coins.PiconeroAmount,
	exchangeRate *coins.ExchangeRate,
	ethAsset types.EthAsset,
	offerID types.Hash,
) (*swapState, error) {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	if inst.swapStates[offerID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	ethBalance, err := inst.backend.ETHClient().Balance(inst.backend.Ctx())
	if err != nil {
		return nil, err
	}

	// Ensure the user's balance is strictly greater than the amount they will provide
	if ethAsset.IsETH() && ethBalance.Cmp(providesAmount.(*coins.WeiAmount)) <= 0 {
		log.Warnf("Account %s needs additional funds for swap balance=%s ETH providesAmount=%s ETH",
			inst.backend.ETHClient().Address(), ethBalance.AsEtherString(), providesAmount.AsStandard())
		return nil, errAssetBalanceTooLow{
			providedAmount: providesAmount.AsStandard(),
			balance:        ethBalance.AsEther(),
			symbol:         "ETH",
		}
	}

	if ethAsset.IsToken() {
		tokenBalance, err := inst.backend.ETHClient().ERC20Balance(inst.backend.Ctx(), ethAsset.Address()) //nolint:govet
		if err != nil {
			return nil, err
		}

		if tokenBalance.AsStandard().Cmp(providesAmount.AsStandard()) <= 0 {
			return nil, errAssetBalanceTooLow{
				providedAmount: providesAmount.AsStandard(),
				balance:        tokenBalance.AsStandard(),
				symbol:         tokenBalance.StandardSymbol(),
			}
		}
	}

	s, err := newSwapStateFromStart(
		inst.backend,
		makerPeerID,
		offerID,
		inst.noTransferBack,
		providesAmount,
		expectedAmount,
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
