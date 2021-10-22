package bob

import (
	"fmt"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/swap-contract"
)

var _ Bob = &bob{}

// Bob contains the functions that will be called by a user who owns XMR
// and wishes to swap for ETH.
type Bob interface {
	// GenerateKeys generates Bob's public spend and view keys (S_b, V_b)
	GenerateKeys() error

	// SetContract sets the contract in which Alice has locked her ETH.
	SetContract(*swap.Swap)

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
	// It accepts Alice's public keys (S_a, V_a) as input, as well as the amount to lock
	// TODO: units
	LockFunds(aliceKeys *monero.PublicKeyPair, amount uint) error

	// RedeemFunds redeem's Bob's funds on ethereum
	// It accepts an instance of the Swap contract (as deployed by Alice)
	RedeemFunds() error
}

type bob struct {
	privkeys *monero.PrivateKeyPair
	pubkeys  *monero.PublicKeyPair
	client monero.Client
	contract *swap.Swap
}

func NewBob() *bob {
	return &bob{}
}

// GenerateKeys generates Bob's public spend and view keys (S_b, V_b)
func (b *bob) GenerateKeys() error {
	var err error
	b.privkeys, err = monero.GenerateKeys()
	if err != nil {
		return err
	}

	b.pubkeys = b.privkeys.PublicKeyPair()
	return nil
}

func (b *bob) SetContract(contract *swap.Swap) {
	b.contract = contract
}

func (b *bob) WatchForRefund() (<-chan *monero.PrivateKeyPair, error) {
	return nil, nil
}

func (b *bob) LockFunds(akp *monero.PublicKeyPair, amount uint) error {
	sk := monero.Sum(akp.SpendKey(), b.pubkeys.SpendKey())
	vk := monero.Sum(akp.ViewKey(), b.pubkeys.ViewKey())

	address := monero.NewPublicKeyPair(sk, vk).Address()
	if err := b.client.Transfer(address, amount); err != nil {
		return err
	}

	fmt.Println("Bob: successfully locked funds")
	fmt.Println("address: ", address)
	return nil
}

func (b *bob) RedeemFunds() error {
	return nil
}