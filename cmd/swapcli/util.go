package main

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/rpcclient"
)

// _tokenCache should only be directly accessed by lookupToken
var _tokenCache = make(map[ethcommon.Address]*coins.ERC20TokenInfo)

func lookupToken(c *rpcclient.Client, tokenAddr ethcommon.Address) (*coins.ERC20TokenInfo, error) {
	token, ok := _tokenCache[tokenAddr]
	if ok {
		return token, nil
	}

	token, err := c.TokenInfo(tokenAddr)
	if err != nil {
		return nil, err
	}

	_tokenCache[tokenAddr] = token

	return token, nil
}

func ethAssetSymbol(c *rpcclient.Client, ethAsset types.EthAsset) (string, error) {
	if ethAsset.IsETH() {
		return "ETH", nil
	}

	token, err := lookupToken(c, ethAsset.Address())
	if err != nil {
		return "", err
	}

	return token.SanitizedSymbol(), nil
}

// providedAndReceivedSymbols returns our provided asset symbol name followed
// by the counterparty's received asset symbol name.
func providedAndReceivedSymbols(
	c *rpcclient.Client,
	provides coins.ProvidesCoin, // determines whether we are the maker or taker
	ethAsset types.EthAsset, // determines provided or received ETH asset symbol
) (string, string, error) {
	ethAssetSymbol, err := ethAssetSymbol(c, ethAsset)
	if err != nil {
		return "", "", err
	}

	switch provides {
	case coins.ProvidesXMR: // we are the maker
		return "XMR", ethAssetSymbol, nil
	case coins.ProvidesETH: // We are the taker
		return ethAssetSymbol, "XMR", nil
	default:
		return "", "", fmt.Errorf("unhandled provides value %q", provides)
	}
}

func printOffer(c *rpcclient.Client, o *types.Offer, index int, indent string) error {
	if index > 0 {
		fmt.Printf("%s---\n", indent)
	}

	xRate := o.ExchangeRate
	var (
		minTake *apd.Decimal
		maxTake *apd.Decimal
		err     error
	)
	if o.EthAsset.IsETH() {
		minTake, err = xRate.ToETH(o.MinAmount)
		if err != nil {
			return err
		}

		maxTake, err = xRate.ToETH(o.MaxAmount)
		if err != nil {
			return err
		}
	} else {
		token, err := lookupToken(c, o.EthAsset.Address()) //nolint:govet
		if err != nil {
			return err
		}

		minTake, err = xRate.ToERC20Amount(o.MinAmount, token)
		if err != nil {
			return err
		}

		maxTake, err = xRate.ToERC20Amount(o.MaxAmount, token)
		if err != nil {
			return err
		}
	}

	// At the current time, offers always have the "Provides" field set to
	// ProvidesXMR, so the Provides/Takes fields below are always from the
	// perspective of the Maker.
	providedCoin, receivedCoin, err := providedAndReceivedSymbols(c, o.Provides, o.EthAsset)
	if err != nil {
		return err
	}

	fmt.Printf("%sOffer ID: %s\n", indent, o.ID)
	fmt.Printf("%sProvides: %s\n", indent, providedCoin)
	fmt.Printf("%sTakes: %s\n", indent, o.EthAsset)
	if o.EthAsset.IsToken() {
		fmt.Printf("%s       %s (self reported symbol)\n", indent, receivedCoin)
	}
	fmt.Printf("%sExchange Rate: %s %s/%s\n", indent, o.ExchangeRate, o.EthAsset, o.Provides)
	fmt.Printf("%sMaker Min: %s %s\n", indent, o.MinAmount.Text('f'), providedCoin)
	fmt.Printf("%sMaker Max: %s %s\n", indent, o.MaxAmount.Text('f'), providedCoin)
	fmt.Printf("%sTaker Min: %s %s\n", indent, minTake.Text('f'), receivedCoin)
	fmt.Printf("%sTaker Max: %s %s\n", indent, maxTake.Text('f'), receivedCoin)
	return nil
}
