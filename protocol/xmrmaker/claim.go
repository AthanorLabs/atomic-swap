package xmrmaker

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/relayer"
)

var numEtherUnitsFloat = big.NewFloat(math.Pow(10, 18))

func (s *swapState) tryClaim() (ethcommon.Hash, error) {
	stage, err := s.Contract().Swaps(s.CallOpts(), s.contractSwapID)
	if err != nil {
		return ethcommon.Hash{}, err
	}
	switch stage {
	case contracts.StageInvalid:
		return ethcommon.Hash{}, errClaimInvalid
	case contracts.StageCompleted:
		return ethcommon.Hash{}, errClaimSwapComplete
	case contracts.StagePending, contracts.StageReady:
		// do nothing
	default:
		panic("Unhandled stage value")
	}

	ts, err := s.LatestBlockTimestamp(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	// The block that our claim transaction goes into needs a timestamp that is strictly less
	// than T1. Since the minimum interval between blocks is 1 second, the current block must
	// be at least 2 seconds before T1 for a non-zero chance of the next block having a
	// timestamp that is strictly less than T1.
	if ts.After(s.t1.Add(-2 * time.Second)) {
		// We've passed t1, so the only way we can regain control of the locked XMR is for
		// XMRTaker to call refund on the contract.
		return ethcommon.Hash{}, errClaimPastTime
	}

	if ts.Before(s.t0) && stage != contracts.StageReady {
		// TODO: t0 could be 24 hours from now. Don't we want to poll the stage periodically? (#163)
		// we need to wait until t0 to claim
		log.Infof("waiting until time %s to claim, time now=%s", s.t0, time.Now())
		err = s.WaitForTimestamp(s.ctx, s.t0)
		if err != nil {
			return ethcommon.Hash{}, err
		}
	}

	return s.claimFunds()
}

// claimFunds redeems XMRMaker's ETH funds by calling Claim() on the contract
func (s *swapState) claimFunds() (ethcommon.Hash, error) {
	addr := s.EthAddress()

	var (
		symbol   string
		decimals uint8
		err      error
	)
	if types.EthAsset(s.contractSwap.Asset) != types.EthAssetETH {
		_, symbol, decimals, err = s.ERC20Info(s.ctx, s.contractSwap.Asset)
		if err != nil {
			return ethcommon.Hash{}, fmt.Errorf("failed to get ERC20 info: %w", err)
		}
	}

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		balance, err := s.BalanceAt(s.ctx, addr, nil) //nolint:govet
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance before claim: %v ETH", common.EtherAmount(*balance).AsEther())
	} else {
		balance, err := s.ERC20BalanceAt(s.ctx, s.contractSwap.Asset, addr, nil) //nolint:govet
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance before claim: %v %s", 
			common.NewERC20TokenAmountFromBigInt(balance, float64(decimals)).AsStandard(), 
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
			return ethcommon.Hash{}, err
		}
	} else {
		// claim and wait for tx to be included
		sc := s.getSecret()
		txHash, _, err = s.sender.Claim(s.contractSwap, sc)
		if err != nil {
			return ethcommon.Hash{}, err
		}
	}

	log.Infof("sent claim transaction, tx hash=%s", txHash)

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		balance, err := s.BalanceAt(s.ctx, addr, nil)
		if err != nil {
			return ethcommon.Hash{}, err
		}
		log.Infof("balance after claim: %v ETH", common.EtherAmount(*balance).AsEther())
	} else {
		balance, err := s.ERC20BalanceAt(s.ctx, s.contractSwap.Asset, addr, nil)
		if err != nil {
			return ethcommon.Hash{}, err
		}

		log.Infof("balance after claim: %v %s", 
			common.NewERC20TokenAmountFromBigInt(balance, float64(decimals)).AsStandard(), 
			symbol,
		)
	}

	return txHash, nil
}

func (s *swapState) claimRelayer() (ethcommon.Hash, error) {
	return claimRelayer(
		s.Ctx(),
		s.EthPrivateKey(),
		s.Contract(),
		s.contractAddr,
		s.EthClient(),
		s.offerExtra.RelayerEndpoint,
		s.offerExtra.RelayerCommission,
		&s.contractSwap,
		s.getSecret(),
	)
}

// claimRelayer claims the ETH funds via relayer.
func claimRelayer(
	ctx context.Context,
	sk *ecdsa.PrivateKey,
	contract *contracts.SwapFactory,
	contractAddr ethcommon.Address,
	ec *ethclient.Client,
	relayerEndpoint string,
	relayerCommission float64,
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

	abi, err := abi.JSON(strings.NewReader(contracts.SwapFactoryABI))
	if err != nil {
		return ethcommon.Hash{}, err
	}

	feeValue, err := calculateRelayerCommissionValue(contractSwap.Value, relayerCommission)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	calldata, err := abi.Pack("claimRelayer", *contractSwap, secret, feeValue)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	txHash, err := rc.SubmitTransaction(contractAddr, calldata)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	// wait for inclusion
	receipt, err := block.WaitForReceipt(ctx, ec, txHash)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	if receipt.Status == 0 {
		return ethcommon.Hash{}, fmt.Errorf("transaction failed")
	}

	if len(receipt.Logs) == 0 {
		return ethcommon.Hash{}, fmt.Errorf("claim transaction had no logs")
	}

	return txHash, nil
}

// swapValue is in wei
// relayerCommission is a percentage (ie must be much less than 1)
// error if it's greater than 0.1 (10%) - arbitrary, just a sanity check
func calculateRelayerCommissionValue(swapValue *big.Int, relayerCommission float64) (*big.Int, error) {
	if relayerCommission > 0.1 {
		return nil, errRelayerCommissionTooHigh
	}

	swapValueF := big.NewFloat(0).SetInt(swapValue)
	relayerCommissionF := big.NewFloat(relayerCommission)
	feeValue := big.NewFloat(0).Mul(swapValueF, relayerCommissionF)
	wei, _ := feeValue.Int(nil)
	return wei, nil
}
