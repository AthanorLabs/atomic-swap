package recovery

import (
	"context"
	"math/big"
	"testing"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/alice"
	"github.com/noot/atomic-swap/protocol/bob"
	"github.com/noot/atomic-swap/swapfactory"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"
)

var defaulTimeout int64 = 5 // 5 seconds

func newRecoverer(t *testing.T) *recoverer {
	r, err := NewRecoverer(common.Development, common.DefaultBobMoneroEndpoint, common.DefaultEthEndpoint)
	require.NoError(t, err)
	return r
}

func newSwap(t *testing.T, claimKey, refundKey [32]byte,
	setReady bool) (ethcommon.Address, *swapfactory.SwapFactory, [32]byte, swapfactory.SwapFactorySwap) {
	tm := big.NewInt(defaulTimeout)

	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)

	txOpts, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(common.GanacheChainID))
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	addr, _, contract, err := swapfactory.DeploySwapFactory(txOpts, ec)
	require.NoError(t, err)

	pkBob, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	nonce := big.NewInt(0)
	bobAddress := common.EthereumPrivateKeyToAddress(pkBob)
	tx, err := contract.NewSwap(txOpts, claimKey, refundKey, bobAddress,
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
		Claimer:      bobAddress,
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

func newAliceInstance(t *testing.T, addr ethcommon.Address, contract *swapfactory.SwapFactory) *alice.Instance {
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyAlice)
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	cfg := &alice.Config{
		Ctx:                  context.Background(),
		Environment:          common.Development,
		EthereumPrivateKey:   pk,
		EthereumClient:       ec,
		ChainID:              big.NewInt(common.GanacheChainID),
		MoneroWalletEndpoint: common.DefaultAliceMoneroEndpoint,
		SwapContract:         contract,
		SwapContractAddress:  addr,
	}

	a, err := alice.NewInstance(cfg)
	require.NoError(t, err)
	return a
}

func newBobInstance(t *testing.T) *bob.Instance {
	pk, err := ethcrypto.HexToECDSA(common.DefaultPrivKeyBob)
	require.NoError(t, err)

	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)

	cfg := &bob.Config{
		Ctx:                  context.Background(),
		Environment:          common.Development,
		EthereumPrivateKey:   pk,
		EthereumClient:       ec,
		ChainID:              big.NewInt(common.GanacheChainID),
		MoneroWalletEndpoint: common.DefaultBobMoneroEndpoint,
		MoneroDaemonEndpoint: common.DefaultMoneroDaemonEndpoint,
	}

	b, err := bob.NewInstance(cfg)
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

func TestRecoverer_RecoverFromBobSecretAndContract_Claim(t *testing.T) {
	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	b := newBobInstance(t)

	claimKey := keys.Secp256k1PublicKey.Keccak256()
	addr, _, swapID, swap := newSwap(t, claimKey, [32]byte{}, true)

	r := newRecoverer(t)
	res, err := r.RecoverFromBobSecretAndContract(b, keys.PrivateKeyPair.SpendKey().Hex(),
		addr.String(), swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestRecoverer_RecoverFromBobSecretAndContract_Claim_afterTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip() // TODO: fails on CI with "no contract code at address"
	}

	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	b := newBobInstance(t)

	claimKey := keys.Secp256k1PublicKey.Keccak256()
	addr, _, swapID, swap := newSwap(t, claimKey, [32]byte{}, false)

	r := newRecoverer(t)
	res, err := r.RecoverFromBobSecretAndContract(b, keys.PrivateKeyPair.SpendKey().Hex(),
		addr.String(), swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Claimed)
}

func TestRecoverer_RecoverFromAliceSecretAndContract_Refund(t *testing.T) {
	if testing.Short() {
		t.Skip() // TODO: fails on CI with "no contract code at address"
	}

	keys, err := pcommon.GenerateKeysAndProof()
	require.NoError(t, err)

	refundKey := keys.Secp256k1PublicKey.Keccak256()
	addr, contract, swapID, swap := newSwap(t, [32]byte{}, refundKey, false)

	a := newAliceInstance(t, addr, contract)

	r := newRecoverer(t)
	res, err := r.RecoverFromAliceSecretAndContract(a, keys.PrivateKeyPair.SpendKey().Hex(),
		swapID, swap)
	require.NoError(t, err)
	require.True(t, res.Refunded)
}
