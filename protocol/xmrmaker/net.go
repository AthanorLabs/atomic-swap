package xmrmaker

import (
	"fmt"
	"math/big"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/fatih/color" //nolint:misspell
)

// EthereumAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthereumAssetAmount interface {
	BigInt() *big.Int
	AsStandard() float64
}

// Provides returns types.ProvidesXMR
func (b *Instance) Provides() types.ProvidesCoin {
	return types.ProvidesXMR
}

func (b *Instance) initiate(
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	providesAmount common.MoneroAmount,
	desiredAmount EthereumAssetAmount,
) (*swapState, error) {
	b.swapMu.Lock()
	defer b.swapMu.Unlock()

	if b.swapStates[offer.ID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	balance, err := b.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	// check user's balance and that they actually have what they will provide
	if balance.UnlockedBalance <= uint64(providesAmount) {
		return nil, errBalanceTooLow{
			unlockedBalance: common.MoneroAmount(balance.UnlockedBalance).AsMonero(),
			providedAmount:  providesAmount.AsMonero(),
		}
	}

	// checks passed, delete offer for now
	b.offerManager.DeleteOffer(offer.ID)

	s, err := newSwapState(b.backend, offer, offerExtra, b.offerManager, providesAmount, desiredAmount)
	if err != nil {
		return nil, err
	}

	go func() {
		<-s.done
		b.swapMu.Lock()
		defer b.swapMu.Unlock()
		delete(b.swapStates, offer.ID)
	}()

	symbol, err := pcommon.AssetSymbol(b.backend, offer.EthAsset)
	if err != nil {
		return nil, err
	}

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%s**", s.info.ID))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	log.Infof(color.New(color.Bold).Sprintf("receiving %v %s for %v XMR",
		s.info.ReceivedAmount,
		symbol,
		s.info.ProvidedAmount),
	)
	b.swapStates[offer.ID] = s
	return s, nil
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
	if types.IsHashZero(id) {
		return nil, nil, errOfferIDNotSet
	}

	offer, offerExtra, err := b.offerManager.GetOffer(id)
	if err != nil {
		return nil, nil, err
	}

	providedAmount := offer.ExchangeRate.ToXMR(msg.ProvidedAmount)

	if providedAmount < offer.MinimumAmount {
		return nil, nil, errAmountProvidedTooLow{providedAmount, offer.MinimumAmount}
	}

	if providedAmount > offer.MaximumAmount {
		return nil, nil, errAmountProvidedTooHigh{providedAmount, offer.MaximumAmount}
	}

	providedPicoXMR := common.MoneroToPiconero(providedAmount)

	// check decimals if ERC20
	// note: this is our counterparty's provided amount, ie. how much we're receiving
	var receivedAmount EthereumAssetAmount
	if offer.EthAsset != types.EthAssetETH {
		_, _, decimals, err := b.backend.ETHClient().ERC20Info(b.backend.Ctx(), offer.EthAsset.Address()) //nolint:govet
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		receivedAmount = common.NewERC20TokenAmountFromDecimals(msg.ProvidedAmount, float64(decimals))
	} else {
		receivedAmount = common.EtherToWei(msg.ProvidedAmount)
	}

	state, err := b.initiate(offer, offerExtra, providedPicoXMR, receivedAmount)
	if err != nil {
		return nil, nil, err
	}

	if err = state.handleSendKeysMessage(msg); err != nil {
		return nil, nil, err
	}

	resp, err := state.SendKeysMessage()
	if err != nil {
		return nil, nil, err
	}

	return state, resp, nil
}
