package xmrtaker

import (
	"math/big"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"

	"github.com/fatih/color" //nolint:misspell
)

// EthereumAssetAmount represents an amount of an Ethereum asset (ie. ether or an ERC20)
type EthereumAssetAmount interface {
	BigInt() *big.Int
	AsStandard() float64
}

// Provides returns types.ProvidesETH
func (a *Instance) Provides() types.ProvidesCoin {
	return types.ProvidesETH
}

// InitiateProtocol is called when an RPC call is made from the user to initiate a swap.
// The input units are ether that we will provide.
func (a *Instance) InitiateProtocol(providesAmount float64, offer *types.Offer) (common.SwapState, error) {
	receivedAmount := offer.ExchangeRate.ToXMR(providesAmount)

	providedAmount, err := pcommon.GetEthereumAssetAmount(
		a.backend.Ctx(),
		a.backend.ETHClient(),
		providesAmount,
		offer.EthAsset,
	)
	if err != nil {
		return nil, err
	}

	state, err := a.initiate(providedAmount, common.MoneroToPiconero(receivedAmount),
		offer.ExchangeRate, offer.EthAsset, offer.ID)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (a *Instance) initiate(providesAmount EthereumAssetAmount, receivedAmount common.MoneroAmount,
	exchangeRate types.ExchangeRate, ethAsset types.EthAsset, offerID types.Hash) (*swapState, error) {
	a.swapMu.Lock()
	defer a.swapMu.Unlock()

	if a.swapStates[offerID] != nil {
		return nil, errProtocolAlreadyInProgress
	}

	balance, err := a.backend.ETHClient().Balance(a.backend.Ctx())
	if err != nil {
		return nil, err
	}

	// Ensure the user's balance is strictly greater than the amount they will provide
	if ethAsset == types.EthAssetETH && balance.Cmp(providesAmount.BigInt()) <= 0 {
		log.Warnf("Account %s needs additional funds for this transaction", a.backend.ETHClient().Address())
		return nil, errBalanceTooLow
	}

	if ethAsset != types.EthAssetETH {
		erc20Contract, err := contracts.NewIERC20(ethAsset.Address(), a.backend.ETHClient().Raw()) //nolint:govet
		if err != nil {
			return nil, err
		}

		balance, err := erc20Contract.BalanceOf(a.backend.ETHClient().CallOpts(a.backend.Ctx()), a.backend.ETHClient().Address()) //nolint:lll
		if err != nil {
			return nil, err
		}

		if balance.Cmp(providesAmount.BigInt()) <= 0 {
			return nil, errBalanceTooLow
		}
	}

	s, err := newSwapState(
		a.backend,
		offerID,
		pcommon.GetSwapInfoFilepath(a.dataDir, offerID.String()),
		a.transferBack,
		providesAmount,
		receivedAmount,
		exchangeRate,
		ethAsset,
	)
	if err != nil {
		return nil, err
	}

	go func() {
		<-s.done
		a.swapMu.Lock()
		defer a.swapMu.Unlock()
		delete(a.swapStates, offerID)
	}()

	log.Info(color.New(color.Bold).Sprintf("**initiated swap with ID=%s**", s.info.ID))
	log.Info(color.New(color.Bold).Sprint("DO NOT EXIT THIS PROCESS OR FUNDS MAY BE LOST!"))
	a.swapStates[offerID] = s
	return s, nil
}
