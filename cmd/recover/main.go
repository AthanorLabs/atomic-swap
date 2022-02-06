package main

import (
	"context"
	"errors"
	"math/big"
	"os"

	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/cmd/utils"
	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	"github.com/noot/atomic-swap/protocol/alice"
	"github.com/noot/atomic-swap/protocol/bob"
	recovery "github.com/noot/atomic-swap/recover"

	logging "github.com/ipfs/go-log"
)

const (
	flagEnv                  = "env"
	flagMoneroWalletEndpoint = "monero-endpoint"
	flagEthereumEndpoint     = "ethereum-endpoint"
	flagEthereumPrivateKey   = "ethereum-privkey"
	flagEthereumChainID      = "ethereum-chain-id"
	flagGasPrice             = "gas-price"
	flagGasLimit             = "gas-limit"
	flagContractSwapID       = "contract-swap-id"
	flagAliceSecret          = "alice-secret"
	flagBobSecret            = "bob-secret"
	flagContractAddr         = "contract-addr"
)

var (
	log = logging.Logger("cmd")
	_   = logging.SetLogLevel("alice", "debug")
	_   = logging.SetLogLevel("bob", "debug")
	_   = logging.SetLogLevel("common", "debug")
	_   = logging.SetLogLevel("cmd", "debug")
	_   = logging.SetLogLevel("net", "debug")
	_   = logging.SetLogLevel("rpc", "debug")
)

var (
	app = &cli.App{
		Name:   "swaprecover",
		Usage:  "A program for recovering swap funds due to unexpected shutdowns",
		Action: runRecover,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEnv,
				Usage: "environment to use: one of mainnet, stagenet, or dev",
			},
			&cli.StringFlag{
				Name:  flagMoneroWalletEndpoint,
				Usage: "monero-wallet-rpc endpoint",
			},
			&cli.StringFlag{
				Name:  flagEthereumEndpoint,
				Usage: "ethereum client endpoint",
			},
			&cli.StringFlag{
				Name:  flagEthereumPrivateKey,
				Usage: "file containing a private key hex string",
			},
			&cli.UintFlag{
				Name:  flagEthereumChainID,
				Usage: "ethereum chain ID; eg. mainnet=1, ropsten=3, rinkeby=4, goerli=5, ganache=1337",
			},
			&cli.UintFlag{
				Name:  flagGasPrice,
				Usage: "ethereum gas price to use for transactions (in gwei). if not set, the gas price is set via oracle.",
			},
			&cli.UintFlag{
				Name:  flagGasLimit,
				Usage: "ethereum gas limit to use for transactions. if not set, the gas limit is estimated for each transaction.",
			},
			&cli.UintFlag{
				Name:  flagContractSwapID,
				Usage: "ID of the swap within the SwapFactory.sol contract",
			},
			&cli.StringFlag{
				Name:  flagAliceSecret,
				Usage: "Alice's swap secret, can be found in the basepath (default ~/.atomicswap), format is a hex-encoded string", //nolint:lll
			},
			&cli.StringFlag{
				Name:  flagBobSecret,
				Usage: "Bob's swap secret, can be found in the basepath (default ~/.atomicswap), format is a hex-encoded string", //nolint:lll
			},
			&cli.StringFlag{
				Name:  flagContractAddr,
				Usage: "address of deployed ethereum swap contract, can be found in the basepath (default ~/.atomicswap)", //nolint:lll
			},
		},
	}
)

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

// Recoverer is implemented by a backend which is able to recover monero
type Recoverer interface {
	WalletFromSecrets(aliceSecret, bobSecret string) (mcrypto.Address, error)
	RecoverFromBobSecretAndContract(b *bob.Instance, bobSecret, contractAddr string, swapID *big.Int) (*bob.RecoveryResult, error)         //nolint:lll
	RecoverFromAliceSecretAndContract(a *alice.Instance, aliceSecret, contractAddr string, swapID *big.Int) (*alice.RecoveryResult, error) //nolint:lll
}

type instance struct {
	getRecovererFunc func(c *cli.Context, env common.Environment) (Recoverer, error)
}

func runRecover(c *cli.Context) error {
	inst := &instance{
		getRecovererFunc: getRecoverer,
	}
	return inst.recover(c)
}

func (inst *instance) recover(c *cli.Context) error {
	as := c.String(flagAliceSecret)
	bs := c.String(flagBobSecret)
	contractAddr := c.String(flagContractAddr)

	env, cfg, err := utils.GetEnvironment(c)
	if err != nil {
		return err
	}

	if as == "" && bs == "" {
		return errors.New("must also provide one of --alice-secret or --bob-secret")
	}

	if as == "" && contractAddr == "" {
		return errors.New("must also provide one of --alice-secret or --contract-addr")
	}

	if contractAddr == "" && bs == "" {
		return errors.New("must also provide one of --contract-addr or --bob-secret")
	}

	swapID := big.NewInt(int64(c.Uint(flagContractSwapID)))
	if swapID.Uint64() == 0 {
		log.Warn("provided contract swap ID of 0, this is likely not correct (unless you deployed the contract)")
	}

	r, err := inst.getRecovererFunc(c, env)
	if err != nil {
		return err
	}

	if as != "" && bs != "" {
		addr, err := r.WalletFromSecrets(as, bs)
		if err != nil {
			return err
		}

		log.Infof("restored wallet from secrets: address=%s", addr)
		return nil
	}

	if bs != "" && contractAddr != "" {
		b, err := createBobInstance(context.Background(), c, env, cfg)
		if err != nil {
			return err
		}

		res, err := r.RecoverFromBobSecretAndContract(b, bs, contractAddr, swapID)
		if err != nil {
			return err
		}

		if res.Claimed {
			log.Info("claimed ether from contract! transaction hash=%s", res.TxHash)
			return nil
		}

		if res.Recovered {
			log.Infof("restored wallet from secrets: address=%s", res.MoneroAddress)
			return nil
		}
	}

	if as != "" && contractAddr != "" {
		a, err := createAliceInstance(context.Background(), c, env, cfg)
		if err != nil {
			return err
		}

		res, err := r.RecoverFromAliceSecretAndContract(a, as, contractAddr, swapID)
		if err != nil {
			return err
		}

		if res.Claimed {
			log.Info("claimed monero! wallet address=%s", res.MoneroAddress)
			return nil
		}

		if res.Refunded {
			log.Infof("refunded ether: transaction hash=%s", res.TxHash)
			return nil
		}
	}

	log.Warnf("unimplemented!")
	return nil
}

func getRecoverer(c *cli.Context, env common.Environment) (Recoverer, error) {
	var (
		moneroEndpoint, ethEndpoint string
	)

	if c.String(flagMoneroWalletEndpoint) != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		moneroEndpoint = common.DefaultBobMoneroEndpoint
	}

	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	log.Infof("created recovery module with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)
	return recovery.NewRecoverer(env, moneroEndpoint, ethEndpoint)
}

func createAliceInstance(ctx context.Context, c *cli.Context, env common.Environment,
	cfg common.Config) (*alice.Instance, error) {
	var (
		moneroEndpoint, ethEndpoint string
	)

	chainID := int64(c.Uint(flagEthereumChainID))
	if chainID == 0 {
		chainID = cfg.EthereumChainID
	}

	if c.String(flagMoneroWalletEndpoint) != "" {
		moneroEndpoint = c.String(flagMoneroWalletEndpoint)
	} else {
		moneroEndpoint = common.DefaultAliceMoneroEndpoint
	}

	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String(flagEthereumEndpoint)
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	ethPrivKey, err := utils.GetEthereumPrivateKey(c, env, false)
	if err != nil {
		return nil, err
	}

	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set
	var gasPrice *big.Int
	if c.Uint(flagGasPrice) != 0 {
		gasPrice = big.NewInt(int64(c.Uint(flagGasPrice)))
	}

	pk, err := ethcrypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	aliceCfg := &alice.Config{
		Ctx:                  ctx,
		Basepath:             cfg.Basepath,
		MoneroWalletEndpoint: moneroEndpoint,
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          env,
		ChainID:              big.NewInt(chainID),
		GasPrice:             gasPrice,
		GasLimit:             uint64(c.Uint(flagGasLimit)),
	}

	return alice.NewInstance(aliceCfg)
}

func createBobInstance(ctx context.Context, c *cli.Context, env common.Environment,
	cfg common.Config) (*bob.Instance, error) {
	var (
		moneroEndpoint, ethEndpoint string
	)

	chainID := int64(c.Uint(flagEthereumChainID))
	if chainID == 0 {
		chainID = cfg.EthereumChainID
	}

	if c.String(flagMoneroWalletEndpoint) != "" {
		moneroEndpoint = c.String(flagMoneroWalletEndpoint)
	} else {
		moneroEndpoint = common.DefaultBobMoneroEndpoint
	}

	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String(flagEthereumEndpoint)
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	ethPrivKey, err := utils.GetEthereumPrivateKey(c, env, true)
	if err != nil {
		return nil, err
	}

	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set
	var gasPrice *big.Int
	if c.Uint(flagGasPrice) != 0 {
		gasPrice = big.NewInt(int64(c.Uint(flagGasPrice)))
	}

	pk, err := ethcrypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	bobCfg := &bob.Config{
		Ctx:                  ctx,
		Basepath:             cfg.Basepath,
		MoneroWalletEndpoint: moneroEndpoint,
		MoneroDaemonEndpoint: common.DefaultMoneroDaemonEndpoint, // TODO: only set if env=development
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          env,
		ChainID:              big.NewInt(chainID),
		GasPrice:             gasPrice,
		GasLimit:             uint64(c.Uint(flagGasLimit)),
	}

	b, err := bob.NewInstance(bobCfg)
	if err != nil {
		return nil, err
	}

	return b, nil
}
