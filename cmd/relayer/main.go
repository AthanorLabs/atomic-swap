// Package main is the entrypoint for the swap-specific Ethereum transaction relayer.
// It's purpose is to allow swaps users (ETH-takers in particular) to submit calls to the
// swap contract to claim their ETH from an account that does not have any ETH in it.
// This improves the swap UX by allowing users to obtain ETH without already having any.
// In this case, the relayer submits the transaction on the user's behalf, paying their
// gas fees, and (optionally) receiving a small percentage of the swap's value as payment.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/apd/v3"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"

	p2pnet "github.com/athanorlabs/go-p2p-net"
	rcommon "github.com/athanorlabs/go-relayer/common"
	rcontracts "github.com/athanorlabs/go-relayer/impls/gsnforwarder"
	net "github.com/athanorlabs/go-relayer/net"
	"github.com/athanorlabs/go-relayer/relayer"
	rrpc "github.com/athanorlabs/go-relayer/rpc"

	"github.com/athanorlabs/atomic-swap/cliutil"
	"github.com/athanorlabs/atomic-swap/common"
	swapnet "github.com/athanorlabs/atomic-swap/net"

	logging "github.com/ipfs/go-log"
)

const (
	flagDataDir          = "data-dir"
	flagEthereumEndpoint = "ethereum-endpoint"
	flagForwarderAddress = "forwarder-address"
	flagKey              = "key"
	flagRPC              = "rpc"
	flagRPCPort          = "rpc-port"
	flagDeploy           = "deploy"
	flagLog              = "log-level"
	// TODO: do we need this, or can we assume all swap relayers will need to be on the p2p network?
	flagWithNetwork       = "with-network"
	flagLibp2pKey         = "libp2p-key"
	flagLibp2pPort        = "libp2p-port"
	flagBootnodes         = "bootnodes"
	flagRelayerCommission = "relayer-commission"

	defaultLibp2pPort = 10900
)

var (
	log = logging.Logger("main")

	flags = []cli.Flag{
		&cli.StringFlag{
			Name:  flagDataDir,
			Usage: "Path to store swap artifacts",
			Value: "{HOME}/.atomicswap/{ENV}", // For --help only, actual default replaces variables
		},
		&cli.StringFlag{
			Name:  flagEthereumEndpoint,
			Value: "http://localhost:8545",
			Usage: "Ethereum RPC endpoint",
		},
		&cli.StringFlag{
			Name:  flagKey,
			Value: "{HOME}/.atomicswap/{ENV}/relayer/eth.key",
			Usage: "Path to file containing Ethereum private key",
		},
		&cli.StringFlag{
			Name:  flagForwarderAddress,
			Usage: "Address of the forwarder contract to use. Defaults to the forwarder address in the chain's swap contract.",
		},
		&cli.UintFlag{
			Name:  flagRPCPort,
			Value: 7799,
			Usage: "Relayer RPC server port",
		},
		&cli.BoolFlag{
			Name:  flagRPC,
			Value: false,
			Usage: "Run the relayer HTTP-RPC server on localhost. Defaults to false",
		},
		&cli.BoolFlag{
			Name:  flagDeploy,
			Usage: "Deploy an instance of the forwarder contract",
		},
		&cli.StringFlag{
			Name:  flagLog,
			Value: "info",
			Usage: "Set log level: one of [error|warn|info|debug]",
		},
		&cli.BoolFlag{
			Name:  flagWithNetwork,
			Value: true,
			Usage: "Run the relayer with p2p network capabilities",
		},
		&cli.StringFlag{
			Name:  flagLibp2pKey,
			Usage: "libp2p private key",
			Value: common.DefaultLibp2pKeyFileName,
		},
		&cli.UintFlag{
			Name:  flagLibp2pPort,
			Usage: "libp2p port to listen on",
			Value: defaultLibp2pPort,
		},
		&cli.StringSliceFlag{
			Name:    flagBootnodes,
			Aliases: []string{"bn"},
			Usage:   "libp2p bootnode, comma separated if passing multiple to a single flag",
			EnvVars: []string{"SWAPD_BOOTNODES"},
		},
		&cli.StringFlag{
			Name: flagRelayerCommission,
			Usage: "Minimum commission percentage (of the swap value) to receive:" +
				" eg. --relayer-commission=0.01 for 1% commission",
			Value: common.DefaultRelayerCommission.Text('f'),
		},
	}

	errInvalidAddress       = errors.New("invalid forwarder address")
	errNoEthereumPrivateKey = errors.New("must provide ethereum private key with --key")
)

func main() {
	app := &cli.App{
		Name:                 "relayer",
		Usage:                "Ethereum transaction relayer",
		Version:              cliutil.GetVersion(),
		Flags:                flags,
		Action:               run,
		EnableBashCompletion: true,
		Suggest:              true,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func setLogLevels(c *cli.Context) error {
	const (
		levelError = "error"
		levelWarn  = "warn"
		levelInfo  = "info"
		levelDebug = "debug"
	)

	level := c.String(flagLog)
	switch level {
	case levelError, levelWarn, levelInfo, levelDebug:
	default:
		return fmt.Errorf("invalid log level %q", level)
	}

	_ = logging.SetLogLevel("main", level)
	_ = logging.SetLogLevel("relayer", level)
	_ = logging.SetLogLevel("rpc", level)
	_ = logging.SetLogLevel("p2pnet", "debug")
	return nil
}

func run(c *cli.Context) error {
	err := setLogLevels(c)
	if err != nil {
		return err
	}

	port := uint16(c.Uint(flagRPCPort))
	endpoint := c.String(flagEthereumEndpoint)
	ec, err := ethclient.Dial(endpoint)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return err
	}

	log.Infof("starting relayer with ethereum endpoint %s and chain ID %s", endpoint, chainID)

	config, err := common.ConfigFromChainID(chainID)
	if err != nil {
		return err
	}

	key, err := getPrivateKey(c.String(flagKey))
	if err != nil {
		return err
	}

	// set the forwarder address from the config, if it exists on that network
	// however, if the --forwarder-address flag is set, that address takes precedence
	var contractAddr string
	if (config.ContractAddress != ethcommon.Address{}) {
		contractAddr = config.ForwarderContractAddress.String()
	}
	addrFromFlag := c.String(flagForwarderAddress)
	if addrFromFlag != "" {
		contractAddr = addrFromFlag
	}

	deploy := c.Bool(flagDeploy)
	if deploy && (addrFromFlag != "") {
		return fmt.Errorf("flags --%s and --%s are mutually exclusive", flagDeploy, flagForwarderAddress)
	}
	if !deploy && (contractAddr == "") {
		return fmt.Errorf("either --%s or --%s is required", flagDeploy, flagForwarderAddress)
	}

	forwarder, forwarderAddr, err := deployOrGetForwarder(
		ctx,
		contractAddr,
		ec,
		key,
		chainID,
	)
	if err != nil {
		return err
	}

	relayerCommission, err := cliutil.ReadUnsignedDecimalFlag(c, flagRelayerCommission)
	if err != nil {
		return err
	}

	if relayerCommission.Cmp(apd.New(1, -1)) > 0 {
		return errors.New("relayer commission is too high: must be less than 0.1 (10%)")
	}

	// TODO: do we need to restrict potential commission values? eg. 1%, 1.25%, 1.5%, etc
	// or should we just require a fixed value for now?
	v := &validator{
		ctx:               ctx,
		ec:                ec,
		relayerCommission: relayerCommission,
		forwarderAddress:  forwarderAddr,
	}

	// the forwarder contract is fixed here; thus it needs to be the same
	// as what's hardcoded in the swap contract addr for that network.
	rcfg := &relayer.Config{
		Ctx:                     ctx,
		EthClient:               ec,
		Forwarder:               rcontracts.NewIForwarderWrapped(forwarder),
		Key:                     key,
		ValidateTransactionFunc: v.validateTransactionFunc,
	}

	r, err := relayer.NewRelayer(rcfg)
	if err != nil {
		return err
	}

	if c.Bool(flagWithNetwork) {
		// cfg.DataDir already has a default set, so only override if the user explicitly set the flag
		var datadir string
		if c.IsSet(flagDataDir) {
			datadir = c.String(flagDataDir) // override the value derived from `flagEnv`
		} else {
			datadir = config.DataDir
		}
		datadir = path.Join(datadir, "relayer")
		if err = common.MakeDir(datadir); err != nil {
			return err
		}

		h, err := setupNetwork(ctx, c, ec, r, datadir) //nolint:govet
		if err != nil {
			return err
		}

		defer func() {
			_ = h.Stop()
		}()
	}

	if c.Bool(flagRPC) {
		go signalHandler(ctx, cancel)
		rpcCfg := &rrpc.Config{
			Ctx:     ctx,
			Address: fmt.Sprintf("127.0.0.1:%d", port),
			Relayer: r,
		}

		server, err := rrpc.NewServer(rpcCfg) //nolint:govet
		if err != nil {
			return err
		}

		err = server.Start()
		if errors.Is(err, context.Canceled) || errors.Is(err, http.ErrServerClosed) {
			return nil
		}
	} else {
		signalHandler(ctx, cancel)
	}

	return err
}

func setupNetwork(
	ctx context.Context,
	c *cli.Context,
	ec *ethclient.Client,
	r *relayer.Relayer,
	datadir string,
) (*net.Host, error) {
	chainID, err := ec.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	var bootnodes []string
	if c.IsSet(flagBootnodes) {
		bootnodes = cliutil.ExpandBootnodes(c.StringSlice(flagBootnodes))
	}

	listenIP := "0.0.0.0"
	netCfg := &p2pnet.Config{
		Ctx:        ctx,
		DataDir:    datadir,
		Port:       uint16(c.Uint(flagLibp2pPort)),
		KeyFile:    path.Join(datadir, c.String(flagLibp2pKey)),
		Bootnodes:  bootnodes,
		ProtocolID: fmt.Sprintf("/%s/%d/%s", swapnet.ProtocolID, chainID.Int64(), net.ProtocolID),
		ListenIP:   listenIP,
	}

	cfg := &net.Config{
		Context:              ctx,
		P2pConfig:            netCfg,
		TransactionSubmitter: r,
		IsRelayer:            true,
	}

	h, err := net.NewHost(cfg)
	if err != nil {
		return nil, err
	}

	err = h.Start()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func getPrivateKey(keyFile string) (*rcommon.Key, error) {
	if keyFile != "" {
		fileData, err := os.ReadFile(filepath.Clean(keyFile))
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %w", err)
		}
		keyHex := strings.TrimSpace(string(fileData))
		return rcommon.NewKeyFromPrivateKeyString(keyHex)
	}
	return nil, errNoEthereumPrivateKey
}
