package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/block"
)

func setupXMRTakerAuth(t *testing.T) (*bind.TransactOpts, *ethclient.Client, *ecdsa.PrivateKey) {
	conn, chainID := NewEthClient(t)
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyXMRTaker)
	require.NoError(t, err)
	auth, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)
	return auth, conn, pk
}

// deploys ERC20Mock.sol and assigns the whole token balance to the XMRTaker default address.
func deployERC20Mock(t *testing.T) ethcommon.Address {
	auth, conn, pkA := setupXMRTakerAuth(t)
	pub := pkA.Public().(*ecdsa.PublicKey)
	addr := ethcrypto.PubkeyToAddress(*pub)

	decimals := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(18), nil)
	balance := big.NewInt(0).Mul(big.NewInt(9999999), decimals)
	erc20Addr, erc20Tx, _, err := contracts.DeployERC20Mock(auth, conn, "ERC20Mock", "MOCK", addr, balance)
	require.NoError(t, err)
	_, err = block.WaitForReceipt(context.Background(), conn, erc20Tx.Hash())
	require.NoError(t, err)
	return erc20Addr
}

func TestXMRTaker_ERC20_Query(t *testing.T) {
	testXMRTakerQuery(t, types.EthAsset(deployERC20Mock(t)))
}

func TestSuccess_ERC20_OneSwap(t *testing.T) {
	testSuccess(t, types.EthAsset(deployERC20Mock(t)))
}

func TestRefund_ERC20_XMRTakerCancels(t *testing.T) {
	testRefundXMRTakerCancels(t, types.EthAsset(deployERC20Mock(t)))
}

func TestAbort_ERC20_XMRTakerCancels(t *testing.T) {
	testAbortXMRTakerCancels(t, types.EthAsset(deployERC20Mock(t)))
}

func TestAbort_ERC20_XMRMakerCancels(t *testing.T) {
	testAbortXMRMakerCancels(t, types.EthAsset(deployERC20Mock(t)))
}

func TestError_ERC20_ShouldOnlyTakeOfferOnce(t *testing.T) {
	testErrorShouldOnlyTakeOfferOnce(t, types.EthAsset(deployERC20Mock(t)))
}

func TestSuccess_ERC20_ConcurrentSwaps(t *testing.T) {
	testSuccessConcurrentSwaps(t, types.EthAsset(deployERC20Mock(t)))
}
