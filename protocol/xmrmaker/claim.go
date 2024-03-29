// Copyright 2023 The AthanorLabs/atomic-swap Authors
// SPDX-License-Identifier: LGPL-3.0-only

package xmrmaker

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/athanorlabs/atomic-swap/coins"
	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
)

// claimFunds redeems XMRMaker's ETH funds by calling Claim() on the contract
func (s *swapState) claimFunds() (*ethtypes.Receipt, error) {
	weiBalance, err := s.ETHClient().Balance(s.ctx)
	if err != nil {
		return nil, err
	}

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		log.Infof("balance before claim: %s ETH", weiBalance.AsEtherString())
	} else {
		balance, err := s.ETHClient().ERC20Balance(s.ctx, s.contractSwap.Asset) //nolint:govet
		if err != nil {
			return nil, err
		}
		log.Infof("balance before claim: %s %s", balance.AsStdString(), balance.StdSymbol())
	}

	hasBalanceToClaim, err := checkForMinClaimBalance(s.ctx, s.ETHClient())
	if err != nil {
		return nil, err
	}

	var receipt *ethtypes.Receipt

	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing XMRMaker's secret spend key
	if s.offerExtra.UseRelayer || !hasBalanceToClaim {
		// relayer fee was set or we had insufficient funds to claim without a relayer
		receipt, err = s.claimWithRelay()
		if err != nil {
			return nil, fmt.Errorf("failed to claim using relayers: %w", err)
		}
		log.Infof("claim transaction was relayed: %s", common.ReceiptInfo(receipt))
	} else {
		// claim and wait for tx to be included
		sc := s.getSecret()
		receipt, err = s.sender.Claim(s.contractSwap, sc)
		if err != nil {
			if strings.Contains(err.Error(), "insufficient funds for gas * price + value") {
				// if we get this error, we need to use a relayer
				receipt, err = s.claimWithRelay()
				if err != nil {
					return nil, fmt.Errorf("failed to claim using relayers: %w", err)
				}
				log.Infof("claim transaction was relayed: %s", common.ReceiptInfo(receipt))
			} else {
				return nil, err
			}
		} else {
			log.Infof("claim transaction %s", common.ReceiptInfo(receipt))
		}
	}
	if err != nil {
		return nil, err
	}

	if types.EthAsset(s.contractSwap.Asset) == types.EthAssetETH {
		balance, err := s.ETHClient().Balance(s.ctx)
		if err != nil {
			return nil, err
		}
		log.Infof("balance after claim: %s ETH", balance.AsEtherString())
	} else {
		balance, err := s.ETHClient().ERC20Balance(s.ctx, s.contractSwap.Asset)
		if err != nil {
			return nil, err
		}

		log.Infof("balance after claim: %s %s", balance.AsStdString(), balance.StdSymbol())
	}

	return receipt, nil
}

// checkForMinClaimBalance check if we have enough balance to call claim.
// return true if we do, false otherwise.
func checkForMinClaimBalance(ctx context.Context, ec extethclient.EthClient) (bool, error) {
	// gas cost for ETH-claim is 42965
	// gas cost for ERC20-claim is 47138
	// add a bit of leeway to allow for sudden gas price spikes
	const claimGas = 50000

	balance, err := ec.Balance(ctx)
	if err != nil {
		return false, err
	}

	if balance.Decimal().IsZero() {
		return false, nil
	}

	gasPrice, err := ec.SuggestGasPrice(ctx)
	if err != nil {
		return false, err
	}

	txCost := new(big.Int).Mul(gasPrice, big.NewInt(claimGas))
	if balance.BigInt().Cmp(txCost) < 0 {
		log.Infof("balance %s ETH is under the minimum %s ETH to call claim, using a relayer",
			balance.AsEtherString(),
			coins.FmtWeiAsETH(txCost),
		)
		return false, nil
	}

	log.Debugf("balance %s ETH is above the minimum %s ETH to call claim",
		balance.AsEtherString(),
		coins.FmtWeiAsETH(txCost),
	)
	return true, nil
}

// relayClaimWithXMRTaker relays the claim to the swap's XMR taker, who should
// process the claim even if they are not relaying claims for everyone.
func (s *swapState) relayClaimWithXMRTaker() (*ethtypes.Receipt, error) {
	secret := s.getSecret()
	relaySwap := &contracts.SwapCreatorRelaySwap{
		Swap:        *s.contractSwap,
		SwapCreator: s.swapCreatorAddr,
		Fee:         coins.RelayerFeeWei,
		// this is set when we receive the relayer's address hash
		RelayerHash: types.Hash{},
	}

	response, err := s.Backend.SubmitClaimToRelayer(s.info.PeerID, &s.offer.ID, relaySwap, secret)
	if err != nil {
		return nil, err
	}

	receipt, err := waitForClaimRelayerReceipt(
		s.ctx,
		s.ETHClient().Raw(),
		response.TxHash,
		s.swapCreatorAddr,
		s.contractSwapID,
		s.getSecret(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt of relayer's tx: %s", err)
	}

	log.Infof("relayer's claim via counterparty included and validated %s", common.ReceiptInfo(receipt))
	return receipt, nil
}

// claimWithAdvertisedRelayers relays the claim to nodes that advertise
// themselves as relayers in the DHT until the claim succeeds, all relayers have
// been tried, or the context is cancelled.
func (s *swapState) claimWithAdvertisedRelayers() (*ethtypes.Receipt, error) {
	secret := s.getSecret()
	relaySwap := &contracts.SwapCreatorRelaySwap{
		Swap:        *s.contractSwap,
		SwapCreator: s.swapCreatorAddr,
		Fee:         coins.RelayerFeeWei,
		// this is set when we receive the relayer's address hash
		RelayerHash: types.Hash{},
	}

	relayers, err := s.Backend.DiscoverRelayers()
	if err != nil {
		return nil, err
	}

	if len(relayers) == 0 {
		return nil, errors.New("no relayers found to submit claim to")
	}
	log.Debugf("Found %d relayers to submit claim to", len(relayers))
	for _, relayerPeerID := range relayers {
		if relayerPeerID == s.info.PeerID {
			log.Debugf("skipping DHT-advertised relayer that is our swap counterparty")
			continue
		}

		log.Debugf("submitting claim to relayer with peer ID %s", relayerPeerID)
		resp, err := s.Backend.SubmitClaimToRelayer(relayerPeerID, nil, relaySwap, secret)
		if err != nil {
			log.Warnf("failed to submit tx to relayer: %s", err)
			continue
		}

		receipt, err := waitForClaimRelayerReceipt(
			s.ctx,
			s.ETHClient().Raw(),
			resp.TxHash,
			s.swapCreatorAddr,
			s.contractSwapID,
			s.getSecret(),
		)
		if err != nil {
			log.Warnf("failed to get receipt of relayer's tx=%s: %s", resp.TxHash.Hex(), err)
			continue
		}

		log.Infof("DHT relayer's claim included and validated %s", common.ReceiptInfo(receipt))

		return receipt, nil
	}

	return nil, errors.New("failed to relay claim with any non-counterparty relayer")
}

// claimWithRelay first tries to relay sequentially with all relayers
// advertising in the DHT that are not the XMR taker and, if that fails, falls
// back to the XMR taker who, if using our software, will act as a relayer of
// last resort for their own swap, even if they are not performing relay
// operations more generally. Note that the receipt returned is for a
// transaction created by the remote relayer, not by us.
func (s *swapState) claimWithRelay() (*ethtypes.Receipt, error) {
	receipt, err := s.claimWithAdvertisedRelayers()
	if err != nil {
		log.Warnf("failed to relay with DHT-advertised relayers: %s", err)
		log.Infof("falling back to swap counterparty as relayer")
		receipt, err = s.relayClaimWithXMRTaker()
		if err != nil {
			return nil, err
		}
	}

	// Save the relayer fee to the database
	s.info.SetRelayerFee(coins.RelayerFeeETH)
	swapManager := s.SwapManager()
	err = swapManager.WriteSwapToDB(s.info)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func waitForClaimRelayerReceipt(
	ctx context.Context,
	ec *ethclient.Client,
	txHash ethcommon.Hash,
	contractAddr ethcommon.Address,
	contractSwapID [32]byte,
	secret [32]byte,
) (*ethtypes.Receipt, error) {
	const (
		checkInterval = 1500 * time.Millisecond // 1.5 seconds between poll attempts
		maxPolls      = 10                      // We'll wait up to 15 seconds
	)

	// The relayer can see the transaction as included in a block and send us
	// the hash before our end sees it included in a block. The synchronization
	// should happen significantly faster, but we allow a full 15 seconds before
	// we give-up and decide that the relayer gave us a bad tx hash.
	for i := 0; i < maxPolls; i++ {
		receipt, err := ec.TransactionReceipt(ctx, txHash)
		if err != nil && !errors.Is(err, ethereum.NotFound) {
			return nil, err
		}
		// If err is still set, the error was ethereum.NotFound, which is returned
		// by TransactionReceipt even if our endpoint sees the TX as pending.
		if err != nil {
			if err = common.SleepWithContext(ctx, checkInterval); err != nil {
				return nil, err // context expired
			}
			continue
		}

		// no error, return the receipt if validation passes
		return validateClaimRelayerReceipt(ctx, ec, receipt, contractAddr, contractSwapID, secret)
	}

	// if we made it here, we exceeded maxPolls of the error ethereum.NotFound
	return nil, ethereum.NotFound
}

func validateClaimRelayerReceipt(
	ctx context.Context,
	ec *ethclient.Client,
	receipt *ethtypes.Receipt,
	contractAddr ethcommon.Address,
	contractSwapID [32]byte,
	secret [32]byte,
) (*ethtypes.Receipt, error) {
	if receipt.Status != ethtypes.ReceiptStatusSuccessful {
		err := fmt.Errorf("relayer's claim transaction failed (gas-lost=%d tx=%s block=%d), %w",
			receipt.GasUsed, receipt.TxHash, receipt.BlockNumber, block.ErrorFromBlock(ctx, ec, receipt))
		return nil, err
	}

	if len(receipt.Logs) == 0 {
		return nil, fmt.Errorf("relayer's claim transaction had no logs (tx=%s block=%d)",
			receipt.TxHash, receipt.BlockNumber)
	}

	if err := checkClaimedLog(receipt.Logs[0], contractAddr, contractSwapID, secret); err != nil {
		return nil, fmt.Errorf("relayer's claim had logs error (tx=%s block=%d): %w",
			receipt.TxHash, receipt.BlockNumber, err)
	}

	return receipt, nil
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
