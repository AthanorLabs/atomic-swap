package bob

import (
	"crypto/ecdsa"
	"fmt"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/swap-contract"
)

var _ Bob = &bob{}

// Bob contains the functions that will be called by a user who owns XMR
// and wishes to swap for ETH.
type Bob interface {
	// GenerateKeys generates Bob's spend and view keys (S_b, V_b)
	// It returns Bob's public spend key and his private view key, so that Alice can see
	// if the funds are locked.
	GenerateKeys() (*monero.PublicKey, *monero.PrivateViewKey, error)

	// SetAlicePublicKeys sets Alice's public spend and view keys
	SetAlicePublicKeys(*monero.PublicKeyPair)

	// SetContract sets the contract in which Alice has locked her ETH.
	SetContract(address ethcommon.Address) error

	// WatchForReady watches for Alice to call Ready() on the swap contract, allowing
	// Bob to call Claim().
	WatchForReady() (<-chan struct{}, error)

	// WatchForRefund watches for the Refund event in the contract.
	// This should be called before LockFunds.
	// If a keypair is sent over this channel, the rest of the protocol should be aborted.
	//
	// If Alice chooses to refund and thus reveals s_a,
	// the private spend and view keys that contain the previously locked monero
	// ((s_a + s_b), (v_a + v_b)) are sent over the channel.
	// Bob can then use these keys to move his funds if he wishes.
	WatchForRefund() (<-chan *monero.PrivateKeyPair, error)

	// LockFunds locks Bob's funds in the monero account specified by public key
	// (S_a + S_b), viewable with (V_a + V_b)
	// It accepts the amount to lock as the input
	// TODO: units
	LockFunds(amount uint) error

	// RedeemFunds redeem's Bob's funds on ethereum
	RedeemFunds() error
}

type bob struct {
	t0, t1 time.Time

	privkeys        *monero.PrivateKeyPair
	pubkeys         *monero.PublicKeyPair
	client          monero.Client
	contract        *swap.Swap
	ethPrivKey      *ecdsa.PrivateKey
	alicePublicKeys *monero.PublicKeyPair
	ethClient       *ethclient.Client
}

// NewBob returns a new instance of Bob.
// It accepts an endpoint to a monero-wallet-rpc instance where account 0 contains Bob's XMR.
func NewBob(moneroEndpoint, ethEndpoint, ethPrivKey string) (*bob, error) {
	pk, err := crypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	return &bob{
		client:     monero.NewClient(moneroEndpoint),
		ethClient:  ec,
		ethPrivKey: pk,
	}, nil
}

// GenerateKeys generates Bob's spend and view keys (S_b, V_b)
func (b *bob) GenerateKeys() (*monero.PublicKey, *monero.PrivateViewKey, error) {
	var err error
	b.privkeys, err = monero.GenerateKeys()
	if err != nil {
		return nil, nil, err
	}

	b.pubkeys = b.privkeys.PublicKeyPair()
	return b.pubkeys.SpendKey(), b.privkeys.ViewKey(), nil
}

func (b *bob) SetAlicePublicKeys(sk *monero.PublicKeyPair) {
	b.alicePublicKeys = sk
}

func (b *bob) SetContract(address ethcommon.Address) error {
	var err error
	b.contract, err = swap.NewSwap(address, b.ethClient)
	return err
}

func (b *bob) WatchForReady() (<-chan struct{}, error) {
	return nil, nil
}

func (b *bob) WatchForRefund() (<-chan *monero.PrivateKeyPair, error) {
	// watch for Refund() event on chain, calculate unlock key as result
	return nil, nil
}

func (b *bob) LockFunds(amount uint) error {
	kp := monero.SumSpendAndViewKeys(b.alicePublicKeys, b.pubkeys)

	address := kp.Address()
	if err := b.client.Transfer(address, 0, amount); err != nil {
		return err
	}

	fmt.Println("Bob: successfully locked funds")
	fmt.Println("address: ", address)
	return nil
}

func (b *bob) RedeemFunds() error {
	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing Bob's secret spend key
	return nil
}
