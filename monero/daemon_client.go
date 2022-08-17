package monero

import (
	"sync"

	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/daemon"
)

// DaemonClient represents a monerod client.
type DaemonClient interface {
	GenerateBlocks(address string, amount uint64) error
}

type daemonClient struct {
	sync.Mutex
	rpc *monerorpc.MoneroRPC
}

// NewDaemonClient returns a new monerod daemonClient.
func NewDaemonClient(endpoint string) *daemonClient {
	return &daemonClient{
		rpc: monerorpc.New(endpoint, nil),
	}
}

func (c *daemonClient) GenerateBlocks(address string, amount uint64) error {
	_, err := c.generateBlocks(address, amount)
	return err
}

func (c *daemonClient) generateBlocks(address string, amount uint64) (*daemon.GenerateBlocksResponse, error) {
	return c.rpc.Daemon.GenerateBlocks(&daemon.GenerateBlocksRequest{
		AmountOfBlocks: amount,
		WalletAddress:  address,
	})
}
