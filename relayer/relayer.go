package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"

	client "github.com/AthanorLabs/go-relayer-client"
	rcommon "github.com/AthanorLabs/go-relayer/common"
	gsnforwarder "github.com/AthanorLabs/go-relayer/examples/gsn_forwarder"
)

// Client is a client that can submit transactions to an Ethereum meta-transaction relayer.
// It has a private key for signing transactions to be forwarded.
type Client struct {
	key              *rcommon.Key
	c                *client.Client
	chainID          *big.Int
	forwarder        *gsnforwarder.IForwarder
	forwarderAddress ethcommon.Address
}

// NewClient returns a new relayer client.
func NewClient(sk *ecdsa.PrivateKey, ec *ethclient.Client, relayerEndpoint, forwarderAddress string) (*Client, error) {
	forwarder, err := forwarderFromAddress(ethcommon.HexToAddress(forwarderAddress), ec)
	if err != nil {
		return nil, err
	}

	chainID, err := ec.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &Client{
		key:              rcommon.NewKeyFromPrivateKey(sk),
		c:                client.NewClient(relayerEndpoint),
		chainID:          chainID,
		forwarder:        forwarder,
		forwarderAddress: ethcommon.HexToAddress(forwarderAddress),
	}, nil
}

func forwarderFromAddress(address ethcommon.Address, ec *ethclient.Client) (*gsnforwarder.IForwarder, error) {
	forwarder, err := gsnforwarder.NewIForwarder(address, ec)
	if err != nil {
		return nil, err
	}

	return forwarder, nil
}

// SubmitTransaction submits a transaction with the given calldata to the relayer.
func (c *Client) SubmitTransaction(
	to ethcommon.Address,
	calldata []byte,
) (ethcommon.Hash, error) {
	nonce, err := c.forwarder.GetNonce(&bind.CallOpts{}, c.key.Address())
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to get nonce from forwarder: %w", err)
	}

	req := &gsnforwarder.IForwarderForwardRequest{
		From:  c.key.Address(),
		To:    to,
		Value: big.NewInt(0),
		Gas:   big.NewInt(679639582), // TODO: fetch from ethclient
		Nonce: nonce,
		Data:  calldata,
	}

	name := "Forwarder"
	version := "0.0.1"

	domainSeparator, err := rcommon.GetEIP712DomainSeparator(name, version, c.chainID, c.forwarderAddress)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to get EIP712 domain separator: %w", err)
	}

	digest, err := rcommon.GetForwardRequestDigestToSign(
		req,
		domainSeparator,
		nil,
	)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to get forward request digest: %w", err)
	}

	sig, err := c.key.Sign(digest)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to sign forward request digest: %w", err)
	}

	rpcReq := &rcommon.SubmitTransactionRequest{
		From:      req.From,
		To:        req.To,
		Value:     req.Value,
		Gas:       req.Gas,
		Nonce:     req.Nonce,
		Data:      req.Data,
		Signature: sig,
	}

	// submit transaction to relayer
	resp, err := c.c.SubmitTransaction(rpcReq)
	if err != nil {
		return ethcommon.Hash{}, fmt.Errorf("failed to submit transaction to relayer: %w", err)
	}

	return resp.TxHash, nil
}
