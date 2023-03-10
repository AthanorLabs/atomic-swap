package relayer

import (
	"context"
	"math/big"

	rcommon "github.com/athanorlabs/go-relayer/common"
	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	rrelayer "github.com/athanorlabs/go-relayer/relayer"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// ValidateAndSendTransaction sends the relayed transaction to the network if it validates successfully.
func ValidateAndSendTransaction(
	ctx context.Context,
	request *message.RelayClaimRequest,
	ec extethclient.EthClient,
	forwarderAddress ethcommon.Address,
) (*message.RelayClaimResponse, error) {

	err := validateClaimRequest(ctx, request, ec.Raw(), forwarderAddress)
	if err != nil {
		return nil, err
	}

	forwarder, domainSeparator, err := getForwarderAndDomainSeparator(ctx, ec.Raw(), forwarderAddress)
	if err != nil {
		return nil, err
	}

	nonce, err := forwarder.GetNonce(&bind.CallOpts{Context: ctx}, request.Swap.Claimer)
	if err != nil {
		return nil, err
	}

	// The size of request.Secret was vetted when it was deserialized
	secret := (*[32]byte)(request.Secret)

	callData, err := getClaimRelayerTxCalldata(request.RelayerFeeWei, request.Swap, secret)
	if err != nil {
		return nil, err
	}

	transSender, err := rrelayer.NewRelayer(&rrelayer.Config{
		Ctx:       ctx,
		EthClient: ec.Raw(), // TODO: Use flashbots to prevent front-running and reverts?
		Forwarder: gsnforwarder.NewIForwarderWrapped(forwarder),
		Key:       rcommon.NewKeyFromPrivateKey(ec.PrivateKey()),
		ValidateTransactionFunc: func(request *rcommon.SubmitTransactionRequest) error {
			// do nothing, we did validation above
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	// Lock the wallet's nonce until we get a receipt
	ec.Lock()
	defer ec.Unlock()

	resp, err := transSender.SubmitTransaction(&rcommon.SubmitTransactionRequest{
		From:            request.Swap.Claimer,
		To:              request.SFContractAddress,
		Value:           big.NewInt(0),
		Gas:             big.NewInt(200000), // TODO: fetch from ethclient?
		Nonce:           nonce,
		Data:            callData,
		Signature:       request.Signature,
		ValidUntilTime:  big.NewInt(0),
		DomainSeparator: *domainSeparator,
		RequestTypeHash: gsnforwarder.ForwardRequestTypehash,
		SuffixData:      nil,
	})
	if err != nil {
		return nil, err
	}

	_, err = block.WaitForReceipt(ctx, ec.Raw(), resp.TxHash)
	if err != nil {
		return nil, err
	}

	return &message.RelayClaimResponse{TxHash: resp.TxHash}, nil
}
