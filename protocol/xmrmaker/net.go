package xmrmaker

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/fatih/color"
)

// EthereumAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthereumAssetAmount interface {
	BigInt() *big.Int
	AsStandard() float64
}

// Provides returns types.ProvidesXMR
func (inst *Instance) Provides() types.ProvidesCoin {
	return types.ProvidesXMR
}

func (inst *Instance) initiate(
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	providesAmount common.PiconeroAmount,
	desiredAmount EthereumAssetAmount,
) (*swapState, error) {
	if inst.swapStates[offer.ID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	balance, err := inst.backend.XMRClient().GetBalance(0)
	if err != nil {
		return nil, err
	}

	// check user's balance and that they actually have what they will provide
	if balance.UnlockedBalance <= uint64(providesAmount) {
		return nil, errBalanceTooLow{
			unlockedBalance: common.PiconeroAmount(balance.UnlockedBalance).AsMonero(),
			providedAmount:  providesAmount.AsMonero(),
		}
	}

	// checks passed, delete offer for now
	inst.offerManager.DeleteOffer(offer.ID)

	s, err := newSwapStateFromStart(
		inst.backend,
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
		return nil, err
	}

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%s**", s.info.ID))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	log.Infof(color.New(color.Bold).Sprintf("receiving %v %s for %v XMR",
		s.info.ReceivedAmount,
		symbol,
		s.info.ProvidedAmount),
	)
	inst.swapStates[offer.ID] = s
	return s, nil
}

// HandleInitiateMessage is called when we receive a network message from a peer that they wish to initiate a swap.
func (inst *Instance) HandleInitiateMessage(msg *net.SendKeysMessage) (net.SwapState, net.Message, error) {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	str := color.New(color.Bold).Sprintf("**incoming take of offer %s with provided amount %v**",
		msg.OfferID,
		msg.ProvidedAmount,
	)
	log.Info(str)

	// get offer and determine expected amount
	if types.IsHashZero(msg.OfferID) {
		return nil, nil, errOfferIDNotSet
	}

	offer, offerExtra, err := inst.offerManager.GetOffer(msg.OfferID)
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
	receivedAmount, err := pcommon.GetEthereumAssetAmount(
		inst.backend.Ctx(),
		inst.backend.ETHClient(),
		msg.ProvidedAmount,
		offer.EthAsset,
	)
	if err != nil {
		return nil, nil, err
	}

	state, err := inst.initiate(offer, offerExtra, providedPicoXMR, receivedAmount)
	if err != nil {
		return nil, nil, err
	}

	if err = state.handleSendKeysMessage(msg); err != nil {
		return nil, nil, err
	}

	resp := state.SendKeysMessage()
	return state, resp, nil
}
