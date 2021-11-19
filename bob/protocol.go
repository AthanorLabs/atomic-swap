package bob

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"

	logging "github.com/ipfs/go-log"
)

var (
	log = logging.Logger("bob")
)

// bob implements the functions that will be called by a user who owns XMR
// and wishes to swap for ETH.
type bob struct {
	ctx    context.Context
	t0, t1 time.Time //nolint

	privkeys     *monero.PrivateKeyPair
	pubkeys      *monero.PublicKeyPair
	client       monero.Client
	daemonClient monero.DaemonClient

	contract     *swap.Swap
	contractAddr ethcommon.Address
	ethClient    *ethclient.Client
	auth         *bind.TransactOpts
	callOpts     *bind.CallOpts

	ethPrivKey      *ecdsa.PrivateKey
	alicePublicKeys *monero.PublicKeyPair

	nextExpectedMessage net.Message

	initiated                     bool
	providesAmount, desiredAmount uint64
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

	pub := pk.Public().(*ecdsa.PublicKey)

	return &bob{
		ctx:          ctx,
		client:       monero.NewClient(moneroEndpoint),
		daemonClient: monero.NewClient(moneroDaemonEndpoint),
		ethClient:    ec,
		ethPrivKey:   pk,
		auth:         auth,
		callOpts: &bind.CallOpts{
			From:    crypto.PubkeyToAddress(*pub),
			Context: ctx,
		},
		nextExpectedMessage: &net.InitiateMessage{},
	}, nil
}

func (b *bob) setNextExpectedMessage(msg net.Message) {
	b.nextExpectedMessage = msg
}

// generateKeys generates Bob's spend and view keys (S_b, V_b)
// It returns Bob's public spend key and his private view key, so that Alice can see
// if the funds are locked.
func (b *bob) generateKeys() (*monero.PublicKey, *monero.PrivateViewKey, error) {
	if b.privkeys != nil {
		return b.pubkeys.SpendKey(), b.privkeys.ViewKey(), nil
	}

	var err error
	b.privkeys, err = monero.GenerateKeys()
	if err != nil {
		return nil, nil, err
	}

	// TODO: configure basepath
	if err := common.WriteKeysToFile("./bob-xmr", b.privkeys); err != nil {
		return nil, nil, err
	}

	b.pubkeys = b.privkeys.PublicKeyPair()
	return b.pubkeys.SpendKey(), b.privkeys.ViewKey(), nil
}

// setAlicePublicKeys sets Alice's public spend and view keys
func (b *bob) setAlicePublicKeys(sk *monero.PublicKeyPair) {
	b.alicePublicKeys = sk
}

// setContract sets the contract in which Alice has locked her ETH.
func (b *bob) setContract(address ethcommon.Address) error {
	var err error
	b.contractAddr = address
	b.contract, err = swap.NewSwap(address, b.ethClient)
	return err
}

// watchForReady watches for Alice to call Ready() on the swap contract, allowing
// Bob to call Claim().
func (b *bob) watchForReady() (<-chan struct{}, error) { //nolint:unused
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

// watchForRefund watches for the Refund event in the contract.
// This should be called before LockFunds.
// If a keypair is sent over this channel, the rest of the protocol should be aborted.
//
// If Alice chooses to refund and thus reveals s_a,
// the private spend and view keys that contain the previously locked monero
// ((s_a + s_b), (v_a + v_b)) are sent over the channel.
// Bob can then use these keys to move his funds if he wishes.
func (b *bob) watchForRefund() (<-chan *monero.PrivateKeyPair, error) { //nolint:unused
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

// lockFunds locks Bob's funds in the monero account specified by public key
// (S_a + S_b), viewable with (V_a + V_b)
// It accepts the amount to lock as the input
// TODO: units
func (b *bob) lockFunds(amount uint64) (monero.Address, error) {
	kp := monero.SumSpendAndViewKeys(b.alicePublicKeys, b.pubkeys)

	log.Debug("public spend keys: ", kp.SpendKey().Hex())
	log.Debug("public view keys: ", kp.ViewKey().Hex())
	log.Infof("going to lock XMR funds, amount=%d", amount)

	balance, err := b.client.GetBalance(0)
	if err != nil {
		return "", err
	}

	log.Debug("XMR balance: ", balance.Balance)
	log.Debug("unlocked XMR balance: ", balance.UnlockedBalance)
	log.Debug("blocks to unlock: ", balance.BlocksToUnlock)

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

	log.Infof("successfully locked XMR funds: address=%s", address)
	return address, nil
}

// claimFunds redeems Bob's ETH funds by calling Claim() on the contract
func (b *bob) claimFunds() (string, error) {
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
