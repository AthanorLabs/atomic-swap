package relayer

import (
	"context"
	"math/big"

	rcommon "github.com/athanorlabs/go-relayer/common"
	"github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	rrelayer "github.com/athanorlabs/go-relayer/relayer"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// SendRelayedTransaction sends the relayed transaction to the network if it validates successfully.
func SendRelayedTransaction(
	ctx context.Context,
	request *message.RelayClaimRequest,
	ec extethclient.EthClient,
	forwarderAddress ethcommon.Address,
	minFee *big.Int,
) (*message.RelayClaimResponse, error) {
	if err := ValidateClaimRequest(ctx, request, ec.Raw(), forwarderAddress, minFee); err != nil {
		return nil, err
	}

	forwarder, err := gsnforwarder.NewIForwarder(forwarderAddress, ec.Raw())
	if err != nil {
		return nil, err
	}

	transSender, err := rrelayer.NewRelayer(&rrelayer.Config{
		Ctx:       ctx,
		EthClient: ec.Raw(), // TODO: Use flashbots to prevent front-running and reverts?
		Forwarder: gsnforwarder.NewIForwarderWrapped(forwarder),
		Key:       rcommon.NewKeyFromPrivateKey(ec.PrivateKey()),
		ValidateTransactionFunc: func(request *rcommon.SubmitTransactionRequest) error {
			// do nothing, we already called ValidateClaimRequest above
			return nil
		},
	})
	if err != nil {
		return nil, err
	}

	ec.Lock()
	defer ec.Unlock()

	resp, err := transSender.SubmitTransaction(&rcommon.SubmitTransactionRequest{
		From:            request.ClaimerAddress,
		To:              request.SFContractAddress,
		Value:           big.NewInt(0),
		Gas:             request.Gas,
		Nonce:           request.Nonce,
		Data:            request.Data,
		Signature:       request.Signature,
		ValidUntilTime:  request.ValidUntilTime,
		DomainSeparator: request.DomainSeparator,
		RequestTypeHash: request.RequestTypeHash,
		SuffixData:      request.SuffixData,
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
