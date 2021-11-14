package bob

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/swap-contract"

	logging "github.com/ipfs/go-log"
)

var (
	_   Bob = &bob{}
	log     = logging.Logger("bob")
)

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
	LockFunds(amount uint64) (monero.Address, error)

	// ClaimFunds redeem's Bob's funds on ethereum
	ClaimFunds() (string, error)
}

// SwapContract is the interface that must be implemented by the Swap contract.
type SwapContract interface {
	WatchIsReady(opts *bind.WatchOpts, sink chan<- *swap.SwapIsReady) (event.Subscription, error)
	WatchRefunded(opts *bind.WatchOpts, sink chan<- *swap.SwapRefunded) (event.Subscription, error)
	Claim(opts *bind.TransactOpts, _s *big.Int) (*ethtypes.Transaction, error)
}

type bob struct {
	ctx    context.Context
	t0, t1 time.Time //nolint

	privkeys     *monero.PrivateKeyPair
	pubkeys      *monero.PublicKeyPair
	client       monero.Client
	daemonClient monero.DaemonClient

	contract     SwapContract
	contractAddr ethcommon.Address
	ethClient    *ethclient.Client
	auth         *bind.TransactOpts

	ethPrivKey      *ecdsa.PrivateKey
	alicePublicKeys *monero.PublicKeyPair
}

// NewBob returns a new instance of Bob.
// It accepts an endpoint to a monero-wallet-rpc instance where account 0 contains Bob's XMR.
func NewBob(ctx context.Context, moneroEndpoint, moneroDaemonEndpoint, ethEndpoint, ethPrivKey string) (*bob, error) {
	pk, err := crypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(1337)) // ganache chainID
	if err != nil {
		return nil, err
	}

	return &bob{
		ctx:          ctx,
		client:       monero.NewClient(moneroEndpoint),
		daemonClient: monero.NewClient(moneroDaemonEndpoint),
		ethClient:    ec,
		ethPrivKey:   pk,
		auth:         auth,
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
	b.contractAddr = address
	b.contract, err = swap.NewSwap(address, b.ethClient)
	return err
}

func (b *bob) WatchForReady() (<-chan struct{}, error) {
	watchOpts := &bind.WatchOpts{
		Context: b.ctx,
	}

	done := make(chan struct{})
	ch := make(chan *swap.SwapIsReady)
	defer close(done)

	// watch for Refund() event on chain, calculate unlock key as result
	sub, err := b.contract.WatchIsReady(watchOpts, ch)
	if err != nil {
		return nil, err
	}

	defer sub.Unsubscribe()

	go func() {
		for {
			select {
			case event := <-ch:
				if !event.B {
					continue
				}

				// contract is ready!!
				close(done)
			case <-b.ctx.Done():
				return
			}
		}
	}()

	return done, nil
}

func (b *bob) WatchForRefund() (<-chan *monero.PrivateKeyPair, error) {
	watchOpts := &bind.WatchOpts{
		Context: b.ctx,
	}

	out := make(chan *monero.PrivateKeyPair)
	ch := make(chan *swap.SwapRefunded)
	defer close(out)

	// watch for Refund() event on chain, calculate unlock key as result
	sub, err := b.contract.WatchRefunded(watchOpts, ch)
	if err != nil {
		return nil, err
	}

	defer sub.Unsubscribe()

	go func() {
		for {
			select {
			case refund := <-ch:
				if refund == nil {
					continue
				}

				// got Alice's secret
				saBytes := refund.S.Bytes()
				var sa [32]byte
				copy(sa[:], saBytes)

				skA, err := monero.NewPrivateSpendKey(sa[:])
				if err != nil {
					log.Info("failed to convert Alice's secret into a key: %w", err)
					return
				}

				vkA, err := skA.View()
				if err != nil {
					log.Info("failed to get view key from Alice's secret spend key: %w", err)
					return
				}

				skAB := monero.SumPrivateSpendKeys(skA, b.privkeys.SpendKey())
				vkAB := monero.SumPrivateViewKeys(vkA, b.privkeys.ViewKey())
				kpAB := monero.NewPrivateKeyPair(skAB, vkAB)
				out <- kpAB
			case <-b.ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (b *bob) LockFunds(amount uint64) (monero.Address, error) {
	kp := monero.SumSpendAndViewKeys(b.alicePublicKeys, b.pubkeys)

	log.Info("public spend keys: ", kp.SpendKey().Hex())
	log.Info("public view keys: ", kp.ViewKey().Hex())

	log.Info("going to lock funds...")

	balance, err := b.client.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Info("balance: ", balance.Balance)
	log.Info("unlocked balance: ", balance.UnlockedBalance)
	log.Info("blocks to unlock: ", balance.BlocksToUnlock)

	address := kp.Address()
	if err := b.client.Transfer(address, 0, uint(amount)); err != nil {
		return "", err
	}

	bobAddr, err := b.client.GetAddress(0)
	if err != nil {
		return "", err
	}

	if err := b.daemonClient.GenerateBlocks(bobAddr.Address, 1); err != nil {
		return "", err
	}

	if err := b.client.Refresh(); err != nil {
		return "", err
	}

	log.Info("Bob: successfully locked funds")
	log.Info("address: ", address)
	return address, nil
}

func (b *bob) ClaimFunds() (string, error) {
	pub := b.ethPrivKey.Public().(*ecdsa.PublicKey)
	addr := ethcrypto.PubkeyToAddress(*pub)

	balance, err := b.ethClient.BalanceAt(b.ctx, addr, nil)
	if err != nil {
		return "", err
	}

	log.Info("Bob's balance before claim: ", balance)

	// call swap.Swap.Claim() w/ b.privkeys.sk, revealing Bob's secret spend key
	secret := b.privkeys.SpendKeyBytes()
	s := big.NewInt(0).SetBytes(secret)

	tx, err := b.contract.Claim(b.auth, s)
	if err != nil {
		return "", err
	}

	log.Info("success! Bob claimed funds")
	log.Info("tx hash: ", tx.Hash())

	receipt, err := b.ethClient.TransactionReceipt(b.ctx, tx.Hash())
	if err != nil {
		return "", err
	}

	//log.Info("tx logs: ", fmt.Sprintf("0x%x", receipt.Logs[0].Data))
	log.Info("included in block number: ", receipt.Logs[0].BlockNumber)
	log.Info("secret: ", fmt.Sprintf("%x", secret))

	balance, err = b.ethClient.BalanceAt(b.ctx, addr, nil)
	if err != nil {
		return "", err
	}

	log.Info("Bob's balance after claim: ", balance)
	return tx.Hash().String(), nil
}
