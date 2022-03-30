package main

import (
	"context"
	"errors"
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
	"github.com/noot/atomic-swap/protocol/alice"
	"github.com/noot/atomic-swap/protocol/bob"
	"github.com/noot/atomic-swap/protocol/swap"
	"github.com/noot/atomic-swap/rpc"
	"github.com/noot/atomic-swap/swapfactory"

	logging "github.com/ipfs/go-log"
)

const (
	// default libp2p ports
	defaultLibp2pPort      = 9900
	defaultAliceLibp2pPort = 9933
	defaultBobLibp2pPort   = 9934

	// default libp2p key files
	defaultLibp2pKey      = "node.key"
	defaultAliceLibp2pKey = "alice.key"
	defaultBobLibp2pKey   = "bob.key"

	// default RPC port
	defaultRPCPort      = 5005
	defaultAliceRPCPort = 5001
	defaultBobRPCPort   = 5002

	defaultWSPort      = 8080
	defaultAliceWSPort = 8081
	defaultBobWSPort   = 8082
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

	flagDevAlice = "dev-alice"
	flagDevBob   = "dev-bob"

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
				Name:  flagDevAlice,
				Usage: "run in development mode and use ETH provider default values",
			},
			&cli.BoolFlag{
				Name:  flagDevBob,
				Usage: "run in development mode and use XMR provider default values",
			},
			&cli.StringFlag{
				Name:  flagLog,
				Usage: "set log level: one of [error|warn|info|debug]",
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

type aliceHandler interface {
	rpc.Alice
	SetMessageSender(net.MessageSender)
}

type bobHandler interface {
	net.Handler
	rpc.Bob
	SetMessageSender(net.MessageSender)
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

	_ = logging.SetLogLevel("alice", level)
	_ = logging.SetLogLevel("bob", level)
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

	devAlice := c.Bool(flagDevAlice)
	devBob := c.Bool(flagDevBob)

	chainID := int64(c.Uint(flagEthereumChainID))
	if chainID == 0 {
		chainID = cfg.EthereumChainID
	}

	sm := swap.NewManager()

	a, b, err := getProtocolInstances(d.ctx, c, env, cfg, chainID, devBob, sm)
	if err != nil {
		return err
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
	case devAlice:
		libp2pKey = defaultAliceLibp2pKey
	case devBob:
		libp2pKey = defaultBobLibp2pKey
	default:
		libp2pKey = defaultLibp2pKey
	}

	switch {
	case p != 0:
		libp2pPort = p
	case devAlice:
		libp2pPort = defaultAliceLibp2pPort
	case devBob:
		libp2pPort = defaultBobLibp2pPort
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
		Handler:     b, // handler handles initiated ("taken") swaps
	}

	host, err := net.NewHost(netCfg)
	if err != nil {
		return err
	}

	// connect network to protocol handlers
	a.SetMessageSender(host)
	b.SetMessageSender(host)

	if err = host.Start(); err != nil {
		return err
	}

	p = uint16(c.Uint(flagRPCPort))
	switch {
	case p != 0:
		rpcPort = p
	case devAlice:
		rpcPort = defaultAliceRPCPort
	case devBob:
		rpcPort = defaultBobRPCPort
	default:
		rpcPort = defaultRPCPort
	}

	wsPort := uint16(c.Uint(flagWSPort))
	switch {
	case wsPort != 0:
	case devAlice:
		wsPort = defaultAliceWSPort
	case devBob:
		wsPort = defaultBobWSPort
	default:
		wsPort = defaultWSPort
	}

	rpcCfg := &rpc.Config{
		Ctx:         d.ctx,
		Port:        rpcPort,
		WsPort:      wsPort,
		Net:         host,
		Alice:       a,
		Bob:         b,
		SwapManager: sm,
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

func getProtocolInstances(ctx context.Context, c *cli.Context, env common.Environment, cfg common.Config,
	chainID int64, devBob bool, sm *swap.Manager) (a aliceHandler, b bobHandler, err error) {
	var (
		moneroEndpoint, daemonEndpoint, ethEndpoint string
	)

	if c.String(flagMoneroWalletEndpoint) != "" {
		moneroEndpoint = c.String(flagMoneroWalletEndpoint)
	} else if devBob {
		moneroEndpoint = common.DefaultBobMoneroEndpoint
	} else {
		moneroEndpoint = common.DefaultAliceMoneroEndpoint
	}

	if c.String(flagEthereumEndpoint) != "" {
		ethEndpoint = c.String(flagEthereumEndpoint)
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	ethPrivKey, err := utils.GetEthereumPrivateKey(c, env, devBob)
	if err != nil {
		return nil, nil, err
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

	pk, err := ethcrypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, nil, err
	}

	ec, err := ethclient.Dial(ethEndpoint)
	if err != nil {
		return nil, nil, err
	}

	var contract *swapfactory.SwapFactory
	if !devBob {
		contract, contractAddr, err = getOrDeploySwapFactory(contractAddr, env, cfg.Basepath,
			big.NewInt(chainID), pk, ec)
		if err != nil {
			return nil, nil, err
		}
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
		SwapManager:          sm,
		SwapContract:         contract,
		SwapContractAddress:  contractAddr,
	}

	a, err = alice.NewInstance(aliceCfg)
	if err != nil {
		return nil, nil, err
	}

	walletFile := c.String("wallet-file")

	// empty password is ok
	walletPassword := c.String("wallet-password")

	bobCfg := &bob.Config{
		Ctx:                  ctx,
		Basepath:             cfg.Basepath,
		MoneroWalletEndpoint: moneroEndpoint,
		MoneroDaemonEndpoint: daemonEndpoint,
		WalletFile:           walletFile,
		WalletPassword:       walletPassword,
		EthereumClient:       ec,
		EthereumPrivateKey:   pk,
		Environment:          env,
		ChainID:              big.NewInt(chainID),
		GasPrice:             gasPrice,
		GasLimit:             uint64(c.Uint(flagGasLimit)),
		SwapManager:          sm,
	}

	b, err = bob.NewInstance(bobCfg)
	if err != nil {
		return nil, nil, err
	}

	log.Infof("created swap protocol module with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)
	return a, b, nil
}
