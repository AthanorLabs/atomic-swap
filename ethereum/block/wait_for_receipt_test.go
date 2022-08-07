package block

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/tests"
)

// stampChecker collects all the pieces of data needed to create a transaction that calls
// check_stamp on the UTContract for unit tests.
type stampChecker struct {
	t        *testing.T
	ec       *ethclient.Client
	ctx      context.Context
	chainID  *big.Int
	fromKey  *ecdsa.PrivateKey
	contract *UTContract
}

// checkStamp performs the setup of creating a check_stamp transaction so unit tests
// can concentrate on the return values of WaitForReceipt.
func (c *stampChecker) checkStamp(epochTimeStamp int64) (*ethtypes.Receipt, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(c.fromKey, c.chainID)
	require.NoError(c.t, err)
	tx, err := c.contract.CheckStamp(auth, big.NewInt(epochTimeStamp))
	require.NoError(c.t, err)
	return WaitForReceipt(c.ctx, c.ec, tx.Hash())
}

// curBlockHeader is a test helper function to get the latest block header
func (c *stampChecker) curBlockHeader() *ethtypes.Header {
	hdr, err := c.ec.HeaderByNumber(c.ctx, nil)
	require.NoError(c.t, err)
	return hdr
}

// createStampChecker deploys the unit test contract (UTContract)
func createStampChecker(t *testing.T) *stampChecker {
	ec, chainID := tests.NewEthClient(t)
	ctx := context.Background()

	fromKey := tests.GetTestKeyByIndex(t, 0)

	auth, err := bind.NewKeyedTransactorWithChainID(fromKey, chainID)
	require.NoError(t, err)

	_, tx, contract, err := DeployUTContract(auth, ec)
	require.NoError(t, err)

	_, err = WaitForReceipt(ctx, ec, tx.Hash())
	require.NoError(t, err)

	return &stampChecker{
		t:        t,
		ec:       ec,
		ctx:      ctx,
		chainID:  chainID,
		fromKey:  fromKey,
		contract: contract,
	}

}

// Test WaitForReceipt with 2 transactions that are successfully mined:
// 1. contract creation
// 2. calling check_stamp with a value >= the mined block's timestamp
func TestWaitForReceipt_success(t *testing.T) {
	checker := createStampChecker(t)
	blockBefore := checker.curBlockHeader()
	futureTime := time.Now().Add(time.Hour).Unix()
	receipt, err := checker.checkStamp(futureTime)
	require.NoError(t, err)
	txBlockNum := receipt.BlockNumber
	require.Greater(t, txBlockNum.Uint64(), blockBefore.Number.Uint64())
}

// Test our errorFromBlock method. This test is designed to create a transaction that will
// successfully be created and sent to the network, but then subsequently fail when it is
// mined into a block.
func TestWaitForReceipt_failWhenTransactionIsMined(t *testing.T) {
	checker := createStampChecker(t)

	// One second in the future rounding up
	oneSecInFuture := time.Now().Add(1500 * time.Millisecond).Unix()

	// We don't want a race condition between getting the current block and creating the
	// check_stamp transaction. The transaction creation needs to be done when the
	// current block timestamp is exactly equal to the time we pass. That way transaction
	// creation will pass, but mining will fail. By waiting for the next block, we have
	// a full second to create the new transaction and avoid any race condition.
	hdr, err := WaitForEthBlockAfterTimestamp(checker.ctx, checker.ec, oneSecInFuture)
	require.NoError(t, err)

	_, err = checker.checkStamp(int64(hdr.Time))
	require.Error(t, err)
	// Ensure that we got the expected error
	require.Contains(t, err.Error(), "revert block.timestamp was not less than stamp")
	// Ensure that the expected error happened when the transaction was mined and not earlier
	require.Contains(t, err.Error(), "gas-lost=")
}
