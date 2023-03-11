package relayer

import (
	"context"
	"fmt"
	"math/big"

	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
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
	expectedForwarderAddress ethcommon.Address,
) error {
	err := validateClaimValues(ctx, request, ec, expectedForwarderAddress, MinRelayerFeeWei)
	if err != nil {
		return err
	}

	return validateClaimSignature(ctx, ec, expectedForwarderAddress, request)
}

// validateClaimValues validates the non-signature aspects of the claim request:
//  1. the claim request swap factory contract is byte compatible with ours
//  2. the forwarder in the claim request swap factory contract has an identical
//     address with our forwarder
//  3. the relayer fee is equal to or greater than the passed minFee
//  4. the swap is for ETH and not an ERC20 token
//  5. the swap value is strictly greater than the relayer fee
//  6. TODO: Validate that the swap exists and is in a claimable state?
func validateClaimValues(
	ctx context.Context,
	req *message.RelayClaimRequest,
	ec *ethclient.Client,
	expectedForwarderAddress ethcommon.Address,
	minFee *big.Int,
) error {
	// Validate that the deployed SwapFactory contract has the same bytecode as
	// the one we use. There is a good chance that it has the same exact address
	// as ours, but we check for binary compatibility regardless.
	requestedForwarderAddr, err := contracts.CheckSwapFactoryContractCode(ctx, ec, req.SFContractAddress)
	if err != nil {
		return err
	}

	// The forwarder used must have the same exact address as ours, so we don't
	// need to check the forwarder contract bytecode.
	if requestedForwarderAddr != expectedForwarderAddress {
		return fmt.Errorf("claim request had expected forwarder address: got %s, expected %s",
			requestedForwarderAddr, expectedForwarderAddress)
	}

	// Relayer fee must be greater than or equal to the minimum fee that we accept
	if req.RelayerFeeWei.Cmp(minFee) < 0 {
		return fmt.Errorf("fee too low: got %s ETH, expected minimum %s ETH",
			coins.FmtWeiAsETH(req.RelayerFeeWei), coins.FmtWeiAsETH(minFee))
	}

	asset := types.EthAsset(req.Swap.Asset)
	if asset != types.EthAssetETH {
		return fmt.Errorf("relaying for ETH Asset %s is not supported", asset)
	}

	// The swap value must be strictly greater than the relayer fee
	if req.Swap.Value.Cmp(req.RelayerFeeWei) <= 0 {
		return fmt.Errorf("swap value of %s ETH is too low to support %s ETH relayer fee",
			coins.FmtWeiAsETH(req.Swap.Value), coins.FmtWeiAsETH(req.RelayerFeeWei))
	}

	return nil
}

func validateClaimSignature(
	ctx context.Context,
	ec *ethclient.Client,
	forwarderAddr ethcommon.Address,
	req *message.RelayClaimRequest,
) error {
	callOpts := &bind.CallOpts{Context: ctx}

	forwarder, domainSeparator, err := getForwarderAndDomainSeparator(ctx, ec, forwarderAddr)
	if err != nil {
		return err
	}

	nonce, err := forwarder.GetNonce(callOpts, req.Swap.Claimer)
	if err != nil {
		return err
	}

	secret := (*[32]byte)(req.Secret)

	forwarderRequest, err := createForwarderRequest(
		nonce,
		req.RelayerFeeWei,
		req.SFContractAddress,
		req.Swap,
		secret,
	)
	if err != nil {
		return err
	}

	err = forwarder.Verify(
		callOpts,
		*forwarderRequest,
		*domainSeparator,
		gsnforwarder.ForwardRequestTypehash,
		nil,
		req.Signature,
	)
	if err != nil {
		return fmt.Errorf("failed to verify signature: %w", err)
	}

	return nil
}
