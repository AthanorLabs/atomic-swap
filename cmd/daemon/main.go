package main

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/cmd/utils"
	"github.com/noot/atomic-swap/common"
	"github.com/noot/atomic-swap/net"
	"github.com/noot/atomic-swap/protocol/backend"
	"github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/protocol/xmrmaker"
	"github.com/noot/atomic-swap/protocol/xmrtaker"
	"github.com/noot/atomic-swap/rpc"

	logging "github.com/ipfs/go-log"
)

const (
	// default libp2p ports
	defaultLibp2pPort         = 9900
	defaultXMRTakerLibp2pPort = 9933
	defaultXMRMakerLibp2pPort = 9934

	// default libp2p key files
	defaultLibp2pKey         = "node.key"
	defaultXMRTakerLibp2pKey = "xmrtaker.key"
	defaultXMRMakerLibp2pKey = "xmrmaker.key"

	// default RPC port
	defaultRPCPort         = 5005
	defaultXMRTakerRPCPort = 5001
	defaultXMRMakerRPCPort = 5002

	defaultWSPort         = 6005
	defaultXMRTakerWSPort = 8081
	defaultXMRMakerWSPort = 8082
)

var (
	log = logging.Logger("cmd")
)

const (
	flagRPCPort    = "rpc-port"
	flagWSPort     = "ws-port"
	flagBasepath   = "basepath"
	flagLibp2pKey  = "libp2p-key"
	flagLibp2pPort = "libp2p-port"
	flagBootnodes  = "bootnodes"

	flagWalletFile           = "wallet-file"
	flagWalletPassword       = "wallet-password"
	flagEnv                  = "env"
	flagMoneroWalletEndpoint = "monero-endpoint"
	flagMoneroDaemonEndpoint = "monero-daemon-endpoint"
	flagEthereumEndpoint     = "ethereum-endpoint"
	flagEthereumPrivKey      = "ethereum-privkey"
	flagEthereumChainID      = "ethereum-chain-id"
	flagContractAddress      = "contract-address"
	flagGasPrice             = "gas-price"
	flagGasLimit             = "gas-limit"
	flagUseExternalSigner    = "external-signer"

	flagDevXMRTaker  = "dev-xmrtaker"
	flagDevXMRMaker  = "dev-xmrmaker"
	flagDeploy       = "deploy"
	flagTransferBack = "transfer-back"

	flagLog = "log"
)

var (
	app = &cli.App{
		Name:   "swapd",
		Usage:  "A program for doing atomic swaps between ETH and XMR",
		Action: runDaemon,
		Flags: []cli.Flag{
			&cli.UintFlag{
				Name:  flagRPCPort,
				Usage: "port for the daemon RPC server to run on; default 5001",
			},
			&cli.UintFlag{
				Name:  flagWSPort,
				Usage: "port for the daemon RPC websockets server to run on; default 8080",
			},
			&cli.StringFlag{
				Name:  flagBasepath,
				Usage: "path to store swap artefacts",
			},
			&cli.StringFlag{
				Name:  flagLibp2pKey,
				Usage: "libp2p private key",
			},
			&cli.UintFlag{
				Name:  flagLibp2pPort,
				Usage: "libp2p port to listen on",
			},
			&cli.StringFlag{
				Name:  flagWalletFile,
				Usage: "filename of wallet file containing XMR to be swapped; required if running as XMR provider",
			},
			&cli.StringFlag{
				Name:  flagWalletPassword,
				Usage: "password of wallet file containing XMR to be swapped",
			},
			&cli.StringFlag{
				Name:  flagEnv,
				Usage: "environment to use: one of mainnet, stagenet, or dev",
			},
			&cli.StringFlag{
				Name:  flagMoneroWalletEndpoint,
				Usage: "monero-wallet-rpc endpoint",
			},
			&cli.StringFlag{
				Name:  flagMoneroDaemonEndpoint,
				Usage: "monerod RPC endpoint; only used if running in development mode",
			},
			&cli.StringFlag{
				Name:  flagEthereumEndpoint,
				Usage: "ethereum client endpoint",
			},
			&cli.StringFlag{
				Name:  flagEthereumPrivKey,
				Usage: "file containing a private key hex string",
			},
			&cli.UintFlag{
				Name:  flagEthereumChainID,
				Usage: "ethereum chain ID; eg. mainnet=1, ropsten=3, rinkeby=4, goerli=5, ganache=1337",
			},
			&cli.StringFlag{
				Name:  flagContractAddress,
				Usage: "address of instance of SwapFactory.sol already deployed on-chain; required if running on mainnet",
			},
			&cli.StringFlag{
				Name:  flagBootnodes,
				Usage: "comma-separated string of libp2p bootnodes",
			},
			&cli.UintFlag{
				Name:  flagGasPrice,
				Usage: "ethereum gas price to use for transactions (in gwei). if not set, the gas price is set via oracle.",
			},
			&cli.UintFlag{
				Name:  flagGasLimit,
				Usage: "ethereum gas limit to use for transactions. if not set, the gas limit is estimated for each transaction.",
			},
			&cli.BoolFlag{
				Name:  flagDevXMRTaker,
				Usage: "run in development mode and use ETH provider default values",
			},
			&cli.BoolFlag{
				Name:  flagDevXMRMaker,
				Usage: "run in development mode and use XMR provider default values",
			},
			&cli.BoolFlag{
				Name:  flagDeploy,
				Usage: "deploy an instance of the swap contract; defaults to false",
			},
			&cli.BoolFlag{
				Name:  flagTransferBack,
				Usage: "when receiving XMR in a swap, transfer it back to the original wallet.",
			},
			&cli.StringFlag{
				Name:  flagLog,
				Usage: "set log level: one of [error|warn|info|debug]",
			},
			&cli.BoolFlag{
				Name:  flagUseExternalSigner,
				Usage: "use external signer, for usage with the swap UI",
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

type xmrtakerHandler interface {
	rpc.XMRTaker
}

type xmrmakerHandler interface {
	net.Handler
	rpc.XMRMaker
}

type daemon struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func setLogLevels(c *cli.Context) error {
	const (
		levelError = "error"
		levelWarn  = "warn"
		levelInfo  = "info"
		levelDebug = "debug"
	)

	level := c.String(flagLog)
	if level == "" {
		level = levelInfo
	}

	switch level {
	case levelError, levelWarn, levelInfo, levelDebug:
	default:
		return errors.New("invalid log level")
	}

	_ = logging.SetLogLevel("xmrtaker", level)
	_ = logging.SetLogLevel("xmrmaker", level)
	_ = logging.SetLogLevel("common", level)
	_ = logging.SetLogLevel("cmd", level)
	_ = logging.SetLogLevel("net", level)
	_ = logging.SetLogLevel("rpc", level)
	return nil
}

func runDaemon(c *cli.Context) error {
	if err := setLogLevels(c); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d := &daemon{
		ctx:    ctx,
		cancel: cancel,
	}

	err := d.make(c)
	if err != nil {
		return err
	}

	d.wait()
	os.Exit(0)
	return nil
}

func (d *daemon) make(c *cli.Context) error {
	env, cfg, err := utils.GetEnvironment(c)
	if err != nil {
		return err
	}

	devXMRTaker := c.Bool(flagDevXMRTaker)
	devXMRMaker := c.Bool(flagDevXMRMaker)

	chainID := int64(c.Uint(flagEthereumChainID))
	if chainID == 0 {
		chainID = cfg.EthereumChainID
	}

	var bootnodes []string
	if c.String(flagBootnodes) != "" {
		bootnodes = strings.Split(c.String(flagBootnodes), ",")
	}

	k := c.String(flagLibp2pKey)
	p := uint16(c.Uint(flagLibp2pPort))
	var (
		libp2pKey  string
		libp2pPort uint16
		rpcPort    uint16
	)

	switch {
	case k != "":
		libp2pKey = k
	case devXMRTaker:
		libp2pKey = defaultXMRTakerLibp2pKey
	case devXMRMaker:
		libp2pKey = defaultXMRMakerLibp2pKey
	default:
		libp2pKey = defaultLibp2pKey
	}

	switch {
	case p != 0:
		libp2pPort = p
	case devXMRTaker:
		libp2pPort = defaultXMRTakerLibp2pPort
	case devXMRMaker:
		libp2pPort = defaultXMRMakerLibp2pPort
	default:
		libp2pPort = defaultLibp2pPort
	}

	netCfg := &net.Config{
		Ctx:         d.ctx,
		Environment: env,
		ChainID:     chainID,
		Port:        libp2pPort,
		KeyFile:     libp2pKey,
		Bootnodes:   bootnodes,
	}

	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}

	sm := swap.NewManager()
	backend, err := newBackend(d.ctx, c, env, cfg, chainID, devXMRMaker, sm, host)
	if err != nil {
		return err
	}

	a, b, err := getProtocolInstances(c, cfg, backend)
	if err != nil {
		return err
	}

	// connect network to protocol handler
	// handler handles initiated ("taken") swap
	host.SetHandler(b)

	if err = host.Start(); err != nil {
		return err
	}

	p = uint16(c.Uint(flagRPCPort))
	switch {
	case p != 0:
		rpcPort = p
	case devXMRTaker:
		rpcPort = defaultXMRTakerRPCPort
	case devXMRMaker:
		rpcPort = defaultXMRMakerRPCPort
	default:
		rpcPort = defaultRPCPort
	}

	wsPort := uint16(c.Uint(flagWSPort))
	switch {
	case wsPort != 0:
	case devXMRTaker:
		wsPort = defaultXMRTakerWSPort
	case devXMRMaker:
		wsPort = defaultXMRMakerWSPort
	default:
		wsPort = defaultWSPort
	}

	rpcCfg := &rpc.Config{
		Ctx:             d.ctx,
		Port:            rpcPort,
		WsPort:          wsPort,
		Net:             host,
		XMRTaker:        a,
		XMRMaker:        b,
		ProtocolBackend: backend,
	}

	s, err := rpc.NewServer(rpcCfg)
	if err != nil {
		return err
	}

	errCh := s.Start()
	go func() {
		select {
		case <-d.ctx.Done():
			return
		case err := <-errCh:
			log.Errorf("failed to start RPC server: %s", err)
			d.cancel()
			os.Exit(1)
		}
	}()

	log.Infof("started swapd with basepath %s",
		cfg.Basepath,
	)
	return nil
}

func newBackend(ctx context.Context, c *cli.Context, env common.Environment, cfg common.Config,
	chainID int64, devXMRMaker bool, sm swap.Manager, net net.Host) (backend.Backend, error) {
	var (
		moneroEndpoint, daemonEndpoint, ethEndpoint string
	)

	if c.String(flagMoneroWalletEndpoint) != "" {
		moneroEndpoint = c.String(flagMoneroWalletEndpoint)
	} else if devXMRMaker {
		moneroEndpoint = common.DefaultXMRMakerMoneroEndpoint
	} else {
		moneroEndpoint = common.DefaultXMRTakerMoneroEndpoint
	}

	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String(flagEthereumEndpoint)
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	ethPrivKey, err := utils.GetEthereumPrivateKey(c, env, devXMRMaker, c.Bool(flagUseExternalSigner))
	if err != nil {
		return nil, err
	}

	var pk *ecdsa.PrivateKey
	if ethPrivKey != "" {
		pk, err = ethcrypto.HexToECDSA(ethPrivKey)
		if err != nil {
			return nil, err
		}
	}

	if c.String(flagMoneroDaemonEndpoint) != "" {
		daemonEndpoint = c.String(flagMoneroDaemonEndpoint)
	} else {
		daemonEndpoint = cfg.MoneroDaemonEndpoint
	}

	// TODO: add configs for different eth testnets + L2 and set gas limit based on those, if not set
	var gasPrice *big.Int
	if c.Uint(flagGasPrice) != 0 {
		gasPrice = big.NewInt(int64(c.Uint(flagGasPrice)))
	}

	var contractAddr ethcommon.Address
	contractAddrStr := c.String(flagContractAddress)
	if contractAddrStr == "" {
		contractAddr = ethcommon.Address{}
	} else {
		contractAddr = ethcommon.HexToAddress(contractAddrStr)
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, err
	}

	deploy := c.Bool(flagDeploy)
	if deploy {
		contractAddr = ethcommon.Address{}
	}

	contract, contractAddr, err := getOrDeploySwapFactory(contractAddr, env, cfg.Basepath,
		big.NewInt(chainID), pk, ec)
	if err != nil {
		return nil, err
	}

	bcfg := &backend.Config{
		Ctx:                  ctx,
		MoneroWalletEndpoint: moneroEndpoint,
		MoneroDaemonEndpoint: daemonEndpoint,
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          env,
		ChainID:              big.NewInt(chainID),
		GasPrice:             gasPrice,
		GasLimit:             uint64(c.Uint(flagGasLimit)),
		SwapManager:          sm,
		SwapContract:         contract,
		SwapContractAddress:  contractAddr,
		Net:                  net,
	}

	b, err := backend.NewBackend(bcfg)
	if err != nil {
		return nil, fmt.Errorf("failed to make backend: %w", err)
	}

	log.Infof("created backend with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)

	return b, nil
}

func getProtocolInstances(c *cli.Context, cfg common.Config,
	b backend.Backend) (xmrtakerHandler, xmrmakerHandler, error) {
	walletFile := c.String("wallet-file")

	// empty password is ok
	walletPassword := c.String("wallet-password")

	xmrtakerCfg := &xmrtaker.Config{
		Backend:              b,
		Basepath:             cfg.Basepath,
		MoneroWalletFile:     walletFile,
		MoneroWalletPassword: walletPassword,
		TransferBack:         c.Bool(flagTransferBack),
	}

	xmrtaker, err := xmrtaker.NewInstance(xmrtakerCfg)
	if err != nil {
		return nil, nil, err
	}

	xmrmakerCfg := &xmrmaker.Config{
		Backend:        b,
		Basepath:       cfg.Basepath,
		WalletFile:     walletFile,
		WalletPassword: walletPassword,
	}

	xmrmaker, err := xmrmaker.NewInstance(xmrmakerCfg)
	if err != nil {
		return nil, nil, err
	}

	return xmrtaker, xmrmaker, nil
}
