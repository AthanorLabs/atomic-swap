package alice

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/noot/atomic-swap/monero"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/swap-contract"

	logging "github.com/ipfs/go-log"
)

var (
	_   Alice = &alice{}
	log       = logging.Logger("alice")
)

func reverse(s []byte) []byte {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

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
	DeployAndLockETH(amount uint64) (ethcommon.Address, error)

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

	// CreateMoneroWallet creates Alice's monero wallet after Bob calls Claim().
	CreateMoneroWallet(*monero.PrivateKeyPair) (monero.Address, error)

	NotifyClaimed(txHash string) (monero.Address, error)
}

type alice struct {
	ctx    context.Context
	t0, t1 time.Time //nolint

	privkeys    *monero.PrivateKeyPair
	pubkeys     *monero.PublicKeyPair
	bobSpendKey *monero.PublicKey
	bobViewKey  *monero.PrivateViewKey
	client      monero.Client

	contract   *swap.Swap
	ethPrivKey *ecdsa.PrivateKey
	ethClient  *ethclient.Client
	auth       *bind.TransactOpts

	nextExpectedMessage net.Message

	initiated                     bool
	providesAmount, desiredAmount uint64
}

// NewAlice returns a new instance of Alice.
// It accepts an endpoint to a monero-wallet-rpc instance where Alice will generate
// the account in which the XMR will be deposited.
func NewAlice(ctx context.Context, moneroEndpoint, ethEndpoint, ethPrivKey string) (*alice, error) {
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

	return &alice{
		ctx:                 ctx, // TODO: add cancel
		ethPrivKey:          pk,
		ethClient:           ec,
		client:              monero.NewClient(moneroEndpoint),
		auth:                auth,
		nextExpectedMessage: &net.InitiateMessage{},
	}, nil
}

func (a *alice) setNextExpectedMessage(msg net.Message) {
	a.nextExpectedMessage = msg
}

func (a *alice) GenerateKeys() (*monero.PublicKeyPair, error) {
	if a.privkeys != nil {
		return nil, nil
	}

	var err error
	a.privkeys, err = monero.GenerateKeys()
	if err != nil {
		return nil, err
	}

	a.pubkeys = a.privkeys.PublicKeyPair()
	return a.pubkeys, nil
}

func (a *alice) SetBobKeys(sk *monero.PublicKey, vk *monero.PrivateViewKey) {
	a.bobSpendKey = sk
	a.bobViewKey = vk
}

func (a *alice) DeployAndLockETH(amount uint64) (ethcommon.Address, error) {
	pkAlice := reverse(a.pubkeys.SpendKey().Bytes())
	pkBob := reverse(a.bobSpendKey.Bytes())

	var pka, pkb [32]byte
	copy(pka[:], reverse(pkAlice))
	copy(pkb[:], reverse(pkBob))

	log.Debug("locking amount: ", amount)
	a.auth.Value = big.NewInt(int64(amount))
	defer func() {
		a.auth.Value = nil
	}()

	address, _, swap, err := swap.DeploySwap(a.auth, a.ethClient, pka, pkb)
	if err != nil {
		return ethcommon.Address{}, err
	}

	balance, err := a.ethClient.BalanceAt(a.ctx, address, nil)
	if err != nil {
		return ethcommon.Address{}, err
	}

	log.Debug("contract balance: ", balance)

	a.contract = swap
	return address, nil
}

func (a *alice) Ready() error {
	_, err := a.contract.SetReady(a.auth)
	return err
}

func (a *alice) WatchForClaim() (<-chan *monero.PrivateKeyPair, error) {
	watchOpts := &bind.WatchOpts{
		Context: a.ctx,
	}

	out := make(chan *monero.PrivateKeyPair)
	ch := make(chan *swap.SwapClaimed)
	defer close(out)

	// watch for Refund() event on chain, calculate unlock key as result
	sub, err := a.contract.WatchClaimed(watchOpts, ch)
	if err != nil {
		return nil, err
	}

	defer sub.Unsubscribe()

	go func() {
		log.Debug("watching for claim...")
		for {
			select {
			case claim := <-ch:
				if claim == nil || claim.S == nil {
					continue
				}

				// got Bob's secret
				sbBytes := claim.S.Bytes()
				var sb [32]byte
				copy(sb[:], sbBytes)

				skB, err := monero.NewPrivateSpendKey(sb[:])
				if err != nil {
					log.Error("failed to convert Bob's secret into a key: ", err)
					return
				}

				vkA, err := skB.View()
				if err != nil {
					log.Error("failed to get view key from Bob's secret spend key: ", err)
					return
				}

				skAB := monero.SumPrivateSpendKeys(skB, a.privkeys.SpendKey())
				vkAB := monero.SumPrivateViewKeys(vkA, a.privkeys.ViewKey())
				kpAB := monero.NewPrivateKeyPair(skAB, vkAB)

				out <- kpAB
				return
			case <-a.ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (a *alice) Refund() error {
	secret := a.privkeys.SpendKeyBytes()
	s := big.NewInt(0).SetBytes(reverse(secret))
	_, err := a.contract.Refund(a.auth, s)
	return err
}

func (a *alice) CreateMoneroWallet(kpAB *monero.PrivateKeyPair) (monero.Address, error) {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	walletName := fmt.Sprintf("alice-swap-wallet-%s", t)
	if err := a.client.GenerateFromKeys(kpAB, walletName, ""); err != nil {
		return "", err
	}

	log.Info("created wallet: ", walletName)

	if err := a.client.Refresh(); err != nil {
		return "", err
	}

	balance, err := a.client.GetBalance(0)
	if err != nil {
		return "", err
	}

	accounts, err := a.client.GetAccounts()
	if err != nil {
		return "", err
	}

	log.Debug(accounts)
	log.Info("wallet balance: ", balance.Balance)
	return kpAB.Address(), nil
}

func (a *alice) NotifyClaimed(txHash string) (monero.Address, error) {
	receipt, err := a.ethClient.TransactionReceipt(a.ctx, ethcommon.HexToHash(txHash))
	if err != nil {
		return "", err
	}

	if len(receipt.Logs) == 0 {
		return "", errors.New("no logs!!!")
	}

	abi, err := swap.SwapMetaData.GetAbi()
	if err != nil {
		return "", err
	}

	data := receipt.Logs[0].Data
	res, err := abi.Unpack("Claimed", data)
	if err != nil {
		return "", err
	}

	log.Debug("got Bob's secret: ", hex.EncodeToString(res[0].(*big.Int).Bytes()))

	// got Bob's secret
	sbBytes := res[0].(*big.Int).Bytes()
	var sb [32]byte
	copy(sb[:], sbBytes)

	skB, err := monero.NewPrivateSpendKey(sb[:])
	if err != nil {
		log.Error("failed to convert Bob's secret into a key: %s\n", err)
		return "", err
	}

	skAB := monero.SumPrivateSpendKeys(skB, a.privkeys.SpendKey())
	// kpAB, err := skAB.AsPrivateKeyPair()
	// if err != nil {
	// 	return "", err
	// }

	vkAB := monero.SumPrivateViewKeys(a.bobViewKey, a.privkeys.ViewKey())
	//log.Debug("private view key: ", vkAB.Hex())
	//log.Debug("public view key: ", vkAB.Public().Hex())
	//log.Debug("private view key from spend key: ", kpAB.ViewKey().Hex())

	kpAB := monero.NewPrivateKeyPair(skAB, vkAB)

	pkAB := kpAB.PublicKeyPair()
	log.Info("public spend keys: ", pkAB.SpendKey().Hex())
	log.Info("public view keys: ", pkAB.ViewKey().Hex())

	return a.CreateMoneroWallet(kpAB)
}
