package main

import (
	"errors"
	"os"

	"github.com/urfave/cli"

	"github.com/noot/atomic-swap/common"
	mcrypto "github.com/noot/atomic-swap/monero/crypto"
	recovery "github.com/noot/atomic-swap/monero/recover"

	logging "github.com/ipfs/go-log"
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
		Name:  "swaprecover",
		Usage: "A program for recovering swap funds due to unexpected shutdowns",
		//Action: runRecover,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "env",
				Usage: "environment to use: one of mainnet, stagenet, or dev",
			},
			&cli.StringFlag{
				Name:  "monero-endpoint",
				Usage: "monero-wallet-rpc endpoint",
			},
			&cli.StringFlag{
				Name:  "monero-daemon-endpoint",
				Usage: "monerod RPC endpoint",
			},
			&cli.StringFlag{
				Name:  "ethereum-endpoint",
				Usage: "ethereum client endpoint",
			},
			&cli.StringFlag{
				Name:  "ethereum-privkey",
				Usage: "file containing a private key hex string",
			},
			&cli.UintFlag{
				Name:  "ethereum-chain-id",
				Usage: "ethereum chain ID; eg. mainnet=1, ropsten=3, rinkeby=4, goerli=5, ganache=1337",
			},
			&cli.UintFlag{
				Name:  "gas-price",
				Usage: "ethereum gas price to use for transactions (in gwei). if not set, the gas price is set via oracle.",
			},
			&cli.UintFlag{
				Name:  "gas-limit",
				Usage: "ethereum gas limit to use for transactions. if not set, the gas limit is estimated for each transaction.",
			},
		},
		Commands: []cli.Command{
			{
				Name:    "monero",
				Aliases: []string{"xmr"},
				Usage:   "recover monero funds from an aborted swap; must provide 2/3 of --alice-secret, --bob-secret, and --contract-addr", //nolint:lll
				Action:  runRecoverMonero,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "alice-secret",
						Usage: "Alice's swap secret, can be found in the basepath (default ~/.atomicswap), format is a hex-encoded string", //nolint:lll
					},
					&cli.StringFlag{
						Name:  "bob-secret",
						Usage: "Bob's swap secret, can be found in the basepath (default ~/.atomicswap), format is a hex-encoded string", //nolint:lll
					},
					&cli.StringFlag{
						Name:  "contract-addr",
						Usage: "address of deployed ethereum swap contract, can be found in the basepath (default ~/.atomicswap)", //nolint:lll
					},
				},
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

// MoneroRecoverer is implemented by a backend which is able to recover monero
type MoneroRecoverer interface {
	WalletFromSecrets(aliceSecret, bobSecret string) (mcrypto.Address, error)
}

func runRecoverMonero(ctx *cli.Context) error {
	as := ctx.String("alice-secret")
	bs := ctx.String("bob-secret")
	contractAddr := ctx.String("contract-addr")

	// TODO: use eth config here for chain ID
	env, _, err := getEnvironment(ctx)
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

	r, err := getRecoverer(ctx, env)
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

	log.Warnf("unimplemented!")
	return nil
}

func getEnvironment(c *cli.Context) (env common.Environment, cfg common.Config, err error) {
	switch c.String("env") {
	case "mainnet":
		env = common.Mainnet
		cfg = common.MainnetConfig
	case "stagenet":
		env = common.Stagenet
		cfg = common.StagenetConfig
	case "dev":
		env = common.Development
		cfg = common.DevelopmentConfig
	case "":
		env = common.Development
		cfg = common.DevelopmentConfig
	default:
		return 0, common.Config{}, errors.New("--env must be one of mainnet, stagenet, or dev")
	}

	return env, cfg, nil
}

func getRecoverer(c *cli.Context, env common.Environment) (MoneroRecoverer, error) {
	var (
		moneroEndpoint, ethEndpoint string
	)

	if c.String("monero-endpoint") != "" {
		moneroEndpoint = c.String("monero-endpoint")
	} else {
		moneroEndpoint = common.DefaultBobMoneroEndpoint
	}

	if c.String("ethereum-endpoint") != "" {
		ethEndpoint = c.String("ethereum-endpoint")
	} else {
		ethEndpoint = common.DefaultEthEndpoint
	}

	log.Info("created recovery module with monero endpoint %s and ethereum endpoint %s",
		moneroEndpoint,
		ethEndpoint,
	)
	return recovery.NewRecoverer(env, moneroEndpoint, ethEndpoint)
}
