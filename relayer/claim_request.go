// Package relayer provides libraries for creating and validating relay requests and responses.
package relayer

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// DefaultRelayerFee is the default fee for swap relayers.
// It's set to 0.009 ETH.
var (
	DefaultRelayerFee     = big.NewInt(9e15)
	minAcceptedRelayerFee = DefaultRelayerFee
)

// CreateRelayClaimRequest fills and returns a RelayClaimRequest ready for
// submission to a relayer.
func CreateRelayClaimRequest(
	ctx context.Context,
	claimerEthKey *ecdsa.PrivateKey,
	ec *ethclient.Client,
	relayerFeeWei *big.Int,
	swapFactoryAddress ethcommon.Address,
	forwarderAddress ethcommon.Address,
	swap *contracts.SwapFactorySwap,
	secret *[32]byte,
) (*message.RelayClaimRequest, error) {

	signature, err := createForwarderSignature(
		ctx,
		claimerEthKey,
		ec,
		relayerFeeWei,
		swapFactoryAddress,
		forwarderAddress,
		swap,
		secret,
	)
	if err != nil {
		return nil, err
	}

	return &message.RelayClaimRequest{
		SFContractAddress: swapFactoryAddress,
		RelayerFeeWei:     relayerFeeWei,
		Swap:              swap,
		Secret:            secret[:],
		Signature:         signature,
	}, nil
}
