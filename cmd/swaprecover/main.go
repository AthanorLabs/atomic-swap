package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	contracts "github.com/athanorlabs/atomic-swap/ethereum"
	"github.com/athanorlabs/atomic-swap/ethereum/extethclient"
	"github.com/athanorlabs/atomic-swap/monero"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker"
	"github.com/athanorlabs/atomic-swap/protocol/xmrtaker"
	recovery "github.com/athanorlabs/atomic-swap/recover"
)

const (
	flagEnv                  = "env"
	flagDataDir              = "data-dir"
	flagMoneroDaemonHost     = "monerod-host"
	flagMoneroDaemonPort     = "monerod-port"
	flagMoneroWalletPath     = "wallet-file"
	flagMoneroWalletPassword = "wallet-password"
	flagEthereumEndpoint     = "ethereum-endpoint"
	flagEthereumPrivKey      = "ethereum-privkey"
	flagGasPrice             = "gas-price"
	flagGasLimit             = "gas-limit"
	flagInfoFile             = "infofile"
	flagXMRMaker             = "xmrmaker"
	flagXMRTaker             = "xmrtaker"
)

var (
	log = logging.Logger("cmd")
	_   = logging.SetLogLevel("xmrtaker", "debug")
	_   = logging.SetLogLevel("xmrmaker", "debug")
	_   = logging.SetLogLevel("common", "debug")
	_   = logging.SetLogLevel("cmd", "debug")
	_   = logging.SetLogLevel("net", "debug")
	_   = logging.SetLogLevel("rpc", "debug")
	_   = logging.SetLogLevel("monero", "debug")
	_   = logging.SetLogLevel("extethclient", "debug")
	_   = logging.SetLogLevel("contracts", "debug")
)

var (
	app = &cli.App{
		Name:                 "swaprecover",
		Usage:                "A program for recovering swap funds due to unexpected shutdowns",
		Version:              cliutil.GetVersion(),
		Action:               runRecover,
		EnableBashCompletion: true,
		Suggest:              true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  flagEnv,
				Usage: "Environment to use: one of mainnet, stagenet, or dev",
				Value: "dev",
			},
			&cli.StringFlag{
				Name:  flagDataDir,
				Usage: "Path to store swap artifacts", //nolint:misspell
				Value: "{HOME}/.atomicswap/{ENV}",     // For --help only, actual default replaces variables
			},
			&cli.StringFlag{
				Name:  flagMoneroDaemonHost,
				Usage: "monerod host",
				Value: "127.0.0.1",
			},
			&cli.UintFlag{
				Name: flagMoneroDaemonPort,
				Usage: fmt.Sprintf("monerod port (--%s=stagenet changes default to %d)",
					flagEnv, common.DefaultMoneroDaemonStagenetPort),
				Value: common.DefaultMoneroDaemonMainnetPort, // at least for now, this is also the dev default
			},
			&cli.StringFlag{
				Name:  flagMoneroWalletPath,
				Usage: "Path to the recovery Monero wallet file, created if missing",
				Value: "{DATA-DIR}/{ENV}/wallet/swap-wallet",
			},
			&cli.StringFlag{
				Name:  flagMoneroWalletPassword,
				Usage: "Password of monero wallet file",
			},
			&cli.StringFlag{
				Name:  flagEthereumEndpoint,
				Usage: "Ethereum client endpoint",
			},
			&cli.StringFlag{
				Name:  flagEthereumPrivKey,
				Usage: "File containing a private key hex string",
			},
			&cli.UintFlag{
				Name:   flagGasPrice,
				Usage:  "Ethereum gas price to use for transactions (in gwei). If not set, the gas price is set via oracle.",
				Hidden: true, // Not well tested, hiding from end users for now
			},
			&cli.UintFlag{
				Name:  flagGasLimit,
				Usage: "Ethereum gas limit to use for transactions. if not set, the gas limit is estimated for each transaction.",
			},
			&cli.StringFlag{
				Name:  flagInfoFile,
				Usage: "Path to swap infofile",
			},
			&cli.BoolFlag{
				Name:  flagXMRMaker,
				Usage: "Use when recovering as an xmr-maker",
			},
			&cli.BoolFlag{
				Name:  flagXMRTaker,
				Usage: "Use when recovering as an xmr-taker",
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

// Recoverer is implemented by a backend which is able to recover swap funds
type Recoverer interface {
	WalletFromSharedSecret(secret *mcrypto.PrivateKeyInfo) (mcrypto.Address, error)
	RecoverFromXMRMakerSecretAndContract(b backend.Backend, dataDir string, xmrmakerSecret, contractAddr string, swapID [32]byte, swap contracts.SwapFactorySwap) (*xmrmaker.RecoveryResult, error) //nolint:lll
	RecoverFromXMRTakerSecretAndContract(b backend.Backend, dataDir string, xmrtakerSecret string, swapID [32]byte, swap contracts.SwapFactorySwap) (*xmrtaker.RecoveryResult, error)               //nolint:lll
}

type instance struct {
	getRecovererFunc func(c *cli.Context, env common.Environment, cfg *common.Config) (Recoverer, error)
}

func runRecover(c *cli.Context) error {
	inst := &instance{
		getRecovererFunc: getRecoverer,
	}
	return inst.recover(c)
}

func (inst *instance) recover(c *cli.Context) error {
	xmrmaker := c.Bool(flagXMRMaker)
	xmrtaker := c.Bool(flagXMRTaker)
	// Either maker or taker must be specified, but not both, so their values must be opposite
	if xmrmaker == xmrtaker {
		return errMustSpecifyXMRMakerOrTaker
	}

	env, cfg, err := cliutil.GetEnvironment(c.String(flagEnv))
	if err != nil {
		return err
	}

	// cfg.DataDir already has a default set, so only override if the user explicitly set the flag
	if c.IsSet(flagDataDir) {
		cfg.DataDir = c.String(flagDataDir) // override the value derived from `flagEnv`
	}

	infofilePath := c.String(flagInfoFile)
	if infofilePath == "" {
		return errMustProvideInfoFile
	}

	infofileBytes, err := os.ReadFile(filepath.Clean(infofilePath))
	if err != nil {
		return err
	}

	var infofile *pcommon.InfoFileContents
	if err = json.Unmarshal(infofileBytes, &infofile); err != nil {
		return err
	}

	r, err := inst.getRecovererFunc(c, env, &cfg)
	if err != nil {
		return err
	}

	if infofile.SharedSwapPrivateKey != nil {
		addr, err := r.WalletFromSharedSecret(infofile.SharedSwapPrivateKey) //nolint:govet
		if err != nil {
			return err
		}

		log.Infof("restored wallet from secrets: address=%s", addr)
		return nil
	}

	contractAddr := infofile.ContractAddress
	addr := ethcommon.HexToAddress(contractAddr)

	b, err := createBackend(context.Background(), c, env, cfg, addr)
	if err != nil {
		return err
	}
	defer b.XMRClient().Close()

	dataDir := filepath.Dir(filepath.Clean(infofilePath))

	if xmrmaker {
		res, err := r.RecoverFromXMRMakerSecretAndContract(b, dataDir, infofile.PrivateKeyInfo.PrivateSpendKey,
			contractAddr, infofile.ContractSwapID, infofile.ContractSwap)
		if err != nil {
			return err
		}

		if res.Claimed {
			log.Infof("claimed ether from contract! transaction hash=%s", res.TxHash)
			return nil
		}

		if res.Recovered {
			log.Infof("restored wallet from secrets: address=%s", res.MoneroAddress)
			return nil
		}
	}

	if xmrtaker {
		res, err := r.RecoverFromXMRTakerSecretAndContract(b, dataDir, infofile.PrivateKeyInfo.PrivateSpendKey,
			infofile.ContractSwapID, infofile.ContractSwap)
		if err != nil {
			return err
		}

		if res.Claimed {
			log.Infof("claimed monero! wallet address=%s", res.MoneroAddress)
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

func getRecoverer(c *cli.Context, env common.Environment, cfg *common.Config) (Recoverer, error) {
	var (
		moneroEndpoint string
	)

	// For the monero wallet related values, keep the default config values unless the end
	// use explicitly set the flag.
	if c.IsSet(flagMoneroDaemonHost) {
		cfg.MoneroDaemonHost = c.String(flagMoneroDaemonHost)
	}
	if c.IsSet(flagMoneroDaemonPort) {
		cfg.MoneroDaemonPort = c.Uint(flagMoneroDaemonPort)
	}
	walletFilePath := cfg.MoneroWalletPath()
	if c.IsSet(flagMoneroWalletPath) {
		walletFilePath = c.String(flagMoneroWalletPath)
	}
	walletClient, err := monero.NewWalletClient(&monero.WalletClientConf{
		Env:                 env,
		WalletFilePath:      walletFilePath,
		MonerodPort:         cfg.MoneroDaemonPort,
		MonerodHost:         cfg.MoneroDaemonHost,
		MoneroWalletRPCPath: "", // look for it in "monero-bin/monero-wallet-rpc" and then the user's path
	})
	if err != nil {
		return nil, err
	}

	ethEndpoint := c.String(flagEthereumEndpoint)
	if ethEndpoint == "" {
		ethEndpoint = common.DefaultEthEndpoint
	}

	log.Infof("created recovery module with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)
	return recovery.NewRecoverer(env, walletClient, ethEndpoint)
}

func createBackend(ctx context.Context, c *cli.Context, env common.Environment,
	cfg common.Config, contractAddr ethcommon.Address) (backend.Backend, error) {
	var (
		ethEndpoint string
	)

	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String(flagEthereumEndpoint)
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	// TODO: add --external-signer option to allow front-end integration (#124)
	ethPrivKeyFile := c.String(flagEthereumPrivKey)
	devXMRMaker := false // Not directly supported, but you can put the Ganache key in a file
	devXMRTaker := false
	ethPrivKey, err := cliutil.GetEthereumPrivateKey(ethPrivKeyFile, env, devXMRMaker, devXMRTaker)
	if err != nil {
		return nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	contract, err := contracts.NewSwapFactory(contractAddr, ec)
	if err != nil {
		return nil, err
	}

	// For the monero wallet related values, keep the default config values unless the end
	// user explicitly set the flag.
	if c.IsSet(flagMoneroDaemonHost) {
		cfg.MoneroDaemonHost = c.String(flagMoneroDaemonHost)
	}
	if c.IsSet(flagMoneroDaemonPort) {
		cfg.MoneroDaemonPort = c.Uint(flagMoneroDaemonPort)
	}
	moneroWalletPath := cfg.MoneroWalletPath()
	if c.IsSet(flagMoneroWalletPath) {
		moneroWalletPath = c.String(flagMoneroWalletPath)
	}
	mc, err := monero.NewWalletClient(&monero.WalletClientConf{
		Env:                 env,
		WalletFilePath:      moneroWalletPath,
		MonerodPort:         cfg.MoneroDaemonPort,
		MonerodHost:         cfg.MoneroDaemonHost,
		MoneroWalletRPCPath: "", // look for it in "monero-bin/monero-wallet-rpc" and then the user's path
	})
	if err != nil {
		return nil, err
	}

	extendedEC, err := extethclient.NewEthClient(ctx, ec, ethPrivKey)
	if err != nil {
		return nil, err
	}
	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set (#153)
	extendedEC.SetGasPrice(uint64(c.Uint(flagGasPrice)))
	extendedEC.SetGasLimit(uint64(c.Uint(flagGasLimit)))

	bcfg := &backend.Config{
		Ctx:                 ctx,
		MoneroClient:        mc,
		EthereumClient:      extendedEC,
		Environment:         env,
		SwapContract:        contract,
		SwapContractAddress: contractAddr,
	}

	return backend.NewBackend(bcfg)
}
