package xmrmaker

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	"github.com/athanorlabs/atomic-swap/relayer"
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

	weiBalance, err := s.ETHClient().Balance(s.ctx)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		log.Infof("balance before claim: %s ETH", coins.NewWeiAmount(weiBalance).AsEther())
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

	var txHash ethcommon.Hash

	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	if s.offerExtra.RelayerFee != nil || weiBalance.Cmp(big.NewInt(0)) == 0 {
		// relayer fee was set or we had insufficient funds to claim without a relayer
		// TODO: Sufficient funds check above should be more specific
		txHash, err = s.discoverRelayersAndClaim()
		if err != nil {
			log.Warnf("failed to claim using relayers: %s", err)
		}
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

	forwarderAddress, err := s.Contract().TrustedForwarder(&bind.CallOpts{Context: s.ctx})
	if err != nil {
		return ethcommon.Hash{}, err
	}

	secret := s.getSecret()

	relayerFeeWei := relayer.DefaultRelayerFee
	if s.offerExtra.RelayerFee != nil {
		relayerFeeWei = coins.ToWeiAmount(s.offerExtra.RelayerFee).BigInt()
	}

	req, err := relayer.CreateRelayClaimRequest(
		s.ctx,
		s.ETHClient().PrivateKey(),
		s.ETHClient().Raw(),
		relayerFeeWei,
		s.contractAddr,
		forwarderAddress,
		s.contractSwap,
		&secret,
	)
	if err != nil {
		return ethcommon.Hash{}, err
	}

	for _, relayer := range relayers {
		resp, err := s.Backend.SubmitClaimToRelayer(relayer, req)
		if err != nil {
			log.Warnf("failed to submit tx to relayer with peer ID %s, trying next relayer: err: %s", relayer, err)
			continue
		}

		err = waitForClaimReceipt(
			s.ctx,
			s.ETHClient().Raw(),
			resp.TxHash,
			s.contractAddr,
			s.contractSwapID,
			s.getSecret(),
		)
		if err != nil {
			log.Warnf("tx %s submitted by relayer with peer ID %s failed, trying next relayer", resp.TxHash, relayer)
			continue
		}

		return resp.TxHash, nil
	}

	return ethcommon.Hash{}, errors.New("failed to submit transaction to any relayer")
}

func waitForClaimReceipt(
	ctx context.Context,
	ec *ethclient.Client,
	txHash ethcommon.Hash,
	contractAddr ethcommon.Address,
	contractSwapID, secret [32]byte,
) error {
	const (
		checkInterval = time.Second // time between transaction polls
		maxWait       = time.Minute // max wait for the tx to be included in a block
		maxNotFound   = 5           // max failures where the tx is not even found in the mempool
	)

	start := time.Now()

	var notFoundCount int
	// wait for inclusion
	for {
		// sleep before the first check, b/c we want to give the tx some time to propagate
		// into the node we're using
		err := common.SleepWithContext(ctx, checkInterval)
		if err != nil {
			return err
		}

		_, isPending, err := ec.TransactionByHash(ctx, txHash)
		if err != nil {
			// allow up to 5 NotFound errors, in case there's some network problems
			if errors.Is(err, ethereum.NotFound) && notFoundCount >= maxNotFound {
				notFoundCount++
				continue
			}

			return err
		}

		if time.Since(start) > maxWait {
			// the tx is taking too long, return an error so we try with another relayer
			return errRelayedTransactionTimeout
		}

		if !isPending {
			break
		}
	}

	receipt, err := ec.TransactionReceipt(ctx, txHash)
	if err != nil {
		return err
	}

	if len(receipt.Logs) == 0 {
		return fmt.Errorf("claim transaction had no logs")
	}

	return checkClaimedLog(receipt.Logs[0], contractAddr, contractSwapID, secret)
}

func checkClaimedLog(log *ethtypes.Log, contractAddr ethcommon.Address, contractSwapID, secret [32]byte) error {
	if log.Address != contractAddr {
		return errClaimedLogInvalidContractAddr
	}

	if len(log.Topics) != 3 {
		return errClaimedLogWrongTopicLength
	}

	if log.Topics[0] != claimedTopic {
		return errClaimedLogWrongEvent
	}

	if log.Topics[1] != contractSwapID {
		return errClaimedLogWrongSwapID
	}

	if log.Topics[2] != secret {
		return errClaimedLogWrongSecret
	}

	return nil
}
