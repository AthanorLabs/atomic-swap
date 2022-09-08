package recovery

import (
	"encoding/hex"
	"fmt"

	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	"github.com/athanorlabs/atomic-swap/monero"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker"
	"github.com/athanorlabs/atomic-swap/protocol/xmrtaker"
	"github.com/athanorlabs/atomic-swap/swapfactory"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type recoverer struct {
	env       common.Environment
	xmrClient monero.WalletClient
	ethClient *ethclient.Client
}

// NewRecoverer ...
func NewRecoverer(env common.Environment, moneroEndpoint, ethEndpoint string) (*recoverer, error) {
	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	return &recoverer{
		env:       env,
		ethClient: ec,
		xmrClient: monero.NewWalletClient(moneroEndpoint),
	}, nil
}

// WalletFromSecrets generates a monero wallet from the given XMRTaker and XMRMaker secrets.
func (r *recoverer) WalletFromSecrets(xmrtakerSecret, xmrmakerSecret string) (mcrypto.Address, error) {
	as, err := hex.DecodeString(xmrtakerSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode xmrmaker's secret: %w", err)
	}

	bs, err := hex.DecodeString(xmrmakerSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode xmrmaker's secret: %w", err)
	}

	ak, err := mcrypto.NewPrivateSpendKey(as)
	if err != nil {
		return "", err
	}

	bk, err := mcrypto.NewPrivateSpendKey(bs)
	if err != nil {
		return "", err
	}

	sk := mcrypto.SumPrivateSpendKeys(ak, bk)
	kp, err := sk.AsPrivateKeyPair()
	if err != nil {
		return "", err
	}

	return monero.CreateWallet("recovered-wallet", r.env, r.xmrClient, kp)
}

// WalletFromSharedSecret generates a monero wallet from the given shared secret.
func (r *recoverer) WalletFromSharedSecret(pk *mcrypto.PrivateKeyInfo) (mcrypto.Address, error) {
	skBytes, err := hex.DecodeString(pk.PrivateSpendKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode spend key: %w", err)
	}

	sk, err := mcrypto.NewPrivateSpendKey(skBytes)
	if err != nil {
		return "", err
	}

	vk, err := mcrypto.NewPrivateViewKeyFromHex(pk.PrivateViewKey)
	if err != nil {
		return "", err
	}

	kp := mcrypto.NewPrivateKeyPair(sk, vk)
	return monero.CreateWallet("recovered-wallet", r.env, r.xmrClient, kp)
}

// RecoverFromXMRMakerSecretAndContract recovers funds by either claiming ether or reclaiming locked monero.
func (r *recoverer) RecoverFromXMRMakerSecretAndContract(b backend.Backend, dataDir string,
	xmrmakerSecret, contractAddr string, swapID [32]byte,
	swap swapfactory.SwapFactorySwap) (*xmrmaker.RecoveryResult, error) {
	bs, err := hex.DecodeString(xmrmakerSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode XMRMaker's secret: %w", err)
	}

	bk, err := mcrypto.NewPrivateSpendKey(bs)
	if err != nil {
		return nil, err
	}

	addr := ethcommon.HexToAddress(contractAddr)
	rs, err := xmrmaker.NewRecoveryState(b, dataDir, bk, addr, swapID, swap)
	if err != nil {
		return nil, err
	}

	return rs.ClaimOrRecover()
}

// RecoverFromXMRTakerSecretAndContract recovers funds by either claiming locked monero or refunding ether.
func (r *recoverer) RecoverFromXMRTakerSecretAndContract(b backend.Backend, dataDir string,
	xmrtakerSecret string, swapID [32]byte, swap swapfactory.SwapFactorySwap) (*xmrtaker.RecoveryResult, error) {
	as, err := hex.DecodeString(xmrtakerSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode XMRTaker's secret: %w", err)
	}

	ak, err := mcrypto.NewPrivateSpendKey(as)
	if err != nil {
		return nil, err
	}

	rs, err := xmrtaker.NewRecoveryState(b, dataDir, ak, swapID, swap)
	if err != nil {
		return nil, err
	}

	return rs.ClaimOrRefund()
}
