package recovery

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"path"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"

	"github.com/athanorlabs/atomic-swap/common"
	"github.com/athanorlabs/atomic-swap/common/types"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/tests"
)

var defaultTimeout int64 = 9 // timeout in seconds

func newRecoverer(t *testing.T) *recoverer {
	r, err := NewRecoverer(common.Development, monero.CreateWalletClient(t), common.DefaultEthEndpoint)
	require.NoError(t, err)
	return r
}

func newSwap(
	t *testing.T,
	claimKey [32]byte,
	refundKey [32]byte,
	setReady bool,
) (
	ethcommon.Address,
	*contracts.SwapFactory,
	[32]byte,
	contracts.SwapFactorySwap,
) {
	tm := big.NewInt(defaultTimeout)

	pk := tests.GetTakerTestKey(t)
	ec, chainID := tests.NewEthClient(t)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, chainID)
	require.NoError(t, err)

	_, tx, contract, err := contracts.DeploySwapFactory(txOpts, ec, ethcommon.Address{})
	require.NoError(t, err)

	addr, err := bind.WaitDeployed(context.Background(), ec, tx)
	require.NoError(t, err)

	pkXMRMaker := tests.GetMakerTestKey(t)

	nonce := big.NewInt(0)
	xmrmakerAddress := common.EthereumPrivateKeyToAddress(pkXMRMaker)
	tx, err = contract.NewSwap(txOpts, claimKey, refundKey, xmrmakerAddress, tm,
		ethcommon.Address(types.EthAssetETH), big.NewInt(0), nonce)
	require.NoError(t, err)
	receipt := tests.MineTransaction(t, ec, tx)

	require.Equal(t, 1, len(receipt.Logs))
	swapID, err := contracts.GetIDFromLog(receipt.Logs[0])
	require.NoError(t, err)

	t0, t1, err := contracts.GetTimeoutsFromLog(receipt.Logs[0])
	require.NoError(t, err)

	swap := contracts.SwapFactorySwap{
		Owner:        txOpts.From,
		Claimer:      xmrmakerAddress,
		PubKeyClaim:  claimKey,
		PubKeyRefund: refundKey,
		Timeout0:     t0,
		Timeout1:     t1,
		Asset:        ethcommon.Address(types.EthAssetETH),
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
	addr ethcommon.Address,
	contract *contracts.SwapFactory,
	privkey *ecdsa.PrivateKey,
) backend.Backend {
	pk := privkey
	ec, _ := tests.NewEthClient(t)

	extendedEC, err := extethclient.NewEthClient(context.Background(), ec, pk)
	require.NoError(t, err)

	cfg := &backend.Config{
		Ctx:                 context.Background(),
		Environment:         common.Development,
		EthereumClient:      extendedEC,
		MoneroClient:        monero.CreateWalletClient(t),
		SwapContract:        contract,
		SwapContractAddress: addr,
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

	claimKey := keys.Secp256k1PublicKey.Keccak256()
	addr, contract, swapID, swap := newSwap(t, claimKey, [32]byte{}, true)
	b := newBackend(t, addr, contract, tests.GetMakerTestKey(t))

	r := newRecoverer(t)
	dataDir := path.Join(t.TempDir(), "test-infofile")
	res, err := r.RecoverFromXMRMakerSecretAndContract(b, dataDir, keys.PrivateKeyPair.SpendKey().Hex(),
		addr.String(), swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestRecoverer_RecoverFromXMRMakerSecretAndContract_Claim_afterTimeout(t *testing.T) {
	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	claimKey := keys.Secp256k1PublicKey.Keccak256()
	addr, contract, swapID, swap := newSwap(t, claimKey, [32]byte{}, false)
	b := newBackend(t, addr, contract, tests.GetMakerTestKey(t))

	r := newRecoverer(t)
	dataDir := path.Join(t.TempDir(), "test-infofile")
	res, err := r.RecoverFromXMRMakerSecretAndContract(b, dataDir, keys.PrivateKeyPair.SpendKey().Hex(),
		addr.String(), swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestRecoverer_RecoverFromXMRTakerSecretAndContract_Refund(t *testing.T) {
	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	refundKey := keys.Secp256k1PublicKey.Keccak256()
	addr, contract, swapID, swap := newSwap(t, [32]byte{}, refundKey, false)
	b := newBackend(t, addr, contract, tests.GetTakerTestKey(t))

	r := newRecoverer(t)
	dataDir := path.Join(t.TempDir(), "test-infofile")
	res, err := r.RecoverFromXMRTakerSecretAndContract(b, dataDir, keys.PrivateKeyPair.SpendKey().Hex(),
		swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Refunded)
}
