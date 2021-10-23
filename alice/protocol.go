package alice

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/swap-contract"
)

var _ Alice = &alice{}

const (
	keyAlice = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"
)

// Alice contains the functions that will be called by a user who owns ETH
// and wishes to swap for XMR.
type Alice interface {
	// GenerateKeys generates Alice's monero spend and view keys (S_b, V_b)
	// It returns Alice's public spend key
	GenerateKeys() (*monero.PublicKeyPair, error)

	// SetBobKeys sets Bob's public spend key (to be stored in the contract) and Bob's
	// private view key (used to check XMR balance before calling Ready())
	SetBobKeys(*monero.PublicKey, *monero.PrivateViewKey)

	// DeployAndLockETH deploys an instance of the Swap contract and locks `amount` ether in it.
	DeployAndLockETH(amount uint) (ethcommon.Address, error)

	// Ready calls the Ready() method on the Swap contract, indicating to Bob he has until time t_1 to
	// call Claim(). Ready() should only be called once Alice sees Bob lock his XMR.
	// If time t_0 has passed, there is no point of calling Ready().
	Ready() error

	// WatchForClaim watches for Bob to call Claim() on the Swap contract.
	// When Claim() is called, revealing Bob's secret s_b, the secret key corresponding
	// to (s_a + s_b) will be sent over this channel, allowing Alice to claim the XMR it contains.
	WatchForClaim() (<-chan *monero.PrivateKeyPair, error)

	// Refund calls the Refund() method in the Swap contract, revealing Alice's secret
	// and returns to her the ether in the contract.
	// If time t_1 passes and Claim() has not been called, Alice should call Refund().
	Refund() error
}

type alice struct {
	t0, t1 time.Time

	privkeys   *monero.PrivateKeyPair
	pubkeys    *monero.PublicKeyPair
	bobpubkeys *monero.PublicKey
	client     monero.Client

	contract   *swap.Swap
	ethPrivKey *ecdsa.PrivateKey
	ethClient  *ethclient.Client
}

// NewAlice returns a new instance of Alice.
// It accepts an endpoint to a monero-wallet-rpc instance where Alice will generate
// the account in which the XMR will be deposited.
func NewAlice(moneroEndpoint, ethEndpoint, ethPrivKey string) (*alice, error) {
	pk, err := crypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	return &alice{
		ethPrivKey: pk,
		ethClient:  ec,
		client:     monero.NewClient(moneroEndpoint),
	}, nil
}

func (a *alice) GenerateKeys() (*monero.PublicKeyPair, error) {
	var err error
	a.privkeys, err = monero.GenerateKeys()
	if err != nil {
		return nil, err
	}

	a.pubkeys = a.privkeys.PublicKeyPair()
	return a.pubkeys, nil
}

func (a *alice) SetBobKeys(*monero.PublicKey, *monero.PrivateViewKey) {

}

func (a *alice) DeployAndLockETH(amount uint) (ethcommon.Address, error) {
	pk_a, err := crypto.HexToECDSA(keyAlice)
	authAlice, err := bind.NewKeyedTransactorWithChainID(pk_a, big.NewInt(1337)) // ganache chainID

	pkAlice := a.pubkeys.SpendKey().Bytes()
	pkBob := a.bobpubkeys.Bytes()

	var pka, pkb [32]byte
	copy(pka[:], pkAlice)
	copy(pkb[:], pkBob)

	address, _, swap, err := swap.DeploySwap(authAlice, a.ethClient, pka, pkb)
	if err != nil {
		return ethcommon.Address{}, err
	}

	a.contract = swap
	return address, nil
}

func (a *alice) Ready() error {
	return nil
}

func (a *alice) WatchForClaim() (<-chan *monero.PrivateKeyPair, error) {
	return nil, nil
}

func (a *alice) Refund() error {
	return nil
}
