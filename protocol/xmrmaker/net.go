package xmrmaker

import (
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/net"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/fatih/color"
)

// EthereumAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthereumAssetAmount interface {
	BigInt() *big.Int
	AsStandard() *apd.Decimal
}

// Provides returns types.ProvidesXMR
func (inst *Instance) Provides() coins.ProvidesCoin {
	return coins.ProvidesXMR
}

func (inst *Instance) initiate(
	offer *types.Offer,
	offerExtra *types.OfferExtra,
	providesAmount *coins.PiconeroAmount,
	desiredAmount EthereumAssetAmount,
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

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with offer ID=%s**", s.info.ID))
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

	str := color.New(color.Bold).Sprintf("**incoming take of offer %s with provided amount %s**",
		msg.OfferID,
		msg.ProvidedAmount,
	)
	log.Info(str)

	// get offer and determine expected amount
	if types.IsHashZero(msg.OfferID) {
		return nil, nil, errOfferIDNotSet
	}

	if err := coins.ValidatePositive("providedAmount", msg.ProvidedAmount); err != nil {
		return nil, nil, err
	}

	offer, offerExtra, err := inst.offerManager.GetOffer(msg.OfferID)
	if err != nil {
		return nil, nil, err
	}

	providedAmount, err := offer.ExchangeRate.ToXMR(msg.ProvidedAmount)
	if err != nil {
		return nil, nil, err
	}

	if providedAmount.Cmp(offer.MinAmount) < 0 {
		// TODO: This message will be confusing to the end-user, since they provided ETH, not XMR
		return nil, nil, errAmountProvidedTooLow{providedAmount, offer.MinAmount}
	}

	if providedAmount.Cmp(offer.MaxAmount) > 0 {
		return nil, nil, errAmountProvidedTooHigh{providedAmount, offer.MaxAmount}
	}

	providedPiconero := coins.MoneroToPiconero(providedAmount)

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

	state, err := inst.initiate(offer, offerExtra, providedPiconero, receivedAmount)
	if err != nil {
		return nil, nil, err
	}

	if err = state.handleSendKeysMessage(msg); err != nil {
		return nil, nil, err
	}

	resp := state.SendKeysMessage()
	return state, resp, nil
}
