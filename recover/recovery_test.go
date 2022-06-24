package recovery

import (
	"context"
	"math/big"
	"path"
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/backend"
	"github.com/noot/atomic-swap/swapfactory"
	"github.com/noot/atomic-swap/tests"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

var defaultTimeout int64 = 5 // 5 seconds

func newRecoverer(t *testing.T) *recoverer {
	r, err := NewRecoverer(common.Development, common.DefaultXMRMakerMoneroEndpoint, common.DefaultEthEndpoint)
	require.NoError(t, err)
	return r
}

func newSwap(
	t *testing.T,
	ec *ethclient.Client,
	claimKey [32]byte,
	refundKey [32]byte,
	setReady bool,
) (
	ethcommon.Address,
	*swapfactory.SwapFactory,
	[32]byte,
	swapfactory.SwapFactorySwap,
) {
	tm := big.NewInt(defaultTimeout)

	pk, err := ethcrypto.HexToECDSA(tests.GetTakerTestKey(t))
	require.NoError(t, err)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	addr, _, contract, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	pkXMRMaker, err := ethcrypto.HexToECDSA(tests.GetMakerTestKey(t))
	require.NoError(t, err)

	nonce := big.NewInt(0)
	xmrmakerAddress := common.EthereumPrivateKeyToAddress(pkXMRMaker)
	tx, err := contract.NewSwap(txOpts, claimKey, refundKey, xmrmakerAddress,
		tm, nonce)
	require.NoError(t, err)

	receipt, err := ec.TransactionReceipt(context.Background(), tx.Hash())
	require.NoError(t, err)
	require.Equal(t, 1, len(receipt.Logs))
	swapID, err := swapfactory.GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := swapfactory.GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := swapfactory.SwapFactorySwap{
		Owner:        txOpts.From,
		Claimer:      xmrmakerAddress,
		PubKeyClaim:  claimKey,
		PubKeyRefund: refundKey,
		Timeout0:     t0,
		Timeout1:     t1,
		Value:        big.NewInt(0),
		Nonce:        nonce,
	}

	if setReady {
		_, err = contract.SetReady(txOpts, swap)
		require.NoError(t, err)
	}

	return addr, contract, swapID, swap
}

func newBackend(
	t *testing.T,
	ec *ethclient.Client,
	addr ethcommon.Address,
	contract *swapfactory.SwapFactory,
	privkey string,
) backend.Backend {
	pk, err := ethcrypto.HexToECDSA(privkey)
	require.NoError(t, err)

	cfg := &backend.Config{
		Ctx:                  context.Background(),
		Environment:          common.Development,
		EthereumPrivateKey:   pk,
		EthereumClient:       ec,
		ChainID:              big.NewInt(common.GanacheChainID),
		MoneroWalletEndpoint: common.DefaultXMRTakerMoneroEndpoint,
		MoneroDaemonEndpoint: common.DefaultMoneroDaemonEndpoint,
		SwapContract:         contract,
		SwapContractAddress:  addr,
	}

	b, err := backend.NewBackend(cfg)
	require.NoError(t, err)
	return b
}

func TestRecoverer_WalletFromSecrets(t *testing.T) {
	r := newRecoverer(t)
	kpA, err := mcrypto.GenerateKeys()
	require.NoError(t, err)
	kpB, err := mcrypto.GenerateKeys()
	require.NoError(t, err)

	addr, err := r.WalletFromSecrets(kpA.SpendKey().Hex(), kpB.SpendKey().Hex())
	require.NoError(t, err)

	skAB := mcrypto.SumPrivateSpendKeys(kpA.SpendKey(), kpB.SpendKey())
	kpAB, err := skAB.AsPrivateKeyPair()
	require.NoError(t, err)
	require.Equal(t, kpAB.Address(common.Development), addr)
}

func TestRecoverer_RecoverFromXMRMakerSecretAndContract_Claim(t *testing.T) {
	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	defer ec.Close()

	claimKey := keys.Secp256k1PublicKey.Keccak256()
	addr, contract, swapID, swap := newSwap(t, ec, claimKey, [32]byte{}, true)
	b := newBackend(t, ec, addr, contract, tests.GetMakerTestKey(t))

	r := newRecoverer(t)
	basePath := path.Join(t.TempDir(), "test-infofile")
	res, err := r.RecoverFromXMRMakerSecretAndContract(b, basePath, keys.PrivateKeyPair.SpendKey().Hex(),
		addr.String(), swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestRecoverer_RecoverFromXMRMakerSecretAndContract_Claim_afterTimeout(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip() // TODO: fails on CI with "no contract code at address"
	// }

	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	defer ec.Close()

	claimKey := keys.Secp256k1PublicKey.Keccak256()
	addr, contract, swapID, swap := newSwap(t, ec, claimKey, [32]byte{}, false)
	b := newBackend(t, ec, addr, contract, tests.GetMakerTestKey(t))

	r := newRecoverer(t)
	basePath := path.Join(t.TempDir(), "test-infofile")
	res, err := r.RecoverFromXMRMakerSecretAndContract(b, basePath, keys.PrivateKeyPair.SpendKey().Hex(),
		addr.String(), swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestRecoverer_RecoverFromXMRTakerSecretAndContract_Refund(t *testing.T) {
	// if testing.Short() {
	// 	t.Skip() // TODO: fails on CI with "no contract code at address"
	// }

	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	defer ec.Close()

	refundKey := keys.Secp256k1PublicKey.Keccak256()
	addr, contract, swapID, swap := newSwap(t, ec, [32]byte{}, refundKey, false)
	b := newBackend(t, ec, addr, contract, tests.GetTakerTestKey(t))

	r := newRecoverer(t)
	basePath := path.Join(t.TempDir(), "test-infofile")
	res, err := r.RecoverFromXMRTakerSecretAndContract(b, basePath, keys.PrivateKeyPair.SpendKey().Hex(),
		swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Refunded)
}
