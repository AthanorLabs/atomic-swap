package xmrtaker

import (
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/fatih/color"
)

// EthereumAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthereumAssetAmount interface {
	BigInt() *big.Int
	AsStandard() *apd.Decimal
}

// Provides returns types.ProvidesETH
func (inst *Instance) Provides() coins.ProvidesCoin {
	return coins.ProvidesETH
}

// InitiateProtocol is called when an RPC call is made from the user to initiate a swap.
// The input units are ether that we will provide.
func (inst *Instance) InitiateProtocol(providesAmount *apd.Decimal, offer *types.Offer) (common.SwapState, error) {
	log.Infof("InitiateProtocol offer=%s", offer)

	expectedAmount, err := offer.ExchangeRate.ToXMR(providesAmount)
	if err != nil {
		return nil, err
	}
	providedAmount, err := pcommon.GetEthereumAssetAmount(
		inst.backend.Ctx(),
		inst.backend.ETHClient(),
		providesAmount,
		offer.EthAsset,
	)
	if err != nil {
		return nil, err
	}

	state, err := inst.initiate(providedAmount, coins.MoneroToPiconero(expectedAmount),
		offer.ExchangeRate, offer.EthAsset, offer.ID)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (inst *Instance) initiate(providesAmount EthereumAssetAmount, expectedAmount *coins.PiconeroAmount,
	exchangeRate *coins.ExchangeRate, ethAsset types.EthAsset, offerID types.Hash) (*swapState, error) {
	inst.swapMu.Lock()
	defer inst.swapMu.Unlock()

	if inst.swapStates[offerID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	balance, err := inst.backend.ETHClient().Balance(inst.backend.Ctx())
	if err != nil {
		return nil, err
	}

	// Ensure the user's balance is strictly greater than the amount they will provide
	if ethAsset == types.EthAssetETH && balance.Cmp(providesAmount.BigInt()) <= 0 {
		log.Warnf("Account %s needs additional funds for this transaction", inst.backend.ETHClient().Address())
		return nil, errBalanceTooLow
	}

	if ethAsset != types.EthAssetETH {
		erc20Contract, err := contracts.NewIERC20(ethAsset.Address(), inst.backend.ETHClient().Raw()) //nolint:govet
		if err != nil {
			return nil, err
		}

		balance, err := erc20Contract.BalanceOf(inst.backend.ETHClient().CallOpts(inst.backend.Ctx()), inst.backend.ETHClient().Address()) //nolint:lll
		if err != nil {
			return nil, err
		}

		if balance.Cmp(providesAmount.BigInt()) <= 0 {
			return nil, errBalanceTooLow
		}
	}

	s, err := newSwapStateFromStart(
		inst.backend,
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

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%s**", s.info.ID))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR THE SWAP MAY BE CANCELLED!"))
	inst.swapStates[offerID] = s
	return s, nil
}
