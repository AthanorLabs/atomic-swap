package recovery

import (
	"encoding/hex"
	"fmt"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
	"github.com/noot/atomic-swap/protocol/alice"
	"github.com/noot/atomic-swap/protocol/bob"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type recoverer struct {
	env       common.Environment
	client    monero.Client
	ethClient *ethclient.Client
}

// NewRecoverer ...
func NewRecoverer(env common.Environment, moneroEndpoint, ethEndpoint string) (*recoverer, error) { //nolint:revive
	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	return &recoverer{
		env:       env,
		ethClient: ec,
		client:    monero.NewClient(moneroEndpoint),
	}, nil
}

// WalletFromSecrets generates a monero wallet from the given Alice and Bob secrets.
func (r *recoverer) WalletFromSecrets(aliceSecret, bobSecret string) (mcrypto.Address, error) {
	as, err := hex.DecodeString(aliceSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode bob's secret: %w", err)
	}

	bs, err := hex.DecodeString(bobSecret)
	if err != nil {
		return "", fmt.Errorf("failed to decode bob's secret: %w", err)
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

	return monero.CreateMoneroWallet("recovered-wallet", r.env, r.client, kp)
}

// RecoverFromBobSecretAndContract recovers funds by either claiming ether or reclaiming locked monero.
func (r *recoverer) RecoverFromBobSecretAndContract(b *bob.Instance,
	bobSecret, contractAddr string) (*bob.RecoveryResult, error) {
	bs, err := hex.DecodeString(bobSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Bob's secret: %w", err)
	}

	bk, err := mcrypto.NewPrivateSpendKey(bs)
	if err != nil {
		return nil, err
	}

	addr := ethcommon.HexToAddress(contractAddr)
	rs, err := bob.NewRecoveryState(b, bk, addr)
	if err != nil {
		return nil, err
	}

	return rs.ClaimOrRecover()
}

// RecoverFromAliceSecretAndContract recovers funds by either claiming locked monero or refunding ether.
func (r *recoverer) RecoverFromAliceSecretAndContract(a *alice.Instance,
	aliceSecret, contractAddr string) (*alice.RecoveryResult, error) {
	as, err := hex.DecodeString(aliceSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Alice's secret: %w", err)
	}

	ak, err := mcrypto.NewPrivateSpendKey(as)
	if err != nil {
		return nil, err
	}

	addr := ethcommon.HexToAddress(contractAddr)
	rs, err := alice.NewRecoveryState(a, ak, addr)
	if err != nil {
		return nil, err
	}

	return rs.ClaimOrRefund()
}
