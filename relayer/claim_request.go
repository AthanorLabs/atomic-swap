// Package relayer provides libraries for creating and validating relay requests and responses.
package relayer

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/coins"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/net/message"
)

// DefaultRelayerFeeWei is the default fee for swap relayers.
// It's set to 0.009 ETH. Currently, the minimum and default
// values are identical.
var (
	DefaultRelayerFeeWei = big.NewInt(9e15)
	MinRelayerFeeWei     = DefaultRelayerFeeWei
	MinRelayerFeeEth     = coins.NewWeiAmount(MinRelayerFeeWei).AsEther()
)

// CreateRelayClaimRequest fills and returns a RelayClaimRequest ready for
// submission to a relayer.
func CreateRelayClaimRequest(
	ctx context.Context,
	claimerEthKey *ecdsa.PrivateKey,
	ec *ethclient.Client,
	relayerFeeEth *apd.Decimal,
	swapFactoryAddress ethcommon.Address,
	forwarderAddress ethcommon.Address,
	swap *contracts.SwapFactorySwap,
	secret *[32]byte,
) (*message.RelayClaimRequest, error) {

	relayerFeeWei := DefaultRelayerFeeWei
	if relayerFeeEth != nil {
		relayerFeeWei = coins.EtherToWei(relayerFeeEth).BigInt()
	}

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
