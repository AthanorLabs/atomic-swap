package main

import (
	"context"
	"encoding/json"
	"math/big"
	"os"
	"path/filepath"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	logging "github.com/ipfs/go-log"
	"github.com/urfave/cli/v2"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	mcrypto "github.com/athanorlabs/atomic-swap/crypto/monero"
	pcommon "github.com/athanorlabs/atomic-swap/protocol"
	"github.com/athanorlabs/atomic-swap/protocol/backend"
	"github.com/athanorlabs/atomic-swap/protocol/xmrmaker"
	"github.com/athanorlabs/atomic-swap/protocol/xmrtaker"
	recovery "github.com/athanorlabs/atomic-swap/recover"
	"github.com/athanorlabs/atomic-swap/swapfactory"
)

const (
	flagEnv                  = "env"
	flagMoneroWalletEndpoint = "monero-endpoint"
	flagEthereumEndpoint     = "ethereum-endpoint"
	flagEthereumPrivKey      = "ethereum-privkey"
	flagEthereumChainID      = "ethereum-chain-id"
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
				Name:  flagMoneroWalletEndpoint,
				Usage: "monero-wallet-rpc endpoint",
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
				Name:  flagEthereumChainID,
				Usage: "Ethereum chain ID; eg. mainnet=1, goerli=5, ganache=1337",
			},
			&cli.UintFlag{
				Name:  flagGasPrice,
				Usage: "Ethereum gas price to use for transactions (in gwei). If not set, the gas price is set via oracle.",
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
	RecoverFromXMRMakerSecretAndContract(b backend.Backend, dataDir string, xmrmakerSecret, contractAddr string, swapID [32]byte, swap swapfactory.SwapFactorySwap) (*xmrmaker.RecoveryResult, error) //nolint:lll
	RecoverFromXMRTakerSecretAndContract(b backend.Backend, dataDir string, xmrtakerSecret string, swapID [32]byte, swap swapfactory.SwapFactorySwap) (*xmrtaker.RecoveryResult, error)               //nolint:lll
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

	r, err := inst.getRecovererFunc(c, env)
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

func getRecoverer(c *cli.Context, env common.Environment) (Recoverer, error) {
	var (
		moneroEndpoint, ethEndpoint string
	)

	if c.String(flagMoneroWalletEndpoint) != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		moneroEndpoint = common.DefaultXMRMakerMoneroEndpoint
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

func createBackend(ctx context.Context, c *cli.Context, env common.Environment,
	cfg common.Config, contractAddr ethcommon.Address) (backend.Backend, error) {
	var (
		moneroEndpoint, ethEndpoint string
	)

	if c.IsSet(flagEthereumChainID) {
		cfg.EthereumChainID = int64(c.Uint(flagEthereumChainID))
	}

	if c.String(flagMoneroWalletEndpoint) != "" {
		moneroEndpoint = c.String(flagMoneroWalletEndpoint)
	} else {
		moneroEndpoint = common.DefaultXMRTakerMoneroEndpoint
	}

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

	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set (#153)
	var gasPrice *big.Int
	if c.Uint(flagGasPrice) != 0 {
		gasPrice = big.NewInt(int64(c.Uint(flagGasPrice)))
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	contract, err := swapfactory.NewSwapFactory(contractAddr, ec)
	if err != nil {
		return nil, err
	}

	bcfg := &backend.Config{
		Ctx:                  ctx,
		MoneroWalletEndpoint: moneroEndpoint,
		EthereumClient:       ec,
		EthereumPrivateKey:   ethPrivKey,
		Environment:          env,
		ChainID:              big.NewInt(cfg.EthereumChainID),
		GasPrice:             gasPrice,
		GasLimit:             uint64(c.Uint(flagGasLimit)),
		SwapContract:         contract,
		SwapContractAddress:  contractAddr,
	}

	if env == common.Development {
		bcfg.MoneroDaemonEndpoint = common.DefaultMoneroDaemonEndpoint
	}

	return backend.NewBackend(bcfg)
}
