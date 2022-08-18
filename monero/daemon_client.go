package monero

import (
	"github.com/MarinX/monerorpc"
	"github.com/MarinX/monerorpc/daemon"
)

// DaemonClient represents a monerod client.
type DaemonClient interface {
	GenerateBlocks(address string, amount uint64) error
}

type daemonClient struct {
	rpc daemon.Daemon // full API with slightly different method signature(s)
}

// NewDaemonClient returns a new monerod daemonClient.
func NewDaemonClient(endpoint string) *daemonClient {
	return &daemonClient{
		rpc: monerorpc.New(endpoint, nil).Daemon,
	}
}

func (c *daemonClient) GenerateBlocks(address string, amount uint64) error {
	// https://github.com/monero-project/monero/blob/v0.18.1.0/src/rpc/core_rpc_server_error_codes.h#L65
	const failedToMineBlock = "Block not accepted"

	prevHeight, err := c.rpc.GetBlockCount()
	if err != nil {
		return err
	}

	for i := 0; i < maxRetries; i++ {
		resp, err := c.generateBlocks(address, amount)
		if err == nil { // no issues, we are done
			break
		}
		if err.Error() != failedToMineBlock {
			return err
		}
		newHeight, err := c.rpc.GetBlockCount()
		if err != nil {
			return err
		}
		if newHeight.Count >= prevHeight.Count+amount {
			break
		}
		oldAmount := amount
		amount -= newHeight.Count - prevHeight.Count
		prevHeight.Count = newHeight.Count
		// TODO: It is possible that resp is non-nil when we get a "Block not accepted" error. If we
		//       succeed in capturing one of these errors and the response is present, we can delete
		//       half the code above and use "amount -= len(resp.Blocks)". If it is not present, we
		//       can remove resp from the log below.
		log.Warnf("GenerateBlocks failure requested=%d, generated=%d, resp=%#v",
			oldAmount, oldAmount-amount, resp)
	}
	return nil
}

func (c *daemonClient) generateBlocks(address string, amount uint64) (*daemon.GenerateBlocksResponse, error) {
	return c.rpc.GenerateBlocks(&daemon.GenerateBlocksRequest{
		AmountOfBlocks: amount,
		WalletAddress:  address,
	})
}
