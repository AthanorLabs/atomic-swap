package recovery

import (
	"encoding/hex"
	"fmt"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"

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
