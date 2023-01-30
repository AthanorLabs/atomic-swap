package xmrmaker

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/relayer"
)

var (
	maxRelayerCommissionRate, _, _ = apd.NewFromString("0.1") // 10%
)

// claimFunds redeems XMRMaker's ETH funds by calling Claim() on the contract
func (s *swapState) claimFunds() (ethcommon.Hash, error) {
	var (
		symbol   string
		decimals uint8
		err      error
	)
	if types.EthAsset(s.contractSwap.Asset) != types.EthAssetETH {
		_, symbol, decimals, err = s.ETHClient().ERC20Info(s.ctx, s.contractSwap.Asset)
		if err != nil {
			return ethcommon.Hash{}, fmt.Errorf("failed to get ERC20 info: %w", err)
		}
	}

	ethBalance, err := s.ETHClient().Balance(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		log.Infof("balance before claim: %s ETH", coins.NewWeiAmount(ethBalance).AsEther())
	} else {
		balance, err := s.ETHClient().ERC20Balance(s.ctx, s.contractSwap.Asset) //nolint:govet
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance before claim: %v %s",
			coins.NewERC20TokenAmountFromBigInt(balance, decimals).AsStandard(),
			symbol,
		)
	}

	var (
		txHash ethcommon.Hash
	)

	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	if s.offerExtra.RelayerEndpoint != "" {
		// relayer endpoint is set, claim using relayer
		// TODO: eventually update when relayer discovery is implemented
		txHash, err = s.claimRelayer()
		if err != nil {
			log.Warnf("failed to claim using relayer at %s, trying relayers via p2p: err: %s",
				s.offerExtra.RelayerEndpoint,
				err,
			)
			txHash, err = s.discoverRelayersAndClaim()
		}
	} else if ethBalance.Uint64() == 0 {
		txHash, err = s.discoverRelayersAndClaim()
	} else {
		// claim and wait for tx to be included
		sc := s.getSecret()
		txHash, _, err = s.sender.Claim(s.contractSwap, sc)
	}
	if err != nil {
		return ethcommon.Hash{}, err
	}

	log.Infof("sent claim transaction, tx hash=%s", txHash)

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		balance, err := s.ETHClient().Balance(s.ctx)
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance after claim: %s ETH", coins.NewWeiAmount(balance).AsEther())
	} else {
		balance, err := s.ETHClient().ERC20Balance(s.ctx, s.contractSwap.Asset)
		if err != nil {
			return ethcommon.Hash{}, err
		}

		log.Infof("balance after claim: %s %s",
			coins.NewERC20TokenAmountFromBigInt(balance, decimals).AsStandard(),
			symbol,
		)
	}

	return txHash, nil
}

// discoverRelayersAndClaim discovers available relayers on the network,
func (s *swapState) discoverRelayersAndClaim() (ethcommon.Hash, error) {
	relayers, err := s.Backend.DiscoverRelayers()
	if err != nil {
		return ethcommon.Hash{}, err
	}

	forwarderAddress, err := s.Contract().TrustedForwarder(&bind.CallOpts{})
	if err != nil {
		return ethcommon.Hash{}, err
	}

	calldata, err := getClaimTxCalldata(common.DefaultRelayerCommission, &s.contractSwap, s.getSecret())
	if err != nil {
		return ethcommon.Hash{}, err
	}

	req, err := relayer.CreateSubmitTransactionRequest(
		s.ETHClient().PrivateKey(),
		s.ETHClient().Raw(),
		forwarderAddress,
		s.contractAddr,
		calldata,
	)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	for _, relayer := range relayers {
		resp, err := s.Backend.SubmitTransactionToRelayer(relayer, req)
		if err != nil {
			log.Warnf("failed to submit tx to relayer with peer ID %s, trying next relayer: err: %s", relayer, err)
			continue
		}

		err = waitAndCheck(s.ctx, s.ETHClient().Raw(), resp.TxHash)
		if err != nil {
			log.Warnf("tx %s submitted by relayer with peer ID %s failed, trying next relayer", resp.TxHash, relayer)
			continue
		}

		return resp.TxHash, nil
	}

	return ethcommon.Hash{}, errors.New("failed to submit transaction to any relayer")
}

func (s *swapState) claimRelayer() (ethcommon.Hash, error) {
	return claimRelayer(
		s.Ctx(),
		s.ETHClient().PrivateKey(),
		s.Contract(),
		s.contractAddr,
		s.ETHClient().Raw(),
		s.offerExtra.RelayerEndpoint,
		s.offerExtra.RelayerCommission,
		&s.contractSwap,
		s.getSecret(),
	)
}

// claimRelayer claims the ETH funds via relayer RPC endpoint.
func claimRelayer(
	ctx context.Context,
	sk *ecdsa.PrivateKey,
	contract *contracts.SwapFactory,
	contractAddr ethcommon.Address,
	ec *ethclient.Client,
	relayerEndpoint string,
	relayerCommission *apd.Decimal,
	contractSwap *contracts.SwapFactorySwap,
	secret [32]byte,
) (ethcommon.Hash, error) {
	forwarderAddress, err := contract.TrustedForwarder(&bind.CallOpts{})
	if err != nil {
		return ethcommon.Hash{}, err
	}

	rc, err := relayer.NewClient(sk, ec, relayerEndpoint, forwarderAddress)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	calldata, err := getClaimTxCalldata(relayerCommission, contractSwap, secret)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	txHash, err := rc.SubmitTransaction(contractAddr, calldata)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	err = waitAndCheck(ctx, ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	return txHash, nil
}

func waitAndCheck(ctx context.Context, ec *ethclient.Client, txHash ethcommon.Hash) error {
	// wait for inclusion
	receipt, err := block.WaitForReceipt(ctx, ec, txHash)
	if err != nil {
		return err
	}

	if receipt.Status == 0 {
		return fmt.Errorf("transaction failed")
	}

	if len(receipt.Logs) == 0 {
		return fmt.Errorf("claim transaction had no logs")
	}

	return nil
}

func getClaimTxCalldata(
	relayerCommission *apd.Decimal,
	contractSwap *contracts.SwapFactorySwap,
	secret [32]byte,
) ([]byte, error) {
	abi, err := abi.JSON(strings.NewReader(contracts.SwapFactoryMetaData.ABI))
	if err != nil {
		return nil, err
	}

	feeValue, err := calculateRelayerCommission(contractSwap.Value, relayerCommission)
	if err != nil {
		return nil, err
	}

	calldata, err := abi.Pack("claimRelayer", *contractSwap, secret, feeValue)
	if err != nil {
		return nil, err
	}

	return calldata, nil
}

// calculateRelayerCommission calculates and returns the amount of wei that the relayer
// will receive as commission. The commissionRate is a multiplier (multiply by 100 to get
// the percent) that must be greater than zero and less than or equal to the 10% maximum.
// The 10% max is an arbitrary sanity check and may be adjusted in the future.
func calculateRelayerCommission(swapWeiAmt *big.Int, commissionRate *apd.Decimal) (*big.Int, error) {
	if commissionRate.Cmp(maxRelayerCommissionRate) > 0 {
		return nil, errRelayerCommissionRateTooHigh
	}

	feeValue := new(apd.Decimal)
	_, err := coins.DecimalCtx().Mul(feeValue, coins.NewWeiAmount(swapWeiAmt).Decimal(), commissionRate)
	if err != nil {
		return nil, err
	}

	return coins.ToWeiAmount(feeValue).BigInt(), nil
}
