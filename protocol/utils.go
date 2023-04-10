// Copyright 2023 Athanor Labs (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package protocol

import (
	"fmt"
	"strconv"

	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/protocol/backend"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

const etherSymbol = "ETH"

// AssetSymbol returns the symbol for the given asset.
func AssetSymbol(b backend.Backend, asset types.EthAsset) (string, error) {
	if asset != types.EthAssetETH {
		tokenInfo, err := b.ETHClient().ERC20Info(b.Ctx(), asset.Address())
		if err != nil {
			return "", fmt.Errorf("failed to get ERC20 info: %w", err)
		}

		return strconv.QuoteToASCII(tokenInfo.Symbol), nil
	}

	return etherSymbol, nil
}

// CheckSwapID checks if the given log is for the given swap ID.
func CheckSwapID(log *ethtypes.Log, eventNameTopic [32]byte, contractSwapID types.Hash) error {
	if len(log.Topics) < 2 {
		return errLogMissingParams
	}

	if log.Topics[0] != eventNameTopic {
		return errInvalidEventTopic
	}

	if log.Topics[1] != contractSwapID {
		return ErrLogNotForUs
	}

	return nil
}
