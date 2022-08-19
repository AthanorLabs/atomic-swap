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

// GenerateBlocks quickly mines the requested number of blocks. This version slightly
// compensates for the non-deterministic behaviour of the raw API call when 2 (or more)
// simultaneous invocations occur. When this happens, the calls generate multiple blocks
// on separate chains, meaning that the call with more blocks usually wins, while the
// call with fewer blocks has their blocks discarded. If 2 calls have a large number
// of blocks, say 512, the result can be worse. The calls fight each other, causing block
// reorganisations in both directions preventing either side from succeeding. As a
// mitigation, this method breaks large amount values up into separate API calls of 32
// blocks.
func (c *daemonClient) GenerateBlocks(address string, amount uint64) error {
	const maxBlocksPerCall = 32
	const maxErrorRetries = 10
	totalErrors := 0
	totalGenerated := uint64(0)

	for totalGenerated < amount {
		reqstAmount := uint64(maxBlocksPerCall)
		if amount-totalGenerated < maxBlocksPerCall {
			reqstAmount = amount - totalGenerated
		}
		resp, err := c.generateBlocks(address, reqstAmount)
		if err != nil {
			totalErrors++
			log.Warnf("GenerateBlocks(%.10s..., %d) failed, total=%d/%d errCount=%d/%d err=%q",
				address, reqstAmount, totalGenerated, amount, totalErrors, maxErrorRetries, err)
			if totalErrors >= maxErrorRetries {
				return err
			}
			continue
		}
		// GenerateBlocks is only used for testing, and we trust the daemon to not put us in an
		// infinite loop with non-error, empty-block responses.
		totalGenerated += uint64(len(resp.Blocks))
	}
	return nil
}

func (c *daemonClient) generateBlocks(address string, amount uint64) (*daemon.GenerateBlocksResponse, error) {
	return c.rpc.GenerateBlocks(&daemon.GenerateBlocksRequest{
		AmountOfBlocks: amount,
		WalletAddress:  address,
	})
}
