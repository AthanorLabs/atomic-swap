package xmrtaker

import (
	"context"
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

// validateMinBalance validates that the Taker has sufficient funds to take an
// offer
func validateMinBalance(
	ctx context.Context,
	ec extethclient.EthClient,
	providesAmt *apd.Decimal,
	asset types.EthAsset,
) error {
	gasPrice, err := ec.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	weiBalance, err := ec.Balance(ctx)
	if err != nil {
		return err
	}

	if asset.IsETH() {
		return validateMinBalForETHSwap(weiBalance, providesAmt, gasPrice)
	}

	tokenBalance, err := ec.ERC20Balance(ctx, asset.Address())
	if err != nil {
		return err
	}

	return validateMinBalForTokenSwap(weiBalance, tokenBalance, providesAmt, gasPrice)
}

// validateMinBalForETHSwap validates that the Taker has sufficient funds to take an
// offer of XMR for ETH
func validateMinBalForETHSwap(weiBalance *coins.WeiAmount, providesAmt *apd.Decimal, gasPriceWei *big.Int) error {
	providedAmtWei := coins.EtherToWei(providesAmt).BigInt()
	neededGas := big.NewInt(contracts.MaxNewSwapETHGas + contracts.MaxSetReadyGas + contracts.MaxRefundETHGas)
	neededWeiForGas := new(big.Int).Mul(neededGas, gasPriceWei)
	neededBalanceWei := new(big.Int).Add(providedAmtWei, neededWeiForGas)

	if weiBalance.BigInt().Cmp(neededBalanceWei) < 0 {
		log.Warnf("Ethereum account needs additional funds, balance=%s ETH, required=%s ETH",
			weiBalance.AsEtherString(), coins.NewWeiAmount(neededBalanceWei).AsEtherString())
		return errETHBalanceTooLow{
			currentBalanceETH:  weiBalance.AsEther(),
			requiredBalanceETH: coins.NewWeiAmount(neededBalanceWei).AsEther(),
		}
	}

	return nil
}

// validateMinBalForTokenSwap validates that the Taker has sufficient tokens for
// to take an XMR->Token swap, and sufficient ETH funds to pay for the gas of
// the swap.
func validateMinBalForTokenSwap(
	weiBalance *coins.WeiAmount,
	tokenBalance *coins.ERC20TokenAmount,
	providesAmt *apd.Decimal, // standard units
	gasPriceWei *big.Int,
) error {
	if tokenBalance.AsStd().Cmp(providesAmt) < 0 {
		return errTokenBalanceTooLow{
			providedAmount: providesAmt,
			tokenBalance:   tokenBalance.AsStd(),
			symbol:         tokenBalance.StdSymbol(),
		}
	}

	// For a token swap, we only need an ETH balance to pay for gas. While we
	// hopefully won't need gas to call refund, we don't want to start the swap
	// if we don't have enough ETH to call it.
	neededGas := big.NewInt(contracts.MaxTokenApproveGas +
		contracts.MaxNewSwapTokenGas +
		contracts.MaxSetReadyGas +
		contracts.MaxRefundTokenGas,
	)
	neededWeiForGas := new(big.Int).Mul(neededGas, gasPriceWei)

	if weiBalance.BigInt().Cmp(neededWeiForGas) < 0 {
		log.Warnf("Ethereum account has infufficient balance to pay for gas, balance=%s ETH, required=%s ETH",
			weiBalance.AsEtherString(), coins.NewWeiAmount(neededWeiForGas).AsEtherString())
		return errETHBalanceTooLow{
			currentBalanceETH:  weiBalance.AsEther(),
			requiredBalanceETH: coins.NewWeiAmount(neededWeiForGas).AsEther(),
		}
	}

	return nil
}
