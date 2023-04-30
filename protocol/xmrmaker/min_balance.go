package xmrmaker

import (
	"context"
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
)

// validateMinBalance validates that the Maker has sufficient funds to make an
// XMR for ETH or XMR for token offer.
func validateMinBalance(
	ctx context.Context,
	mc monero.WalletClient,
	ec extethclient.EthClient,
	offerMaxAmt *apd.Decimal,
	ethAsset types.EthAsset,
) error {
	piconeroBalance, err := mc.GetBalance(0)
	if err != nil {
		return err
	}

	// Maker needs a sufficient XMR balance regardless if it is an ETH or toke swap
	unlockedBalance := coins.NewPiconeroAmount(piconeroBalance.UnlockedBalance).AsMonero()
	if unlockedBalance.Cmp(offerMaxAmt) <= 0 {
		return errUnlockedBalanceTooLow{offerMaxAmt, unlockedBalance}
	}

	// For a token swap, we also check if the maker has sufficient ETH funds to make a
	// claim at the end of the swap.
	if ethAsset.IsToken() {
		gasPriceWei, err := ec.SuggestGasPrice(ctx)
		if err != nil {
			return err
		}

		requiredETHToClaimTokens := coins.NewWeiAmount(
			new(big.Int).Mul(big.NewInt(contracts.MaxClaimTokenGas), gasPriceWei),
		).AsEther()

		weiBalance, err := ec.Balance(ctx)
		if err != nil {
			return err
		}

		ethBalance := weiBalance.AsEther()

		if ethBalance.Cmp(requiredETHToClaimTokens) <= 0 {
			log.Warnf("Ethereum account has insufficient funds for token claim, balance=%s ETH, required=%s ETH",
				ethBalance.Text('f'), requiredETHToClaimTokens.Text('f'))
			return errETHBalanceTooLowForTokenSwap{
				ethBalance:         ethBalance,
				requiredETHToClaim: requiredETHToClaimTokens,
			}
		}
	}

	return nil
}
