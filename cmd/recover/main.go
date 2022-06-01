package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/cmd/utils"
	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/crypto/monero"
	pcommon "github.com/noot/atomic-swap/protocol"
	"github.com/noot/atomic-swap/protocol/xmrmaker"
	"github.com/noot/atomic-swap/protocol/xmrtaker"
	recovery "github.com/noot/atomic-swap/recover"
	"github.com/noot/atomic-swap/swapfactory"

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
			&cli.StringFlag{
				Name:  flagInfoFile,
				Usage: "path to swap infofile",
			},
			&cli.BoolFlag{
				Name:  flagXMRMaker,
				Usage: "true if recovering as an xmr-maker",
			},
			&cli.BoolFlag{
				Name:  flagXMRTaker,
				Usage: "true if recovering as an xmr-taker",
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
	RecoverFromXMRMakerSecretAndContract(b *xmrmaker.Instance, xmrmakerSecret, contractAddr string, swapID [32]byte, swap swapfactory.SwapFactorySwap) (*xmrmaker.RecoveryResult, error) //nolint:lll
	RecoverFromXMRTakerSecretAndContract(a *xmrtaker.Instance, xmrtakerSecret string, swapID [32]byte, swap swapfactory.SwapFactorySwap) (*xmrtaker.RecoveryResult, error)               //nolint:lll
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
	if !xmrmaker && !xmrtaker {
		return errMustSpecifyXMRMakerOrTaker
	}

	env, cfg, err := utils.GetEnvironment(c)
	if err != nil {
		return err
	}

	infofilePath := c.String(flagInfoFile)
	if infofilePath == "" {
		return errMustProvideInfoFile
	}

	infofileBytes, err := ioutil.ReadFile(filepath.Clean(infofilePath))
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
		addr, err := r.WalletFromSharedSecret(infofile.SharedSwapPrivateKey)
		if err != nil {
			return err
		}

		log.Infof("restored wallet from secrets: address=%s", addr)
		return nil
	}

	contractAddr := infofile.ContractAddress

	if xmrmaker {
		b, err := createXMRMakerInstance(context.Background(), c, env, cfg)
		if err != nil {
			return err
		}

		res, err := r.RecoverFromXMRMakerSecretAndContract(b, infofile.PrivateKeyInfo.PrivateSpendKey,
			contractAddr, infofile.ContractSwapID, infofile.ContractSwap)
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

	if xmrtaker {
		addr := ethcommon.HexToAddress(contractAddr)
		a, err := createXMRTakerInstance(context.Background(), c, env, cfg, addr)
		if err != nil {
			return err
		}

		res, err := r.RecoverFromXMRTakerSecretAndContract(a, infofile.PrivateKeyInfo.PrivateSpendKey,
			infofile.ContractSwapID, infofile.ContractSwap)
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

func createXMRTakerInstance(ctx context.Context, c *cli.Context, env common.Environment,
	cfg common.Config, contractAddr ethcommon.Address) (*xmrtaker.Instance, error) {
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
		moneroEndpoint = common.DefaultXMRTakerMoneroEndpoint
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

	contract, err := swapfactory.NewSwapFactory(contractAddr, ec)
	if err != nil {
		return nil, err
	}

	xmrtakerCfg := &xmrtaker.Config{
		Ctx:                  ctx,
		Basepath:             cfg.Basepath,
		MoneroWalletEndpoint: moneroEndpoint,
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          env,
		ChainID:              big.NewInt(chainID),
		GasPrice:             gasPrice,
		GasLimit:             uint64(c.Uint(flagGasLimit)),
		SwapContract:         contract,
		SwapContractAddress:  contractAddr,
	}

	return xmrtaker.NewInstance(xmrtakerCfg)
}

func createXMRMakerInstance(ctx context.Context, c *cli.Context, env common.Environment,
	cfg common.Config) (*xmrmaker.Instance, error) {
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
		moneroEndpoint = common.DefaultXMRMakerMoneroEndpoint
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

	xmrmakerCfg := &xmrmaker.Config{
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

	b, err := xmrmaker.NewInstance(xmrmakerCfg)
	if err != nil {
		return nil, err
	}

	return b, nil
}
