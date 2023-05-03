// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package relayer

import (
	"context"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

func validateClaimRequest(
	ctx context.Context,
	request *message.RelayClaimRequest,
	ec *ethclient.Client,
	ourAddress ethcommon.Address,
	salt [4]byte,
	ourSwapCreatorAddr ethcommon.Address,
) error {
	err := validateClaimValues(ctx, request, ec, ourAddress, salt, ourSwapCreatorAddr)
	if err != nil {
		return err
	}

	return validateClaimSignature(request)
}

// validateClaimValues validates the non-signature aspects of the claim request:
//  1. the claim request's SwapCreator bytecode matches ours
//  2. the swap is for ETH and not an ERC20 token
//  3. the swap value is strictly greater than the relayer fee
//  4. the claim request's relayer hash matches keccak256(ourAddress || salt)
//  5. the relayer fee is greater than or equal the expected relayer fee
func validateClaimValues(
	ctx context.Context,
	request *message.RelayClaimRequest,
	ec *ethclient.Client,
	ourAddress ethcommon.Address,
	salt [4]byte,
	ourSwapCreatorAddr ethcommon.Address,
) error {
	isTakerRelay := request.OfferID != nil

	// Validate the requested SwapCreator contract, if it is not at the same address
	// as our own.
	if request.RelaySwap.SwapCreator != ourSwapCreatorAddr {
		if isTakerRelay {
			return fmt.Errorf("taker claim swap creator mismatch found=%s expected=%s",
				request.RelaySwap.SwapCreator, ourSwapCreatorAddr)
		}
		err := contracts.CheckSwapCreatorContractCode(ctx, ec, request.RelaySwap.SwapCreator)
		if err != nil {
			return err
		}
	}

	asset := types.EthAsset(request.RelaySwap.Swap.Asset)
	if asset != types.EthAssetETH {
		return fmt.Errorf("relaying for ETH Asset %s is not supported", asset)
	}

	// The relayer fee must be strictly less than the swap value
	if coins.RelayerFeeWei.Cmp(request.RelaySwap.Swap.Value) >= 0 {
		return fmt.Errorf("swap value of %s ETH is too low to support %s ETH relayer fee",
			coins.FmtWeiAsETH(request.RelaySwap.Swap.Value), coins.RelayerFeeETH.Text('f'))
	}

	hash := ethcrypto.Keccak256Hash(append(ourAddress.Bytes(), salt[:]...))
	if request.RelaySwap.RelayerHash != hash {
		return fmt.Errorf("relay request payout address hash %s does not match expected (%s)",
			request.RelaySwap.RelayerHash,
			hash,
		)
	}

	// the relayer fee must be greater than or equal the expected relayer fee
	if coins.RelayerFeeWei.Cmp(request.RelaySwap.Fee) > 0 {
		return fmt.Errorf("relayer fee of %s ETH is less than expected %s ETH",
			coins.FmtWeiAsETH(request.RelaySwap.Fee),
			coins.RelayerFeeETH.Text('f'),
		)
	}

	return nil
}

// validateClaimSignature validates the claim signature. It is assumed that the
// request fields have already been validated.
func validateClaimSignature(
	request *message.RelayClaimRequest,
) error {
	msg := request.RelaySwap.Hash()
	var sig [65]byte
	copy(sig[:], request.Signature)
	sig[64] -= 27 // ecrecover requires 0/1 while EVM requires 27/28

	signer, err := ethcrypto.Ecrecover(msg[:], sig[:])
	if err != nil {
		return err
	}

	pubkey, err := ethcrypto.UnmarshalPubkey(signer)
	if err != nil {
		return err
	}

	if ethcrypto.PubkeyToAddress(*pubkey) != request.RelaySwap.Swap.Claimer {
		return fmt.Errorf("signer of message is not swap claimer")
	}

	return nil
}
