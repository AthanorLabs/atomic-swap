package xmrmaker

import (
	"fmt"
	"math"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/relayer"
)

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
		log.Infof("balance before claim: %v %s", common.EtherAmount(*balance).ToDecimals(decimals), symbol)
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

		log.Infof("balance after claim: %v %s", common.EtherAmount(*balance).ToDecimals(decimals), symbol)
	}

	return txHash, nil
}

// claimRelayer claims the ETH funds via relayer.
func (s *swapState) claimRelayer() (ethcommon.Hash, error) {
	sk := s.EthPrivateKey()
	forwarderAddress, err := s.Contract().TrustedForwarder(&bind.CallOpts{})
	if err != nil {
		return ethcommon.Hash{}, err
	}

	rc, err := relayer.NewClient(sk, s.EthClient(), s.offerExtra.RelayerEndpoint, forwarderAddress)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	abi, err := abi.JSON(strings.NewReader(contracts.SwapFactoryABI))
	if err != nil {
		return ethcommon.Hash{}, err
	}

	sc := s.getSecret()
	feeValue, err := calculateRelayerCommissionValue(s.contractSwap.Value, s.offerExtra.RelayerCommission)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	calldata, err := abi.Pack("claimRelayer", s.contractSwap, sc, feeValue)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	txHash, err := rc.SubmitTransaction(s.ContractAddr(), calldata)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	// wait for inclusion
	_, err = s.WaitForReceipt(s.Ctx(), txHash)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	return txHash, nil
}

var numEtherUnitsFloat = big.NewFloat(math.Pow(10, 18))

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
